package model

type SizeResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	SellPrice uint   `json:"sell_price,omitempty"`
	BuyPrice  uint   `json:"buy_price,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	// Product   ProductResponse `json:"product"`
}

type CreateSizeRequest struct {
	ProductSKU string `json:"-" validate:"required,max=100"`
	Name       string `json:"name" validate:"required,max=100"`
	SellPrice  uint   `json:"sell_price,omitempty" validate:"required"`
	BuyPrice   uint   `json:"buy_price,omitempty" validate:"required"`
}

type SearchSizeRequest struct {
	ProductSKU string `json:"-" validate:"required,max=100"`
	Name       string `json:"name" validate:"max=100"`
	Page       int    `json:"page" validate:"min=1"`
	Size       int    `json:"size" validate:"min=1,max=100"`
}

type GetSizeRequest struct {
	ID         uint   `json:"-" id:"required,max=100"`
	ProductSKU string `json:"-" validate:"required,max=100"`
}

type UpdateSizeRequest struct {
	ProductSKU string `json:"-" validate:"required,max=100"`
	ID         uint   `json:"-" id:"required,max=100"`
	Name       string `json:"name,omitempty" validate:"max=100"`
	SellPrice  uint   `json:"sell_price,omitempty"`
	BuyPrice   uint   `json:"buy_price,omitempty"`
}

type DeleteSizeRequest struct {
	ProductSKU string `json:"-" validate:"required,max=100"`
	ID         uint   `json:"-" id:"required,max=100"`
}
