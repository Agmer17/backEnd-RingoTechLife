package productimage

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"mime/multipart"

	"github.com/google/uuid"
)

const productImagePlace = "products"

type ProductImageService struct {
	productImageRepo ProductImageRepoInterface
	fileStorage      *storage.FileStorage
}

func NewProductImageService(repo *ProductImageRepoImpl, serverStorage *storage.FileStorage) *ProductImageService {
	return &ProductImageService{
		productImageRepo: repo,
		fileStorage:      serverStorage,
	}
}

func (p *ProductImageService) SaveAllImages(ctx context.Context, files []*multipart.FileHeader, productId uuid.UUID) ([]*model.ProductImage, *common.ErrorResponse) {

	savedFileNames, err := p.processImageToServer(ctx, files)
	if err != nil {
		p.fileStorage.DeleteAllPublicFile(savedFileNames, productImagePlace)
		return []*model.ProductImage{}, common.NewErrorResponse(500, "gagal saat menyimpan file ke server "+err.Error())
	}

	var imageModel []*model.ProductImage = make([]*model.ProductImage, len(savedFileNames))

	for i, v := range savedFileNames {
		tmpModel := model.ProductImage{
			ProductID:    productId,
			ImageURL:     v,
			IsPrimary:    i == 0,
			DisplayOrder: i,
		}

		imageModel[i] = &tmpModel

	}

	saved, err := p.productImageRepo.CreateBulk(ctx, imageModel)

	if err != nil {
		p.fileStorage.DeleteAllPublicFile(savedFileNames, productImagePlace)
		return []*model.ProductImage{}, common.NewErrorResponse(500, "gagal menyimpan data di database! "+err.Error())
	}

	return saved, nil

}

func (p *ProductImageService) processImageToServer(ctx context.Context, files []*multipart.FileHeader) ([]string, error) {

	var filesExt []string = make([]string, len(files))

	for i, f := range files {

		mimetype, err := p.fileStorage.DetectFileType(f)

		if err != nil {
			return []string{}, err
		}

		ext, ok := p.fileStorage.IsTypeSupportted(mimetype)

		if !ok {
			return []string{}, err
		}
		filesExt[i] = ext
	}

	return p.fileStorage.SaveAllPublicFiles(ctx, files, filesExt, productImagePlace)
}

func (p *ProductImageService) UpdateProductsImage(
	ctx context.Context,
	productId uuid.UUID,
	fUpdate []*multipart.FileHeader,
	imgIds []uuid.UUID,
) ([]*model.ProductImage, *common.ErrorResponse) {

	if len(fUpdate) != len(imgIds) {
		return nil, common.NewErrorResponse(400, "jumlah id dan gambar tidak sama")
	}

	savedFileNames, err := p.processImageToServer(ctx, fUpdate)
	if err != nil {
		return nil, common.NewErrorResponse(500, err.Error())
	}

	currentImages, err := p.productImageRepo.GetAllByIDs(ctx, imgIds)
	if err != nil {
		p.fileStorage.DeleteAllPublicFile(savedFileNames, productImagePlace)
		return nil, common.NewErrorResponse(500, err.Error())
	}

	for i := range currentImages {

		if currentImages[i].ProductID != productId {
			p.fileStorage.DeleteAllPublicFile(savedFileNames, productImagePlace)
			return nil, common.NewErrorResponse(403, "image tidak sesuai product")
		}

		currentImages[i].ImageURL = savedFileNames[i]
	}

	updated, err := p.productImageRepo.UpdateBulk(ctx, currentImages)
	if err != nil {
		p.fileStorage.DeleteAllPublicFile(savedFileNames, productImagePlace)
		return nil, common.NewErrorResponse(500, err.Error())
	}

	return updated, nil
}
