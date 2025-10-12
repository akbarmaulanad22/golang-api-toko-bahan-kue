package helper

import "strconv"

func ParseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}

	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	return def
}
