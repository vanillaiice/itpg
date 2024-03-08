package db

import (
	"database/sql"
	"fmt"
	"itpg/responses"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/zeebo/xxh3"
	_ "modernc.org/sqlite"
)

// NullFloat64 represents a special value for float64 indicating null or undefined.
const NullFloat64 = -1

// MaxRowReturn represents the maximum number of rows returned by a query
const MaxRowReturn = 100

// RoundPrecision is the number decimals to use when rounding
const RoundPrecision = 2

// Hasher hashes using the xxh3 algorithm
var Hasher = xxh3.New()

// DB is a struct representing the database and its methods.
type DB struct {
	db *sql.DB
}

// Course represents a course with its code and name.
type Course struct {
	Code string `json:"code"` // Code of the course
	Name string `json:"name"` // Name of the course
}

// Professor represents a professor with surname, middle name, and name.
type Professor struct {
	UUID string `json:"uuid"` // UUID of the professor
	Name string `json:"name"` // Name of the professor
}

// Score represents a score for a course and its professor
type Score struct {
	ProfessorUUID   string  `json:"profUUID"`        // UUID of the professor
	ProfessorName   string  `json:"profName"`        // Name of the professor
	CourseCode      string  `json:"courseCode"`      // Code of the course
	CourseName      string  `json:"courseName"`      // Name of the course
	ScoreTeaching   float32 `json:"scoreTeaching"`   // Score related to the Teaching style/method of the professor
	ScoreCourseWork float32 `json:"scoreCoursework"` // Score related to the homeworks, quizes, and exams given by the professor
	ScoreLearning   float32 `json:"scoreLearning"`   // Score related to the learning outcomes of the course
	ScoreAverage    float32 `json:"scoreAverage"`    // Average score of the teaching, coursework, and learning scores
	Count           int     `json:"count"`           // Numbero of students who graded this course
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.db.Close()
}

// NewDB initializes a new database connection and sets up the necessary tables if they don't exist.
func NewDB(path string, speed bool) (*DB, error) {
	var (
		err   error
		sqldb *sql.DB
	)

	if speed {
		// FIXME: using cache=shared makes tests fails for some reason
		sqldb, err = sql.Open("sqlite", fmt.Sprintf("file:%s?journal_mode=memory&sync_mode=off&mode=rwc", path))
	} else {
		sqldb, err = sql.Open("sqlite", path)
	}

	if err != nil {
		return nil, err
	}

	if err = sqldb.Ping(); err != nil {
		return nil, err
	}

	db := &DB{db: sqldb}

	stmt := []string{
		"PRAGMA foreign_keys = ON",
		"CREATE TABLE IF NOT EXISTS Courses(code TEXT PRIMARY KEY NOT NULL CHECK(code != ''), name TEXT NOT NULL CHECK(name != ''), inserted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
		"CREATE TABLE IF NOT EXISTS Professors(uuid TEXT(36) PRIMARY KEY NOT NULL, name TEXT NOT NULL CHECK(name != ''), inserted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
		"CREATE TABLE IF NOT EXISTS Scores(professor_uuid TEXT(36) NOT NULL, course_code TEXT NOT NULL, score_teaching REAL CHECK(score_teaching >= 0 AND score_teaching <= 5), score_coursework REAL CHECK(score_coursework >= 0 AND score_coursework <= 5), score_learning REAL CHECK(score_learning >= 0 AND score_learning <= 5), count INTEGER NOT NULL, inserted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, UNIQUE(professor_uuid, course_code), FOREIGN KEY(professor_uuid) REFERENCES Professors(uuid), FOREIGN KEY(course_code) REFERENCES Courses(code))",
		"CREATE TABLE IF NOT EXISTS GradeHashes(id INTEGER PRIMARY KEY NOT NULL, hash INTEGER NOT NULL)",
	}

	for _, s := range stmt {
		if _, err := execStmt(sqldb, s); err != nil {
			return nil, err
		}
	}

	return db, err
}

