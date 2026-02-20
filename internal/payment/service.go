package payment

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"errors"
	"mime/multipart"

	"github.com/google/uuid"
)

const paymentImagePlace = "payments"

type PayementService struct {
	paymentRepo PaymentRepositoryInterface
	fileStorage *storage.FileStorage
}

func NewPaymentService(repo *PaymentRepositoryImpl, storage *storage.FileStorage) *PayementService {
	return &PayementService{
		paymentRepo: repo,
		fileStorage: storage,
	}
}

func (ps *PayementService) SubmitProof(ctx context.Context, submitReq dto.SubmitPaymentRequest, currentUser uuid.UUID) (model.Payment, *common.ErrorResponse) {

	orderId, err := uuid.Parse(submitReq.OrderId)
	// fmt.Println(orderId)
	if err != nil {
		return model.Payment{}, common.NewErrorResponse(400, "id payment tidak valid!")
	}

	//save dulu bg image nya!
	savedFileNames, svImgErr := ps.processPaymentImages(submitReq.ProofImage)
	if svImgErr != nil {
		return model.Payment{}, svImgErr
	}

	val, err := ps.paymentRepo.GetPaymentValidationData(ctx, orderId)
	if err != nil {
		ps.fileStorage.DeletePublicFile(savedFileNames, paymentImagePlace)

		if errors.Is(err, ErrNoPaymentfound) {
			return model.Payment{}, common.NewErrorResponse(404, "data order tidak ditemukan!")
		}

		return model.Payment{}, common.NewErrorResponse(500, "gagal memproses order! "+err.Error())
	}

	if val.IssuerId != currentUser {
		ps.fileStorage.DeletePublicFile(savedFileNames, paymentImagePlace)
		return model.Payment{}, common.NewErrorResponse(401, "kamu tidak mengakses data ini")
	}

	if val.OrderStatus != string(model.OrderStatusPending) {
		ps.fileStorage.DeletePublicFile(savedFileNames, paymentImagePlace)
		return model.Payment{}, common.NewErrorResponse(400, "waktu pembayaran untuk order ini sudah habis")
	}

	tempData := model.Payment{
		OrderID:    orderId,
		ProofImage: &savedFileNames,
		Amount:     val.Amount,
	}

	subErr := ps.paymentRepo.SubmitProof(ctx, &tempData)
	if subErr != nil {
		return model.Payment{}, common.NewErrorResponse(500, "gagal memproses transaksi! "+subErr.Error())
	}

	return tempData, nil

}

func (ps *PayementService) processPaymentImages(fileheader *multipart.FileHeader) (string, *common.ErrorResponse) {

	mimeType, err := ps.fileStorage.DetectFileType(fileheader)
	if err != nil {
		return "", common.NewErrorResponse(500, "gagal memproses file! mungkin file tidak didukung")
	}

	ext, ok := ps.fileStorage.IsTypeSupportted(mimeType)
	if !ok {
		return "", common.NewErrorResponse(400, "format file tidak didukung!")
	}

	savedFileName, err := ps.fileStorage.SavePublicFile(fileheader, ext, paymentImagePlace)
	if err != nil {
		return "", common.NewErrorResponse(500, "gagal menyimpan gambar di database! "+err.Error())
	}

	return savedFileName, nil

}
