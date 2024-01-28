package itpg

import (
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "modernc.org/sqlite"
)

var courses = []*Course{
	{Code: "S209", Name: "How to replace head gaskets"},
	{Code: "CN9A", Name: "Controlling the Anti Lag System"},
	{Code: "AE86", Name: "How to beat any car"},
}

var professors = []*Professor{
	{Surname: "great", MiddleName: "teacher", Name: "onizuka"},
	{Surname: "professor", Name: "oak"},
	{Surname: "pippy", MiddleName: "peepee", Name: "poopypants"},
}

var scores = []*Score{
	{ProfessorId: 1, CourseCode: "CN9A", Score: NullFloat64},
	{ProfessorId: 2, CourseCode: "AE86", Score: NullFloat64},
	{ProfessorId: 3, CourseCode: "S209", Score: NullFloat64},
}

func newDB(path ...string) (*DB, error) {
	if len(path) == 0 {
		path = append(path, ":memory:")
	}
	db := &DB{}
	sqldb, err := sql.Open("sqlite", path[0])
	if err != nil {
		return nil, err
	}
	db.db = sqldb

	stmt := []string{
		"PRAGMA foreign_keys = ON",
		"CREATE TABLE IF NOT EXISTS Courses(code TEXT PRIMARY KEY NOT NULL CHECK(code != ''), name TEXT NOT NULL CHECK(name != ''))",
		"CREATE TABLE IF NOT EXISTS Professors(id INTEGER PRIMARY KEY NOT NULL, surname TEXT NOT NULL CHECK(surname != ''), middlename TEXT NOT NULL, name TEXT NOT NULL CHECK(name != ''), UNIQUE(surname, middlename, name))",
		"CREATE TABLE IF NOT EXISTS Scores(professorid INTEGER NOT NULL, coursecode TEXT NOT NULL, score REAL CHECK(score >= 0 AND score <= 5), FOREIGN KEY(professorid) REFERENCES Professors(id), FOREIGN KEY(coursecode) REFERENCES Courses(code))",
	}
	for _, s := range stmt {
		_, err := execStmt(db.db, s)
		if err != nil {
			return nil, err
		}
	}

	for _, c := range courses {
		_, err := db.AddCourse(c.Code, c.Name)
		if err != nil {
			return nil, err
		}
	}
	for _, p := range professors {
		_, err := db.AddProfessor(p.Surname, p.MiddleName, p.Name)
		if err != nil {
			return nil, err
		}
	}
	for _, s := range scores {
		_, err := db.AddCourseProfessor(s.ProfessorId, s.CourseCode)
		if err != nil {
			return nil, err
		}
	}

	return db, err
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf(err.Error())
	}
	db.Close()
}

func TestAddCourse(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	_, err = db.AddCourse("FC3S", "How to BRAPPPPPP")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddCourse("FC3S", "How to BRAPPPPPP")
	if err == nil {
		t.Error("expected failure for TestAddCourse")
	}
	_, err = db.AddCourse("FD3S", "")
	if err == nil {
		t.Error("expected failure for TestAddCourse")
	}
	_, err = db.AddCourse("", "How to BRAPPPPPP (second edition)")
	if err == nil {
		t.Error("expected failure for TestAddCourse")
	}
}

func TestAddProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	_, err = db.AddProfessor("master", "", "roshi")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddProfessor("", "", "roshi")
	if err == nil {
		t.Error("expected failure for AddProfessor")
	}
	_, err = db.AddProfessor("master", "", "")
	if err == nil {
		t.Error("expected failure for AddProfessor")
	}
	_, err = db.AddProfessor("", "roshi", "")
	if err == nil {
		t.Error("expected failure for AddProfessor")
	}
}

func TestAddCourseProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	_, err = db.AddCourseProfessor(1, "S209")
	if err != nil {
		t.Error(err)
	}
	/*
		_, err = db.AddCourseProfessor(1, "S209")
		if err == nil {
			t.Error("expected failure for AddCourseProfessor")
		}
	*/
	_, err = db.AddCourseProfessor(1, "GC8F")
	if err == nil {
		t.Error("expected failure for AddCourseProfessor")
	}
}

func TestRemoveCourse(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()

	_, err = db.RemoveCourse("CN9A", false)
	if err == nil {
		t.Error("expected failure for RemoveCourse")
	}
	_, err = db.RemoveCourse("CN9A", true)
	if err != nil {
		t.Error(err)
	}
	/*
		_, err = db.RemoveCourse("GC8F", false)
		if err == nil {
			t.Error("expected failure for RemoveCourse")
		}
	*/
}

func TestRemoveCourseProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	_, err = db.RemoveCourseProfessor(1, "CN9A")
	if err != nil {
		t.Error(err)
	}
	/*
		_, err = db.RemoveCourseProfessor(1, "CN9A")
		if err == nil {
			t.Error("expected failure for RemoveCourseProfessor")
		}
	*/
}

func TestRemoveProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = db.RemoveProfessor(1, false)
	if err == nil {
		t.Error("expected failure for RemoveProfessor")
	}
	_, err = db.RemoveProfessor(1, true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAllCourses(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allCourses, err := db.GetAllCourses()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses, courses) {
		t.Errorf("TestGetAllCourses: got %v, want %v", allCourses, courses)
	}
}

func TestGetAllProfessors(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allProfessors, err := db.GetAllProfessors()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors, professors) {
		t.Errorf("TestGetAllProfessors: got %v, want %v", allProfessors, professors)
	}
}

func TestGetAllScores(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allScores, err := db.GetAllScores()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores, scores) {
		t.Errorf("TestGetAllScores: got %v, want %v", allScores, scores)
	}
}

func TestGetCoursesByProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allCourses, err := db.GetCoursesByProfessor(1)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses[0], courses[1]) {
		t.Errorf("TestGetAllCourses: got %v, want %v", allCourses, courses[1])
	}
}

func TestGetProfessorsByCourse(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allProfessors, err := db.GetProfessorsByCourse("CN9A")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors[0], professors[0]) {
		t.Errorf("TestGetProfessorByCourse: got %v, want %v", allProfessors[0], professors[0])
	}
}

func TestGetScoresByProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allScores, err := db.GetScoresByProfessor(1)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("TestGetScoresByProfessor: got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourse(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourse("CN9A")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("TestGetScoresByProfessor: got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetLastScore(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	_, err = db.GradeCourseProfessor(1, "CN9A", 5.00)
	if err != nil {
		t.Error(err)
	}
	score, err := getLastScore(db.db, 1, "CN9A")
	if err != nil {
		t.Error(err)
	}
	if score != 5.00 {
		t.Errorf("TestGetLastScore: got %.2f, want %.2f", score, 5.00)
	}
}

func TestGradeCourseProfessor(t *testing.T) {
	db, err := newDB()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer db.Close()
	n, err := db.GradeCourseProfessor(1, "CN9A", 4.2)
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Error("expected 1 row to be affected")
	}
	/*
		n, err = db.GradeCourseProfessor(1, "GC8F", 5)
		if err == nil {
			t.Error("expected failure for TestGradeCourseProfessor")
		}
	*/
}