// AddCourse adds a new course to the database.
func (db *DB) AddCourse(course *Course) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Courses(code, name) VALUES(%q, %q)", course.Code, course.Name)
	return execStmt(db.db, stmt)
}

// AddCourseMany adds new courses to the database.
func (db *DB) AddCourseMany(courses []*Course) (err error) {
	stmt, err := db.db.Prepare("INSERT INTO Courses(code, name) VALUES(?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, c := range courses {
		if _, err = stmt.Exec(c.Code, c.Name); err != nil {
			return
		}
	}

	return
}

// AddProfessor adds a new professor to the database.
func (db *DB) AddProfessor(name string) (n int64, err error) {
	professorUUID, err := uuid.NewV4()
	if err != nil {
		return
	}
	stmt := fmt.Sprintf("INSERT INTO Professors(uuid, name) VALUES(%q, %q)", professorUUID, name)
	return execStmt(db.db, stmt)
}

// AddProfessorMany adds new professors to the database.
func (db *DB) AddProfessorMany(names []string) (err error) {
	stmt, err := db.db.Prepare("INSERT INTO Professors(uuid, name) VALUES(?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, n := range names {
		professorUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}

		if _, err = stmt.Exec(professorUUID, n); err != nil {
			return err
		}
	}

	return
}

// AddCourseProfessor adds a course to a professor in the database.
func (db *DB) AddCourseProfessor(professorUUID, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("INSERT INTO Scores(professor_uuid, course_code, count) VALUES(%q, %q, 0)", professorUUID, courseCode)
	return execStmt(db.db, stmt)
}

// AddCourseProfessorMany adds courses to professors in the database.
func (db *DB) AddCourseProfessorMany(professorUUIDS, courseCodes []string) (err error) {
	if len(professorUUIDS) != len(courseCodes) {
		return fmt.Errorf("unequal slice length")
	}

	stmt, err := db.db.Prepare("INSERT INTO Scores(professor_uuid, course_code, count) VALUES(?, ?, 0)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for i := 0; i < len(professorUUIDS); i++ {
		if _, err = stmt.Exec(professorUUIDS[i], courseCodes[i]); err != nil {
			return err
		}
	}

	return
}

// RemoveCourse removes a course from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveCourse(code string, forceDelete bool) (n int64, err error) {
	stmt := []string{fmt.Sprintf("DELETE FROM Courses WHERE code = %q", code)}
	if forceDelete {
		stmt = append([]string{fmt.Sprintf("DELETE FROM Scores WHERE course_code = %q", code)}, stmt...)
	}

	for _, s := range stmt {
		if _, err = execStmt(db.db, s); err != nil {
			return n, err
		}
	}

	return
}

// RemoveCourseProfessor removes a course from a professor in the database.
func (db *DB) RemoveCourseProfessor(professorUUID, courseCode string) (n int64, err error) {
	stmt := fmt.Sprintf("DELETE FROM Scores WHERE course_code = %q AND professor_uuid = %q", courseCode, professorUUID)
	return execStmt(db.db, stmt)
}

// RemoveProfessor removes a professor from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveProfessor(professorUUID string, forceDelete bool) (n int64, err error) {
	stmt := []string{fmt.Sprintf("DELETE FROM Professors WHERE uuid = %q", professorUUID)}
	if forceDelete {
		stmt = append([]string{fmt.Sprintf("DELETE FROM Scores WHERE professor_uuid = %q", professorUUID)}, stmt...)
	}

	for _, s := range stmt {
		if _, err = execStmt(db.db, s); err != nil {
			return n, err
		}
	}

	return
}

// GetLastCourses retrieves the last 100 courses from the database.
func (db *DB) GetLastCourses() (courses []*Course, err error) {
	stmt := fmt.Sprintf("SELECT code, name FROM Courses ORDER BY inserted_at DESC LIMIT %d", MaxRowReturn)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return nil, err
		}
		courses = append(courses, &course)
	}

	return
}

