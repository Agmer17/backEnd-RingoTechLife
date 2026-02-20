package order

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/products"
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type OrderService struct {
	muTransactionsData sync.Mutex
	transactionsData   map[uuid.UUID]*time.Timer
	orderRepo          OrderRepositoryInterface
	productService     *products.ProductsService
	appContext         context.Context
}

func NewOrderService(ord *OrderRepositoryImpl, psvc *products.ProductsService, ctx context.Context) *OrderService {
	return &OrderService{
		transactionsData: make(map[uuid.UUID]*time.Timer, 0),
		orderRepo:        ord,
		productService:   psvc,
		appContext:       ctx,
	}
}

func (o *OrderService) CreateOneOrder(ctx context.Context, productId uuid.UUID, q int, userId uuid.UUID) (*model.Order, *common.ErrorResponse) {

	productData, getErr := o.productService.GetById(ctx, productId)

	if getErr != nil {
		return nil, getErr
	}

	totalPrice := productData.Price * float64(q)
	order := model.Order{
		UserID:      userId,
		Status:      model.OrderStatusPending,
		Subtotal:    totalPrice,
		TotalAmount: totalPrice,
		Notes:       nil,
	}

	orders := make([]model.OrderItem, 1)

	orders[0] = model.OrderItem{
		ProductID:       productId,
		ProductName:     productData.Name,
		ProductSKU:      productData.SKU,
		PriceAtPurchase: productData.Price,
		Quantity:        q,
		Subtotal:        totalPrice,
	}

	result, err := o.orderRepo.Create(ctx, &order, orders)
	if err != nil {

		return nil, common.NewErrorResponse(500, "gagal membuat order! "+err.Error())
	}

	o.muTransactionsData.Lock()
	defer o.muTransactionsData.Unlock()
	// buat timer untuk cancel ordernya!
	orderDeadline := time.AfterFunc(2*time.Hour, func() {
		o.muTransactionsData.Lock()
		defer o.muTransactionsData.Unlock()

		if _, ok := o.transactionsData[result.ID]; !ok {
			return
		}

		opertionContext, cancel := context.WithTimeout(o.appContext, 10*time.Second)
		defer cancel()

		err := o.orderRepo.Cancel(opertionContext, result.ID)
		if err != nil {
			log.Println("failed auto cancel:", err)
			return
		}
		delete(o.transactionsData, result.ID)
	})

	o.transactionsData[result.ID] = orderDeadline

	return result, nil
}

func (o *OrderService) GetAllOrderByUserId(ctx context.Context, userId uuid.UUID) ([]model.Order, *common.ErrorResponse) {

	data, err := o.orderRepo.GetByUserIDWithDetails(ctx, userId)
	if err != nil {
		return []model.Order{}, common.NewErrorResponse(500, "gagal mengambil data di database!"+err.Error())
	}
	return data, nil
}

func (o *OrderService) GetByOrderId(ctx context.Context, orderId uuid.UUID, userID uuid.UUID) (model.Order, *common.ErrorResponse) {

	data, err := o.orderRepo.GetByIDWithDetails(ctx, orderId)

	if err != nil {
		if errors.Is(err, ErrNoOrderFound) {
			return model.Order{}, common.NewErrorResponse(404, "order tidak ditemukan!")
		}
		return model.Order{}, common.NewErrorResponse(500, "gagal mengambil data di database "+err.Error())
	}

	if data.UserID != userID {
		return model.Order{}, common.NewErrorResponse(401, "kamu tidak bisa mengakses data ini!")
	}

	return *data, nil

}
