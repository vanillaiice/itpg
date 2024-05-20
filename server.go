package itpg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/negroni"
	"github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/db/postgres"
	"github.com/vanillaiice/itpg/db/sqlite"
	"github.com/vanillaiice/itpg/responses"
	"github.com/xyproto/permissionbolt/v2"
	"github.com/xyproto/pinterface"
)

// PathType is the type of the path (admin, user, public).
type PathType int

// Enum for path types
const (
	UserPath   PathType = 0 // UserPath is a path only accessible by users.
	PublicPath PathType = 1 // PublicPath is a path accessible by anyone.
	AdminPath  PathType = 2 // AdminPath is a path accessible by admins.
)

// DatabaseBackend is the type of database backend to use.
type DatabaseBackend string

// Enum for datbase backend
const (
	Sqlite   DatabaseBackend = "sqlite"
	Postgres DatabaseBackend = "postgres"
)

// LogLevel is the log level to use.
type LogLevel string

// Enum for log levels.
const (
	LogLevelDisabled LogLevel = "disabled"
	LogLevelInfo     LogLevel = "info"
	LogLevelError    LogLevel = "error"
	LogLevelFatal    LogLevel = "fatal"
)

// LimitHandlerFunc is executed when the request limit is reached.
var LimitHandlerFunc = httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	responses.ErrRequestLimitReached.WriteJSON(w)
})

// LimiterLenient is a limiter that allows 1000 requests per second per IP.
var LimiterLenient = httprate.Limit(
	1000,
	time.Second,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	LimitHandlerFunc,
)

// LimiterModerate is a limiter that allows 1000 requests per minute per IP.
var LimiterModerate = httprate.Limit(
	1000,
	time.Minute,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	LimitHandlerFunc,
)

// LimiterStrict is a limiter that allows 500 requests per hour per IP.
var LimiterStrict = httprate.Limit(
	500,
	time.Hour,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	LimitHandlerFunc,
)

// LimiterVeryStrict is a limiter that allows 100 requests per hour per IP.
var LimiterVeryStrict = httprate.Limit(
	100,
	1*time.Hour,
	httprate.WithKeyFuncs(httprate.KeyByIP),
	LimitHandlerFunc,
)

// HandlerInfo represents a struct containing information about an HTTP handler.
type HandlerInfo struct {
	Path     string                                   // Path specifies the URL pattern for which the handler is responsible.
	Handler  func(http.ResponseWriter, *http.Request) // Handler is the function that will be called to handle HTTP requests.
	Method   string                                   // Method specifies the HTTP method associated with the handler.
	PathType PathType                                 // PathType is the type of the path (admin, user, public).
	Limiter  func(http.Handler) http.Handler          // Limiter is the limiter used to limit requests.
}

// DataDB represents a database connection,
// storing professor names, course codes and names,
// and professor scores.
var DataDB db.DB

// UserState stores the state of all users.
var UserState pinterface.IUserState

// PasswordResetURL is the URL of the password reset web page.
// An example URL would be: https://demo.itpg.cc/changepass.
// The backend server will then append the following to the previous URL:
// ?code=foobarbaz, and send it to the user's email.
// Then, the website should get the email and new password of the user,
// and make the following example POST request to the api server:
// curl https://api.itpg.cc/resetpass -d '{"code": "foobarbaz", "email": "foo@bar.com", "password": "fizzbuzz"}'
var PasswordResetWebsiteURL string

// CookieTimeout represents the duration after which a session cookie expires.
var CookieTimeout time.Duration

// Logger is the logger used by the server.
var Logger = log.Logger

