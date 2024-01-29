package itpg

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// NullFloat64 represents a special value for float64 indicating null or undefined.
const NullFloat64 = -1

// DB is a struct representing the database and its methods.
type DB struct {
	db *sql.DB
}

// Course represents a course with its code and name.
type Course struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Professor represents a professor with surname, middle name, and name.
type Professor struct {
	Surname    string `json:"surname"`
	MiddleName string `json:"middleName"`
	Name       string `json:"name"`
}

// Score represents a score with professor ID, course code, and the score value.
type Score struct {
	ProfessorId int     `json:"professorId"`
	CourseCode  string  `json:"courseCode"`
	Score       float32 `json:"score"`
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.db.Close()
}

// NewDB initializes a new database connection and sets up the necessary tables if they don't exist.
func NewDB(path string) (*DB, error) {
	db := &DB{}
	sqldb, err := sql.Open("sqlite", path)
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
			return db, err
		}
	}
	return db, err
}

// AddCourse adds a new course to the database.
func (db *DB) AddCourse(code, name string) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Courses(code, name) VALUES(%q, %q)", code, name)
	_, err = execStmt(db.db, stmt)
	return
}

// AddProfessor adds a new professor to the database.
func (db *DB) AddProfessor(surname, middleName, name string) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Professors(surname, middlename, name) VALUES(%q, %q, %q)", surname, middleName, name)
	_, err = execStmt(db.db, stmt)
	return
}

// AddCourseProfessor adds a course to a professor in the database.
func (db *DB) AddCourseProfessor(professorId int, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Scores(professorid, coursecode) VALUES(%d, %q)", professorId, courseCode)
	n, err = execStmt(db.db, stmt)
	return
}

// RemoveCourse removes a course from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveCourse(code string, forceDelete bool) (n int64, err error) {
	stmt := []string{fmt.Sprintf("DELETE FROM Courses WHERE code = %q", code)}
	if forceDelete {
		stmt = append([]string{fmt.Sprintf("DELETE FROM Scores WHERE coursecode = %q", code)}, stmt...)
	}
	for _, s := range stmt {
		_, err = execStmt(db.db, s)
		if err != nil {
			return n, err
		}
	}
	return
}

// RemoveCourseProfessor removes a course from a professor in the database.
func (db *DB) RemoveCourseProfessor(professorId int, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("DELETE FROM Scores WHERE coursecode = %q AND professorid = %d", courseCode, professorId)
	n, err = execStmt(db.db, stmt)
	return
}

// RemoveProfessor removes a professor from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveProfessor(professorId int, forceDelete bool) (n int64, err error) {
	stmt := []string{fmt.Sprintf("DELETE FROM Professors WHERE id = %d", professorId)}
	if forceDelete {
		stmt = append([]string{fmt.Sprintf("DELETE FROM Scores WHERE professorid = %d", professorId)}, stmt...)
	}
	for _, s := range stmt {
		_, err = execStmt(db.db, s)
		if err != nil {
			return n, err
		}
	}
	return
}

// GetAllCourses retrieves all courses from the database.
func (db *DB) GetAllCourses() (courses []*Course, err error) {
	stmt := "SELECT * FROM Courses"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var code, name string
	for rows.Next() {
		err = rows.Scan(&code, &name)
		if err != nil {
			return nil, err
		}
		courses = append(courses, &Course{Code: code, Name: name})
	}
	return
}

// GetAllProfessors retrieves all professors from the database.
func (db *DB) GetAllProfessors() (professors []*Professor, err error) {
	stmt := "SELECT surname, middlename, name FROM Professors"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var sname, mname, name string
	for rows.Next() {
		err = rows.Scan(&sname, &mname, &name)
		if err != nil {
			return nil, err
		}
		professors = append(professors, &Professor{Surname: sname, MiddleName: mname, Name: name})
	}
	return
}

