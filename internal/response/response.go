// Package response renders every API response in the standard envelope used
// by the Laravel app:
//
//	{ success, message, data, error_code? (errors only), meta? (pagination), errors? (validation) }
package response

import (
	"math"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Error codes (App\Enums\ErrorCode).
const (
	CodeServerError        = 1000
	CodeBadRequest         = 1001
	CodeTooManyRequests    = 1006
	CodeUnauthenticated    = 1101
	CodeForbidden          = 1102
	CodeValidationFailed   = 1103
	CodeNotFound           = 1105
	CodeMethodNotAllowed   = 1106
	CodeInvalidCredentials = 1301
)

// Messages (App\Enums\ResponseMessage).
const (
	MsgInternalServerError = "Internal Server Error"
	MsgFound               = "Item found"
	MsgNoItem              = "Item not found"
	MsgSuccessCreate       = "Successfully created item"
	MsgFailedCreate        = "Failed to create item"
	MsgSuccessUpdate       = "Successfully updated item"
	MsgFailedUpdate        = "Failed to update item"
	MsgSuccessDelete       = "Successfully deleted item"
	MsgFailedDelete        = "Failed to delete item"
	MsgLoginSuccess        = "Login Successfully"
	MsgUnauthorized        = "Unauthorized"
	MsgLogoutSuccess       = "Logout Successfully"
)

func CodeFromStatus(status int) int {
	switch status {
	case 400:
		return CodeBadRequest
	case 401:
		return CodeUnauthenticated
	case 403:
		return CodeForbidden
	case 404:
		return CodeNotFound
	case 405:
		return CodeMethodNotAllowed
	case 422:
		return CodeValidationFailed
	case 429:
		return CodeTooManyRequests
	default:
		return CodeServerError
	}
}

type successBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type errorBody struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
	Data      any    `json:"data"`
	Errors    any    `json:"errors,omitempty"`
}

type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

// Success writes a success envelope. An empty message falls back to
// "Successfully created item" for 201 and "Item found" otherwise.
func Success(c *gin.Context, status int, message string, data any) {
	write(c, status, message, data, nil)
}

func write(c *gin.Context, status int, message string, data any, meta *Meta) {
	if message == "" {
		if status == 201 {
			message = MsgSuccessCreate
		} else {
			message = MsgFound
		}
	}
	c.JSON(status, successBody{Success: true, Message: message, Data: orEmpty(data), Meta: meta})
}

// Data writes data with the default message.
func Data(c *gin.Context, status int, data any) { Success(c, status, "", data) }

func Paginated(c *gin.Context, data any, page, limit int, total int64) {
	write(c, 200, "", data, &Meta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		LastPage:    int(math.Ceil(float64(total) / float64(limit))),
	})
}

// Error writes an error envelope. code 0 derives the error code from status.
func Error(c *gin.Context, status int, message string, errors any, code int) {
	if message == "" {
		message = MsgInternalServerError
	}
	if code == 0 {
		code = CodeFromStatus(status)
	}
	c.AbortWithStatusJSON(status, errorBody{
		Success: false, Message: message, ErrorCode: code, Data: struct{}{}, Errors: errors,
	})
}

// Message mirrors ResponseHelper::responseMessageRequest.
func Message(c *gin.Context, status int, success bool, message string) {
	if success {
		Success(c, status, message, nil)
	} else {
		Error(c, status, message, nil, 0)
	}
}

// ValidationErrors mirrors ResponseHelper::responseMessageRequestWithErrors.
func ValidationErrors(c *gin.Context, status int, message string, errors any) {
	Error(c, status, message, errors, CodeValidationFailed)
}

func orEmpty(data any) any {
	if data == nil {
		return struct{}{}
	}
	v := reflect.ValueOf(data)
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Map) && v.IsNil() {
		return struct{}{}
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		return []any{}
	}
	return data
}
