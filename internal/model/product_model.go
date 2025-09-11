package model

type ProductResponse struct {
	SKU       string            `json:"sku,omitempty"`
	Name      string            `json:"name,omitempty"`
	CreatedAt int64             `json:"created_at,omitempty"`
	UpdatedAt int64             `json:"updated_at,omitempty"`
	Category  *CategoryResponse `json:"category,omitempty"`
	Sizes     []SizeResponse    `json:"sizes,omitempty"`
}

type CreateProductRequest struct {
	CategoryID uint   `json:"category_id" validate:"required,max=100"`
	SKU        string `json:"sku" validate:"required,max=100"`
	Name       string `json:"name" validate:"required,max=100"`
}

type SearchProductRequest struct {
	BranchID uint   `json:"-" validate:"required"`
	Search   string `json:"search"`
	// SKU      string `json:"sku" validate:"max=100"`
	// Name     string `json:"name" validate:"max=100"`
	Page int `json:"page" validate:"min=1"`
	Size int `json:"size" validate:"min=1,max=100"`
}

type GetProductRequest struct {
	SKU string `json:"-" validate:"required,max=100"`
}

type UpdateProductRequest struct {
	CategoryID uint   `json:"category_id,omitempty" validate:"max=100"`
	SKU        string `json:"-" validate:"required,max=100"`
	Name       string `json:"name,omitempty" validate:"max=100"`
}

type DeleteProductRequest struct {
	SKU string `json:"-" validate:"required,max=100"`
}
