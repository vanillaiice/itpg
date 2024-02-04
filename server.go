package itpg

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// Status represents the status of a response.
type Status string

// Constants representing the status values for a response.
const (
	// Ok indicates a successful response.
	Ok Status = "ok"
	// Err indicates an error.
	Err Status = "error"
)

// Message represents the standard response format for API endpoints.
type Message struct {
	Message interface{} `json:"message"`
}

// db represents a pointer to a database connection,
// storing professor names, courses codes and names,
// and professor scores.
var db *DB

// authDB represents a pointer to a database connection,
// storing user credentials and session tokens.
var authDB *AuthDB

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(port, dbPath, authDBPath string, allowedOrigins []string) (err error) {
	db, err = NewDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	authDB, err = NewAuthDB(authDBPath)
	if err != nil {
		log.Fatal(err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "DELETE", "POST"},
	})

	n := negroni.Classic()
	n.Use(c)

	router := mux.NewRouter()
	handlers := []struct {
		Path    string
		Handler func(http.ResponseWriter, *http.Request)
		Method  string
	}{
		{"/courses", GetAllCourses, "GET"},
		{"/professors", GetAllProfessors, "GET"},
		{"/scores", GetAllScores, "GET"},
		{"/courses/{uuid}", GetCoursesByProfessor, "GET"},
		{"/professors/{code}", GetProfessorsByCourse, "GET"},
		{"/scores/prof/{uuid}", GetScoresByProfessor, "GET"},
		{"/scores/course/{uuid}", GetScoresByCourse, "GET"},
		{"/courses/add", AddCourse, "POST"},
		{"/courses/addprof", AddCourseProfessor, "POST"},
		{"/professors/add", AddProfessor, "POST"},
		{"/courses/grade", GradeCourseProfessor, "POST"},
		{"/courses/remove", RemoveCourse, "DELETE"},
		{"/courses/removeforce", RemoveCourseForce, "DELETE"},
		{"/courses/removeprof", RemoveCourseProfessor, "DELETE"},
		{"/professors/remove", RemoveProfessor, "DELETE"},
		{"/professors/removeforce", RemoveProfessorForce, "DELETE"},
		{"/signup", Signup, "POST"},
		{"/login", Login, "POST"},
		{"/logout", Logout, "POST"},
		{"/refresh", Refresh, "POST"},
		{"/greet", Greet, "GET"},
	}
	for _, h := range handlers {
		router.HandleFunc(h.Path, h.Handler).Methods(h.Method)
	}
	n.UseHandler(router)

	log.Printf("itpg-backend listening on port %q\n", port)
	return http.ListenAndServe(":"+port, n)
}
