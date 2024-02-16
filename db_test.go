package itpg

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
)

var courses = []*Course{
	{Code: "S209", Name: "How to replace head gaskets"},
	{Code: "CN9A", Name: "Controlling the Anti Lag System"},
	{Code: "AE86", Name: "How to beat any car"},
}
var professors = []*Professor{}
var scores = []*Score{}

var professorNames = []string{
	"Great Teacher Onizuka",
	"Pippy Peepee Poopypants",
	"Professor Oak",
}

func initDb(path ...string) (*DB, error) {
	if len(path) == 0 {
		path = append(path, ":memory:")
	}
	db, err := NewDB(path[0])
	if err != nil {
		return nil, err
	}

	for _, c := range courses {
		_, err := db.AddCourse(c.Code, c.Name)
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
	professors, err = db.GetAllProfessors()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(professors) && i < len(courses); i++ {
		_, err := db.AddCourseProfessor(professors[i].UUID, courses[i].Code)
		if err != nil {
			return nil, err
		}
	}
	scores, err = db.GetAllScores()
	if err != nil {
		return nil, err
	}

	return db, err
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestAddCourse(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.AddCourse("FC3S", "How to BRAPPPPPP")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddCourse("FC3S", "How to BRAPPPPPP")
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.AddCourse("FD3S", "")
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.AddCourse("", "How to BRAPPPPPP (second edition)")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.AddProfessor("Master Roshi")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddProfessor("")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddCourseProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.AddCourseProfessor(professors[1].UUID, "S209")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddCourseProfessor(professors[1].UUID, "S209")
	if err == nil {
		t.Error("expected failure")
	}
	UUID, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.AddCourseProfessor(UUID.String(), "GC8F")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestRemoveCourse(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.RemoveCourse("CN9A", false)
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.RemoveCourse("CN9A", true)
	if err != nil {
		t.Error(err)
	}
	n, err := db.RemoveCourse("GC8F", false)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("got %d, want %d", n, 0)
	}
}

func TestRemoveCourseProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.RemoveCourseProfessor(professors[1].UUID, "CN9A")
	if err != nil {
		t.Error(err)
	}
	n, err := db.RemoveCourseProfessor(professors[1].UUID, "CN9A")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("got %d, want %d", n, 0)
	}
}

func TestRemoveProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.RemoveProfessor(professors[0].UUID, false)
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.RemoveProfessor(professors[0].UUID, true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAllCourses(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allCourses, err := db.GetAllCourses()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses, courses) {
		t.Errorf("got %v, want %v", allCourses, courses)
	}
}

func TestGetAllProfessors(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allProfessors, err := db.GetAllProfessors()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors, professors) {
		t.Errorf("got %v, want %v", allProfessors, professors)
	}
}

func TestGetAllScores(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetAllScores()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores, scores) {
		t.Errorf("got %v, want %v", allScores, scores)
	}
}

func TestGetCoursesByProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allCourses, err := db.GetCoursesByProfessor(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses[0], courses[0]) {
		t.Errorf("got %v, want %v", allCourses, courses[1])
	}
}

func TestGetProfessorsByCourse(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allProfessors, err := db.GetProfessorsByCourse("CN9A")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors[0], professors[1]) {
		t.Errorf("got %v, want %v", allProfessors[0], professors[0])
	}
}

func TestGetScoresByProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByProfessor(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourse(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourse("S209")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetLastScore(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.GradeCourseProfessor(professors[1].UUID, "CN9A", 5.00)
	if err != nil {
		t.Error(err)
	}
	score, err := getLastScore(db.db, professors[1].UUID, "CN9A")
	if err != nil {
		t.Error(err)
	}
	if score != 5.00 {
		t.Errorf("got %.2f, want %.2f", score, 5.00)
	}
}

func TestGradeCourseProfessor(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	n, err := db.GradeCourseProfessor(professors[1].UUID, "CN9A", 4.2)
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Error("expected 1 row to be affected")
	}
	n, err = db.GradeCourseProfessor("1", "GC8F", 5)
	if err == nil {
		t.Error("expected failure")
	}
}
