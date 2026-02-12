package products

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/productimage"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ProductsService struct {
	repo                ProductRepositoryInterface
	fileStorage         *storage.FileStorage
	productImageService *productimage.ProductImageService
}

func NewProductsService(rp *ProductRepositoryImpl, fs *storage.FileStorage, img *productimage.ProductImageService) *ProductsService {
	return &ProductsService{
		repo:                rp,
		fileStorage:         fs,
		productImageService: img,
	}
}

func (p *ProductsService) Create(ctx context.Context, product dto.CreateProductRequest) (model.Product, []*model.ProductImage, *common.ErrorResponse) {
	productModel, err := dto.NewProductFromCreateRequest(product)
	if err != nil {
		return model.Product{}, []*model.ProductImage{}, common.NewErrorResponse(400, "gagal convert form data!"+err.Error())
	}

	fmt.Println(productModel)
	data, err := p.repo.Create(ctx, &productModel)
	if err != nil {
		if errors.Is(err, ErrConflictSlugName) {
			return model.Product{}, []*model.ProductImage{}, common.NewErrorResponse(409, err.Error())
		}

		if errors.Is(err, ErrConflicSku) {
			return model.Product{}, []*model.ProductImage{}, common.NewErrorResponse(409, err.Error())
		}

		if errors.Is(err, ErrFkTagsConstraint) {
			return model.Product{}, []*model.ProductImage{}, common.NewErrorResponse(404, err.Error()+" operasi dibatalkan")
		}

		return model.Product{}, []*model.ProductImage{}, common.NewErrorResponse(500, "gagal insert ke database! : "+err.Error())
	}

	savedImgModel, imgErr := p.productImageService.SaveAllImages(ctx, product.ProductImages, data.ID)

	if imgErr != nil {
		// batalin save productsnya! soalnya gagal banh!
		p.DeleteProducts(ctx, data.ID)
		return model.Product{}, []*model.ProductImage{}, imgErr
	}

	return *data, savedImgModel, nil
}

func (p *ProductsService) GetAllProducts(ctx context.Context) ([]model.Product, *common.ErrorResponse) {

	data, err := p.repo.GetAllProducts(ctx)

	if err != nil {
		return []model.Product{}, common.NewErrorResponse(500, "gagal mengambil data dari database :"+err.Error())
	}

	return data, nil
}

func (p *ProductsService) DeleteProducts(ctx context.Context, id uuid.UUID) *common.ErrorResponse {

	err := p.repo.Delete(ctx, id)
	if err != nil {

		if errors.Is(err, ErrProductNotFound) {
			return common.NewErrorResponse(404, "produk tidak ditemukan!")
		}

		return common.NewErrorResponse(500, "gagal menghapus data!")
	}

	return nil
}

func (p *ProductsService) GetById(ctx context.Context, id uuid.UUID) (model.Product, *common.ErrorResponse) {

	data, err := p.repo.GetByID(ctx, id)

	if err != nil {

		if errors.Is(err, ErrProductNotFound) {
			return model.Product{}, common.NewErrorResponse(404, "produk tidak ditemukan!")
		}

		return model.Product{}, common.NewErrorResponse(500, "gagal mengambil data di database! "+err.Error())
	}

	return data, nil

}

func (p *ProductsService) GetBySlug(ctx context.Context, slug string) (model.Product, *common.ErrorResponse) {

	data, err := p.repo.GetBySlug(ctx, slug)

	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return model.Product{}, common.NewErrorResponse(404, "produk tidak ditemukan!")
		}
		return model.Product{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
	}

	return data, nil
}

func (p *ProductsService) GetByCategorySlug(ctx context.Context, catSlug string) ([]model.Product, *common.ErrorResponse) {

	data, err := p.repo.GetProductsByCategorySlug(ctx, catSlug)

	if err != nil {
		return []model.Product{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
	}

	return data, nil

}

func (p *ProductsService) UpdateProducts() {

}
