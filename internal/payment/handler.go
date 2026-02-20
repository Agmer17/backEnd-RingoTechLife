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
)

const maxFormSize = 5 << 20

type PaymentHandler struct {
	payementService *PayementService
	decoder         *form.Decoder
}

func NewPaymentHandler(pay *PayementService, dec *form.Decoder) *PaymentHandler {
	return &PaymentHandler{
		payementService: pay,
		decoder:         dec,
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
	})

}
