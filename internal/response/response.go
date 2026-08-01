package response

import (
	"net/http"

	"forum/internal/apperror"
	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Code: 200, Msg: "success", Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Code: 200, Msg: "success", Data: data})
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, Envelope{Code: 200, Msg: "success", Data: nil})
}

func Error(c *gin.Context, err error) {
	appErr := apperror.As(err)
	if appErr.HTTPStatus >= http.StatusInternalServerError {
		_ = c.Error(err)
	}
	c.JSON(appErr.HTTPStatus, Envelope{Code: appErr.Code, Msg: appErr.Message, Data: nil})
}

func Abort(c *gin.Context, err error) {
	Error(c, err)
	c.Abort()
}
