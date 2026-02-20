package order

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/products"
	"context"
	"errors"
	"fmt"
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
	orderDeadline := time.AfterFunc(2*time.Minute, func() {
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

	fmt.Println(o.transactionsData)

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

func (o *OrderService) GetAllOrders(ctx context.Context) ([]model.Order, *common.ErrorResponse) {

	data, err := o.orderRepo.GetAllWithDetails(ctx)

	if err != nil {
		return []model.Order{}, common.NewErrorResponse(500, "gagal mengambil data di database! "+err.Error())
	}

	return data, nil

}

func (o *OrderService) GetAllOrdersByStatus(ctx context.Context, status string) ([]model.Order, *common.ErrorResponse) {
	_, ok := model.AllowedOrderStatus[status]
	fmt.Println(ok)
	if !ok {
		// default aja ya dulu!
		status = string(model.OrderStatusPending)
	}
	fmt.Println(status)
	data, err := o.orderRepo.GetByStatus(ctx, model.OrderStatus(status))
	if err != nil {
		return []model.Order{}, common.NewErrorResponse(500, "gagal membaca data dari database "+err.Error())
	}
	return data, nil
}

func (o *OrderService) UpdateOrderStatus(ctx context.Context, prodId uuid.UUID, status string) *common.ErrorResponse {

	err := o.orderRepo.UpdateStatus(ctx, prodId, model.OrderStatus(status))
	if err != nil {

		if errors.Is(err, ErrNoOrderFound) {
			return common.NewErrorResponse(404, "order tidak ditemukan!")
		}

		return common.NewErrorResponse(500, "terjadi kesalahan di database "+err.Error())
	}
	return nil
}

func (o *OrderService) DeleteTransactionDeadline(orderId uuid.UUID) {

	o.muTransactionsData.Lock()
	defer o.muTransactionsData.Unlock()

	deadline, ok := o.transactionsData[orderId]

	if !ok {
		// udah terlanjur jalan ygy
		return
	}
	deadline.Stop()
	delete(o.transactionsData, orderId)

	fmt.Println(o.transactionsData)
}
