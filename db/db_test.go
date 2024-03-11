package db

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/zeebo/xxh3"
)

var professorNames = []string{
	"Great Teacher Onizuka",
	"Pippy Peepee Poopypants",
	"Professor Oak",
	"Takahashi Keisuke",
}

var courses = []*Course{
	{Code: "S209", Name: "How to replace head gaskets"},
	{Code: "CN9A", Name: "Controlling the Anti Lag System"},
	{Code: "AE86", Name: "How to beat any car"},
	{Code: "FD3S", Name: "How to BRAAAP"},
}
var professors = []*Professor{}
var scores = []*Score{}

func initDB(path ...string) (*DB, error) {
	if len(path) == 0 {
		path = append(path, ":memory:")
	}
	db, err := NewDB(path[0], true)
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

	for i := 0; i < len(professors); i++ {
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

func TestNewDB(t *testing.T) {
	db, err := NewDB(":memory:", true)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestAddCourse(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.AddCourse(&Course{"FC3S", "How to BRAPPPPPP"})
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddCourse(&Course{"FC3S", "How to BRAPPPPPP"})
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.AddCourse(&Course{"FD3S", ""})
	if err == nil {
		t.Error("expected failure")
	}
	_, err = db.AddCourse(&Course{"", "How to BRAPPPPPP (second edition)"})
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddCourseMany(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cs := []*Course{
		{"FC3S", "How to BRAPPPPPP"},
		{"AP1", "One Hand Driving 101"},
		{"EK9", "Art of VTEC"},
	}

	err = db.AddCourseMany(cs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddProfessor(t *testing.T) {
	db, err := initDB()
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

func TestAddProfessorMany(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ps := []string{
		"foo",
		"bar",
		"baz",
	}

	err = db.AddProfessorMany(ps)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddCourseProfessor(t *testing.T) {
	db, err := initDB()
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

func TestAddCourseProfessorMany(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uuids, codes := []string{}, []string{}
	for i := len(professors) - 1; i >= 0; i-- {
		uuids = append(uuids, professors[i].UUID)
	}
	for _, c := range courses {
		codes = append(codes, c.Code)
	}

	err = db.AddCourseProfessorMany(uuids, codes)
	if err != nil {
		t.Error(err)
	}

	uuids = []string{}
	err = db.AddCourseProfessorMany(uuids, codes)
	if err == nil {
		t.Error("expected failure")
	}
}

func TestRemoveCourse(t *testing.T) {
	db, err := initDB()
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
	db, err := initDB()
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
	db, err := initDB()
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

func TestGetLastCourses(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allCourses, err := db.GetLastCourses()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses, courses) {
		t.Errorf("got %v, want %v", allCourses, courses)
	}
}

func TestGetLastProfessors(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allProfessors, err := db.GetLastProfessors()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors, professors) {
		t.Errorf("got %v, want %v", allProfessors, professors)
	}
}

func TestGetLastScores(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetLastScores()
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores, scores) {
		t.Errorf("got %v, want %v", allScores, scores)
	}
}

func TestGetCoursesByProfessorUUID(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allCourses, err := db.GetCoursesByProfessorUUID(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allCourses[0], courses[0]) {
		t.Errorf("got %v, want %v", allCourses, courses[1])
	}
}

func TestGetProfessorsByCourseCode(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allProfessors, err := db.GetProfessorsByCourseCode("CN9A")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allProfessors[0], professors[1]) {
		t.Errorf("got %v, want %v", allProfessors[0], professors[0])
	}
}

func TestGetProfessorUUIDByName(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.GetProfessorUUIDByName(professors[0].Name)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetScoresByProfessorUUID(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByProfessorUUID(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByProfessorName(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByProfessorName(professors[0].Name)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByProfessorNameLike(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByProfessorNameLike(professors[0].Name[:5])
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourseName(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourseName("How to replace head gaskets")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourseNameLike(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourseNameLike("How to rep")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourseCode(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourseCode("S209")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourseCodeLike(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	allScores, err := db.GetScoresByCourseCodeLike("S2")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGradeCourseProfessor(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	profScores := [3]float32{5.00, 4.00, 3.00}
	n, err := db.GradeCourseProfessor(professors[1].UUID, "CN9A", "joe", profScores)
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Error("expected 1 row to be affected")
	}

	_, err = db.GradeCourseProfessor(professors[1].UUID, "CN9A", "joe", profScores)
	if err == nil {
		t.Error("expected failure")
	}

	n, err = db.GradeCourseProfessor("1", "GC8F", "joe", profScores)
	if err == nil {
		t.Error("expected failure")
	}
}

func TestCheckGraded(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hasher := xxh3.New()
	hasher.WriteString("joe" + courses[0].Code + professors[0].UUID)
	hash := hasher.Sum64()

	graded, err := db.checkGraded(hash)
	if err != nil {
		t.Error(err)
	}
	if graded {
		t.Errorf("got %v, want %v", graded, false)
	}

	grades := [3]float32{5.00, 4.00, 3.00}
	_, err = db.GradeCourseProfessor(professors[0].UUID, courses[0].Code, "joe", grades)
	if err != nil {
		t.Fatal(err)
	}

	graded, err = db.checkGraded(hash)
	if err != nil {
		t.Error(err)
	}
	if !graded {
		t.Errorf("got %v, want %v", graded, true)
	}
}

func TestGetLastScore(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	profScores := [3]float32{5.00, 4.00, 3.00}
	_, err = db.GradeCourseProfessor(professors[1].UUID, "CN9A", "joe", profScores)
	if err != nil {
		t.Error(err)
	}
	lastScores, err := lastScores(db.db, professors[1].UUID, "CN9A")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(profScores, lastScores) {
		t.Errorf("got %.2f, want %.2f", lastScores, profScores)
	}
}

func TestAverageScore(t *testing.T) {
	scores := []float32{5, 4, 3}
	avgScore := averageScore(scores...)
	expected := float32((5 + 4 + 3) / 3)
	if avgScore != float32(expected) {
		t.Errorf("got %f, want %f", avgScore, expected)
	}
	scores = []float32{5, NullFloat64, 3}
	expected = NullFloat64
	avgScore = averageScore(scores...)
	if avgScore != NullFloat64 {
		t.Errorf("got %f, want %f", avgScore, expected)
	}
}

func TestExecStmt(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = execStmt(db.db, "SELECT * FROM Courses")
	if err != nil {
		t.Error(err)
	}
}
