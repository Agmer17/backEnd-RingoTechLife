package dto

import (
	"backEnd-RingoTechLife/internal/common/model"
	"encoding/json"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateProductRequest struct {
	CategoryId     string  `form:"product_category_id" validate:"required,uuid"`
	Name           string  `form:"product_name" validate:"required,min=3,alphanumspace"`
	Slug           string  `form:"product_slug" validate:"required,max=150,slug"`
	Description    *string `form:"product_description" validate:"omitempty"`
	Brand          *string `form:"product_brand" validate:"omitempty,max=100"`
	Condition      string  `form:"product_condition" validate:"required,oneof=new used refurbished"`
	Sku            string  `form:"product_sku" validate:"required"`
	Price          float32 `form:"product_price" validate:"required"`
	Stock          int     `form:"product_initial_stock" validate:"required"`
	Specifications *string `form:"product_specification" validate:"omitempty,json"`
	Status         string  `form:"product_status" validate:"required,oneof=draft active inactive out_of_stock"`
	IsFeatured     *bool   `form:"product_featured" validate:"omitempty"`
	Weight         *int    `form:"product_weight" validate:"omitempty,min=1"`
	ProductImages  []*multipart.FileHeader
}

type UpdateProductsRequest struct {
	CategoryId        *string  `form:"product_category_id" validate:"omitempty,uuid"`
	Name              *string  `form:"product_name" validate:"omitempty,min=3,alphanumspace"`
	Slug              *string  `form:"product_slug" validate:"omitempty,max=150,slug"`
	Description       *string  `form:"product_description" validate:"omitempty"`
	Brand             *string  `form:"product_brand" validate:"omitempty,max=100"`
	Condition         *string  `form:"product_condition" validate:"omitempty,oneof=new used refurbished"`
	Sku               *string  `form:"product_sku" validate:"omitempty"`
	Price             *float32 `form:"product_price" validate:"omitempty"`
	Stock             *int     `form:"product_initial_stock" validate:"omitempty"`
	Specifications    *string  `form:"product_specification" validate:"omitempty,json"`
	Status            *string  `form:"product_status" validate:"omitempty,oneof=draft active inactive out_of_stock"`
	IsFeatured        *bool    `form:"product_featured" validate:"omitempty"`
	Weight            *int     `form:"product_weight" validate:"omitempty,min=1"`
	NewProductImages  []*multipart.FileHeader
	DeletedImage      []string `form:"product_deleted_image"`
	UpdatedImage      []string `form:"product_updated_image_id"`
	UpdatedImageFiles []*multipart.FileHeader
}

type ProductDetailResponse struct {
	ID             uuid.UUID              `json:"product_id"`
	CategoryID     uuid.UUID              `json:"product_category_id"`
	Name           string                 `json:"product_name"`
	Slug           string                 `json:"product_slug"`
	Description    *string                `json:"product_description"`
	Brand          *string                `json:"product_brand"`
	Condition      model.ProductCondition `json:"product_condition"`
	Price          float64                `json:"product_price"`
	Stock          int                    `json:"product_stock"`
	SKU            *string                `json:"product_sku"`
	Specifications model.JSONB            `json:"product_specification"`
	Status         model.ProductStatus    `json:"product_status"`
	IsFeatured     bool                   `json:"product_is_featured"`
	Weight         *int                   `json:"product_weight"`
	Images         []model.ProductImage   `json:"product_images"`
	Category       model.Category         `json:"category"`
	CreatedAt      time.Time              `json:"product_created_at"`
	Reviews        []ReviewDetail         `json:"reviews"` // tambah ini
}

func NewProductFromCreateRequest(req CreateProductRequest) (model.Product, error) {
	var specs model.JSONB
	if req.Specifications != nil && *req.Specifications != "" {
		if err := json.Unmarshal([]byte(*req.Specifications), &specs); err != nil {
			specs = nil
		}
	}

	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	catId, err := uuid.Parse(req.CategoryId)

	if err != nil {
		return model.Product{}, err
	}

	return model.Product{
		ID:             uuid.Nil, // diisi DB
		CategoryID:     catId,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		Brand:          req.Brand,
		Condition:      model.ProductCondition(req.Condition),
		Price:          float64(req.Price),
		Stock:          req.Stock,
		SKU:            &req.Sku, // belum ada di request
		Specifications: specs,
		Status:         model.ProductStatus(req.Status),
		IsFeatured:     isFeatured,
		Weight:         req.Weight,
		CreatedAt:      time.Time{}, // diisi DB
	}, nil
}
