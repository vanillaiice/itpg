package itpg

import (
	"itpg/db"
	"itpg/responses"
	"log"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/didip/tollbooth_negroni"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"github.com/xyproto/permissionbolt/v2"
	"github.com/xyproto/pinterface"
)

type PathType int

const (
	AdminPath  PathType = 0
	UserPath   PathType = 1
	PublicPath PathType = 2
)

// HandlerInfo represents a struct containing information about an HTTP handler.
type HandlerInfo struct {
	Path     string                                   // Path specifies the URL pattern for which the handler is responsible.
	Handler  func(http.ResponseWriter, *http.Request) // Handler is the function that will be called to handle HTTP requests.
	Method   string                                   // Method specifies the HTTP method associated with the handler.
	PathType PathType
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
		log.Fatal(err)
	}
	AllowedMailDomains = config.AllowedMailDomains

	if err = InitCredsSMTP(config.SMTPEnvPath, !config.UseSMTP); err != nil {
		log.Fatal(err)
	}

	DataDB, err = db.NewDB(config.DBPath, config.Speed)
	if err != nil {
		log.Fatal(err)
	}
	defer DataDB.Close()

	perm, err := permissionbolt.NewWithConf(config.UsersDBPath)
	if err != nil {
		log.Fatal(err)
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

	lmt := tollbooth.NewLimiter(100, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})
	lmt.SetMessageContentType("application/json")
	lmt.SetMessage(responses.ErrRequestLimitReached.Error())
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		responses.ErrRequestLimitReached.WriteJSON(w)
	})

	handlers := []*HandlerInfo{
		// Admin
		{"/courses/add", AddCourse, http.MethodPost, AdminPath},
		{"/professors/add", AddProfessor, http.MethodPost, AdminPath},
		{"/courses/addprof", AddCourseProfessor, http.MethodPost, AdminPath},
		{"/courses/remove", RemoveCourse, http.MethodDelete, AdminPath},
		{"/courses/removeforce", RemoveCourseForce, http.MethodDelete, AdminPath},
		{"/courses/removeprof", RemoveCourseProfessor, http.MethodDelete, AdminPath},
		{"/professors/remove", RemoveProfessor, http.MethodDelete, AdminPath},
		{"/professors/removeforce", RemoveProfessorForce, http.MethodDelete, AdminPath},
		// User
		{"/courses/grade", GradeCourseProfessor, http.MethodPost, UserPath},
		{"/refresh", RefreshCookie, http.MethodPost, UserPath},
		{"/logout", Logout, http.MethodPost, UserPath},
		{"/clear", ClearCookie, http.MethodPost, UserPath},
		{"/changepass", ChangePassword, http.MethodPost, UserPath},
		{"/ping", Ping, http.MethodGet, UserPath},
		// Public
		{"/courses", GetLastCourses, http.MethodGet, PublicPath},
		{"/professors", GetLastProfessors, http.MethodGet, PublicPath},
		{"/scores", GetLastScores, http.MethodGet, PublicPath},
		{"/courses/{uuid}", GetCoursesByProfessorUUID, http.MethodGet, PublicPath},
		{"/professors/{code}", GetProfessorsByCourseCode, http.MethodGet, PublicPath},
		{"/scores/prof/{uuid}", GetScoresByProfessorUUID, http.MethodGet, PublicPath},
		{"/scores/profname/{name}", GetScoresByProfessorName, http.MethodGet, PublicPath},
		{"/scores/profnamelike/{name}", GetScoresByProfessorNameLike, http.MethodGet, PublicPath},
		{"/scores/coursename/{name}", GetScoresByCourseName, http.MethodGet, PublicPath},
		{"/scores/coursenamelike/{name}", GetScoresByCourseNameLike, http.MethodGet, PublicPath},
		{"/scores/coursecode/{code}", GetScoresByCourseCode, http.MethodGet, PublicPath},
		{"/scores/coursecodelike/{code}", GetScoresByCourseCodeLike, http.MethodGet, PublicPath},
		{"/login", Login, http.MethodPost, PublicPath},
		{"/register", Register, http.MethodPost, PublicPath},
		{"/confirm", Confirm, http.MethodPost, PublicPath},
		{"/newconfirmationcode", SendNewConfirmationCode, http.MethodPost, PublicPath},
		{"/sendresetlink", SendResetLink, http.MethodPost, PublicPath},
		{"/resetpass", ResetPassword, http.MethodPost, PublicPath},
	}

	router := mux.NewRouter()

	for _, h := range handlers {
		switch h.PathType {
		case AdminPath:
			router.Handle(h.Path, checkCookieExpiryMiddleware(h.Handler)).Methods(h.Method)
			perm.AddAdminPath(h.Path)
		case UserPath:
			router.Handle(h.Path, checkCookieExpiryMiddleware(checkConfirmedMiddleware(h.Handler))).Methods(h.Method)
			perm.AddUserPath(h.Path)
		case PublicPath:
			router.HandleFunc(h.Path, h.Handler).Methods(h.Method)
			perm.AddPublicPath(h.Path)
		}
	}

	n := negroni.Classic()
	n.Use(c)
	n.Use(perm)
	n.Use(tollbooth_negroni.LimitHandler(lmt))
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
