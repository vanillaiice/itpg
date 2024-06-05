package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	itpgDB "github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/responses"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/zeebo/xxh3"
)

// maxRowReturn represents the maximum number of rows returned by a query
const maxRowReturn = 100

// roundPrecision is the number decimals to use when rounding
const roundPrecision = 2

// defaultHash is the hash value used when adding course to a professor
const defaultHash = ""

// DB is a struct contaning a SQL database connection
type DB struct {
	conn *sql.DB
	ctx  context.Context
}

// New initializes a new database connection and sets up the necessary tables if they don't exist.
func New(url string, ctx context.Context) (db *DB, err error) {
	var conn *sql.DB

	conn, err = sql.Open("sqlite", url)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	stmt := `
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS Courses(
			code TEXT PRIMARY KEY NOT NULL
			CHECK(code <> ''),
			name TEXT NOT NULL
			CHECK(name <> ''),
			inserted_at TIMESTAMP
			DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(code, name)
		);

		CREATE TABLE IF NOT EXISTS Professors(
			uuid VARCHAR(36) PRIMARY KEY NOT NULL,
			name TEXT NOT NULL
			CHECK(name <> ''),
			inserted_at TIMESTAMP
			DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name)
		);

		CREATE TABLE IF NOT EXISTS Scores(
			id INTEGER PRIMARY KEY,
			hash TEXT NOT NULL,
			professor_uuid VARCHAR(36) NOT NULL,
			course_code TEXT NOT NULL,
			score_teaching REAL
			CHECK(score_teaching BETWEEN 0 AND 5),
			score_coursework REAL
			CHECK(score_coursework BETWEEN 0 AND 5),
			score_learning REAL
			CHECK(score_learning BETWEEN 0 AND 5),
			inserted_at TIMESTAMP
			DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(professor_uuid)
			REFERENCES Professors(uuid),
			FOREIGN KEY(course_code)
			REFERENCES Courses(code)
		);
	`

	if err := execStmtContext(conn, ctx, stmt); err != nil {
		return nil, err
	}

	db = &DB{conn: conn, ctx: ctx}

	return
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// AddCourse adds a new course to the database.
func (db *DB) AddCourse(course *itpgDB.Course) (err error) {
	stmt := "INSERT INTO Courses(code, name, inserted_at) VALUES(?, ?, ?)"
	return execStmtContext(db.conn, db.ctx, stmt, course.Code, course.Name, time.Now().UnixNano())
}

// AddCourseMany adds new courses to the database.
func (db *DB) AddCourseMany(courses []*itpgDB.Course) (err error) {
	stmt, err := db.conn.PrepareContext(db.ctx, "INSERT INTO Courses(code, name, inserted_at) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, c := range courses {
		if _, err = stmt.Exec(c.Code, c.Name, time.Now().UnixNano()); err != nil {
			return
		}
	}

	return
}

// AddProfessor adds a new professor to the database.
func (db *DB) AddProfessor(name string) (err error) {
	professorUUID, err := uuid.NewV4()
	if err != nil {
		return
	}
	stmt := "INSERT INTO Professors(uuid, name, inserted_at) VALUES(?, ?, ?)"
	return execStmtContext(db.conn, db.ctx, stmt, professorUUID, name, time.Now().UnixNano())
}

// AddProfessorMany adds new professors to the database.
func (db *DB) AddProfessorMany(names []string) (err error) {
	stmt, err := db.conn.PrepareContext(db.ctx, "INSERT INTO Professors(uuid, name, inserted_at) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, n := range names {
		professorUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}

		if _, err = stmt.Exec(professorUUID, n, time.Now().UnixNano()); err != nil {
			return err
		}
	}

	return
}

// AddCourseProfessor adds a course to a professor in the database.
func (db *DB) AddCourseProfessor(professorUUID, courseCode string) (err error) {
	stmt := "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES(?, ?, ?)"
	return execStmtContext(db.conn, db.ctx, stmt, defaultHash, professorUUID, courseCode)
}