// GetLastProfessors retrieves the last 100 professors from the database.
func (db *DB) GetLastProfessors() (professors []*Professor, err error) {
	stmt := fmt.Sprintf("SELECT uuid, name FROM Professors ORDER BY inserted_at DESC LIMIT %d", MaxRowReturn)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return nil, err
		}
		professors = append(professors, &professor)
	}

	return
}

// GetLastScores retrieves the last 100 scores from the database.
func (db *DB) GetLastScores() (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Scores.professor_uuid, Professors.name, Scores.course_code, Courses.name, IFNULL(score_teaching, %d), IFNULL(score_coursework, %d), IFNULL(score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code ORDER BY Scores.inserted_at DESC LIMIT %d", NullFloat64, NullFloat64, NullFloat64, MaxRowReturn)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.ProfessorUUID, &score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetCoursesByProfessor retrieves all courses associated with a professor from the database.
func (db *DB) GetCoursesByProfessorUUID(UUID string) (courses []*Course, err error) {
	stmt := fmt.Sprintf("SELECT code, name FROM Courses JOIN Scores ON Courses.code = Scores.course_code WHERE Scores.professor_uuid = %q", UUID)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return nil, err
		}
		courses = append(courses, &course)
	}

	return
}

// GetProfessorsByCourse retrieves all professors associated with a course from the database.
func (db *DB) GetProfessorsByCourseCode(code string) (professors []*Professor, err error) {
	stmt := fmt.Sprintf("SELECT uuid, name FROM Professors JOIN Scores ON Professors.uuid = Scores.professor_uuid WHERE Scores.course_code = %q", code)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return nil, err
		}
		professors = append(professors, &professor)
	}

	return
}

// GetProfessorUUIDByName retrieves the UUID of the professor that matches the specified name.
func (db *DB) GetProfessorUUIDByName(name string) (uuid string, err error) {
	stmt := fmt.Sprintf("SELECT uuid FROM Professors WHERE name = %q", name)
	row := db.db.QueryRow(stmt)
	if err = row.Scan(&uuid); err != nil {
		return
	}
	return
}

