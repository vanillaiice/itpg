package postgres

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	itpgDB "github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/responses"
	"github.com/zeebo/xxh3"
)

// MaxRowReturn represents the maximum number of rows returned by a query
const MaxRowReturn = 100

// RoundPrecision is the number decimals to use when rounding
const RoundPrecision = 2

// DefaultHash is the hash value used when adding course to a professor
const DefaultHash = ""

// DB is a struct contaning a SQL database connection
type DB struct {
	conn *pgx.Conn
	ctx  context.Context
}

// NewDB initializes a new database connection and sets up the necessary tables if they don't exist.
func New(url string) (db *DB, err error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, err
	}

	stmt := `
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
			id SERIAL PRIMARY KEY,
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

	if err := execStmt(ctx, conn, stmt); err != nil {
		return nil, err
	}

	db = &DB{conn: conn, ctx: ctx}

	return
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close(db.ctx)
}

// AddCourse adds a new course to the database.
func (db *DB) AddCourse(course *itpgDB.Course) (err error) {
	stmt := "INSERT INTO Courses(code, name) VALUES($1, $2)"
	return execStmt(db.ctx, db.conn, stmt, course.Code, course.Name)
}

// AddCourseMany adds new courses to the database.
func (db *DB) AddCourseMany(courses []*itpgDB.Course) (err error) {
	stmt, err := db.conn.Prepare(db.ctx, "add_course_many", "INSERT INTO Courses(code, name) VALUES($1, $2)")
	if err != nil {
		return
	}

	for _, c := range courses {
		if _, err = db.conn.Exec(db.ctx, stmt.Name, c.Code, c.Name); err != nil {
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
	stmt := "INSERT INTO Professors(uuid, name) VALUES($1, $2)"
	return execStmt(db.ctx, db.conn, stmt, professorUUID, name)
}

// AddProfessorMany adds new professors to the database.
func (db *DB) AddProfessorMany(names []string) (err error) {
	stmt, err := db.conn.Prepare(db.ctx, "add_professor_many", "INSERT INTO Professors(uuid, name) VALUES($1, $2)")
	if err != nil {
		return
	}

	for _, n := range names {
		professorUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}

		if _, err = db.conn.Exec(db.ctx, stmt.Name, professorUUID, n); err != nil {
			return err
		}
	}

	return
}

// AddCourseProfessor adds a course to a professor in the database.
func (db *DB) AddCourseProfessor(professorUUID, courseCode string) (err error) {
	stmt := "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES($1, $2, $3)"
	return execStmt(db.ctx, db.conn, stmt, DefaultHash, professorUUID, courseCode)
}

// AddCourseProfessorMany adds courses to professors in the database.
func (db *DB) AddCourseProfessorMany(professorUUIDS, courseCodes []string) (err error) {
	if len(professorUUIDS) != len(courseCodes) {
		return fmt.Errorf("unequal slice length")
	}

	stmt, err := db.conn.Prepare(db.ctx, "add_course_professor_many", "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES($1, $2, $3)")
	if err != nil {
		return
	}

	for i := 0; i < len(professorUUIDS); i++ {
		if _, err = db.conn.Exec(db.ctx, stmt.Name, DefaultHash, professorUUIDS[i], courseCodes[i]); err != nil {
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
		{s: "DELETE FROM Scores WHERE course_code = $1", args: code, skip: !forceDelete},
		{s: "DELETE FROM Courses WHERE code = $1", args: code, skip: false},
	}

	for _, s := range stmt {
		if s.skip {
			continue
		}

		if err = execStmt(db.ctx, db.conn, s.s, s.args); err != nil {
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
		{s: "DELETE FROM Scores WHERE professor_uuid = $1", args: professorUUID, skip: !forceDelete},
		{s: "DELETE FROM Professors WHERE uuid = $1", args: professorUUID, skip: false},
	}

	for _, s := range stmt {
		if s.skip {
			continue
		}

		if err = execStmt(db.ctx, db.conn, s.s, s.args); err != nil {
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
		LIMIT $1
	`

	rows, err := db.conn.Query(db.ctx, stmt, MaxRowReturn)
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
		LIMIT $1
	`

	rows, err := db.conn.Query(db.ctx, stmt, MaxRowReturn)
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
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Courses.name, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
		LIMIT $1
	`

	rows, err := db.conn.Query(db.ctx, stmt, MaxRowReturn)
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
		WHERE Scores.professor_uuid = $1
		ORDER BY Courses.inserted_at
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, UUID)
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
		WHERE Scores.course_code = $1
		ORDER BY Professors.inserted_at
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, code)
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
		WHERE name = $1
	`

	row := db.conn.QueryRow(db.ctx, stmt, name)
	if err = row.Scan(&uuid); err != nil {
		return
	}
	return
}

// GetScoresByProfessorUUID retrieves all scores associated with a professor's UUID from the database.
func (db *DB) GetScoresByProfessorUUID(UUID string) (scores []*itpgDB.Score, err error) {
	stmt := `
		SELECT 
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Courses.name, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE
			Scores.professor_uuid = $1
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, UUID)
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
			STRING_AGG(DISTINCT Courses.name, ', '),
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code 
		WHERE Professors.name = $1
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, name)
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
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Courses.name, ', '),
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Professors.name
		LIKE @name_like
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
		LIMIT @max_row_return
	`

	args := pgx.NamedArgs{
		"name_like":      fmt.Sprintf("%%%s%%", nameLike),
		"max_row_return": MaxRowReturn,
	}

	rows, err := db.conn.Query(db.ctx, stmt, args)
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
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Courses.name = $1
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, name)
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
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Courses.name, ', '),
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Courses.name
		LIKE @name_like
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
		LIMIT @max_row_return
	`

	args := pgx.NamedArgs{
		"name_like":      fmt.Sprintf("%%%s%%", nameLike),
		"max_row_return": MaxRowReturn,
	}

	rows, err := db.conn.Query(db.ctx, stmt, args)
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
			STRING_AGG(DISTINCT Professors.name, ', '),
			STRING_AGG(DISTINCT Courses.name, ', '),
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Scores.course_code = $1
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
	`

	rows, err := db.conn.Query(db.ctx, stmt, code)
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
			STRING_AGG(DISTINCT Professors.name, ', '),
			Scores.course_code,
			STRING_AGG(DISTINCT Courses.name, ', '),
			STRING_AGG(DISTINCT Scores.professor_uuid, ', '),
			COALESCE(AVG(Scores.score_teaching), 0),
			COALESCE(AVG(Scores.score_coursework), 0),
			COALESCE(AVG(Scores.score_learning), 0)
		FROM
			Scores
			LEFT JOIN Professors ON Scores.professor_uuid = Professors.uuid
			LEFT JOIN Courses ON Scores.course_code = Courses.code
		WHERE Scores.course_code
		LIKE @code_like
		GROUP BY Scores.course_code
		ORDER BY MAX(Scores.inserted_at)
		DESC
		LIMIT @max_row_return
	`

	args := pgx.NamedArgs{
		"code_like":      fmt.Sprintf("%%%s%%", codeLike),
		"max_row_return": MaxRowReturn,
	}

	rows, err := db.conn.Query(db.ctx, stmt, args)
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
			score_learning
		)
		VALUES (
			@hash,
			@professor_uuid,
			@course_code,
			@score_teaching,
			@score_coursework,
			@score_learning
		)
	`

	args := pgx.NamedArgs{
		"professor_uuid":   professorUUID,
		"hash":             fmt.Sprintf("%d", hash),
		"course_code":      courseCode,
		"score_teaching":   grades[0],
		"score_coursework": grades[1],
		"score_learning":   grades[2],
	}

	return execStmt(db.ctx, db.conn, stmt, args)
}

// CheckGraded checks if a user graded a course.
// The hash parameter is obtained by hashing
// the concatenation of the username, course code,
// and professor uuid using the xxh3 algorithm.
func (db *DB) checkGraded(hash uint64) (graded bool, err error) {
	var count int

	stmt := "SELECT COUNT(*) FROM Scores WHERE hash = $1"
	if err = db.conn.QueryRow(db.ctx, stmt, fmt.Sprintf("%d", hash)).Scan(&count); err != nil {
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

	avg := sum / float32(len(scores))

	return float32(decimal.NewFromFloat32(avg).Round(RoundPrecision).InexactFloat64())
}

// execStmt executes a SQL statement.
func execStmt(ctx context.Context, conn *pgx.Conn, stmt string, args ...any) (err error) {
	_, err = conn.Exec(ctx, stmt, args...)
	return
}
