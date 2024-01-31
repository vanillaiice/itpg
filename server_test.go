package itpg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const ErrForeignKeyConstraint = "constraint failed: FOREIGN KEY constraint failed (787)"
const ErrNoRows = "sql: no rows in result set"

func dbInit() (err error) {
	db, err = newDB()
	if err != nil {
		return
	}
	return
}

func TestServerAddCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("PUT", "/course/add?code=GC8F&name=Showing%20your%20son%20whose%20the%20boss", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourse(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerAddCourse: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerAddProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatal(err.Error())
	}
	r, err := http.NewRequest("PUT", "/professors/add?fullname=Gintoki%20Sakata%20Senpai", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("TestServerAddProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerAddCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("PUT", fmt.Sprintf("/courses/addprof?uuid=%s&code=S209", professors[1].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerAddCourseProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerRemoveCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", "/courses/remove?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourse(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("TestServerRemoveCourse: got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != Err {
		t.Error("TestServerRemoveCourse: expected constraint failed error")
	}
	if resp.Message != ErrForeignKeyConstraint {
		fmt.Println(rr.Body)
		t.Error("TestServerRemoveCourse: expected constraint failed error")
	}
}

func TestServerRemoveCourseForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", "/courses/removeforce?code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveCourseForce: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerRemoveCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", "/courses/removeprof?id=1&code=CN9A", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveCourseProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerRemoveProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/remove?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessor(rr, r)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("TestServerRemoveProfessor: got %v, want %v", rr.Code, http.StatusInternalServerError)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != Err {
		t.Error("TestServerRemoveProfessor: expected constraint failed error")
	}
	if resp.Message != ErrForeignKeyConstraint {
		fmt.Println(rr.Body)
		t.Error("TestServerRemoveProfessor: expected constraint failed error")
	}
}

func TestServerRemoveProfessorForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", fmt.Sprintf("/professors/removeforce?uuid=%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessorForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveProfessorForce: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerGetAllCourses(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/courses", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllCourses(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetAllCourses: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestServerGetAllProfessors(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/professors", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllProfessors(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetAllProfessors: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGetAllScores(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/scores", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetAllScores(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetAllScores: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGetCoursesByProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", fmt.Sprintf("/courses/%s", professors[0].UUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetCoursesByProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestGetCoursesByProfessors: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGetProfessorsByCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/professors/S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetProfessorsByCourse(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestGetCoursesByProfessors: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGetScoresByProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/scores/prof/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetScoresByProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetScoresByProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGetScoresByCourse(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/scores/course/S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GetScoresByCourse(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetScoresByCourse: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerGradeCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("UPDATE", fmt.Sprintf("/courses/grade?uuid=%s&code=%s&grade=3.5", professors[0].UUID, courses[0].Code), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GradeCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetScoresByCourse: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status == Err {
		t.Error(resp.Message)
	}
}

func TestRun(t *testing.T) {
	//t.Fatal(Run("5555", ":memory:"))
}
