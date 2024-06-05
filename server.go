package itpg

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

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

// DatabaseBackend is the type of database backend to use.
type DatabaseBackend string

// Enum for database backend
const (
	sqliteBackend   DatabaseBackend = "sqlite"
	postgresBackend DatabaseBackend = "postgres"
)

// LogLevel is the log level to use.
type LogLevel string

// logLevelMap is the map of log levels.
var logLevelMap = map[string]zerolog.Level{
	"disabled": zerolog.Disabled,
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
}

// dataDb represents a database connection,
// storing professor names, course codes and names,
// and professor scores.
var dataDb db.DB

// userState stores the state of all users.
var userState pinterface.IUserState

// passwordResetURL is the URL of the password reset web page.
// An example URL would be: https://demo.itpg.cc/changepass.
// The backend server will then append the following to the previous URL:
// ?code=foobarbaz, and send it to the user's email.
// Then, the website should get the email and new password of the user,
// and make the following example POST request to the api server:
// curl https://api.itpg.cc/resetpass -d '{"code": "foobarbaz", "email": "foo@bar.com", "password": "fizzbuzz"}'
var passwordResetWebsiteURL string

// cookieTimeout represents the duration after which a session cookie expires.
var cookieTimeout time.Duration

// logger is the logger used by the server.
var logger = log.Logger

// RunConfig defines the server's confiuration.
type RunConfig struct {
	Port                    string          // Port on which the server will run.
	DbURL                   string          // Path to the SQLite database file.
	DbBackend               DatabaseBackend // Database backend type.
	LogLevel                LogLevel        // Log level.
	UsersDBPath             string          // Path to the users BOLT database file.
	SMTPEnvPath             string          // Path to the .env file containing SMTP cfguration.
	PasswordResetWebsiteURL string          // URL to the password reset website page.
	AllowedOrigins          []string        // List of allowed origins for CORS.
	AllowedMailDomains      []string        // List of allowed mail domains for registering with the service.
	UseSMTP                 bool            // Whether to use SMTP (false for SMTPS).
	UseHTTP                 bool            // Whether to use HTTP (false for HTTPS).
	CertFilePath            string          // Path to the certificate file (required for HTTPS).
	KeyFilePath             string          // Path to the key file (required for HTTPS).
	CookieTimeout           int             // Duration in minute after which a session cookie expires.
	CodeValidityMinute      int             // Duration in minute after which a code is invalid.
	CodeLength              int             // Length of generated codes.
	MinPasswordScore        int             // Minimum acceptable score of a password scores computed by zxcvbn.
	HandlerCfg              string          // Handler config json file.
}

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(cfg *RunConfig) (err error) {
	if err = validAllowedDomains(cfg.AllowedMailDomains); err != nil {
		return err
	}
	allowedMailDomains = cfg.AllowedMailDomains

	if err = initCredsSmtp(cfg.SMTPEnvPath, !cfg.UseSMTP); err != nil {
		return err
	}

	logLevel, ok := logLevelMap[string(cfg.LogLevel)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}
	zerolog.SetGlobalLevel(logLevel)

	switch cfg.DbBackend {
	case sqliteBackend:
		dataDb, err = sqlite.New(cfg.DbURL, context.Background())
	case postgresBackend:
		dataDb, err = postgres.New(cfg.DbURL, context.Background())
	default:
		return fmt.Errorf("invalid database backend: %s", cfg.DbBackend)
	}

	if err != nil {
		return err
	}

	defer dataDb.Close()

	perm, err := permissionbolt.NewWithConf(cfg.UsersDBPath)
	if err != nil {
		return err
	}
	perm.SetDenyFunction(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrPermissionDenied.WriteJSON(w)
	})

	if cfg.CodeLength > 32 || cfg.CodeLength < 8 {
		return fmt.Errorf("invalid code length: %d (should be between 8 and 32)", cfg.CodeLength)
	}
	codeLength = cfg.CodeLength

	if minPasswordScore < 0 || minPasswordScore > 4 {
		return fmt.Errorf("invalid min password score: %d (should be between 0 and 4)", minPasswordScore)
	}
	minPasswordScore = cfg.MinPasswordScore

	if cfg.CodeValidityMinute <= 0 {
		return fmt.Errorf("invalid code validity: %d (should be greater than 0)", cfg.CodeValidityMinute)
	}
	confirmationCodeValidityTime = time.Minute * time.Duration(cfg.CodeValidityMinute)

	handlerCfg, err := os.ReadFile(cfg.HandlerCfg)
	if err != nil {
		return err
	}

	handlers, err := parseHandlers(bytes.NewReader(handlerCfg))
	if err != nil {
		return err
	}

	userState = perm.UserState()

	cookieTimeout = time.Minute * time.Duration(cfg.CookieTimeout)

	userState.SetCookieTimeout(int64(cookieTimeout.Seconds()))

	router := mux.NewRouter()
	for _, h := range handlers {
		switch h.PathType {
		case adminPath:
			router.Handle(h.Path, h.limiter(checkCookieExpiryMiddleware(h.Handler))).Methods(h.Method)
			perm.AddAdminPath(h.Path)
		case userPath:
			router.Handle(h.Path, h.limiter(checkCookieExpiryMiddleware(checkConfirmedMiddleware(h.Handler)))).Methods(h.Method)
			perm.AddUserPath(h.Path)
		case publicPath:
			router.Handle(h.Path, h.limiter(DummyMiddleware(h.Handler))).Methods(h.Method)
			perm.AddPublicPath(h.Path)
		default:
			return fmt.Errorf("invalid path type: %d", h.PathType)
		}
	}

	passwordResetWebsiteURL = cfg.PasswordResetWebsiteURL

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	})

	n := negroni.Classic()

	n.Use(c)
	n.Use(perm)
	n.UseHandler(router)

	s := fmt.Sprintf("itpg-backend (%s) listening on port %s", cfg.DbBackend, cfg.Port)
	if !cfg.UseSMTP {
		s += " with SMTPS,"
	} else {
		s += " with SMTP,"
	}

	if !cfg.UseHTTP {
		log.Info().Msgf("%s with HTTPS\n", s)
		return http.ListenAndServeTLS(":"+cfg.Port, cfg.CertFilePath, cfg.KeyFilePath, n)
	} else {
		log.Info().Msgf("%s with HTTP\n", s)
		return http.ListenAndServe(":"+cfg.Port, n)
	}
}
