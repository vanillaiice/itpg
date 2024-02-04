package itpg

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// AddCourse handles the HTTP request to add a new course.
func AddCourse(w http.ResponseWriter, r *http.Request) {
	courseCode, courseName := r.FormValue("code"), r.FormValue("name")
	if err := isEmptyStr(w, courseCode, courseName); err != nil {
		return
	}
	_, err := db.AddCourse(courseCode, courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// AddProfessor handles the HTTP request to add a new professor.
func AddProfessor(w http.ResponseWriter, r *http.Request) {
	fullName := r.FormValue("fullname")
	if err := isEmptyStr(w, fullName); err != nil {
		return
	}
	_, err := db.AddProfessor(fullName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// AddCourseProfessor handles the HTTP request to associate a course with a professor.
func AddCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, professorUUID, courseCode); err != nil {
		return
	}
	_, err := db.AddCourseProfessor(professorUUID, courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RemoveCourse handles the HTTP request to remove a course.
func RemoveCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}
	_, err := db.RemoveCourse(courseCode, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RemoveCourseForce handles the HTTP request to forcefully remove a course.
func RemoveCourseForce(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}
	_, err := db.RemoveCourse(courseCode, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RemoveCourseProfessor handles the HTTP request to disassociate a course from a professor.
func RemoveCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, professorUUID, courseCode); err != nil {
		return
	}
	_, err := db.RemoveCourseProfessor(professorUUID, courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RemoveProfessor handles the HTTP request to remove a professor.
func RemoveProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}
	_, err := db.RemoveProfessor(professorUUID, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RemoveProfessorForce handles the HTTP request to forcefully remove a professor.
func RemoveProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}
	_, err := db.RemoveProfessor(professorUUID, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GetAllCourses handles the HTTP request to get all courses.
func GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := db.GetAllCourses()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: courses})
}

// GetAllProfessors handles the HTTP request to get all professors.
func GetAllProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := db.GetAllProfessors()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: professors})
}

// GetAllScores handles the HTTP request to get all scores.
func GetAllScores(w http.ResponseWriter, r *http.Request) {
	scores, err := db.GetAllScores()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: scores})
}

// GetCoursesByProfessor handles the HTTP request to get courses associated with a professor.
func GetCoursesByProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}
	courses, err := db.GetCoursesByProfessor(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: courses})
}

// GetProfessorsByCourse handles the HTTP request to get professors associated with a course.
func GetProfessorsByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}
	professors, err := db.GetProfessorsByCourse(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: professors})
}

// GetScoresByProfessor handles the HTTP request to get scores associated with a professor.
func GetScoresByProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		return
	}
	scores, err := db.GetScoresByProfessor(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: scores})
}

// GetScoresByCourse handles the HTTP request to get scores associated with a course.
func GetScoresByCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		return
	}
	scores, err := db.GetScoresByCourse(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{Message: scores})
}

// GradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func GradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	courseCode := r.FormValue("code")
	grade := r.FormValue("grade")
	if err := isEmptyStr(w, professorUUID, grade, courseCode); err != nil {
		return
	}

	fgrade, err := strconv.ParseFloat(r.FormValue("grade"), 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.GradeCourseProfessor(professorUUID, courseCode, float32(fgrade))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
