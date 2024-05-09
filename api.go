package itpg

import (
	"errors"
	"itpg/db"
	"itpg/responses"
	"net/http"

	"github.com/gorilla/mux"
)

// GradeData contains data needed to grade a course.
type GradeData struct {
	CourseCode      string  `json:"code"`
	ProfUUID        string  `json:"uuid"`
	GradeTeaching   float32 `json:"teaching"`
	GradeCoursework float32 `json:"coursework"`
	GradeLearning   float32 `json:"learning"`
}

// AddCourse handles the HTTP request to add a new course.
func AddCourse(w http.ResponseWriter, r *http.Request) {
	courseCode, courseName := r.FormValue("code"), r.FormValue("name")
	if err := isEmptyStr(w, courseCode, courseName); err != nil {
		return
	}

	if err := DataDB.AddCourse(&db.Course{Code: courseCode, Name: courseName}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// AddProfessor handles the HTTP request to add a new professor.
func AddProfessor(w http.ResponseWriter, r *http.Request) {
	fullName := r.FormValue("fullname")
	if err := isEmptyStr(w, fullName); err != nil {
		return
	}

	if err := DataDB.AddProfessor(fullName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// RemoveCourse handles the HTTP request to remove a course.
func RemoveCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	if err := DataDB.RemoveCourse(courseCode, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// RemoveCourseForce handles the HTTP request to forcefully remove a course.
func RemoveCourseForce(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}

	if err := DataDB.RemoveCourse(courseCode, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// RemoveProfessor handles the HTTP request to remove a professor.
func RemoveProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	if err := DataDB.RemoveProfessor(professorUUID, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// RemoveProfessorForce handles the HTTP request to forcefully remove a professor.
func RemoveProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}

	if err := DataDB.RemoveProfessor(professorUUID, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// GetLastCourses handles the HTTP request to get all courses.
func GetLastCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := DataDB.GetLastCourses()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: courses}).WriteJSON(w)
}

// GetLastProfessors handles the HTTP request to get all professors.
func GetLastProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := DataDB.GetLastProfessors()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: professors}).WriteJSON(w)
}

// GetLastScores handles the HTTP request to get all scores.
func GetLastScores(w http.ResponseWriter, r *http.Request) {
	scores, err := DataDB.GetLastScores()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: courses}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: professors}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByCourseName handles the HTTP request to get scores associated with a course.
func GetScoresByCourseName(w http.ResponseWriter, r *http.Request) {
	courseName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, courseName); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByCourseName(courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// GetScoresByCourseNameLike handles the HTTP request to get scores associated with a course.
func GetScoresByCourseNameLike(w http.ResponseWriter, r *http.Request) {
	courseName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, courseName); err != nil {
		return
	}

	scores, err := DataDB.GetScoresByCourseNameLike(courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
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
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// GradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func GradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(UsernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	gradeData, err := decodeGradeData(w, r)
	if err != nil {
		return
	}

	grades := [3]float32{gradeData.GradeTeaching, gradeData.GradeCoursework, gradeData.GradeLearning}
	if err := DataDB.GradeCourseProfessor(gradeData.ProfUUID, gradeData.CourseCode, username, grades); err != nil {
		if errors.Is(err, responses.ErrCourseGraded) {
			w.WriteHeader(http.StatusForbidden)
			responses.ErrCourseGraded.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			responses.ErrInternal.WriteJSON(w)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}
