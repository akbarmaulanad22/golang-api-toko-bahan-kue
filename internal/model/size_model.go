package model

type SizeResponse struct {
	ID        uint    `json:"id,omitempty"`
	Name      string  `json:"name,omitempty"`
	SellPrice float64 `json:"sell_price,omitempty"`
	BuyPrice  float64 `json:"buy_price,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
	UpdatedAt int64   `json:"updated_at,omitempty"`
}

type CreateSizeRequest struct {
	ProductSKU string  `json:"-" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	SellPrice  float64 `json:"sell_price" validate:"required"`
	BuyPrice   float64 `json:"buy_price" validate:"required"`
}

type SearchSizeRequest struct {
	ProductSKU string `json:"-" validate:"required,max=100"`
	Name       string `json:"name" validate:"max=100"`
	Page       int    `json:"page" validate:"min=1"`
	Size       int    `json:"size" validate:"min=1,max=100"`
}

type GetSizeRequest struct {
	ID         uint   `json:"-" validate:"required"`
	ProductSKU string `json:"-" validate:"required"`
}

type UpdateSizeRequest struct {
	ProductSKU string  `json:"-" validate:"required"`
	ID         uint    `json:"-" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	SellPrice  float64 `json:"sell_price" validate:"required"`
	BuyPrice   float64 `json:"buy_price" validate:"required"`
}

type DeleteSizeRequest struct {
	ProductSKU string `json:"-" validate:"required"`
	ID         uint   `json:"-" validate:"required"`
}

type SizeWithProduct struct {
	ID          uint
	Size        string
	SellPrice   float64
	ProductName string
}
