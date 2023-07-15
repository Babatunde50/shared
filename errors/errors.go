package errors

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Babatunde50/shared/leveledlog"
	"github.com/Babatunde50/shared/response"
	"github.com/Babatunde50/shared/validator"
)

func ReportError(err error) {
	trace := debug.Stack()

	logger := leveledlog.NewLogger(os.Stdout, leveledlog.LevelAll, true)

	logger.Error(err, trace)

}

func errorMessage(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	message = strings.ToUpper(message[:1]) + message[1:]

	err := response.JSONWithHeaders(w, status, map[string]string{"Error": message}, headers)
	if err != nil {
		ReportError(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	ReportError(err)

	message := "The server encountered a problem and could not process your request"
	errorMessage(w, r, http.StatusInternalServerError, message, nil)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	errorMessage(w, r, http.StatusNotFound, message, nil)
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	errorMessage(w, r, http.StatusMethodNotAllowed, message, nil)
}

func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	errorMessage(w, r, http.StatusBadRequest, err.Error(), nil)
}

func FailedValidation(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	err := response.JSON(w, http.StatusBadRequest, v)
	if err != nil {
		ServerError(w, r, err)
	}
}

func InvalidCredentials(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	message := "Invalid email or password"
	errorMessage(w, r, http.StatusBadRequest, message, nil)
}

func EmailInUse(w http.ResponseWriter, r *http.Request) {
	message := "The email is already in use"
	errorMessage(w, r, http.StatusConflict, message, nil)
}

func Unauthenticated(w http.ResponseWriter, r *http.Request) {
	message := "You need to be logged in to access this resource"
	errorMessage(w, r, http.StatusUnauthorized, message, nil)
}
