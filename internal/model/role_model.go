package model

type RoleResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type CreateRoleRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type SearchRoleRequest struct {
	Name string `json:"name" validate:"max=100"`
	Page int    `json:"page" validate:"min=1"`
	Size int    `json:"size" validate:"min=1,max=100"`
}

type GetRoleRequest struct {
	ID uint `json:"id" validate:"required,max=100"`
}

type UpdateRoleRequest struct {
	ID   uint   `json:"-" validate:"required,max=100"`
	Name string `json:"name,omitempty" validate:"max=100"`
}

type DeleteRoleRequest struct {
	ID uint `json:"-" validate:"required,max=100"`
}
