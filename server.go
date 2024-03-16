package itpg

import (
	"itpg/db"
	"itpg/responses"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"github.com/xyproto/permissionbolt/v2"
	"github.com/xyproto/pinterface"
)

// PathType is the type of the path (admin, user, public).
type PathType int

// Enum for path types
const (
	UserPath   PathType = 0 // UserPath is a path only accessible by users.
	PublicPath PathType = 1 // PublicPath is a path accessible by anyone.
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
	PathType PathType                                 // PathType is the type of the path (admin, user, public)
	Limiter  func(http.Handler) http.Handler          // Limiter is the limiter used to limit requests
}

// DataDB represents a pointer to a database connection,
// storing professor names, course codes and names,
// and professor scores.
var DataDB *db.DB

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

// RunConfig defines the server's configuration settings.
type RunConfig struct {
	Port                    string   // Port on which the server will run
	DBPath                  string   // Path to the SQLite database file
	Speed                   bool     // Whether to use prioritize database transaction speed at the cost of data integrity
	UsersDBPath             string   // Path to the users BOLT database file
	SMTPEnvPath             string   // Path to the .env file containing SMTP configuration
	PasswordResetWebsiteURL string   // URL to the password reset website page
	AllowedOrigins          []string // List of allowed origins for CORS
	AllowedMailDomains      []string // List of allowed mail domains for registering with the service
	UseSMTP                 bool     // Whether to use SMTP (false for SMTPS)
	UseHTTP                 bool     // Whether to use HTTP (false for HTTPS)
	CertFilePath            string   // Path to the certificate file (required for HTTPS)
	KeyFilePath             string   // Path to the key file (required for HTTPS)
	CookieTimeout           int      // Duration in minute after which a session cookie expires
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

	DataDB, err = db.NewDB(config.DBPath, config.Speed)
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
		{"/courses/grade", GradeCourseProfessor, http.MethodPost, UserPath, LimiterModerate},
		{"/refresh", RefreshCookie, http.MethodPost, UserPath, LimiterLenient},
		{"/logout", Logout, http.MethodPost, UserPath, LimiterLenient},
		{"/clear", ClearCookie, http.MethodPost, UserPath, LimiterLenient},
		{"/changepass", ChangePassword, http.MethodPost, UserPath, LimiterStrict},
		{"/delete", DeleteAccount, http.MethodPost, UserPath, LimiterVeryStrict},
		{"/ping", Ping, http.MethodGet, UserPath, LimiterLenient},
		// Public
		{"/courses", GetLastCourses, http.MethodGet, PublicPath, LimiterLenient},
		{"/professors", GetLastProfessors, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores", GetLastScores, http.MethodGet, PublicPath, LimiterLenient},
		{"/courses/{uuid}", GetCoursesByProfessorUUID, http.MethodGet, PublicPath, LimiterLenient},
		{"/professors/{code}", GetProfessorsByCourseCode, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/prof/{uuid}", GetScoresByProfessorUUID, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/profname/{name}", GetScoresByProfessorName, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/profnamelike/{name}", GetScoresByProfessorNameLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/coursename/{name}", GetScoresByCourseName, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/coursenamelike/{name}", GetScoresByCourseNameLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/coursecode/{code}", GetScoresByCourseCode, http.MethodGet, PublicPath, LimiterLenient},
		{"/scores/coursecodelike/{code}", GetScoresByCourseCodeLike, http.MethodGet, PublicPath, LimiterLenient},
		{"/login", Login, http.MethodPost, PublicPath, LimiterLenient},
		{"/register", Register, http.MethodPost, PublicPath, LimiterModerate},
		{"/confirm", Confirm, http.MethodPost, PublicPath, LimiterModerate},
		{"/newconfirmationcode", SendNewConfirmationCode, http.MethodPost, PublicPath, LimiterStrict},
		{"/sendresetlink", SendResetLink, http.MethodPost, PublicPath, LimiterVeryStrict},
		{"/resetpass", ResetPassword, http.MethodPost, PublicPath, LimiterVeryStrict},
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

	s := "itpg-backend listening on port " + config.Port
	if !config.UseSMTP {
		s += " with SMTPS,"
	} else {
		s += " with SMTP,"
	}
	if !config.UseHTTP {
		log.Printf("%s with HTTPS\n", s)
		return http.ListenAndServeTLS(":"+config.Port, config.CertFilePath, config.KeyFilePath, n)
	} else {
		log.Printf("%s with HTTP\n", s)
		return http.ListenAndServe(":"+config.Port, n)
	}
}