// GetAllScores retrieves all scores from the database.
func (db *DB) GetAllScores() (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT professorid, coursecode, IFNULL(score, %d) FROM Scores", NullFloat64)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorId int
	var courseCode string
	var score float64
	for rows.Next() {
		err = rows.Scan(&professorId, &courseCode, &score)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorId: professorId, CourseCode: courseCode, Score: float32(score)})
	}
	return
}

// GetCoursesByProfessor retrieves all courses associated with a professor from the database.
func (db *DB) GetCoursesByProfessor(professorId int) (courses []*Course, err error) {
	stmt := fmt.Sprintf("SELECT code, name FROM Courses JOIN Scores ON Courses.code = Scores.coursecode WHERE Scores.professorid = %d", professorId)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var code, name string
	for rows.Next() {
		err = rows.Scan(&code, &name)
		if err != nil {
			return nil, err
		}
		courses = append(courses, &Course{Code: code, Name: name})
	}
	return
}

// GetProfessorsByCourse retrieves all professors associated with a course from the database.
func (db *DB) GetProfessorsByCourse(courseCode string) (professors []*Professor, err error) {
	stmt := fmt.Sprintf("SELECT surname, middlename, name FROM Professors JOIN Scores ON Professors.id = Scores.professorid WHERE Scores.coursecode = %q", courseCode)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var sname, mname, name string
	for rows.Next() {
		err = rows.Scan(&sname, &mname, &name)
		if err != nil {
			return nil, err
		}
		professors = append(professors, &Professor{Surname: sname, MiddleName: mname, Name: name})
	}
	return
}

// GetScoresByProfessor retrieves all scores associated with a professor from the database.
func (db *DB) GetScoresByProfessor(professorId int) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT professorid, coursecode, IFNULL(score, %d) FROM Scores WHERE professorid = %d", NullFloat64, professorId)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var profId int
	var courseCode string
	var score float64
	for rows.Next() {
		err = rows.Scan(&profId, &courseCode, &score)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorId: profId, CourseCode: courseCode, Score: float32(score)})
	}
	return
}

// GetScoresByCourse retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourse(courseCode string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT professorid, coursecode, IFNULL(score, %d) FROM Scores WHERE coursecode = %q", NullFloat64, courseCode)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var profId int
	var cCode string
	var score float64
	for rows.Next() {
		err = rows.Scan(&profId, &cCode, &score)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorId: profId, CourseCode: cCode, Score: float32(score)})
	}
	return
}

// GradeCourseProfessor updates the score of a professor for a specific course in the database.
func (db *DB) GradeCourseProfessor(professorId int, courseCode string, grade float32) (n int64, err error) {
	lastGrade, err := getLastScore(db.db, professorId, courseCode)
	if err != nil {
		return
	}
	if lastGrade == NullFloat64 {
		lastGrade = grade
	} else {
		lastGrade = (lastGrade + grade) / 2
	}
	stmt := fmt.Sprintf("UPDATE scores SET Score = %0.2f WHERE professorid = %d AND coursecode = %q", lastGrade, professorId, courseCode)
	n, err = execStmt(db.db, stmt)
	return
}

// getLastScore retrieves the last score of a professor for a specific course from the database.
func getLastScore(db *sql.DB, professorId int, courseCode string) (score float32, err error) {
	var s sql.NullFloat64
	stmt := fmt.Sprintf("SELECT score FROM Scores WHERE professorid = %d AND coursecode = %q", professorId, courseCode)
	err = db.QueryRow(stmt).Scan(&s)
	if err != nil {
		return
	}
	if s.Valid {
		return float32(s.Float64), nil
	}
	return NullFloat64, nil
}

// execStmt executes a SQL statement and returns the number of affected rows.
func execStmt(db *sql.DB, stmt string) (n int64, err error) {
	res, err := db.Exec(stmt)
	if err != nil {
		return
	}
	n, err = res.RowsAffected()
	if err != nil {
		return
	}
	return
}
