package model

type DistributorResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type CreateDistributorRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
}

type SearchDistributorRequest struct {
	Search string `json:"search" validate:"max=100"`
	Page   int    `json:"page" validate:"min=1"`
	Size   int    `json:"size" validate:"min=1,max=100"`
}

type GetDistributorRequest struct {
	ID uint `json:"id" validate:"required,max=100"`
}

type UpdateDistributorRequest struct {
	ID      uint   `json:"-" id:"required"`
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
}

type DeleteDistributorRequest struct {
	ID uint `json:"-" validate:"required"`
}
