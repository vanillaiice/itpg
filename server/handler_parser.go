package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/vanillaiice/itpg/responses"
)

// Handler holds data for a handler.
type Handler struct {
	Handlers []struct {
		Path     string `json:"path"`
		PathType string `json:"pathType"`
		Handler  string `json:"handler"`
		Limiter  string `json:"limiter"`
		Method   string `json:"method"`
	} `json:"handlers"`
}

// HandlerInfo represents a struct containing information about an HTTP handler.
type HandlerInfo struct {
	path     string                                   // Path specifies the URL pattern for which the handler is responsible.
	handler  func(http.ResponseWriter, *http.Request) // Handler is the function that will be called to handle HTTP requests.
	method   string                                   // Method specifies the HTTP method associated with the handler.
	pathType PathType                                 // PathType is the type of the path (admin, user, public).
	limiter  func(http.Handler) http.Handler          // Limiter is the limiter used to limit requests.
}

// PathType is the type of the path (admin, user, public).
type PathType int

// Enum for path types
const (
	userPath   PathType = 0 // UserPath is a path only accessible by users.
	publicPath PathType = 1 // publicPath is a path accessible by anyone.
	adminPath  PathType = 2 // adminPath is a path only accessible by admins.
)

// limitHandlerFunc is executed when the request limit is reached.
var limitHandlerFunc = httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	responses.ErrRequestLimitReached.WriteJSON(w)
})

// limiterLenient is a limiter that allows 1000 requests per second per IP.
var limiterLenient = httprate.Limit(
	1000,
	time.Second,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	limitHandlerFunc,
)

// limiterModerate is a limiter that allows 1000 requests per minute per IP.
var limiterModerate = httprate.Limit(
	1000,
	time.Minute,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	limitHandlerFunc,
)

// limiterStrict is a limiter that allows 500 requests per hour per IP.
var limiterStrict = httprate.Limit(
	500,
	time.Hour,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	limitHandlerFunc,
)

// limiterVeryStrict is a limiter that allows 100 requests per hour per IP.
var limiterVeryStrict = httprate.Limit(
	100,
	1*time.Hour,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	limitHandlerFunc,
)

// limiterMap is a map of limiter functions to their names.
var limiterMap = map[string]func(http.Handler) http.Handler{
	"lenient":    limiterLenient,
	"moderate":   limiterModerate,
	"strict":     limiterStrict,
	"veryStrict": limiterVeryStrict,
}

// pathTypeMap is a map of path types to their names.
var pathTypeMap = map[string]PathType{
	"admin":  adminPath,
	"user":   userPath,
	"public": publicPath,
}

// methodMap is a map of HTTP methods to their names.
var methodMap = map[string]string{
	"GET":    http.MethodGet,
	"POST":   http.MethodPost,
	"PUT":    http.MethodPut,
	"DELETE": http.MethodDelete,
}

// handlerFuncMap is a map of handler functions to their names.
var handlerFuncMap = map[string]func(http.ResponseWriter, *http.Request){
	"gradeCourseProfessor":         gradeCourseProfessor,
	"refreshCookie":                refreshCookie,
	"logout":                       logout,
	"clearCookie":                  clearCookie,
	"changePassword":               changePassword,
	"deleteAccount":                deleteAccount,
	"ping":                         ping,
	"getLastCourses":               getLastCourses,
	"getLastProfessors":            getLastProfessors,
	"getLastScores":                getLastScores,
	"getCoursesByProfessorUUID":    getCoursesByProfessorUUID,
	"getProfessorsByCourseCode":    getProfessorsByCourseCode,
	"getScoresByProfessorUUID":     getScoresByProfessorUUID,
	"getScoresByProfessorName":     getScoresByProfessorName,
	"getScoresByProfessorNameLike": getScoresByProfessorNameLike,
	"getScoresByCourseName":        getScoresByCourseName,
	"getScoresByCourseNameLike":    getScoresByCourseNameLike,
	"getScoresByCourseCode":        getScoresByCourseCode,
	"getScoresByCourseCodeLike":    getScoresByCourseCodeLike,
	"login":                        login,
	"register":                     register,
	"confirm":                      confirm,
	"sendNewConfirmationCode":      sendNewConfirmationCode,
	"sendResetLink":                sendResetLink,
	"resetPassword":                resetPassword,
	"addCourse":                    addCourse,
	"removeCourse":                 removeCourse,
	"removeCourseForce":            removeCourseForce,
	"addCourseProfessor":           addCourseProfessor,
	"addProfessor":                 addProfessor,
	"removeProfessor":              removeProfessor,
	"removeProfessorForce":         removeProfessorForce,
}

// parseHandlers parses a handlers.json file and returns a slice of HandlerInfo.
func parseHandlers(reader *bytes.Reader) ([]*HandlerInfo, error) {
	var handlers Handler
	var handlersInfo []*HandlerInfo

	if err := json.NewDecoder(reader).Decode(&handlers); err != nil {
		return nil, err
	}

	for _, h := range handlers.Handlers {
		handlerFunc, ok := handlerFuncMap[h.Handler]
		if !ok {
			return nil, fmt.Errorf("handler %s not found", h.Handler)
		}

		method, ok := methodMap[h.Method]
		if !ok {
			return nil, fmt.Errorf("method %s not found", h.Method)
		}

		pathType, ok := pathTypeMap[h.PathType]
		if !ok {
			return nil, fmt.Errorf("path type %s not found", h.PathType)
		}

		limiter, ok := limiterMap[h.Limiter]
		if !ok {
			return nil, fmt.Errorf("limiter %s not found", h.Limiter)
		}

		handlersInfo = append(handlersInfo, &HandlerInfo{
			path:     h.Path,
			handler:  handlerFunc,
			method:   method,
			pathType: pathType,
			limiter:  limiter,
		})
	}

	return handlersInfo, nil
}
