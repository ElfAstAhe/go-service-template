package transport

type ErrorDTO struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewErrorDTO(status int, message string) *ErrorDTO {
	return &ErrorDTO{
		Code:    status,
		Message: message,
	}
}

func NewErrorDTOFromError(code int, err error) *ErrorDTO {
	return NewErrorDTO(code, err.Error())
}