// RunConfig defines the server's configuration settings.
type RunConfig struct {
	Port                    string          // Port on which the server will run.
	DbURL                   string          // Path to the SQLite database file.
	DbBackend               DatabaseBackend // Database backend type.
	LogLevel                LogLevel        // Log level.
	UsersDBPath             string          // Path to the users BOLT database file.
	SMTPEnvPath             string          // Path to the .env file containing SMTP configuration.
	PasswordResetWebsiteURL string          // URL to the password reset website page.
	AllowedOrigins          []string        // List of allowed origins for CORS.
	AllowedMailDomains      []string        // List of allowed mail domains for registering with the service.
	UseSMTP                 bool            // Whether to use SMTP (false for SMTPS).
	UseHTTP                 bool            // Whether to use HTTP (false for HTTPS).
	CertFilePath            string          // Path to the certificate file (required for HTTPS).
	KeyFilePath             string          // Path to the key file (required for HTTPS).
	CookieTimeout           int             // Duration in minute after which a session cookie expires.
}

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(config *RunConfig) (err error) {
	if err = validAllowedDomains(config.AllowedMailDomains); err != nil {
		return err
	}
	AllowedMailDomains = config.AllowedMailDomains

	if err = InitCredsSMTP(config.SMTPEnvPath, !config.UseSMTP); err != nil {
		return err
	}

	switch config.LogLevel {
	case LogLevelDisabled:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	case LogLevelInfo:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case LogLevelError:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case LogLevelFatal:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		return fmt.Errorf("invalid log level: %s", string(config.LogLevel))
	}

	switch config.DbBackend {
	case Sqlite:
		DataDB, err = sqlite.New(config.DbURL)
	case Postgres:
		DataDB, err = postgres.New(config.DbURL)
	default:
		return fmt.Errorf("invalid database backend: %s", string(config.DbBackend))
	}

	if err != nil {
		return err
	}

	defer DataDB.Close()

	perm, err := permissionbolt.NewWithConf(config.UsersDBPath)
	if err != nil {
		return err
	}
	perm.SetDenyFunction(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrPermissionDenied.WriteJSON(w)
	})

	CookieTimeout = time.Minute * time.Duration(config.CookieTimeout)
	UserState = perm.UserState()
	UserState.SetCookieTimeout(int64(CookieTimeout.Seconds()))

	PasswordResetWebsiteURL = config.PasswordResetWebsiteURL

	c := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	})

	handlers := []*HandlerInfo{
		// User
		{"/course/grade", GradeCourseProfessor, http.MethodPost, UserPath, LimiterModerate},
		{"/refresh", RefreshCookie, http.MethodPost, UserPath, LimiterLenient},
		{"/logout", Logout, http.MethodPost, UserPath, LimiterLenient},
		{"/clear", ClearCookie, http.MethodPost, UserPath, LimiterLenient},
		{"/changepass", ChangePassword, http.MethodPost, UserPath, LimiterStrict},
		{"/delete", DeleteAccount, http.MethodPost, UserPath, LimiterVeryStrict},
		{"/ping", Ping, http.MethodGet, UserPath, LimiterLenient},
		// Public
		{"/course/all", GetLastCourses, http.MethodGet, PublicPath, LimiterLenient},
		{"/professor/all", GetLastProfessors, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/all", GetLastScores, http.MethodGet, PublicPath, LimiterLenient},
		{"/course/{uuid}", GetCoursesByProfessorUUID, http.MethodGet, PublicPath, LimiterLenient},
		{"/professor/{code}", GetProfessorsByCourseCode, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/prof/{uuid}", GetScoresByProfessorUUID, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/profname/{name}", GetScoresByProfessorName, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/profnamelike/{name}", GetScoresByProfessorNameLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/coursename/{name}", GetScoresByCourseName, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/coursenamelike/{name}", GetScoresByCourseNameLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/coursecode/{code}", GetScoresByCourseCode, http.MethodGet, PublicPath, LimiterLenient},
		{"/score/coursecodelike/{code}", GetScoresByCourseCodeLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/login", Login, http.MethodPost, PublicPath, LimiterLenient},
		{"/register", Register, http.MethodPost, PublicPath, LimiterModerate},
		{"/confirm", Confirm, http.MethodPost, PublicPath, LimiterModerate},
		{"/newconfirmationcode", SendNewConfirmationCode, http.MethodPost, PublicPath, LimiterStrict},
		{"/sendresetlink", SendResetLink, http.MethodPost, PublicPath, LimiterVeryStrict},
		{"/resetpass", ResetPassword, http.MethodPost, PublicPath, LimiterVeryStrict},
		// Admin
		{"course/add", AddCourse, http.MethodPost, AdminPath, LimiterLenient},
		{"course/remove", RemoveCourse, http.MethodDelete, AdminPath, LimiterLenient},
		{"course/removeforce", RemoveCourseForce, http.MethodDelete, AdminPath, LimiterLenient},
		{"course/addprof", AddCourseProfessor, http.MethodPost, AdminPath, LimiterLenient},
		{"professor/add", AddProfessor, http.MethodDelete, AdminPath, LimiterLenient},
		{"professor/remove", RemoveProfessor, http.MethodDelete, AdminPath, LimiterLenient},
		{"professor/removeforce", RemoveProfessorForce, http.MethodDelete, AdminPath, LimiterLenient},
	}

	router := mux.NewRouter()
	for _, h := range handlers {
		switch h.PathType {
		case UserPath:
			router.Handle(h.Path, h.Limiter(checkCookieExpiryMiddleware(checkConfirmedMiddleware(h.Handler)))).Methods(h.Method)
			perm.AddUserPath(h.Path)
		case PublicPath:
			router.Handle(h.Path, h.Limiter(DummyMiddleware(h.Handler))).Methods(h.Method)
			perm.AddPublicPath(h.Path)
		}
	}

	n := negroni.Classic()
	n.Use(c)
	n.Use(perm)
	n.UseHandler(router)

	s := fmt.Sprintf("itpg-backend (%s) listening on port %s", config.DbBackend, config.Port)
	if !config.UseSMTP {
		s += " with SMTPS,"
	} else {
		s += " with SMTP,"
	}

	if !config.UseHTTP {
		log.Info().Msgf("%s with HTTPS\n", s)
		return http.ListenAndServeTLS(":"+config.Port, config.CertFilePath, config.KeyFilePath, n)
	} else {
		log.Info().Msgf("%s with HTTP\n", s)
		return http.ListenAndServe(":"+config.Port, n)
	}
}