// GetScoresByProfessorUUID retrieves all scores associated with a professor's UUID from the database.
func (db *DB) GetScoresByProfessorUUID(UUID string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.name, Scores.course_code, Courses.name, IFNULL(Scores.score_teaching, %d), IFNULL(Scores.score_coursework, %d), IFNULL(Scores.score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code WHERE Scores.professor_uuid = %q", NullFloat64, NullFloat64, NullFloat64, UUID)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.ProfessorUUID = UUID
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByProfessorName retrieves all scores associated with a professor's name from the database.
func (db *DB) GetScoresByProfessorName(name string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Scores.course_code, Courses.name, Scores.professor_uuid, IFNULL(Scores.score_teaching, %d), IFNULL(Scores.score_coursework, %d), IFNULL(Scores.score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code WHERE Professors.name = %q", NullFloat64, NullFloat64, NullFloat64, name)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.ProfessorName = name
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByProfessorNameLike retrieves the last 100 scores for courses taught by professors whose names contain the given search string.
func (db *DB) GetScoresByProfessorNameLike(nameLike string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.name, Scores.course_code, Courses.name, Scores.professor_uuid, IFNULL(Scores.score_teaching, %d), IFNULL(Scores.score_coursework, %d), IFNULL(Scores.score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code WHERE Professors.name LIKE %q ORDER BY Scores.inserted_at LIMIT %d", NullFloat64, NullFloat64, NullFloat64, fmt.Sprintf("%%%s%%", nameLike), MaxRowReturn)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourse retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourseCode(code string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.name, Courses.name, Scores.professor_uuid, IFNULL(Scores.score_teaching, %d), IFNULL(Scores.score_coursework, %d), IFNULL(Scores.score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code WHERE Scores.course_code = %q ", NullFloat64, NullFloat64, NullFloat64, code)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.CourseCode = code
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseLike retrieves the last 100 scores associated with a course code from the database that matches the given search string
func (db *DB) GetScoresByCourseCodeLike(codeLike string) (scores []*Score, err error) {
	stmt := fmt.Sprintf("SELECT Professors.name, Scores.course_code, Courses.name, Scores.professor_uuid, IFNULL(Scores.score_teaching, %d), IFNULL(Scores.score_coursework, %d), IFNULL(Scores.score_learning, %d), Scores.count FROM Scores LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid LEFT JOIN Courses ON Scores.course_code = Courses.code WHERE Scores.course_code LIKE %q ORDER BY Scores.inserted_at LIMIT %d", NullFloat64, NullFloat64, NullFloat64, fmt.Sprintf("%%%s%%", codeLike), MaxRowReturn)

	rows, err := db.db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning, &score.Count); err != nil {
			return nil, err
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GradeCourseProfessor updates the scores of a professor for a specific course in the database.
func (db *DB) GradeCourseProfessor(professorUUID, courseCode, username string, grades [3]float32) (n int64, err error) {
	Hasher.Reset()
	Hasher.WriteString(username + courseCode + professorUUID)
	hash := Hasher.Sum64()

	if graded, err := db.checkGraded(hash); err != nil {
		return n, err
	} else {
		if graded {
			return n, responses.ErrCourseGraded
		}
	}

	oldScores, err := lastScores(db.db, professorUUID, courseCode)
	if err != nil {
		return
	}
	newScores := [3]float32{}

	for i, s := range oldScores {
		if s == NullFloat64 {
			newScores[i] = grades[i]
		} else {
			newScores[i] = (grades[i] + s) / 2
		}
	}

	stmt := fmt.Sprintf("UPDATE Scores SET score_teaching = %0.2f, score_coursework = %0.2f, score_learning = %0.2f, count = count + 1 WHERE professor_uuid = %q AND course_code = %q; INSERT INTO GradeHashes(hash) VALUES(%d)", newScores[0], newScores[1], newScores[2], professorUUID, courseCode, hash)

	return execStmt(db.db, stmt)
}

// CheckGraded checks if a user graded a course.
// The hash parameter is obtained by hashing
// the concatenation of the username, course code,
// and professor uuid using the xxh3 algorithm.
func (db *DB) checkGraded(hash uint64) (graded bool, err error) {
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM GradeHashes WHERE hash = %d", hash)

	var count int
	if err = db.db.QueryRow(stmt).Scan(&count); err != nil {
		return
	}

	if count > 0 {
		return !graded, nil
	} else {
		return graded, nil
	}
}

// lastScores retrieves the last scores of a professor for a specific course from the database.
func lastScores(db *sql.DB, professorUUID, courseCode string) (scores [3]float32, err error) {
	stmt := fmt.Sprintf("SELECT score_teaching, score_coursework, score_learning FROM Scores WHERE professor_uuid = %q AND course_code = %q", professorUUID, courseCode)

	var ss [3]sql.NullFloat64
	if err = db.QueryRow(stmt).Scan(&ss[0], &ss[1], &ss[2]); err != nil {
		return
	}

	for i, s := range ss {
		if !s.Valid {
			scores[i] = NullFloat64
		} else {
			scores[i] = float32(s.Float64)
		}
	}

	return scores, nil
}

// averageScore calculates the average score from a slice of flaots.
func averageScore(scores ...float32) float32 {
	sum := float32(0)
	for _, s := range scores {
		if s != NullFloat64 {
			sum += s
		} else {
			return NullFloat64
		}
	}
	avgScore := sum / float32(len(scores))
	return float32(decimal.NewFromFloat32(avgScore).Round(RoundPrecision).InexactFloat64())
}

// execStmt executes a SQL statement and returns the number of affected rows.
func execStmt(db *sql.DB, stmt string) (n int64, err error) {
	res, err := db.Exec(stmt)
	if err != nil {
		return
	}
	return res.RowsAffected()
}
