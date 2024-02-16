package itpg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func dbInit() (err error) {
	db, err = initDb()
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
	r, err := http.NewRequest("POST", "/course/add?code=GC8F&name=Showing%20your%20son%20whose%20the%20boss", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourse(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerAddProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err.Error())
	}
	r, err := http.NewRequest("POST", "/professors/add?fullname=Gintoki%20Sakata%20Senpai", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerAddCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("POST", fmt.Sprintf("/courses/addprof?uuid=%s&code=S209", professors[1].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerRemoveCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", "/courses/remove?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourse(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != errInternal.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerRemoveCourseForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", "/courses/removeforce?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerRemoveCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", "/courses/removeprof?uuid=1&code=CN9A", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerRemoveProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/remove?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessor(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != errInternal.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerRemoveProfessorForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/removeforce?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessorForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}

func TestServerGetAllCourses(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", "/courses", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllCourses(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 0 {
		t.Errorf("got len = 0, want %s", "> 0")
	}
	if resp.Code != successCode {
		t.Errorf("got %d, want %d", resp.Code, successCode)
	}
}

func TestServerGetAllProfessors(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", "/professors", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllProfessors(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 0 {
		t.Errorf("got len = 0, want %s", "> 0")
	}
}

func TestServerGetAllScores(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", "/scores", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllScores(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp == 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetCoursesByProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", fmt.Sprintf("/courses/%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/courses/{uuid}", GetCoursesByProfessor)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetProfessorsByCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", fmt.Sprintf("/professors/%s", courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/professors/{code}", GetProfessorsByCourse)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/prof/%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/prof/{uuid}", GetScoresByProfessor)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	lresp := len(resp.Message.([]interface{}))
	if lresp != 1 {
		t.Errorf("got %d, want %d", lresp, 1)
	}
}

func TestServerGetScoresByCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", fmt.Sprintf("/scores/course/%s", courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/scores/course/{code}", GetScoresByCourse)
	router.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
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
	r, err := http.NewRequest("POST", fmt.Sprintf("/courses/grade?uuid=%s&code=%s&grade=3.5", professors[0].UUID, courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GradeCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("got %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != success.String() {
		t.Errorf("got %s, want %s", rr.Body.String(), success.String())
	}
}