// AddCourseProfessorMany adds courses to professors in the database.
func (db *DB) AddCourseProfessorMany(professorUUIDS, courseCodes []string) (err error) {
	if len(professorUUIDS) != len(courseCodes) {
		return fmt.Errorf("unequal slice length")
	}

	stmt, err := db.conn.PrepareContext(db.ctx, "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for i := 0; i < len(professorUUIDS); i++ {
		if _, err = stmt.Exec(defaultHash, professorUUIDS[i], courseCodes[i]); err != nil {
			return err
		}
	}

	return
}

// RemoveCourse removes a course from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveCourse(code string, forceDelete bool) (err error) {
	stmt := []struct {
		s    string
		args string
		skip bool
	}{
		{s: "DELETE FROM Scores WHERE course_code = ?", args: code, skip: !forceDelete},
		{s: "DELETE FROM Courses WHERE code = ?", args: code, skip: false},
	}

	for _, s := range stmt {
		if s.skip {
			continue
		}

		if err = execStmtContext(db.conn, db.ctx, s.s, s.args); err != nil {
			return
		}
	}

	return
}

// RemoveProfessor removes a professor from the database. If forceDelete is true, associated scores are also deleted.
func (db *DB) RemoveProfessor(professorUUID string, forceDelete bool) (err error) {
	stmt := []struct {
		s    string
		args string
		skip bool
	}{
		{s: "DELETE FROM Scores WHERE professor_uuid = ?", args: professorUUID, skip: !forceDelete},
		{s: "DELETE FROM Professors WHERE uuid = ?", args: professorUUID, skip: false},
	}

	for _, s := range stmt {
		if s.skip {
			continue
		}

		if err = execStmtContext(db.conn, db.ctx, s.s, s.args); err != nil {
			return
		}
	}

	return
}

// GetLastCourses retrieves the last 100 courses from the database.
func (db *DB) GetLastCourses() (courses []*itpgDB.Course, err error) {
	stmt := `
		SELECT code, name
		FROM Courses
		ORDER BY inserted_at
		DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := itpgDB.Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return
		}
		courses = append(courses, &course)
	}

	return
}

// GetLastProfessors retrieves the last 100 professors from the database.
func (db *DB) GetLastProfessors() (professors []*itpgDB.Professor, err error) {
	stmt := `
		SELECT uuid, name
		FROM Professors
		ORDER BY inserted_at
		DESC 
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := itpgDB.Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return
		}
		professors = append(professors, &professor)
	}

	return
}

// GetLastScores retrieves the last 100 scores from the database.
func (db *DB) GetLastScores() (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Scores.professor_uuid,
			Professors.name,
			Scores.course_code,
			Courses.name,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorUUID, &score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetCoursesByProfessor retrieves all courses associated with a professor from the database.
func (db *DB) GetCoursesByProfessorUUID(UUID string) (courses []*itpgDB.Course, err error) {
	stmt := `
		SELECT code, name
		FROM Courses
		JOIN Scores ON Courses.code = Scores.course_code
		WHERE Scores.professor_uuid = ?
		ORDER BY Courses.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, UUID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := itpgDB.Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return
		}
		courses = append(courses, &course)
	}

	return
}

// GetProfessorsByCourse retrieves all professors associated with a course from the database.
func (db *DB) GetProfessorsByCourseCode(code string) (professors []*itpgDB.Professor, err error) {
	stmt := `
		SELECT uuid, name
		FROM Professors
		JOIN Scores ON Professors.uuid = Scores.professor_uuid
		WHERE Scores.course_code = ?
		ORDER BY Professors.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, code)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := itpgDB.Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return
		}
		professors = append(professors, &professor)
	}

	return
}

// GetProfessorUUIDByName retrieves the UUID of the professor that matches the specified name.
func (db *DB) GetProfessorUUIDByName(name string) (uuid string, err error) {
	stmt := `
		SELECT uuid
		FROM Professors
		WHERE name = ?
	`

	row := db.conn.QueryRowContext(db.ctx, stmt, name)
	if err = row.Scan(&uuid); err != nil {
		return
	}
	return
}

