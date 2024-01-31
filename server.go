package itpg

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
	// Err indicates an error or unsuccessful response.
	Err Status = "error"
)

// Response represents the standard response format for API endpoints.
type Response struct {
	Status  Status      `json:"status"`
	Message interface{} `json:"message"`
}

// db represents a pointer to a database connection.
var db *DB

// AddCourse handles the HTTP request to add a new course.
func AddCourse(w http.ResponseWriter, r *http.Request) {
	courseCode, courseName := r.FormValue("code"), r.FormValue("name")
	_, err := db.AddCourse(courseCode, courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// AddProfessor handles the HTTP request to add a new professor.
func AddProfessor(w http.ResponseWriter, r *http.Request) {
	fullName := r.FormValue("fullname")
	_, err := db.AddProfessor(fullName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// AddCourseProfessor handles the HTTP request to associate a course with a professor.
func AddCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	courseCode := r.FormValue("code")
	_, err := db.AddCourseProfessor(professorUUID, courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// RemoveCourse handles the HTTP request to remove a course.
func RemoveCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	_, err := db.RemoveCourse(courseCode, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// RemoveCourseForce handles the HTTP request to forcefully remove a course.
func RemoveCourseForce(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	_, err := db.RemoveCourse(courseCode, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// RemoveCourseProfessor handles the HTTP request to disassociate a course from a professor.
func RemoveCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	courseCode := r.FormValue("code")
	_, err := db.RemoveCourseProfessor(professorUUID, courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// RemoveProfessor handles the HTTP request to remove a professor.
func RemoveProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	_, err := db.RemoveProfessor(professorUUID, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// RemoveProfessorForce handles the HTTP request to forcefully remove a professor.
func RemoveProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	_, err := db.RemoveProfessor(professorUUID, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// GetAllCourses handles the HTTP request to get all courses.
func GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := db.GetAllCourses()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: courses})
}

// GetAllProfessors handles the HTTP request to get all professors.
func GetAllProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := db.GetAllProfessors()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: professors})
}

// GetAllScores handles the HTTP request to get all scores.
func GetAllScores(w http.ResponseWriter, r *http.Request) {
	scores, err := db.GetAllScores()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: scores})
}

// GetCoursesByProfessor handles the HTTP request to get courses associated with a professor.
func GetCoursesByProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["id"]
	courses, err := db.GetCoursesByProfessor(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: courses})
}

// GetProfessorsByCourse handles the HTTP request to get professors associated with a course.
func GetProfessorsByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	professors, err := db.GetProfessorsByCourse(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: professors})
}

// GetScoresByProfessor handles the HTTP request to get scores associated with a professor.
func GetScoresByProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["id"]
	scores, err := db.GetScoresByProfessor(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: scores})
}

// GetScoresByCourse handles the HTTP request to get scores associated with a course.
func GetScoresByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	scores, err := db.GetScoresByCourse(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: scores})
}

// GradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func GradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	grade, err := strconv.ParseFloat(r.FormValue("grade"), 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	courseCode := r.FormValue("code")
	_, err = db.GradeCourseProfessor(professorUUID, courseCode, float32(grade))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{Status: Ok, Message: "success"})
}

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(port, dbPath string, allowedOrigins []string) (err error) {
	db, err = NewDB(dbPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"PUT", "GET", "DELETE"},
	})

	n := negroni.Classic()

	r := mux.NewRouter()

	r.HandleFunc("/courses", GetAllCourses).Methods("GET")
	r.HandleFunc("/professors", GetAllProfessors).Methods("GET")
	r.HandleFunc("/scores", GetAllScores).Methods("GET")

	r.HandleFunc("/courses/{uuid}", GetCoursesByProfessor).Methods("GET")
	r.HandleFunc("/professors/{code}", GetProfessorsByCourse).Methods("GET")
	r.HandleFunc("/scores/prof/{uuid}", GetScoresByProfessor).Methods("GET")
	r.HandleFunc("/scores/course/{uuid}", GetScoresByCourse).Methods("GET")

	r.HandleFunc("/courses/add", AddCourse).Methods("PUT")
	r.HandleFunc("/courses/addprof", AddCourseProfessor).Methods("PUT")
	r.HandleFunc("/professors/add", AddProfessor).Methods("PUT")

	r.HandleFunc("/courses/grade", GradeCourseProfessor).Methods("UPDATE")

	r.HandleFunc("/courses/remove", RemoveCourse).Methods("DELETE")
	r.HandleFunc("/courses/removeforce", RemoveCourseForce).Methods("DELETE")
	r.HandleFunc("/courses/removeprof", RemoveCourseProfessor).Methods("DELETE")
	r.HandleFunc("/professors/remove", RemoveProfessor).Methods("DELETE")
	r.HandleFunc("/professors/removeforce", RemoveProfessorForce).Methods("DELETE")

	n.Use(c)
	n.UseHandler(r)

	log.Printf("itpg-backend listening on port %q\n", port)
	return http.ListenAndServe(":"+port, n)
}
