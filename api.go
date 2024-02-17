package itpg

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// AddCourse handles the HTTP request to add a new course.
func AddCourse(w http.ResponseWriter, r *http.Request) {
	courseCode, courseName := r.FormValue("code"), r.FormValue("name")
	if err := isEmptyStr(w, courseCode, courseName); err != nil {
		return
	}

	if _, err := DataDB.AddCourse(courseCode, courseName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// AddProfessor handles the HTTP request to add a new professor.
func AddProfessor(w http.ResponseWriter, r *http.Request) {
	fullName := r.FormValue("fullname")
	if err := isEmptyStr(w, fullName); err != nil {
		return
	}

	if _, err := DataDB.AddProfessor(fullName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// AddCourseProfessor handles the HTTP request to associate a course with a professor.
func AddCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID, courseCode := r.FormValue("uuid"), r.FormValue("code")
	if err := isEmptyStr(w, professorUUID, courseCode); err != nil {
		return
	}

	if _, err := DataDB.AddCourseProfessor(professorUUID, courseCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// RemoveCourse handles the HTTP request to remove a course.
func RemoveCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	if _, err := DataDB.RemoveCourse(courseCode, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// RemoveCourseForce handles the HTTP request to forcefully remove a course.
func RemoveCourseForce(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	if _, err := DataDB.RemoveCourse(courseCode, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// RemoveCourseProfessor handles the HTTP request to disassociate a course from a professor.
func RemoveCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID, courseCode := r.FormValue("uuid"), r.FormValue("code")
	if err := isEmptyStr(w, professorUUID, courseCode); err != nil {
		return
	}

	if _, err := DataDB.RemoveCourseProfessor(professorUUID, courseCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// RemoveProfessor handles the HTTP request to remove a professor.
func RemoveProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	if _, err := DataDB.RemoveProfessor(professorUUID, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// RemoveProfessorForce handles the HTTP request to forcefully remove a professor.
func RemoveProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	if _, err := DataDB.RemoveProfessor(professorUUID, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// GetAllCourses handles the HTTP request to get all courses.
func GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := DataDB.GetAllCourses()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: courses}).WriteJSON(w)
}

// GetAllProfessors handles the HTTP request to get all professors.
func GetAllProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := DataDB.GetAllProfessors()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: professors}).WriteJSON(w)
}

// GetAllScores handles the HTTP request to get all scores.
func GetAllScores(w http.ResponseWriter, r *http.Request) {
	scores, err := DataDB.GetAllScores()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GetCoursesByProfessor handles the HTTP request to get courses associated with a professor.
func GetCoursesByProfessorUUID(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["uuid"]
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	courses, err := DataDB.GetCoursesByProfessorUUID(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: courses}).WriteJSON(w)
}

// GetProfessorsByCourse handles the HTTP request to get professors associated with a course.
func GetProfessorsByCourseCode(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	professors, err := DataDB.GetProfessorsByCourseCode(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: professors}).WriteJSON(w)
}

// GetScoresByProfessorUUID handles the HTTP request to get scores associated with a professor.
func GetScoresByProfessorUUID(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["uuid"]
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByProfessorUUID(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByProfessorName handles the HTTP request to get scores associated with a professor's name.
func GetScoresByProfessorName(w http.ResponseWriter, r *http.Request) {
	professorName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, professorName); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByProfessorName(professorName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByProfessorNameLike handles the HTTP request to get scores associated with a professor's name.
func GetScoresByProfessorNameLike(w http.ResponseWriter, r *http.Request) {
	professorName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, professorName); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByProfessorNameLike(professorName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByCourseCode handles the HTTP request to get scores associated with a course.
func GetScoresByCourseCode(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByCourseCode(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByCourseCodeLike handles the HTTP request to get scores associated with a course.
func GetScoresByCourseCodeLike(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByCourseCodeLike(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&Response{Code: SuccessCode, Message: scores}).WriteJSON(w)
}

// GradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func GradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID, courseCode, grade := r.FormValue("uuid"), r.FormValue("code"), r.FormValue("grade")
	if err := isEmptyStr(w, professorUUID, grade, courseCode); err != nil {
		return
	}

	fgrade, err := strconv.ParseFloat(r.FormValue("grade"), 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	if _, err = DataDB.GradeCourseProfessor(professorUUID, courseCode, float32(fgrade)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}
