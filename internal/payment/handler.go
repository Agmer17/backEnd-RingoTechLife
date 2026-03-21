package payment

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const maxFormSize = 5 << 20

type PaymentHandler struct {
	payementService *PayementService
	decoder         *form.Decoder
	Validator       *validator.Validate
}

func NewPaymentHandler(pay *PayementService, dec *form.Decoder, vld *validator.Validate) *PaymentHandler {
	return &PaymentHandler{
		payementService: pay,
		decoder:         dec,
		Validator:       vld,
	}
}

func (p *PaymentHandler) SubmitPaymentHandler(w http.ResponseWriter, r *http.Request) {

	currentUser, _ := middleware.GetUserID(r.Context())

	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		pkg.JSONError(w, 400, "gagal parse form data "+err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll()

	var submitReq dto.SubmitPaymentRequest
	if err := p.decoder.Decode(&submitReq, r.MultipartForm.Value); err != nil {
		fmt.Println(err)
		pkg.JSONError(w, 400, "form data tidak valid")
		return
	}

	submitReq.ProofImage = r.MultipartForm.File["proof_image"][0]

	if submitReq.ProofImage == nil {
		pkg.JSONError(w, 400, "Bukti gambar tidak ditemukan! harap isi data dengan benar!")
		return

	}

	fmt.Println(submitReq.OrderId)
	result, subErr := p.payementService.SubmitProof(r.Context(), submitReq, currentUser)
	if subErr != nil {
		pkg.JSONError(w, subErr.Code, subErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menyimpan pembayaran!", result)

}

func (p *PaymentHandler) AcceptPaymentHandler(w http.ResponseWriter, r *http.Request) {

	var req dto.UpdatePaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Harap isi data dengan benar!")
		return
	}

	if err := p.Validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	adminId, _ := middleware.GetUserID(r.Context())

	paymentUuid, err := uuid.Parse(req.PaymentId)
	if err != nil {
		pkg.JSONError(w, 401, "Kamu tidak bisa mengakses ini!")
		return
	}

	uptErr := p.payementService.AcceptPayment(r.Context(), paymentUuid, adminId, req.Note)
	if uptErr != nil {
		pkg.JSONError(w, uptErr.Code, uptErr.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "Berhasil mengupdate status pembayaran!", nil)

}

func (p *PaymentHandler) RejectPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Harap isi data dengan benar!")
		return
	}

	if err := p.Validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	adminId, _ := middleware.GetUserID(r.Context())

	paymentUuid, err := uuid.Parse(req.PaymentId)
	if err != nil {
		pkg.JSONError(w, 401, "Kamu tidak bisa mengakses ini!")
		return
	}

	uptErr := p.payementService.RejectPayment(r.Context(), paymentUuid, adminId, "gagal melakukan pembayaran")
	if uptErr != nil {
		pkg.JSONError(w, uptErr.Code, uptErr.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "Berhasil mengupdate status pembayaran!", nil)

}

func (p *PaymentHandler) SetupRoute(router chi.Router) {

	router.Route("/payments", func(r chi.Router) {
		r.Use(httprate.Limit(
			20,
			time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Terlalu banyak request, coba lagi nanti")
				errorResJson, _ := json.Marshal(errorRes)
				w.Write(errorResJson)
			}),
		))
		r.Use(middleware.AuthMiddleware)

		r.Post("/order", p.SubmitPaymentHandler)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware(middleware.RoleAdmin))
			r.Post("/accept", p.AcceptPaymentHandler)
			r.Post("/reject", p.RejectPaymentHandler)
		})
	})

}
