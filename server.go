package itpg

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"github.com/xyproto/permissionbolt/v2"
	"github.com/xyproto/pinterface"
)

// Message represents the standard response format for API endpoints.
type Message struct {
	Message interface{} `json:"message"`
}

// Handler represents a struct containing information about an HTTP handler.
type Handler struct {
	Path    string                                   // Path specifies the URL pattern for which the handler is responsible.
	Handler func(http.ResponseWriter, *http.Request) // Handler is the function that will be called to handle HTTP requests.
	Method  string                                   // Method specifies the HTTP method (e.g., GET, POST, PUT, DELETE) associated with the handler.
}

// db represents a pointer to a database connection,
// storing professor names, courses codes and names,
// and professor scores.
var db *DB

// userState stores the state of all users.
var userState pinterface.IUserState

// cookieTimeout represents the duration after which a session cookie expires.
const cookieTimeout = 30 * time.Minute

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(port, dbPath string, allowedOrigins []string) (err error) {
	db, err = NewDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	perm, err := permissionbolt.New()
	if err != nil {
		log.Fatal(err)
	}
	userState = perm.UserState()
	userState.SetCookieTimeout(int64(cookieTimeout.Seconds()))

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodDelete, http.MethodPost},
		AllowCredentials: true,
	})

	adminHandlers := []Handler{
		{"/courses/add", AddCourse, http.MethodPost},
		{"/courses/addprof", AddCourseProfessor, http.MethodPost},
		{"/professors/add", AddProfessor, http.MethodPost},
		{"/courses/remove", RemoveCourse, http.MethodDelete},
		{"/courses/removeforce", RemoveCourseForce, http.MethodDelete},
		{"/courses/removeprof", RemoveCourseProfessor, http.MethodDelete},
		{"/professors/remove", RemoveProfessor, http.MethodDelete},
		{"/professors/removeforce", RemoveProfessorForce, http.MethodDelete},
	}
	userHandlers := []Handler{
		{"/courses/grade", GradeCourseProfessor, http.MethodPost},
		{"/refresh", RefreshCookie, http.MethodPost},
		{"/logout", Logout, http.MethodPost},
		{"/clear", ClearCookie, http.MethodPost},
		{"/delete", DeleteAccount, http.MethodPost},
		{"/greet", Greet, http.MethodGet},
	}
	publicHandlers := []Handler{
		{"/courses", GetAllCourses, http.MethodGet},
		{"/professors", GetAllProfessors, http.MethodGet},
		{"/scores", GetAllScores, http.MethodGet},
		{"/courses/{uuid}", GetCoursesByProfessor, http.MethodGet},
		{"/professors/{code}", GetProfessorsByCourse, http.MethodGet},
		{"/scores/prof/{uuid}", GetScoresByProfessor, http.MethodGet},
		{"/scores/course/{code}", GetScoresByCourse, http.MethodGet},
		{"/login", Login, http.MethodPost},
		{"/register", Register, http.MethodPost},
	}

	router := mux.NewRouter()
	for _, h := range adminHandlers {
		router.Handle(h.Path, checkCookieExpiryMiddleware(http.HandlerFunc(h.Handler))).Methods(h.Method)
		perm.AddAdminPath(h.Path)
	}
	for _, h := range userHandlers {
		if h.Path == "/courses/grade" {
			router.Handle(h.Path, checkCookieExpiryMiddleware(checkUserAlreadyGradedMiddleware(h.Handler))).Methods(h.Method)
			perm.AddUserPath(h.Path)
			continue
		}
		router.Handle(h.Path, checkCookieExpiryMiddleware(http.HandlerFunc(h.Handler))).Methods(h.Method)
		perm.AddUserPath(h.Path)
	}
	for _, h := range publicHandlers {
		router.HandleFunc(h.Path, h.Handler).Methods(h.Method)
		perm.AddPublicPath(h.Path)
	}

	n := negroni.Classic()
	n.Use(c)
	n.Use(perm)
	n.UseHandler(router)

	log.Printf("itpg-backend listening on port %q\n", port)
	return http.ListenAndServe(":"+port, n)
}
