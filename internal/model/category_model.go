package model

type CategoryResponse struct {
	Slug      string `json:"slug,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type SearchCategoryRequest struct {
	Name string `json:"name" validate:"max=100"`
	Page int    `json:"page" validate:"min=1"`
	Size int    `json:"size" validate:"min=1,max=100"`
}

type GetCategoryRequest struct {
	Slug string `json:"slug" validate:"required,max=100"`
}

type UpdateCategoryRequest struct {
	Slug string `json:"-" validate:"required,max=100"`
	Name string `json:"name,omitempty" validate:"max=100"`
}

type DeleteCategoryRequest struct {
	Slug string `json:"-" validate:"required,max=100"`
}
