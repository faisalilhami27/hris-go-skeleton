package constants

import "errors"

const (
	Success = "success"
	Error   = "error"
)

var (
	ErrInternal                = errors.New("internal server error please try again later")
	ErrSQLError                = errors.New("database server failed to execute, please try again")
	ErrOrderDate               = errors.New("order date must be greater than now")
	ErrStatus                  = errors.New("invalid status")
	ErrTooManyRequest          = errors.New("too many request, please try again later")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInvalidToken            = errors.New("invalid token")
	ErrForbidden               = errors.New("you don't have access for this resource")
	ErrInvalidUploadedFile     = errors.New(`invalid uploaded file`)
	ErrSizeTooBig              = errors.New(`file size is too big`)
	ErrInvalidUploadedFileType = errors.New(`extension is not allowed`)
)

var GeneralErrors = []error{
	ErrInternal,
	ErrSQLError,
	ErrOrderDate,
	ErrStatus,
	ErrTooManyRequest,
	ErrUnauthorized,
	ErrForbidden,
}
