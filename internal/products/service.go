package products

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/productimage"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"encoding/json"
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

func (p *ProductsService) UpdateProducts(ctx context.Context, reqData dto.UpdateProductsRequest, id uuid.UUID) (model.Product, *common.ErrorResponse) {

	oldData, err := p.repo.GetByID(ctx, id)

	if err != nil {

		if errors.Is(err, ErrProductNotFound) {
			return model.Product{}, common.NewErrorResponse(400, "product tidak ditemukan!")
		}
		return model.Product{}, common.NewErrorResponse(500, "gagal mengambil data di database")
	}

	err = applyUpdateProductRequest(&oldData, &reqData)
	if err != nil {
		return model.Product{}, common.NewErrorResponse(400, "data yang kamu kirim tidak valid! "+err.Error())
	}

	// udah di apply changes di old data
	updatedData, err := p.repo.Update(ctx, &oldData)

	if err != nil {
		if errors.Is(err, ErrConflictSlugName) {
			return model.Product{}, common.NewErrorResponse(409, err.Error())
		}

		if errors.Is(err, ErrConflicSku) {
			return model.Product{}, common.NewErrorResponse(409, err.Error())
		}

		if errors.Is(err, ErrFkTagsConstraint) {
			return model.Product{}, common.NewErrorResponse(404, err.Error()+" operasi dibatalkan")
		}

		return model.Product{}, common.NewErrorResponse(500, "gagal mengupdate data di database!")
	}

	if len(reqData.UpdatedImage) != 0 {
		if len(reqData.UpdatedImage) != len(reqData.UpdatedImageFiles) {
			return model.Product{}, common.NewErrorResponse(400, "jumlah gambar yang dikirim tidak sama!")
		}

		var imgToUpdate []uuid.UUID = make([]uuid.UUID, len(reqData.UpdatedImage))

		for i, v := range reqData.UpdatedImage {
			temp, _ := uuid.Parse(v)

			imgToUpdate[i] = temp
		}

		updatedImageData, err := p.productImageService.UpdateProductsImage(ctx, updatedData.ID, reqData.UpdatedImageFiles, imgToUpdate)

		if err != nil {
			return model.Product{}, err
		}

		updatedData.Images = mergeUpdatedImages(updatedData.Images, updatedImageData)

		return *updatedData, nil
	}

	// todo impl delete image

	return *updatedData, nil

}

func applyUpdateProductRequest(p *model.Product, req *dto.UpdateProductsRequest) error {
	// CategoryID (string → uuid.UUID)
	if req.CategoryId != nil {
		parsedUUID, err := uuid.Parse(*req.CategoryId)
		if err != nil {
			return fmt.Errorf("invalid category id: %w", err)
		}
		p.CategoryID = parsedUUID
	}

	// Name
	if req.Name != nil {
		p.Name = *req.Name
	}

	// Slug
	if req.Slug != nil {
		p.Slug = *req.Slug
	}

	// Description
	if req.Description != nil {
		p.Description = req.Description
	}

	// Brand
	if req.Brand != nil {
		p.Brand = req.Brand
	}

	// Condition (string → enum)
	if req.Condition != nil {
		p.Condition = model.ProductCondition(*req.Condition)
	}

	// SKU
	if req.Sku != nil {
		p.SKU = req.Sku
	}

	// Price (float32 → float64)
	if req.Price != nil {
		p.Price = float64(*req.Price)
	}

	// Stock
	if req.Stock != nil {
		p.Stock = *req.Stock
	}

	// Specifications (string JSON → JSONB)
	if req.Specifications != nil {
		var jsonMap model.JSONB
		if err := json.Unmarshal([]byte(*req.Specifications), &jsonMap); err != nil {
			return fmt.Errorf("invalid specification json: %w", err)
		}
		p.Specifications = jsonMap
	}

	// Status (string → enum)
	if req.Status != nil {
		p.Status = model.ProductStatus(*req.Status)
	}

	// IsFeatured
	if req.IsFeatured != nil {
		p.IsFeatured = *req.IsFeatured
	}

	// Weight
	if req.Weight != nil {
		p.Weight = req.Weight
	}

	return nil
}

func mergeUpdatedImages(
	oldImages []model.ProductImage,
	updatedImages []*model.ProductImage,
) []model.ProductImage {

	updatedMap := make(map[uuid.UUID]*model.ProductImage, len(updatedImages))

	for _, img := range updatedImages {
		updatedMap[img.ID] = img
	}

	for i, old := range oldImages {
		if newImg, ok := updatedMap[old.ID]; ok {
			oldImages[i] = *newImg
		}
	}

	return oldImages
}
