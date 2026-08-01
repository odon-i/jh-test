package apperror

import (
	"errors"
	"net/http"
)

type Error struct {
	HTTPStatus int
	Code       int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func New(status, code int, message string) *Error {
	return &Error{HTTPStatus: status, Code: code, Message: message}
}

func Wrap(status, code int, message string, err error) *Error {
	return &Error{HTTPStatus: status, Code: code, Message: message, Err: err}
}

var (
	ErrBadRequest   = New(http.StatusBadRequest, http.StatusBadRequest, "参数校验失败")
	ErrUnauthorized = New(http.StatusUnauthorized, http.StatusUnauthorized, "未登录或令牌无效")
	ErrForbidden    = New(http.StatusForbidden, http.StatusForbidden, "没有操作权限")
	ErrNotFound     = New(http.StatusNotFound, http.StatusNotFound, "资源不存在")
	ErrConflict     = New(http.StatusConflict, http.StatusConflict, "资源已存在")
	ErrInternal     = New(http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
)

func As(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误", err)
}
