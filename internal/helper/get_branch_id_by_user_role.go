package helper

// import (
// 	"net/url"
// 	"strconv"
// 	"strings"
// 	"tokobahankue/internal/model"
// )

// func GetBranchID(auth *model.Auth, params url.Values, request *model.SearchSalesReportRequest) {
// 	if strings.ToUpper(auth.Role) == "OWNER" {
// 		branchID := params.Get("branch_id")
// 		if branchID == "" {
// 			branchID = "0"
// 		}
// 		branchIDInt, _ := strconv.Atoi(branchID)
// 		branchIDUint := uint(branchIDInt)

// 		request.BranchID = &branchIDUint
// 	} else {
// 		request.BranchID = auth.BranchID
// 	}
// }
