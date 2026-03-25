package servicerequest

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/internal/order"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type DeviceService struct {
	DeviceServiceRepo ServiceRequestRepositoryInterface
	FileStorage       storage.FileStorage
	OrderService      *order.OrderService
}

func NewDeviceService(drp *ServiceRequestRepository, serverStorage storage.FileStorage, ord *order.OrderService) *DeviceService {
	return &DeviceService{
		DeviceServiceRepo: drp,
		FileStorage:       serverStorage,
		OrderService:      ord,
	}
}

func (ds *DeviceService) CreateNew(ctx context.Context, newData dto.CreateServiceRequestDTO, userId uuid.UUID) (model.ServiceRequest, *common.ErrorResponse) {
	savedImages, err := ds.processDeviceImage(ctx, newData.ProductPictures)
	if err != nil {
		ds.FileStorage.DeleteAllPublicFile(savedImages, "device_service")
		return model.ServiceRequest{}, common.NewErrorResponse(500, "gagal menyimpan gambar ke server")
	}

	var photo1, photo2, photo3 *string

	if len(savedImages) > 0 {
		photo1 = &savedImages[0]
	}
	if len(savedImages) > 1 {
		photo2 = &savedImages[1]
	}
	if len(savedImages) > 2 {
		photo3 = &savedImages[2]
	}

	newModel := model.ServiceRequest{
		DeviceType:         newData.DeviceType,
		DeviceBrand:        newData.DeviceBrand,
		DeviceModel:        newData.DeviceModel,
		ProblemDescription: newData.ProblemDescription,
		Photo1:             photo1,
		Photo2:             photo2,
		Photo3:             photo3,
		UserID:             userId,
	}

	err = ds.DeviceServiceRepo.Create(ctx, &newModel)
	if err != nil {
		return model.ServiceRequest{}, common.NewErrorResponse(500, "gagal menyimpan data ke database")
	}

	return newModel, nil
}

func (ds *DeviceService) GetAllServiceRequest(ctx context.Context) ([]model.ServiceRequest, *common.ErrorResponse) {

	data, err := ds.DeviceServiceRepo.GetAll(ctx)
	if err != nil {
		return []model.ServiceRequest{}, common.NewErrorResponse(500, err.Error())
	}

	respData := make([]model.ServiceRequest, len(data))

	for i, v := range data {
		respData[i] = *v
	}
	return respData, nil
}

func (ds *DeviceService) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]model.ServiceRequest, *common.ErrorResponse) {

	data, err := ds.DeviceServiceRepo.GetByUserID(ctx, userId)
	if err != nil {
		return []model.ServiceRequest{}, common.NewErrorResponse(500, err.Error())
	}

	respData := make([]model.ServiceRequest, len(data))

	for i, v := range data {
		respData[i] = *v
	}
	return respData, nil
}

func (ds *DeviceService) GetByServiceID(ctx context.Context, id uuid.UUID, userId uuid.UUID, role string) (model.ServiceRequest, *common.ErrorResponse) {
	data, err := ds.DeviceServiceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, notFoundError) {
			return model.ServiceRequest{}, common.NewErrorResponse(404, "data tidak ditemukan!")
		}
	}

	if data.UserID != userId && role != middleware.RoleAdmin {
		return model.ServiceRequest{}, common.NewErrorResponse(403, "Kamu tidak dapat mengakses fitur ini!")
	}

	return *data, nil
}

func (ds *DeviceService) processDeviceImage(ctx context.Context, files []*multipart.FileHeader) ([]string, error) {

	var filesExt []string = make([]string, len(files))

	for i, f := range files {

		mimetype, err := ds.FileStorage.DetectFileType(f)

		if err != nil {
			return []string{}, err
		}

		ext, ok := ds.FileStorage.IsTypeSupportted(mimetype)

		if !ok {
			return []string{}, err
		}
		filesExt[i] = ext
	}

	return ds.FileStorage.SaveAllPublicFiles(ctx, files, filesExt, "device_service")
}

func (ds *DeviceService) GetCurrentUserHistory(ctx context.Context, userId uuid.UUID) ([]model.ServiceRequest, *common.ErrorResponse) {
	data, err := ds.DeviceServiceRepo.GetByUserID(ctx, userId)
	if err != nil {
		return []model.ServiceRequest{}, common.NewErrorResponse(500, "Gagal mengambil data di database!")
	}

	respData := make([]model.ServiceRequest, len(data))

	for i, v := range data {
		respData[i] = *v
	}
	return respData, nil
}

func (ds *DeviceService) QuoteService(ctx context.Context, serviceId uuid.UUID, d dto.AdminQuoteServiceRequestDTO, adminId uuid.UUID) *common.ErrorResponse {
	err := ds.DeviceServiceRepo.AdminQuote(ctx, serviceId, &d, adminId)
	if err != nil {
		return common.NewErrorResponse(500, "terjadi kesalahan di server")
	}

	return nil
}

func (ds *DeviceService) RejectService(ctx context.Context, serviceId uuid.UUID, d dto.AdminRejectServiceRequestDTO, adminId uuid.UUID) *common.ErrorResponse {
	err := ds.DeviceServiceRepo.AdminReject(ctx, serviceId, &d, adminId)
	if err != nil {
		return common.NewErrorResponse(500, "terjadi kesalahan di server")
	}
	return nil
}

func (ds *DeviceService) AcceptServiceByUser(ctx context.Context, serviceId uuid.UUID, userId uuid.UUID) *common.ErrorResponse {
	oldData, err := ds.DeviceServiceRepo.GetByID(ctx, serviceId)
	if err != nil {
		return common.NewErrorResponse(500, "gagal mengambil data di database")

	}
	if oldData.UserID != userId {
		return common.NewErrorResponse(403, "Kamu tidak dapat mengakses ini")
	}

	expiresAt := time.Now().UTC().Add(12 * time.Hour)
	newInserted, insErr := ds.OrderService.CreateOrderWithoutProduct(ctx, userId, "Silahkan kirim produk ke a;amat ringotechlife atau bisa datang langsung", *oldData.QuotedPrice, expiresAt)
	if insErr != nil {
		return insErr
	}

	err = ds.DeviceServiceRepo.UserAccept(ctx, serviceId, newInserted.ID)
	if err != nil {
		return common.NewErrorResponse(500, "terjadi kesalahan di server")
	}
	return nil
}

func (ds *DeviceService) RejectServiceByUser(ctx context.Context, serviceId uuid.UUID, userId uuid.UUID) *common.ErrorResponse {
	oldData, err := ds.DeviceServiceRepo.GetByID(ctx, serviceId)
	if err != nil {
		return common.NewErrorResponse(500, "gagal mengambil data di database")

	}

	if oldData.UserID != userId {
		return common.NewErrorResponse(403, "Kamu tidak dapat mengakses ini")
	}

	err = ds.DeviceServiceRepo.UserReject(ctx, serviceId)
	if err != nil {
		return common.NewErrorResponse(500, "terjadi kesalahan di server")
	}
	return nil
}
