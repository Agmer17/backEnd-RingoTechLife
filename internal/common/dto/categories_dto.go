package dto

type CreateCategoryRequest struct {
	Name string  `json:"category_name" validate:"required,min=2,max=100"`
	Slug string  `json:"category_slug" validate:"required,slug,max=100"`
	Desc *string `json:"category_description" validate:"omitempty,max=255"`
}

type UpdateCategoryRequest struct {
	Name *string `json:"category_name" validate:"omitempty,min=2,max=100"`
	Slug *string `json:"category_slug" validate:"omitempty,slug,max=100"`
	Desc *string `json:"category_description" validate:"omitempty,max=255"`
}
