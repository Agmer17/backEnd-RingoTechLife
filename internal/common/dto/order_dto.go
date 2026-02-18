package dto

type CreateOrderRequest struct {
	ProductId string  `json:"product_id" validate:"required,uuid"`
	Quantity  int     `json:"product_quantity" validate:"required,min=1,max=1000"`
	Notes     *string `json:"order_notes" validate:"omitempty"`
}