// GetScoresByProfessorUUID retrieves all scores associated with a professor's UUID from the database.
func (db *DB) GetScoresByProfessorUUID(UUID string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Scores.course_code,
			Courses.name,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE
			Scores.professor_uuid = ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, UUID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ProfessorUUID = UUID
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByProfessorName retrieves all scores associated with a professor's name from the database.
func (db *DB) GetScoresByProfessorName(name string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Scores.course_code,
			Courses.name,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code 
		WHERE Professors.name = ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ProfessorName = name
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByProfessorNameLike retrieves the last 100 scores for courses taught by professors whose names contain the given search string.
func (db *DB) GetScoresByProfessorNameLike(nameLike string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Scores.course_code,
			Courses.name,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Professors.name
		LIKE ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, fmt.Sprintf("%%%s%%", nameLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseName retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourseName(name string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Scores.course_code,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Courses.name = ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.CourseName = name
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseNameLike retrieves the last 100 scores associated with a course code from the database that matches the given search string
func (db *DB) GetScoresByCourseNameLike(nameLike string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Scores.course_code,
			Courses.name,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Courses.name
		LIKE ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, fmt.Sprintf("%%%s%%", nameLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseCode retrieves all scores associated with a course from the database.
func (db *DB) GetScoresByCourseCode(code string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Courses.name,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Scores.course_code = ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, code)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.CourseCode = code
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseCodeLike retrieves the last 100 scores associated with a course code from the database that matches the given search string
func (db *DB) GetScoresByCourseCodeLike(codeLike string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			Professors.name,
			Scores.course_code,
			Courses.name,
			Scores.professor_uuid,
			IFNULL(AVG(Scores.score_teaching), 0),
			IFNULL(AVG(Scores.score_coursework), 0),
			IFNULL(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Scores.course_code
		LIKE ?
		GROUP BY Scores.course_code, Scores.professor_uuid
		ORDER BY Scores.inserted_at
		DESC
		LIMIT ?
	`

	rows, err := db.conn.QueryContext(db.ctx, stmt, fmt.Sprintf("%%%s%%", codeLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := itpgDB.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GradeCourseProfessor updates the scores of a professor for a specific course in the database.
func (db *DB) GradeCourseProfessor(professorUUID, courseCode, username string, grades [3]float32) (err error) {
	var Hasher = xxh3.New()
	Hasher.WriteString(username + courseCode + professorUUID)
	hash := Hasher.Sum64()

	if graded, err := db.checkGraded(hash); err != nil {
		return err
	} else {
		if graded {
			return responses.ErrCourseGraded
		}
	}

	stmt := `
		INSERT INTO Scores (
			hash,
			professor_uuid,
			course_code,
			score_teaching,
			score_coursework,
			score_learning,
			inserted_at
		) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	return execStmtContext(db.conn, db.ctx, stmt, fmt.Sprintf("%d", hash), professorUUID, courseCode, grades[0], grades[1], grades[2], time.Now().UnixNano())
}

// CheckGraded checks if a user graded a course.
// The hash parameter is obtained by hashing
// the concatenation of the username, course code,
// and professor uuid using the xxh3 algorithm.
func (db *DB) checkGraded(hash uint64) (graded bool, err error) {
	var count int

	stmt := "SELECT COUNT(*) FROM Scores WHERE hash = ?"
	if err = db.conn.QueryRowContext(db.ctx, stmt, fmt.Sprintf("%d", hash)).Scan(&count); err != nil {
		return
	}

	if count > 0 {
		return !graded, nil
	} else {
		return graded, nil
	}
}

// averageScore calculates the average score from a slice of floats.
func averageScore(scores ...float32) float32 {
	var sum float32
	for _, s := range scores {
		sum += s
	}

	avgScore := sum / float32(len(scores))

	return float32(decimal.NewFromFloat32(avgScore).Round(roundPrecision).InexactFloat64())
}

// execStmtContext executes a SQL statement.
func execStmtContext(conn *sql.DB, ctx context.Context, stmt string, args ...any) (err error) {
	_, err = conn.ExecContext(ctx, stmt, args...)
	return
}
