package servicerequest

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ServiceRequestHandler struct {
	service   *DeviceService
	decoder   *form.Decoder
	validator *validator.Validate
}

const maxDeviceFormSize = 6 << 20
const maxFileSize = 2 << 20

func NewServiceRequestHandler(svc *DeviceService, dcd *form.Decoder, vld *validator.Validate) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		service:   svc,
		decoder:   dcd,
		validator: vld,
	}
}

func (sr *ServiceRequestHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sr.service.GetAllServiceRequest(r.Context())
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (sr *ServiceRequestHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(maxDeviceFormSize); err != nil {
		pkg.JSONError(w, 400, "gagal parse form data")
		return
	}

	defer r.MultipartForm.RemoveAll()

	var req dto.CreateServiceRequestDTO

	if err := sr.decoder.Decode(&req, r.MultipartForm.Value); err != nil {
		// fmt.Println(err)
		pkg.JSONError(w, 400, "form data tidak valid")
		return
	}

	req.ProductPictures = r.MultipartForm.File["device_images"]
	for _, fileHeader := range req.ProductPictures {
		if fileHeader.Size > maxFileSize {
			pkg.JSONError(w, 400, "gambar terlalu besar! maksimal 4mb")
			return
		}
	}

	if len(req.ProductPictures) > 3 {
		pkg.JSONError(w, 400, "minmal 1 dan maksimal 3 gambar dalam sekali upload!")
		return
	}

	if err := sr.validator.Struct(req); err != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	userId, _ := middleware.GetUserID(r.Context())

	insertData, insErr := sr.service.CreateNew(r.Context(), req, userId)
	if insErr != nil {
		pkg.JSONError(w, insErr.Code, insErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "Berhasil menambahkan data", insertData)
}

func (sr *ServiceRequestHandler) GetMyServiceHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userId, _ := middleware.GetUserID(r.Context())

	data, err := sr.service.GetCurrentUserHistory(r.Context(), userId)
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "Berhasil mengambil data", data)

}
func (sr *ServiceRequestHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	userId, _ := middleware.GetUserID(r.Context())
	role, _ := middleware.GetRole(r.Context())
	idStr := chi.URLParam(r, "id")
	serviceId, err := uuid.Parse(idStr)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	data, getErr := sr.service.GetByServiceID(r.Context(), serviceId, userId, role)
	if getErr != nil {
		pkg.JSONError(w, getErr.Code, getErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "Berhasil mengambil data", data)
}

func (sr *ServiceRequestHandler) QuoteServiceHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userId, _ := middleware.GetUserID(r.Context())
	serviceId, err := uuid.Parse(idStr)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}
	var req dto.AdminQuoteServiceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Body tidak valid! harap masukan data dengan benar")
		return
	}

	if err := sr.validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	qerr := sr.service.QuoteService(r.Context(), serviceId, req, userId)
	if qerr != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	pkg.JSONSuccess(w, 200, "Berhasil mengambil data", nil)
}

func (sr *ServiceRequestHandler) RejectServiceHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	adminId, _ := middleware.GetUserID(r.Context())
	serviceId, err := uuid.Parse(idStr)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}
	var req dto.AdminRejectServiceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Body tidak valid! harap masukan data dengan benar")
		return
	}
	if err := sr.validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)
		pkg.JSONError(w, 400, validationErr)
		return
	}
	if rerr := sr.service.RejectService(r.Context(), serviceId, req, adminId); rerr != nil {
		pkg.JSONError(w, rerr.Code, rerr.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "Request service berhasil ditolak", nil)
}

func (sr *ServiceRequestHandler) UserDecisionHandler(w http.ResponseWriter, r *http.Request) {
	userId, _ := middleware.GetUserID(r.Context())
	idStr := chi.URLParam(r, "id")
	serviceId, err := uuid.Parse(idStr)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}
	var req dto.UserDecisionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Body tidak valid! harap masukan data dengan benar")
		return
	}
	if err := sr.validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)
		pkg.JSONError(w, 400, validationErr)
		return
	}

	if req.Accept {
		if aerr := sr.service.AcceptServiceByUser(r.Context(), serviceId, userId); aerr != nil {
			pkg.JSONError(w, aerr.Code, aerr.Message)
			return
		}
		pkg.JSONSuccess(w, 200, "Penawaran berhasil diterima", nil)
		return
	}

	if rerr := sr.service.RejectServiceByUser(r.Context(), serviceId, userId); rerr != nil {
		pkg.JSONError(w, rerr.Code, rerr.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "Penawaran berhasil ditolak", nil)
}

func (sr *ServiceRequestHandler) SetUpRoute(router chi.Router) {

	router.Route("/device-service", func(r chi.Router) {
		r.Use(httprate.Limit(
			50,
			time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Tolong jangan lakukan request terlalu banyak")

				errorResJson, _ := json.Marshal(errorRes)

				w.Write(errorResJson)

			}),
		))
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RoleMiddleware(middleware.RoleAdmin, middleware.RoleUser))

		r.Get("/get-my-service", sr.GetMyServiceHistoryHandler)
		r.Get("/details/{id}", sr.GetDetails)
		r.Put("/status-service/{id}", sr.UserDecisionHandler)
		r.Post("/new", sr.CreateHandler)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware(middleware.RoleAdmin))
			r.Get("/get-all", sr.GetAllHandler)
			r.Post("/quote-service/{id}", sr.QuoteServiceHandler)
			r.Put("/admin-reject/{id}", sr.RejectServiceHandler)
		})
	})
}
