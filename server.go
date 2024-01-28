package itpg

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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
	Status Status `json:"status"`
	Msg    string `json:"message"`
}

// db represents a pointer to a database connection.
var db *DB

// AddCourse handles the HTTP request to add a new course.
func AddCourse(w http.ResponseWriter, r *http.Request) {
	code, name := r.FormValue("code"), r.FormValue("name")
	_, err := db.AddCourse(code, name)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "added course"})
}

// AddProfessor handles the HTTP request to add a new professor.
func AddProfessor(w http.ResponseWriter, r *http.Request) {
	sname := r.FormValue("surname")
	mname := r.FormValue("middlename")
	name := r.FormValue("name")
	_, err := db.AddProfessor(sname, mname, name)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "added professor"})
}

// AddCourseProfessor handles the HTTP request to associate a course with a professor.
func AddCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	code := r.FormValue("code")
	_, err = db.AddCourseProfessor(professorId, code)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "added course to professor"})
}

// RemoveCourse handles the HTTP request to remove a course.
func RemoveCourse(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	_, err := db.RemoveCourse(code, false)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "removed course"})
}

// RemoveCourseForce handles the HTTP request to forcefully remove a course.
func RemoveCourseForce(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	_, err := db.RemoveCourse(code, true)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "removed course"})
}

// RemoveCourseProfessor handles the HTTP request to disassociate a course from a professor.
func RemoveCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	code := r.FormValue("code")
	_, err = db.RemoveCourseProfessor(professorId, code)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "removed course from professor"})
}

// RemoveProfessor handles the HTTP request to remove a professor.
func RemoveProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	_, err = db.RemoveProfessor(professorId, false)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "removed professor"})
}

// RemoveProfessorForce handles the HTTP request to forcefully remove a professor.
func RemoveProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	_, err = db.RemoveProfessor(professorId, true)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "removed professor"})
}

// GetAllCourses handles the HTTP request to get all courses.
func GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := db.GetAllCourses()
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(courses)
}

// GetAllProfessors handles the HTTP request to get all professors.
func GetAllProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := db.GetAllProfessors()
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
	}
	json.NewEncoder(w).Encode(professors)
}

// GetAllScores handles the HTTP request to get all scores.
func GetAllScores(w http.ResponseWriter, r *http.Request) {
	scores, err := db.GetAllScores()
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
	}
	json.NewEncoder(w).Encode(scores)
}

// GetCoursesByProfessor handles the HTTP request to get courses associated with a professor.
func GetCoursesByProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	courses, err := db.GetCoursesByProfessor(professorId)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(courses)
}

// GetProfessorsByCourse handles the HTTP request to get professors associated with a course.
func GetProfessorsByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	professors, err := db.GetProfessorsByCourse(courseCode)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(professors)
}

// GetScoresByProfessor handles the HTTP request to get scores associated with a professor.
func GetScoresByProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	scores, err := db.GetScoresByProfessor(professorId)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(scores)
}

// GetScoresByCourse handles the HTTP request to get scores associated with a course.
func GetScoresByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	scores, err := db.GetScoresByCourse(courseCode)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(scores)
}

// GradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func GradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	grade, err := strconv.ParseFloat(r.FormValue("grade"), 32)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	code := r.FormValue("code")
	_, err = db.GradeCourseProfessor(professorId, code, float32(grade))
	if err != nil {
		json.NewEncoder(w).Encode(&Response{Status: Err, Msg: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(&Response{Status: Ok, Msg: "graded professor"})
}

// Run starts the HTTP server on the specified port and connects to the specified database.
func Run(port, dbPath string) (err error) {
	db, err = NewDB(dbPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	n := negroni.Classic()

	r := mux.NewRouter()

	r.HandleFunc("/courses", GetAllCourses).Methods("GET")
	r.HandleFunc("/professors", GetAllProfessors).Methods("GET")
	r.HandleFunc("/scores", GetAllScores).Methods("GET")

	r.HandleFunc("/courses/{id}", GetCoursesByProfessor).Methods("GET")
	r.HandleFunc("/professors/{code}", GetProfessorsByCourse).Methods("GET")
	r.HandleFunc("/scores/prof/{id}", GetScoresByProfessor).Methods("GET")
	r.HandleFunc("/scores/course/{id}", GetScoresByCourse).Methods("GET")

	r.HandleFunc("/courses/add", AddCourse).Methods("PUT")
	r.HandleFunc("/courses/addprof", AddCourseProfessor).Methods("PUT")
	r.HandleFunc("/courses/grade", GradeCourseProfessor).Methods("PUT")
	r.HandleFunc("/professors/add", AddProfessor).Methods("PUT")

	r.HandleFunc("/courses/remove", RemoveCourse).Methods("DELETE")
	r.HandleFunc("/courses/removeforce", RemoveCourseForce).Methods("DELETE")
	r.HandleFunc("/courses/removeprof", RemoveCourseProfessor).Methods("DELETE")
	r.HandleFunc("/professors/remove", RemoveProfessor).Methods("DELETE")
	r.HandleFunc("/professors/removeforce", RemoveProfessorForce).Methods("DELETE")

	n.UseHandler(r)

	return http.ListenAndServe(port, n)
}
