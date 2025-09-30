package errors

type HTTPError struct {
	HTTPStatus int    `json:"httpStatus"`
	ErrorKey   string `json:"errorKey"`
}

func (httpError HTTPError) Error() string {
	return httpError.ErrorKey
}

func NewHTTPError(key string, statuscode int) *HTTPError {
	return &HTTPError{
		HTTPStatus: statuscode,
		ErrorKey:   key,
	}
}
