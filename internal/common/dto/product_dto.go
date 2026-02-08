package dto

import (
	"backEnd-RingoTechLife/internal/common/model"
	"encoding/json"
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
