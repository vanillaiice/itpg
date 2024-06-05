package server

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/responses"
)

// GradeData contains data needed to grade a course.
type GradeData struct {
	CourseCode      string  `json:"code"`
	ProfUUID        string  `json:"uuid"`
	GradeTeaching   float32 `json:"teaching"`
	GradeCoursework float32 `json:"coursework"`
	GradeLearning   float32 `json:"learning"`
}

// addCourse handles the HTTP request to add a new course.
func addCourse(w http.ResponseWriter, r *http.Request) {
	courseCode, courseName := r.FormValue("code"), r.FormValue("name")
	if err := isEmptyStr(w, courseCode, courseName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.AddCourse(&db.Course{Code: courseCode, Name: courseName}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// addProfessor handles the HTTP request to add a new professor.
func addProfessor(w http.ResponseWriter, r *http.Request) {
	fullName := r.FormValue("fullname")
	if err := isEmptyStr(w, fullName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.AddProfessor(fullName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// removeCourse handles the HTTP request to remove a course.
func removeCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.RemoveCourse(courseCode, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// removeCourseForce handles the HTTP request to forcefully remove a course.
func removeCourseForce(w http.ResponseWriter, r *http.Request) {
	courseCode := r.FormValue("code")
	if err := isEmptyStr(w, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.RemoveCourse(courseCode, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// removeProfessor handles the HTTP request to remove a professor.
func removeProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.RemoveProfessor(professorUUID, false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// removeProfessorForce handles the HTTP request to forcefully remove a professor.
func removeProfessorForce(w http.ResponseWriter, r *http.Request) {
	professorUUID := r.FormValue("uuid")
	if err := isEmptyStr(w, professorUUID); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.RemoveProfessor(professorUUID, true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// addCourseProfessor handles the HTTP request to associate a course with a professor.
func addCourseProfessor(w http.ResponseWriter, r *http.Request) {
	professorUUID, courseCode := r.FormValue("uuid"), r.FormValue("code")
	if err := isEmptyStr(w, professorUUID, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	if err := dataDb.AddCourseProfessor(professorUUID, courseCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// getLastCourses handles the HTTP request to get all courses.
func getLastCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := dataDb.GetLastCourses()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: courses}).WriteJSON(w)
}

// getLastProfessors handles the HTTP request to get all professors.
func getLastProfessors(w http.ResponseWriter, r *http.Request) {
	professors, err := dataDb.GetLastProfessors()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: professors}).WriteJSON(w)
}

// getLastScores handles the HTTP request to get all scores.
func getLastScores(w http.ResponseWriter, r *http.Request) {
	scores, err := dataDb.GetLastScores()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getCoursesByProfessor handles the HTTP request to get courses associated with a professor.
func getCoursesByProfessorUUID(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["uuid"]
	if err := isEmptyStr(w, professorUUID); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	courses, err := dataDb.GetCoursesByProfessorUUID(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: courses}).WriteJSON(w)
}

// getProfessorsByCourse handles the HTTP request to get professors associated with a course.
func getProfessorsByCourseCode(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	professors, err := dataDb.GetProfessorsByCourseCode(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: professors}).WriteJSON(w)
}

// getScoresByProfessorUUID handles the HTTP request to get scores associated with a professor.
func getScoresByProfessorUUID(w http.ResponseWriter, r *http.Request) {
	professorUUID := mux.Vars(r)["uuid"]
	if err := isEmptyStr(w, professorUUID); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByProfessorUUID(professorUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByProfessorName handles the HTTP request to get scores associated with a professor's name.
func getScoresByProfessorName(w http.ResponseWriter, r *http.Request) {
	professorName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, professorName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByProfessorName(professorName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByProfessorNameLike handles the HTTP request to get scores associated with a professor's name.
func getScoresByProfessorNameLike(w http.ResponseWriter, r *http.Request) {
	professorName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, professorName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByProfessorNameLike(professorName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByCourseName handles the HTTP request to get scores associated with a course.
func getScoresByCourseName(w http.ResponseWriter, r *http.Request) {
	courseName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, courseName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByCourseName(courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByCourseNameLike handles the HTTP request to get scores associated with a course.
func getScoresByCourseNameLike(w http.ResponseWriter, r *http.Request) {
	courseName := mux.Vars(r)["name"]
	if err := isEmptyStr(w, courseName); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByCourseNameLike(courseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByCourseCode handles the HTTP request to get scores associated with a course.
func getScoresByCourseCode(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByCourseCode(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// getScoresByCourseCodeLike handles the HTTP request to get scores associated with a course.
func getScoresByCourseCodeLike(w http.ResponseWriter, r *http.Request) {
	courseCode := mux.Vars(r)["code"]
	if err := isEmptyStr(w, courseCode); err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	scores, err := dataDb.GetScoresByCourseCodeLike(courseCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		logger.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	(&responses.Response{Code: responses.SuccessCode, Message: scores}).WriteJSON(w)
}

// gradeCourseProfessor handles the HTTP request to grade a professor for a specific course.
func gradeCourseProfessor(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(usernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	gradeData, err := decodeGradeData(w, r)
	if err != nil {
		logger.Error().Msg(err.Error())
		return
	}

	grades := [3]float32{gradeData.GradeTeaching, gradeData.GradeCoursework, gradeData.GradeLearning}
	if err := dataDb.GradeCourseProfessor(gradeData.ProfUUID, gradeData.CourseCode, username, grades); err != nil {
		if errors.Is(err, responses.ErrCourseGraded) {
			w.WriteHeader(http.StatusForbidden)
			responses.ErrCourseGraded.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			responses.ErrInternal.WriteJSON(w)
			logger.Error().Msg(err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}
