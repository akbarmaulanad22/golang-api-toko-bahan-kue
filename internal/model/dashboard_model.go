package model

type DashboardResponse struct {
	TotalEmployees    int `json:"total_employees"`
	TotalProducts     int `json:"total_products"`
	TotalDistributors int `json:"total_distributors"`
	TotalBranches     int `json:"total_branches"`
}
