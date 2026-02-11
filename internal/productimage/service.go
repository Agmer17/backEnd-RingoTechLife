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

	var filesExt []string = make([]string, len(files))

	for i, f := range files {

		mimetype, err := p.fileStorage.DetectFileType(f)

		if err != nil {
			return []*model.ProductImage{}, common.NewErrorResponse(500, "gagal saat membaca file")
		}

		ext, ok := p.fileStorage.IsTypeSupportted(mimetype)

		if !ok {
			return []*model.ProductImage{}, common.NewErrorResponse(400, "format file tidak didukung! harap hanya masukan gambar!")
		}
		filesExt[i] = ext
	}

	savedFileNames, err := p.fileStorage.SaveAllPublicFiles(ctx, files, filesExt, productImagePlace)

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
