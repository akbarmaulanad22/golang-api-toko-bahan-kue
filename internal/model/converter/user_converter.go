package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func UserToResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{
		Username:  user.Username,
		Name:      user.Name,
		Address:   user.Address,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Role:      *RoleToResponse(&user.Role),
		Branch:    *BranchToResponse(&user.Branch),
	}
}

func UserToTokenResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{
		Token: user.Token,
	}
}
