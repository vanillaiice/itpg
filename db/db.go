package db

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
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
	UUID     string `json:"uuid"`
	FullName string `json:"name"`
}

// Score represents a score with professor ID, course code, and the score value.
type Score struct {
	ProfessorUUID string  `json:"profUUID"`
	ProfessorName string  `json:"profName"`
	CourseCode    string  `json:"courseCode"`
	CourseName    string  `json:"courseName"`
	Score         float32 `json:"score"`
	Count         int     `json:"count"`
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
		"CREATE TABLE IF NOT EXISTS Professors(uuid TEXT(36) PRIMARY KEY NOT NULL, fullname TEXT NOT NULL CHECK(fullname != ''))",
		"CREATE TABLE IF NOT EXISTS Scores(professoruuid TEXT(36) NOT NULL, coursecode TEXT NOT NULL, score REAL CHECK(score >= 0 AND score <= 5), count INTEGER NOT NULL, UNIQUE(professoruuid, coursecode), FOREIGN KEY(professoruuid) REFERENCES Professors(uuid), FOREIGN KEY(coursecode) REFERENCES Courses(code))",
	}
	for _, s := range stmt {
		_, err := execStmt(db.db, s)
		if err != nil {
			return nil, err
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
func (db *DB) AddProfessor(fullName string) (n int64, err error) {
	professorUUID, err := uuid.NewV4()
	if err != nil {
		return
	}
	stmt := fmt.Sprintf("INSERT INTO Professors(uuid, fullname) VALUES(%q, %q)", professorUUID, fullName)
	_, err = execStmt(db.db, stmt)
	return
}

// AddCourseProfessor adds a course to a professor in the database.
func (db *DB) AddCourseProfessor(professorUUID, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Scores(professoruuid, coursecode, count) VALUES(%q, %q, 0)", professorUUID, courseCode)
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
func (db *DB) RemoveCourseProfessor(professorUUID, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("DELETE FROM Scores WHERE coursecode = %q AND professoruuid = %q", courseCode, professorUUID)
	n, err = execStmt(db.db, stmt)
	return
}

// RemoveProfessor removes a professor from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveProfessor(professorUUID string, forceDelete bool) (n int64, err error) {
	stmt := []string{fmt.Sprintf("DELETE FROM Professors WHERE uuid = %q", professorUUID)}
	if forceDelete {
		stmt = append([]string{fmt.Sprintf("DELETE FROM Scores WHERE professoruuid = %q", professorUUID)}, stmt...)
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
	stmt := "SELECT * FROM Professors"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, fullName string
	for rows.Next() {
		err = rows.Scan(&professorUUID, &fullName)
		if err != nil {
			return nil, err
		}
		professors = append(professors, &Professor{UUID: professorUUID, FullName: fullName})
	}
	return
}

// GetAllScores retrieves all scores from the database.
func (db *DB) GetAllScores() (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Scores.professoruuid, Professors.fullName, Courses.name, Scores.coursecode, IFNULL(score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code", NullFloat64)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, professorFullName, courseCode, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&professorUUID, &professorFullName, &courseName, &courseCode, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorFullName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// GetCoursesByProfessor retrieves all courses associated with a professor from the database.
func (db *DB) GetCoursesByProfessorUUID(professorUUID string) (courses []*Course, err error) {
	stmt := fmt.Sprintf("SELECT code, name FROM Courses JOIN Scores ON Courses.code = Scores.coursecode WHERE Scores.professoruuid = %q", professorUUID)
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
func (db *DB) GetProfessorsByCourseCode(courseCode string) (professors []*Professor, err error) {
	stmt := fmt.Sprintf("SELECT uuid, fullname FROM Professors JOIN Scores ON Professors.uuid = Scores.professoruuid WHERE Scores.coursecode = %q", courseCode)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, fullName string
	for rows.Next() {
		err = rows.Scan(&professorUUID, &fullName)
		if err != nil {
			return nil, err
		}
		professors = append(professors, &Professor{UUID: professorUUID, FullName: fullName})
	}
	return
}

// GetScoresByProfessorUUID retrieves all scores associated with a professor's UUID from the database.
func (db *DB) GetScoresByProfessorUUID(professorUUID string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.fullname, Courses.name, Scores.coursecode, IFNULL(Scores.score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code WHERE Scores.professoruuid = %q", NullFloat64, professorUUID)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorFullName, courseCode, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&professorFullName, &courseName, &courseCode, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorFullName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// rename global vars

// GetScoresByProfessorName retrieves all scores associated with a professor's name from the database.
func (db *DB) GetScoresByProfessorName(professorName string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Courses.name, Scores.coursecode, Scores.professoruuid, IFNULL(Scores.score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code WHERE Professors.fullname LIKE %q", NullFloat64, fmt.Sprintf("%%%s%%", professorName))
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, courseCode, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&courseName, &courseCode, &professorUUID, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// GetScoresByProfessorNameLike retrieves scores for courses taught by professors whose names contain the given search string.
func (db *DB) GetScoresByProfessorNameLike(professorNameLike string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.fullname, Courses.name, Scores.coursecode, Scores.professoruuid, IFNULL(Scores.score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code WHERE Professors.fullname LIKE %q", NullFloat64, fmt.Sprintf("%%%s%%", professorNameLike))
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, professorName, courseCode, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&professorName, &courseName, &courseCode, &professorUUID, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// GetScoresByCourse retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourseCode(courseCode string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.fullname, Courses.name, Scores.professoruuid, IFNULL(Scores.score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code WHERE Scores.coursecode = %q", NullFloat64, courseCode)
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, professorFullName, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&professorFullName, &courseName, &professorUUID, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorFullName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// GetScoresByCourseLike retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourseCodeLike(courseCodeLike string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.fullname, Courses.name, Scores.coursecode, Scores.professoruuid, IFNULL(Scores.score, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professoruuid = Professors.uuid LEFT JOIN Courses ON Scores.coursecode = Courses.code WHERE Scores.coursecode LIKE %q", NullFloat64, fmt.Sprintf("%%%s%%", courseCodeLike))
	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()
	var professorUUID, professorFullName, courseCode, courseName string
	var score float64
	var count int
	for rows.Next() {
		err = rows.Scan(&professorFullName, &courseName, &courseCode, &professorUUID, &score, &count)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &Score{ProfessorUUID: professorUUID, ProfessorName: professorFullName, CourseCode: courseCode, CourseName: courseName, Score: float32(score), Count: count})
	}
	return
}

// GradeCourseProfessor updates the score of a professor for a specific course in the database.
func (db *DB) GradeCourseProfessor(professorUUID, courseCode string, grade float32) (n int64, err error) {
	lastGrade, err := getLastScore(db.db, professorUUID, courseCode)
	if err != nil {
		return
	}
	if lastGrade == NullFloat64 {
		lastGrade = grade
	} else {
		lastGrade = (lastGrade + grade) / 2
	}
	stmt := fmt.Sprintf("UPDATE Scores SET score = %0.2f, count = count + 1 WHERE professoruuid = %q AND coursecode = %q", lastGrade, professorUUID, courseCode)
	n, err = execStmt(db.db, stmt)
	return
}

// getLastScore retrieves the last score of a professor for a specific course from the database.
func getLastScore(db *sql.DB, professorUUID, courseCode string) (score float32, err error) {
	var s sql.NullFloat64
	stmt := fmt.Sprintf("SELECT score FROM Scores WHERE professoruuid = %q AND coursecode = %q", professorUUID, courseCode)
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
