package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ProductCondition string
type ProductStatus string

const (
	ProductConditionNew         ProductCondition = "new"
	ProductConditionUsed        ProductCondition = "used"
	ProductConditionRefurbished ProductCondition = "refurbished"
)

// , 'active', 'inactive', 'out_of_stock'
const (
	ProductStatusInActive    ProductStatus = "inactive"
	ProductsStatusActive     ProductStatus = "active"
	ProductsStatusOutOfStock ProductStatus = "out_of_stock"
)

type Product struct {
	ID             uuid.UUID        `json:"product_id"`
	CategoryID     uuid.UUID        `json:"product_category_id"`
	Name           string           `json:"product_name"`
	Slug           string           `json:"product_slug"`
	Description    *string          `json:"product_description"`
	Brand          *string          `json:"product_brand"`
	Condition      ProductCondition `json:"product_condition"`
	Price          float64          `json:"product_price"`
	Stock          int              `json:"product_stock"`
	SKU            *string          `json:"product_sku"`
	Specifications JSONB            `json:"product_specification"`
	Status         ProductStatus    `json:"product_status"`
	IsFeatured     bool             `json:"product_is_featured"`
	Weight         *int             `json:"product_weight"`
	Images         []ProductImage   `json:"product_images"`
	Category       Category         `json:"category"`
	CreatedAt      time.Time        `json:"product_created_at"`
}

type JSONB map[string]any

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	result := make(map[string]any)
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
