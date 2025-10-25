package model

type RoleResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

type CreateRoleRequest struct {
	Name string `json:"name" validate:"required"`
}

type SearchRoleRequest struct {
	Name string `json:"name" validate:"max=100"`
	Page int    `json:"page" validate:"min=1"`
	Size int    `json:"size" validate:"min=1,max=100"`
}

type GetRoleRequest struct {
	ID uint `json:"id" validate:"required"`
}

type UpdateRoleRequest struct {
	ID   uint   `json:"-" validate:"required"`
	Name string `json:"name,omitempty" validate:"required"`
}

type DeleteRoleRequest struct {
	ID uint `json:"-" validate:"required"`
}
