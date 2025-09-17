package model

type BranchResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type CreateBranchRequest struct {
	Name    string `json:"name" validate:"required,max=100"`
	Address string `json:"address" validate:"required,max=100"`
}

type SearchBranchRequest struct {
	Search string `json:"search" validate:"max=100"`
	Page   int    `json:"page" validate:"min=1"`
	Size   int    `json:"size" validate:"min=1,max=100"`
}

type GetBranchRequest struct {
	ID uint `json:"id" validate:"required,max=100"`
}

type UpdateBranchRequest struct {
	ID      uint   `json:"-" validate:"required,max=100"`
	Name    string `json:"name,omitempty" validate:"max=100"`
	Address string `json:"address,omitempty" validate:"max=100"`
}

type DeleteBranchRequest struct {
	ID uint `json:"-" validate:"required,max=100"`
}
