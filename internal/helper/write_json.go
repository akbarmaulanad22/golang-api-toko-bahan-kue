package helper

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Supaya gak error kalau data nil
	if data == nil {
		return nil
	}

	// Encode data ke JSON
	return json.NewEncoder(w).Encode(data)
	// if err := json.NewEncoder(w).Encode(data); err != nil {
	// 	http.Error(w, `{"error":{"message":"failed to encode response"}}`, http.StatusInternalServerError)
	// }
}
