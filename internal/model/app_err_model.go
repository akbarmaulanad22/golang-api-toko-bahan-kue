package model

type AppError struct {
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppErr(message string, details any) *AppError {
	return &AppError{Message: message, Details: details}
}
