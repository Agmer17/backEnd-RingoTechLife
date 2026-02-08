package products

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type ProductsService struct {
	repo ProductRepositoryInterface
}

func NewProductsService(rp *ProductRepositoryImpl) *ProductsService {
	return &ProductsService{
		repo: rp,
	}
}

func (p *ProductsService) Create(ctx context.Context, product dto.CreateProductRequest) (model.Product, *common.ErrorResponse) {
	productModel, err := dto.NewProductFromCreateRequest(product)
	if err != nil {
		return model.Product{}, common.NewErrorResponse(400, "gagal convert form data!"+err.Error())
	}

	fmt.Println(productModel)
	data, err := p.repo.Create(ctx, &productModel)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "products_slug_key":
					return model.Product{}, common.NewErrorResponse(409, "nama slug sudah ada! silahkan pilih nama lain")
				case "products_sku_key":
					return model.Product{}, common.NewErrorResponse(409, "sku sudah ada! silahkan masukan yang lain")
				}
			}
			if pgErr.Code == "23503" {
				return model.Product{}, common.NewErrorResponse(404, "tag id tidak ditemukan! harap masukan data dengan benar "+err.Error())
			}
		}

		return model.Product{}, common.NewErrorResponse(500, "gagal insert ke database! : "+err.Error())
	}
	return *data, nil
}

func (p *ProductsService) GetAllProducts(ctx context.Context) ([]model.Product, *common.ErrorResponse) {

	data, err := p.repo.GetAllProducts(ctx)

	if err != nil {
		return []model.Product{}, common.NewErrorResponse(500, "gagal mengambil data dari database :"+err.Error())
	}

	return data, nil
}
