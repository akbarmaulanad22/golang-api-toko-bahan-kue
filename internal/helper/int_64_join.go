package helper

import (
	"fmt"
	"strings"
)

func Int64Join(arr []int64, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	strs := make([]string, len(arr))
	for i, v := range arr {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, sep)
}
