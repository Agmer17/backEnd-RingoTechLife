package category

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"

	"github.com/google/uuid"
)

type CategoryService struct {
	repository CategoryRepositoryInterface
}

func NewCategoryService(repo *CategoryRepositoryImpl) *CategoryService {
	return &CategoryService{
		repository: repo,
	}
}

func (cs *CategoryService) GetAllCategories(ctx context.Context) ([]model.Category, *common.ErrorResponse) {
	data, err := cs.repository.GetAllCategories(ctx)
	if err != nil {
		return []model.Category{}, common.NewErrorResponse(500, "internal server error "+err.Error())
	}
	return data, nil
}

func (cs *CategoryService) CreateNewCategory(ctx context.Context, data model.Category) (model.Category, *common.ErrorResponse) {

	exist, err := cs.repository.ExistByNameOrSlug(ctx, data.Name, data.Slug, nil)

	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "gagal melakukan operasi di database "+err.Error())
	}

	if exist {
		return model.Category{}, common.NewErrorResponse(409, "nama atau slug sudah ada di databse")

	}
	// anggep data dah valid dari fe
	// soalnya tinggal pake validator nanti di handler buat cek
	result, err := cs.repository.Create(ctx, &data)

	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "gagal melakukan operasi di database "+err.Error())
	}

	return *result, nil

}

func (cs *CategoryService) UpdateCategories(ctx context.Context, id uuid.UUID, updateData dto.UpdateCategoryRequest) (model.Category, *common.ErrorResponse) {

	existData, err := cs.repository.GetByID(ctx, id)

	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
	}

	if updateData.Name != nil && *updateData.Name != "" {

		nameExist, err := cs.repository.ExistsByName(ctx, *updateData.Name, &existData.ID)

		if err != nil {
			return model.Category{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
		}

		if nameExist {
			return model.Category{}, common.NewErrorResponse(409, "nama category sudah ada di database!")

		}

		existData.Name = *updateData.Name
	}

	if updateData.Slug != nil && *updateData.Slug != "" {

		nameExist, err := cs.repository.ExistsByName(ctx, *updateData.Slug, &existData.ID)

		if err != nil {
			return model.Category{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
		}

		if nameExist {
			return model.Category{}, common.NewErrorResponse(409, "nama category sudah ada di database!")

		}

		existData.Name = *updateData.Slug
	}

	if updateData.Desc != nil && *updateData.Desc != "" {
		existData.Description = updateData.Desc
	}

	updatedData, err := cs.repository.Update(ctx, &existData)

	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "gagal mengupdate data di database! "+err.Error())

	}

	return *updatedData, nil
}

func (cs *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) *common.ErrorResponse {
	exist, err := cs.repository.ExistsById(ctx, id)
	if err != nil {
		return common.NewErrorResponse(500, "internal server error! "+err.Error())
	}
	if !exist {
		return common.NewErrorResponse(404, "id category tidak ditemukan!")
	}

	err = cs.repository.Delete(ctx, id)
	if err != nil {
		return common.NewErrorResponse(500, "internal server error! "+err.Error())
	}
	return nil
}

func (cs *CategoryService) GetById(ctx context.Context, id uuid.UUID) (model.Category, *common.ErrorResponse) {

	data, err := cs.repository.GetByID(ctx, id)
	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "internal server error! "+err.Error())
	}
	return data, nil
}

func (cs *CategoryService) GetBySlug(ctx context.Context, slug string) (model.Category, *common.ErrorResponse) {

	data, err := cs.repository.GetBySlug(ctx, slug)
	if err != nil {
		return model.Category{}, common.NewErrorResponse(500, "internal server error! "+err.Error())
	}
	return data, nil
}

func (cs *CategoryService) IsCategoryIdExist(ctx context.Context, id uuid.UUID) (bool, *common.ErrorResponse) {

	exist, err := cs.repository.ExistsById(ctx, id)

	if err != nil {
		return false, common.NewErrorResponse(500, "gagal mengambil data di datbase!"+err.Error())
	}

	return exist, nil

}
