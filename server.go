package itpg

import (
	"itpg/db"
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

// HandlerInfo represents a struct containing information about an HTTP handler.
type HandlerInfo struct {
	Path    string                                   // Path specifies the URL pattern for which the handler is responsible.
	Handler func(http.ResponseWriter, *http.Request) // Handler is the function that will be called to handle HTTP requests.
	Method  string                                   // Method specifies the HTTP method associated with the handler.
}

// DataDB represents a pointer to a database connection,
// storing professor names, course codes and names,
// and professor scores.
var DataDB *db.DB

// UserState stores the state of all users.
var UserState pinterface.IUserState

// CookieTimeout represents the duration after which a session cookie expires.
const CookieTimeout = 30 * time.Minute

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(port, dbPath, usersDbPath, envPath string, speed bool, allowedOrigins, allowedMailDomains []string, useSMTP, useHTTP bool, certFile, keyFile string) (err error) {
	if err = InitCredsSMTP(envPath, !useSMTP); err != nil {
		log.Fatal(err)
	}

	DataDB, err = db.NewDB(dbPath, speed)
	if err != nil {
		log.Fatal(err)
	}
	defer DataDB.Close()

	perm, err := permissionbolt.NewWithConf(usersDbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = validAllowedDomains(allowedMailDomains); err != nil {
		log.Fatal(err)
	}
	AllowedMailDomains = allowedMailDomains

	perm.SetDenyFunction(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		ErrPermissionDenied.WriteJSON(w)
	})

	UserState = perm.UserState()
	UserState.SetCookieTimeout(int64(CookieTimeout.Seconds()))

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	})

	lmt := tollbooth.NewLimiter(10, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})
	lmt.SetMessageContentType("application/json")
	lmt.SetMessage(ErrRequestLimitReached.String())
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		ErrRequestLimitReached.WriteJSON(w)
	})

	adminHandlers := []HandlerInfo{
		{"/courses/add", AddCourse, http.MethodPost},
		{"/professors/add", AddProfessor, http.MethodPost},
		{"/courses/addprof", AddCourseProfessor, http.MethodPost},
		{"/courses/remove", RemoveCourse, http.MethodDelete},
		{"/courses/removeforce", RemoveCourseForce, http.MethodDelete},
		{"/courses/removeprof", RemoveCourseProfessor, http.MethodDelete},
		{"/professors/remove", RemoveProfessor, http.MethodDelete},
		{"/professors/removeforce", RemoveProfessorForce, http.MethodDelete},
	}
	userHandlers := []HandlerInfo{
		{"/courses/grade", GradeCourseProfessor, http.MethodPost},
		{"/refresh", RefreshCookie, http.MethodPost},
		{"/logout", Logout, http.MethodPost},
		{"/clear", ClearCookie, http.MethodPost},
		{"/changepass", ChangePassword, http.MethodPost},
		{"/delete", DeleteAccount, http.MethodPost},
		{"/ping", Ping, http.MethodGet},
	}
	publicHandlers := []HandlerInfo{
		{"/courses", GetAllCourses, http.MethodGet},
		{"/professors", GetAllProfessors, http.MethodGet},
		{"/scores", GetAllScores, http.MethodGet},
		{"/courses/{uuid}", GetCoursesByProfessorUUID, http.MethodGet},
		{"/professors/{code}", GetProfessorsByCourseCode, http.MethodGet},
		{"/scores/prof/{uuid}", GetScoresByProfessorUUID, http.MethodGet},
		{"/scores/name/{name}", GetScoresByProfessorName, http.MethodGet},
		{"/scores/namelike/{name}", GetScoresByProfessorNameLike, http.MethodGet},
		{"/scores/course/{code}", GetScoresByCourseCode, http.MethodGet},
		{"/scores/courselike/{code}", GetScoresByCourseCodeLike, http.MethodGet},
		{"/login", Login, http.MethodPost},
		{"/register", Register, http.MethodPost},
		{"/confirm", Confirm, http.MethodPost},
		{"/newconfirmationcode", SendNewConfirmationCode, http.MethodPost},
	}

	router := mux.NewRouter()
	for _, h := range adminHandlers {
		router.Handle(h.Path, checkCookieExpiryMiddleware(h.Handler)).Methods(h.Method)
		perm.AddAdminPath(h.Path)
	}
	for _, h := range userHandlers {
		if h.Path == "/courses/grade" {
			router.Handle(h.Path, checkConfirmedMiddleware(checkCookieExpiryMiddleware(checkUserAlreadyGradedMiddleware(h.Handler)))).Methods(h.Method)
			perm.AddUserPath(h.Path)
			continue
		}
		router.Handle(h.Path, checkConfirmedMiddleware(checkCookieExpiryMiddleware(h.Handler))).Methods(h.Method)
		perm.AddUserPath(h.Path)
	}
	for _, h := range publicHandlers {
		router.HandleFunc(h.Path, h.Handler).Methods(h.Method)
		perm.AddPublicPath(h.Path)
	}

	n := negroni.Classic()
	n.Use(c)
	n.Use(perm)
	n.Use(tollbooth_negroni.LimitHandler(lmt))
	n.UseHandler(router)

	if !useHTTP {
		log.Printf("itpg-backend listening on port %s with HTTPS\n", port)
		return http.ListenAndServeTLS(":"+port, certFile, keyFile, n)
	} else {
		log.Printf("itpg-backend listening on port %s\n", port)
		return http.ListenAndServe(":"+port, n)
	}
}
