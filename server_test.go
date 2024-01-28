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
}

func TestServerAddProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("PUT", "/professors/add?surname=senpai&middlename=sakata&name=gintoki", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerAddProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestServerAddCourseProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("PUT", "/courses/addprof?id=1&code=S209", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	AddCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerAddCourseProfessor: got %v, want %v", rr.Code, http.StatusOK)
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
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveCourse: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Msg != ErrForeignKeyConstraint {
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
}

func TestServerRemoveProfessor(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", "/professors/remove?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveProfessor: got %v, want %v", rr.Code, http.StatusOK)
	}
	resp := &Response{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Msg != ErrForeignKeyConstraint {
		fmt.Println(rr.Body)
		t.Error("TestServerRemoveCourse: expected constraint failed error")
	}
}

func TestServerRemoveProfessorForce(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("DELETE", "/professors/removeforce?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	RemoveProfessorForce(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerRemoveProfessorForce: got %v, want %v", rr.Code, http.StatusOK)
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

func TestGetCoursesByProfessors(t *testing.T) {
	err := dbInit()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r, err := http.NewRequest("GET", "/courses/1", nil)
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
	r, err := http.NewRequest("PUT", "/courses/grade?id=1&code=CN9A&grade=3.5", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	GradeCourseProfessor(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("TestServerGetScoresByCourse: got %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestRun(t *testing.T) {
	//t.Fatal(Run(":5555", ":memory:"))
}
