package model

type UserResponse struct {
	Username  string          `json:"username,omitempty"`
	Token     string          `json:"token,omitempty"`
	Name      string          `json:"name,omitempty"`
	Address   string          `json:"address,omitempty"`
	CreatedAt int64           `json:"created_at,omitempty"`
	Role      *RoleResponse   `json:"role,omitempty"`
	Branch    *BranchResponse `json:"branch,omitempty"`
}

type VerifyUserRequest struct {
	Token string `validate:"required,max=100"`
}

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
	BranchID *uint  `json:"branch_id" validate:"required"`
}

type LoginUserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LogoutUserRequest struct {
	Username string `json:"username" validate:"required"`
}

type SearchUserRequest struct {
	Search   string `json:"search"`
	RoleID   uint   `json:"role_id"`
	BranchID uint   `json:"branch_id"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}

type GetUserRequest struct {
	Username string `json:"username" validate:"required"`
}

type UpdateUserRequest struct {
	Username string `json:"-" validate:"required"`
	Password string `json:"password"`
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
	BranchID *uint  `json:"branch_id" validate:"required"`
}

type DeleteUserRequest struct {
	Username string `json:"-" validate:"required"`
}
