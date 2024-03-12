package itpg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"itpg/db"
	"itpg/responses"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

var professorNames = []string{
	"Great Teacher Onizuka",
	"Pippy Peepee Poopypants",
	"Professor Oak",
}
var courses = []*db.Course{
	{Code: "S209", Name: "How to replace head gaskets"},
	{Code: "CN9A", Name: "Controlling the Anti Lag System"},
	{Code: "AE86", Name: "How to beat any car"},
}
var professors = []*db.Professor{}
var scores = []*db.Score{}

func initDB(path ...string) (*db.DB, error) {
	if len(path) == 0 {
		path = append(path, ":memory:")
	}
	db, err := db.NewDB(path[0], true)
	if err != nil {
		return nil, err
	}

	for _, c := range courses {
		_, err := db.AddCourse(c)
		if err != nil {
			return nil, err
		}
	}

	for _, p := range professorNames {
		_, err := db.AddProfessor(p)
		if err != nil {
			return nil, err
		}
	}
	professors, err = db.GetLastProfessors()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(professors) && i < len(courses); i++ {
		_, err := db.AddCourseProfessor(professors[i].UUID, courses[i].Code)
		if err != nil {
			return nil, err
		}
	}
	scores, err = db.GetLastScores()
	if err != nil {
		return nil, err
	}

	return db, err
}

func dbInit() (err error) {
	DataDB, err = initDB()
	if err != nil {
		return
	}
	return
}

func TestServerAddCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("POST", "/course/add?code=GC8F&name=Showing%20your%20son%20whose%20the%20boss", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourse(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerAddProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer DataDB.Close()
	r, err := http.NewRequest("POST", "/professors/add?fullname=Gintoki%20Sakata%20Senpai", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerAddCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("POST", fmt.Sprintf("/courses/addprof?uuid=%s&code=S209", professors[1].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerRemoveCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("DELETE", "/courses/remove?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourse(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != responses.ErrInternal.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerRemoveCourseForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("DELETE", "/courses/removeforce?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerRemoveCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("DELETE", "/courses/removeprof?uuid=1&code=CN9A", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerRemoveProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/remove?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessor(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != responses.ErrInternal.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerRemoveProfessorForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/removeforce?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessorForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}

func TestServerGetLastCourses(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", "/courses", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetLastCourses(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 0 {
		t.Errorf("got len = 0, want %s", "> 0")
	}
	if resp.Code != responses.SuccessCode {
		t.Errorf("got %d, want %d", resp.Code, responses.SuccessCode)
	}
}

func TestServerGetLastProfessors(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", "/professors", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetLastProfessors(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 0 {
		t.Errorf("got len = 0, want %s", "> 0")
	}
}

func TestServerGetLastScores(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", "/scores", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetLastScores(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetCoursesByProfessorUUID(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/courses/%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/courses/{uuid}", GetCoursesByProfessorUUID)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetProfessorsByCourseCode(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/professors/%s", courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/professors/{code}", GetProfessorsByCourseCode)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByProfessorUUID(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/prof/%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/prof/{uuid}", GetScoresByProfessorUUID)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByProfessorName(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/name/%s", professors[0].Name), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/name/{name}", GetScoresByProfessorName)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByProfessorNameLike(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/namelike/%s", professors[0].Name[:5]), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/namelike/{name}", GetScoresByProfessorNameLike)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByCourseName(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/coursename/%s", courses[0].Name), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/coursename/{name}", GetScoresByCourseName)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByCourseNameLike(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/coursenamelike/%s", courses[0].Name[:5]), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/coursenamelike/{name}", GetScoresByCourseNameLike)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 2 {
		t.Errorf("got %d, want %d", lresp, 2)
	}
}

func TestServerGetScoresByCourseCode(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/coursecode/%s", courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/coursecode/{code}", GetScoresByCourseCode)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByCourseCodeLike(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/coursecodelike/%s", courses[0].Code[:3]), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/coursecodelike/{code}", GetScoresByCourseCodeLike)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &responses.Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGradeCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	defer DataDB.Close()

	data, _ := json.Marshal(&GradeData{CourseCode: courses[0].Code, ProfUUID: professors[0].UUID, GradeTeaching: 5, GradeCoursework: 4, GradeLearning: 3})
	r := httptest.NewRequest("POST", "/courses/grade", bytes.NewReader(data))
	rr := httptest.NewRecorder()
	err = initTestUserState()
	if err != nil {
		log.Fatal(err)
	}
	defer removeUserState()
	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.Confirm(creds.Email)
	UserState.Login(rr, creds.Email)
	cookie := rr.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)
	r = r.WithContext(context.WithValue(r.Context(), "username", creds.Email))
	GradeCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != responses.Success.Error() {
		t.Errorf("got %s, want %s", rr.Body.String(), responses.Success.Error())
	}
}
