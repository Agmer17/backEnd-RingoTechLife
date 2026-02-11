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
	ID             uuid.UUID
	CategoryID     uuid.UUID
	Name           string
	Slug           string
	Description    *string
	Brand          *string
	Condition      ProductCondition
	Price          float64
	Stock          int
	SKU            *string
	Specifications JSONB
	Status         ProductStatus
	IsFeatured     bool
	Weight         *int
	Images         []ProductImage
	Category       Category
	CreatedAt      time.Time
}

type JSONB map[string]interface{}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
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
