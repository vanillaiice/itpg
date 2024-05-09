package postgres

import (
	"fmt"
	itpgDB "itpg/db"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/zeebo/xxh3"
)

var TestDB *DB

var TestDBUrl string

var professorNames = []string{
	"Great Teacher Onizuka",
	"Pippy Peepee Poopypants",
	"Professor Oak",
	"Takahashi Keisuke",
}

var courses = []*itpgDB.Course{
	{Code: "S209", Name: "How to replace head gaskets"},
	{Code: "CN9A", Name: "Controlling the Anti Lag System"},
	{Code: "AE86", Name: "How to beat any car"},
	{Code: "FD3S", Name: "How to BRAAAP"},
}

var professors = []*itpgDB.Professor{}

var scores = []*itpgDB.Score{}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatal(err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16.2-alpine3.19",
		Env: []string{
			"POSTGRES_PASSWORD=pazzword",
			"POSTGRES_USER=uzer",
			"POSTGRES_DB=db",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatal(err)
	}

	addr := resource.GetHostPort("5432/tcp")
	TestDBUrl = fmt.Sprintf("postgres://uzer:pazzword@%s/db?sslmode=disable", addr)

	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		TestDB, err = New(TestDBUrl, false)
		return err
	}); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func initDB() (err error) {
	err = execStmt(TestDB.ctx, TestDB.conn, "DROP TABLE IF EXISTS Courses, Professors, Scores")
	if err != nil {
		return
	}

	err = TestDB.Close()
	if err != nil {
		return
	}

	TestDB, err = New(TestDBUrl, true)
	if err != nil {
		return
	}

	for i := len(courses) - 1; i >= 0; i-- {
		err = TestDB.AddCourse(courses[i])
		if err != nil {
			return
		}
	}

	for _, p := range professorNames {
		err = TestDB.AddProfessor(p)
		if err != nil {
			return
		}
	}

	professors, err = TestDB.GetLastProfessors()
	if err != nil {
		return
	}

	for i, j := len(professors)-1, 0; i >= 0 && j < len(courses); i, j = i-1, j+1 {
		profScores := [3]float32{rand.Float32() * 5, rand.Float32() * 5, rand.Float32() * 5}
		err = TestDB.GradeCourseProfessor(professors[i].UUID, courses[j].Code, "jim", profScores)
		if err != nil {
			return
		}
	}

	scores, err = TestDB.GetLastScores()
	if err != nil {
		return
	}

	return
}

