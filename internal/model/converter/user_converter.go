package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func UserToResponse(user *entity.User) *model.UserResponse {
	branch := new(model.BranchResponse)
	if user.Branch != nil {
		branch = BranchToResponse(user.Branch)
	}

	return &model.UserResponse{
		Username:  user.Username,
		Name:      user.Name,
		Address:   user.Address,
		CreatedAt: user.CreatedAt,
		Role:      RoleToResponse(&user.Role),
		Branch:    branch,
	}
}

func UserToTokenResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{
		Token: user.Token,
	}
}
