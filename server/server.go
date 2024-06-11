package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/negroni"
	"github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/db/postgres"
	"github.com/vanillaiice/itpg/db/sqlite"
	"github.com/vanillaiice/itpg/mail"
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
	pgBackend       DatabaseBackend = "pg"
)

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

// mailer is the client used to send mail.
var mailer *mail.SmtpClient

// dataDb represents a database connection,
// storing professor names, course codes and names,
// and professor scores.
var dataDb db.DB

// userState stores the state of all users.
var userState pinterface.IUserState

// passwordResetUrl is the URL of the password reset web page.
// An example URL would be: https://demo.itpg.cc/changepass.
// The backend server will then append the following to the previous URL:
// ?code=foobarbaz, and send it to the user's email.
// Then, the website should get the email and new password of the user,
// and make the following example POST request to the api server:
// curl https://api.itpg.cc/resetpass -d '{"code": "foobarbaz", "email": "foo@bar.com", "password": "fizzbuzz"}'
var passwordResetUrl string

// cookieTimeout represents the duration after which a session cookie expires.
var cookieTimeout time.Duration

// RunCfg defines the server's configuration.
type RunCfg struct {
	Port               string          // Port on which the server will run.
	DbUrl              string          // Path to the SQLite database file.
	DbBackend          DatabaseBackend // Database backend type.
	CacheDbUrl         string          // URL to the redis cache database.
	CacheTtl           int             // Time-to-live of the cache in seconds.
	UsersDbPath        string          // Path to the users BOLT database file.
	AllowedOrigins     []string        // List of allowed origins for CORS.
	AllowedMailDomains []string        // List of allowed mail domains for registering with the service.
	PasswordResetUrl   string          // URL to the password reset website page.
	SmtpEnvPath        string          // Path to the .env file containing SMTP cfguration.
	UseSmtp            bool            // Whether to use SMTP (false for SMTPS).
	UseHttp            bool            // Whether to use HTTP (false for HTTPS).
	HandlersFilePath   string          // Handler config json file.
	CertFilePath       string          // Path to the certificate file (required for HTTPS).
	KeyFilePath        string          // Path to the key file (required for HTTPS).
	CookieTimeout      int             // Duration in minute after which a session cookie expires.
	CodeValidityMinute int             // Duration in minute after which a code is invalid.
	CodeLength         int             // Length of generated codes.
	MinPasswordScore   int             // Minimum acceptable score of a password scores computed by zxcvbn.
	LogLevel           LogLevel        // Log level.
}

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(cfg *RunCfg) (err error) {
	if err = validAllowedDomains(cfg.AllowedMailDomains); err != nil {
		return
	}
	allowedMailDomains = cfg.AllowedMailDomains

	mailer, err = mail.NewClient(cfg.SmtpEnvPath, !cfg.UseSmtp)
	if err != nil {
		return
	}

	logLevel, ok := logLevelMap[string(cfg.LogLevel)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}
	zerolog.SetGlobalLevel(logLevel)

	ctx := context.Background()

	cacheTtl := time.Duration(cfg.CacheTtl) * time.Second

	switch cfg.DbBackend {
	case sqliteBackend:
		dataDb, err = sqlite.New(cfg.DbUrl, cfg.CacheDbUrl, cacheTtl, ctx)
	case postgresBackend, pgBackend:
		dataDb, err = postgres.New(cfg.DbUrl, cfg.CacheDbUrl, cacheTtl, ctx)
	default:
		return fmt.Errorf("invalid database backend: %s", cfg.DbBackend)
	}

	if err != nil {
		return
	}

	defer dataDb.Close()

	var initUsersDbAdmin bool
	if _, err := os.Stat(cfg.UsersDbPath); errors.Is(err, os.ErrNotExist) {
		initUsersDbAdmin = true
	}

	perm, err := permissionbolt.NewWithConf(cfg.UsersDbPath)
	if err != nil {
		return
	}

	perm.SetDenyFunction(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrPermissionDenied.WriteJSON(w)
	})

	userState = perm.UserState()

	if initUsersDbAdmin {
		log.Info().Msgf("Initializing users database %s", cfg.UsersDbPath)

		if err = godotenv.Load(); err != nil {
			return
		}

		var adminUsername, adminPassword, adminEmail string

		if os.Getenv("ADMIN_USERNAME") != "" {
			log.Debug().Msg("found environment variable ADMIN_USERNAME")
			adminUsername = os.Getenv("ADMIN_USERNAME")
		} else {
			fmt.Println("enter admin username:")
			if _, err = fmt.Scanln(&adminUsername); err != nil {
				removeUsersDb(cfg.UsersDbPath)
				return
			}
		}

		if os.Getenv("ADMIN_PASSWORD") != "" {
			log.Debug().Msg("found environment variable ADMIN_PASSWORD")
			adminPassword = os.Getenv("ADMIN_PASSWORD")
		} else {
			fmt.Println("enter admin password:")
			if _, err = fmt.Scanln(&adminPassword); err != nil {
				removeUsersDb(cfg.UsersDbPath)
				return
			}
		}

		if os.Getenv("ADMIN_EMAIL") != "" {
			log.Debug().Msg("found environment variable ADMIN_EMAIL")
			adminEmail = os.Getenv("ADMIN_EMAIL")
		} else {
			fmt.Println("enter admin email:")
			if _, err = fmt.Scanln(&adminEmail); err != nil {
				removeUsersDb(cfg.UsersDbPath)
				return
			}
		}

		if err = permissionbolt.ValidUsernamePassword(adminUsername, adminPassword); err != nil {
			removeUsersDb(cfg.UsersDbPath)
			return
		}

		userState.AddUser(adminUsername, adminPassword, adminEmail)

		userState.SetAdminStatus(adminUsername)

		userState.SetBooleanField(adminUsername, "super", true)

		log.Info().Msgf("Initialized users database %s with super admin %s", cfg.UsersDbPath, adminUsername)
	}

	cookieTimeout = time.Minute * time.Duration(cfg.CookieTimeout)

	userState.SetCookieTimeout(int64(cookieTimeout.Seconds()))

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

	router := mux.NewRouter()

	handlerCfg, err := os.ReadFile(cfg.HandlersFilePath)
	if err != nil {
		return
	}

	handlers, err := parseHandlers(bytes.NewReader(handlerCfg))
	if err != nil {
		return
	}

	for _, h := range handlers {
		switch h.pathType {
		case superPath:
			router.Handle(h.path, h.limiter(checkCookieExpiryMiddleware(checkSuperAdminMiddleware(h.handler)))).Methods(h.method)
			perm.AddAdminPath(h.path)
		case adminPath:
			router.Handle(h.path, h.limiter(checkCookieExpiryMiddleware(checkAdminMiddleware(h.handler)))).Methods(h.method)
			perm.AddAdminPath(h.path)
		case userPath:
			router.Handle(h.path, h.limiter(checkCookieExpiryMiddleware(checkConfirmedMiddleware(h.handler)))).Methods(h.method)
			perm.AddUserPath(h.path)
		case publicPath:
			router.Handle(h.path, h.limiter(DummyMiddleware(h.handler))).Methods(h.method)
			perm.AddPublicPath(h.path)
		default:
			return fmt.Errorf("invalid path type: %d", h.pathType)
		}
	}

	passwordResetUrl = cfg.PasswordResetUrl

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	})

	n := negroni.Classic()

	n.Use(c)
	n.Use(perm)
	n.UseHandler(router)

	sigChan := make(chan os.Signal, 1)
	errChan := make(chan error)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		errChan <- fmt.Errorf("%v signal received, shutting down", sig)
	}()

	s := fmt.Sprintf("itpg-backend (%s) listening on port %s", cfg.DbBackend, cfg.Port)
	if !cfg.UseSmtp {
		s += " with SMTPS,"
	} else {
		s += " with SMTP,"
	}

	go func() {
		if !cfg.UseHttp {
			log.Info().Msgf("%s with HTTPS", s)
			errChan <- http.ListenAndServeTLS(":"+cfg.Port, cfg.CertFilePath, cfg.KeyFilePath, n)
		} else {
			log.Info().Msgf("%s with HTTP", s)
			errChan <- http.ListenAndServe(":"+cfg.Port, n)
		}
	}()

	return <-errChan
}

func removeUsersDb(path string) {
	if err := os.Remove(path); err != nil {
		panic(err)
	}
}