func TestNew(t *testing.T) {
	db, err := New(TestDBUrl, false)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestAddCourse(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	err = TestDB.AddCourse(&itpgDB.Course{"FC3S", "How to BRAPPPPPP"})
	if err != nil {
		t.Error(err)
	}

	err = TestDB.AddCourse(&itpgDB.Course{"FC3S", "How to BRAPPPPPP"})
	if err == nil {
		t.Error("expected failure")
	}

	err = TestDB.AddCourse(&itpgDB.Course{"FD3S", ""})
	if err == nil {
		t.Error("expected failure")
	}

	err = TestDB.AddCourse(&itpgDB.Course{"", "How to BRAPPPPPP (second edition)"})
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddCourseMany(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	cs := []*itpgDB.Course{
		{"FC3S", "How to BRAPPPPPP"},
		{"AP1", "One Hand Driving 101"},
		{"EK9", "Art of VTEC"},
	}

	err = TestDB.AddCourseMany(cs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddProfessor(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	err = TestDB.AddProfessor("Master Roshi")
	if err != nil {
		t.Error(err)
	}

	err = TestDB.AddProfessor("")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddProfessorMany(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	ps := []string{
		"foo",
		"bar",
		"baz",
	}

	err = TestDB.AddProfessorMany(ps)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveCourse(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	err = TestDB.RemoveCourse("CN9A", false)
	if err == nil {
		t.Error("expected failure")
	}

	err = TestDB.RemoveCourse("CN9A", true)
	if err != nil {
		t.Error(err)
	}

	err = TestDB.RemoveCourse("GC8F", false)
	if err != nil {
		t.Error(err)
	}
}

func TestRemoveProfessor(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	err = TestDB.RemoveProfessor(professors[0].UUID, false)
	if err == nil {
		t.Error("expected failure")
	}

	err = TestDB.RemoveProfessor(professors[0].UUID, true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLastCourses(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allCourses, err := TestDB.GetLastCourses()
	if err != nil {
		t.Error(err)
	}

	if len(allCourses) == 0 {
		t.Fatal("got 0 courses")
	}

	if len(allCourses) != len(courses) {
		t.Fatal("slices len unequal")
	}

	if !cmp.Equal(allCourses, courses) {
		t.Errorf("got %v, want %v", allCourses, courses)
	}
}

func TestGetLastProfessors(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allProfessors, err := TestDB.GetLastProfessors()
	if err != nil {
		t.Error(err)
	}

	if len(allProfessors) == 0 {
		t.Fatal("got 0 professors")
	}

	if len(allProfessors) != len(professors) {
		t.Fatal("slices len unequal")
	}

	if !cmp.Equal(allProfessors, professors) {
		t.Errorf("got %v, want %v", allProfessors, professors)
	}
}

func TestGetLastScores(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetLastScores()
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if len(allScores) != len(scores) {
		t.Fatal("slices len unequal")
	}

	if !cmp.Equal(allScores, scores) {
		t.Errorf("got %v, want %v", allScores, scores)
	}
}

func TestGetCoursesByProfessorUUID(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allCourses, err := TestDB.GetCoursesByProfessorUUID(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}

	if len(allCourses) == 0 {
		t.Fatal("got 0 courses")
	}

	if !cmp.Equal(allCourses[0], courses[len(courses)-1]) {
		t.Errorf("got %v, want %v", allCourses[0], courses[len(courses)-1])
	}
}

func TestGetProfessorsByCourseCode(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allProfessors, err := TestDB.GetProfessorsByCourseCode("S209")
	if err != nil {
		t.Error(err)
	}

	if len(allProfessors) == 0 {
		t.Fatal("got 0 professors")
	}

	if !cmp.Equal(allProfessors[0], professors[len(professors)-1]) {
		t.Errorf("got %v, want %v", allProfessors[0], professors[len(professors)-1])
	}
}

func TestGetProfessorUUIDByName(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	uuid, err := TestDB.GetProfessorUUIDByName(professors[0].Name)
	if err != nil {
		t.Fatal(err)
	}

	if uuid != professors[0].UUID {
		t.Errorf("got %s, want %s", uuid, professors[0].UUID)
	}
}

func TestGetScoresByProfessorUUID(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByProfessorUUID(professors[0].UUID)
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByProfessorName(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByProfessorName(professors[0].Name)
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByProfessorNameLike(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByProfessorNameLike(professors[0].Name[:5])
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[0]) {
		t.Errorf("got %v, want %v", allScores[0], scores[0])
	}
}

func TestGetScoresByCourseName(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByCourseName("How to replace head gaskets")
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[len(scores)-1]) {
		t.Errorf("got %v, want %v", allScores[0], scores[len(scores)-1])
	}
}

func TestGetScoresByCourseNameLike(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByCourseNameLike("How to rep")
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[len(scores)-1]) {
		t.Errorf("got %v, want %v", allScores[0], scores[len(scores)-1])
	}
}

func TestGetScoresByCourseCode(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByCourseCode("S209")
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[len(scores)-1]) {
		t.Errorf("got %v, want %v", allScores[0], scores[len(scores)-1])
	}
}

func TestGetScoresByCourseCodeLike(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	allScores, err := TestDB.GetScoresByCourseCodeLike("S2")
	if err != nil {
		t.Error(err)
	}

	if len(allScores) == 0 {
		t.Fatal("got 0 scores")
	}

	if !cmp.Equal(allScores[0], scores[len(scores)-1]) {
		t.Errorf("got %v, want %v", allScores[0], scores[len(scores)-1])
	}
}

func TestGradeCourseProfessor(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	profScores := [3]float32{5.00, 4.00, 3.00}
	err = TestDB.GradeCourseProfessor(professors[1].UUID, "CN9A", "joe", profScores)
	if err != nil {
		t.Error(err)
	}

	err = TestDB.GradeCourseProfessor(professors[1].UUID, "CN9A", "joe", profScores)
	if err == nil {
		t.Error("expected failure")
	}

	err = TestDB.GradeCourseProfessor("1", "GC8F", "joe", profScores)
	if err == nil {
		t.Error("expected failure")
	}
}

func TestCheckGraded(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	hasher := xxh3.New()
	hasher.WriteString("joe" + courses[0].Code + professors[0].UUID)
	hash := int64(hasher.Sum64())

	graded, err := TestDB.checkGraded(hash)
	if err != nil {
		t.Error(err)
	}

	if graded {
		t.Errorf("got %v, want %v", graded, false)
	}

	grades := [3]float32{5.00, 4.00, 3.00}
	err = TestDB.GradeCourseProfessor(professors[0].UUID, courses[0].Code, "joe", grades)
	if err != nil {
		t.Fatal(err)
	}

	graded, err = TestDB.checkGraded(hash)
	if err != nil {
		t.Error(err)
	}

	if !graded {
		t.Errorf("got %v, want %v", graded, true)
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
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	err = execStmt(TestDB.ctx, TestDB.conn, "SELECT * FROM Courses")
	if err != nil {
		t.Error(err)
	}
}
