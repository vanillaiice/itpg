package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/vanillaiice/itpg/db"
	"github.com/vanillaiice/itpg/db/cache"
	"github.com/vanillaiice/itpg/responses"
	"github.com/zeebo/xxh3"
	_ "modernc.org/sqlite"
)

// maxRowReturn represents the maximum number of rows returned by a query
const maxRowReturn = 100

// roundPrecision is the number decimals to use when rounding
const roundPrecision = 2

// defaultHash is the hash value used when adding course to a professor
const defaultHash = ""

// defaultCacheTtl is the default cache TTL.
var defaultCacheTtl time.Duration

// DB is a struct contaning a SQL database connection
type DB struct {
	conn  *sql.DB         // conn is the sqlite database connection.
	cache *cache.Cache    // cache is the cache database connection.
	ctx   context.Context // ctx is the context for database connections.
}

// New initializes a new database connection and sets up the necessary tables if they don't exist.
func New(url, cacheUrl string, cacheTtl time.Duration, ctx context.Context) (db *DB, err error) {
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

	if cacheUrl != "" {
		db.cache, err = cache.New(cacheUrl, ctx)
		if err != nil {
			return nil, err
		}
		defaultCacheTtl = cacheTtl
	}

	return
}

// Close closes the database connection.
func (d *DB) Close() (err error) {
	if err = d.conn.Close(); err != nil {
		return
	}

	if d.cache != nil {
		err = d.cache.Close()
	}

	return
}

// AddCourse adds a new course to the database.
func (d *DB) AddCourse(course *db.Course) (err error) {
	stmt := "INSERT INTO Courses(code, name, inserted_at) VALUES(?, ?, ?)"
	return execStmtContext(d.conn, d.ctx, stmt, course.Code, course.Name, time.Now().UnixNano())
}

// AddCourseMany adds new courses to the database.
func (d *DB) AddCourseMany(courses []*db.Course) (err error) {
	stmt, err := d.conn.PrepareContext(d.ctx, "INSERT INTO Courses(code, name, inserted_at) VALUES(?, ?, ?)")
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
func (d *DB) AddProfessor(name string) (err error) {
	professorUUID, err := uuid.NewV4()
	if err != nil {
		return
	}
	stmt := "INSERT INTO Professors(uuid, name, inserted_at) VALUES(?, ?, ?)"
	return execStmtContext(d.conn, d.ctx, stmt, professorUUID, name, time.Now().UnixNano())
}

// AddProfessorMany adds new professors to the database.
func (d *DB) AddProfessorMany(names []string) (err error) {
	stmt, err := d.conn.PrepareContext(d.ctx, "INSERT INTO Professors(uuid, name, inserted_at) VALUES(?, ?, ?)")
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
func (d *DB) AddCourseProfessor(professorUUID, courseCode string) (err error) {
	stmt := "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES(?, ?, ?)"
	return execStmtContext(d.conn, d.ctx, stmt, defaultHash, professorUUID, courseCode)
}

// AddCourseProfessorMany adds courses to professors in the database.
func (d *DB) AddCourseProfessorMany(professorUUIDS, courseCodes []string) (err error) {
	if len(professorUUIDS) != len(courseCodes) {
		return fmt.Errorf("unequal slice length")
	}

	stmt, err := d.conn.PrepareContext(d.ctx, "INSERT INTO Scores(hash, professor_uuid, course_code) VALUES(?, ?, ?)")
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
func (d *DB) RemoveCourse(code string, forceDelete bool) (err error) {
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

		if err = execStmtContext(d.conn, d.ctx, s.s, s.args); err != nil {
			return
		}
	}

	return
}

// RemoveProfessor removes a professor from the database. If forceDelete is true, associated scores are also deleted.
func (d *DB) RemoveProfessor(professorUUID string, forceDelete bool) (err error) {
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

		if err = execStmtContext(d.conn, d.ctx, s.s, s.args); err != nil {
			return
		}
	}

	return
}

// GetLastCourses retrieves the last 100 courses from the database.
func (d *DB) GetLastCourses() (courses []*db.Course, err error) {
	if d.cache != nil {
		key := "GetLastCourses"
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(courses)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return courses, json.Unmarshal([]byte(cached), &courses)
		}
	}

	stmt := `
		SELECT code, name
		FROM Courses
		ORDER BY inserted_at
		DESC
		LIMIT ?
	`

	rows, err := d.conn.QueryContext(d.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := db.Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return
		}
		courses = append(courses, &course)
	}

	return
}

// GetLastProfessors retrieves the last 100 professors from the database.
func (d *DB) GetLastProfessors() (professors []*db.Professor, err error) {
	if d.cache != nil {
		key := "GetLastProfessors"
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(professors)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return professors, json.Unmarshal([]byte(cached), &professors)
		}
	}

	stmt := `
		SELECT uuid, name
		FROM Professors
		ORDER BY inserted_at
		DESC 
		LIMIT ?
	`

	rows, err := d.conn.QueryContext(d.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := db.Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return
		}
		professors = append(professors, &professor)
	}

	return
}

// GetLastScores retrieves the last 100 scores from the database.
func (d *DB) GetLastScores() (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetLastScores"
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
		if err = rows.Scan(&score.ProfessorUUID, &score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetCoursesByProfessor retrieves all courses associated with a professor from the database.
func (d *DB) GetCoursesByProfessorUUID(UUID string) (courses []*db.Course, err error) {
	if d.cache != nil {
		key := "GetCoursesByProfessorUUID" + UUID
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(courses)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return courses, json.Unmarshal([]byte(cached), &courses)
		}
	}

	stmt := `
		SELECT code, name
		FROM Courses
		JOIN Scores ON Courses.code = Scores.course_code
		WHERE Scores.professor_uuid = ?
		ORDER BY Courses.inserted_at
		DESC
	`

	rows, err := d.conn.QueryContext(d.ctx, stmt, UUID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		course := db.Course{}
		if err = rows.Scan(&course.Code, &course.Name); err != nil {
			return
		}
		courses = append(courses, &course)
	}

	return
}

// GetProfessorsByCourse retrieves all professors associated with a course from the database.
func (d *DB) GetProfessorsByCourseCode(code string) (professors []*db.Professor, err error) {
	if d.cache != nil {
		key := "GetProfessorsByCourseCode" + code
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(professors)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return professors, json.Unmarshal([]byte(cached), &professors)
		}
	}

	stmt := `
		SELECT uuid, name
		FROM Professors
		JOIN Scores ON Professors.uuid = Scores.professor_uuid
		WHERE Scores.course_code = ?
		ORDER BY Professors.inserted_at
		DESC
	`

	rows, err := d.conn.QueryContext(d.ctx, stmt, code)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		professor := db.Professor{}
		if err = rows.Scan(&professor.UUID, &professor.Name); err != nil {
			return
		}
		professors = append(professors, &professor)
	}

	return
}

// GetProfessorUUIDByName retrieves the UUID of the professor that matches the specified name.
func (d *DB) GetProfessorUUIDByName(name string) (uuid string, err error) {
	if d.cache != nil {
		key := "GetProfessorUUIDByName" + name
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				if err = d.cache.Set(key, uuid, defaultCacheTtl); err != nil {
					log.Println(err)
				}
			}()
		} else if err == nil {
			return cached, nil
		}
	}

	stmt := `
		SELECT uuid
		FROM Professors
		WHERE name = ?
		LIMIT 1
	`

	row := d.conn.QueryRowContext(d.ctx, stmt, name)
	if err = row.Scan(&uuid); err != nil {
		return
	}
	return
}

// GetScoresByProfessorUUID retrieves all scores associated with a professor's UUID from the database.
func (d *DB) GetScoresByProfessorUUID(UUID string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByProfessorUUID" + UUID
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, UUID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
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
func (d *DB) GetScoresByProfessorName(name string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByProfessorName" + name
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
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
func (d *DB) GetScoresByProfessorNameLike(nameLike string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByProfessorNameLike" + nameLike
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, fmt.Sprintf("%%%s%%", nameLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseName retrieves all scores associated with a course from the database.
func (d *DB) GetScoresByCourseName(name string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByCourseName" + name
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
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
func (d *DB) GetScoresByCourseNameLike(nameLike string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByCourseNameLike" + nameLike
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, fmt.Sprintf("%%%s%%", nameLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GetScoresByCourseCode retrieves all scores associated with a course from the database.
func (d *DB) GetScoresByCourseCode(code string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByCourseCode" + code
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, code)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
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
func (d *DB) GetScoresByCourseCodeLike(codeLike string) (scores []*db.Score, err error) {
	if d.cache != nil {
		key := "GetScoresByCourseCodeLike" + codeLike
		cached, err := d.cache.Get(key)
		if err == cache.ErrRedisNil {
			defer func() {
				data, err := json.Marshal(scores)
				if err == nil {
					if err = d.cache.Set(key, data, defaultCacheTtl); err != nil {
						log.Println(err)
					}
				}
			}()
		} else if err == nil {
			return scores, json.Unmarshal([]byte(cached), &scores)
		}
	}

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

	rows, err := d.conn.QueryContext(d.ctx, stmt, fmt.Sprintf("%%%s%%", codeLike), maxRowReturn)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		score := db.Score{}
		if err = rows.Scan(&score.ProfessorName, &score.CourseCode, &score.CourseName, &score.ProfessorUUID, &score.ScoreTeaching, &score.ScoreCourseWork, &score.ScoreLearning); err != nil {
			return
		}
		score.ScoreAverage = averageScore(score.ScoreTeaching, score.ScoreCourseWork, score.ScoreLearning)
		scores = append(scores, &score)
	}

	return
}

// GradeCourseProfessor updates the scores of a professor for a specific course in the database.
func (d *DB) GradeCourseProfessor(professorUUID, courseCode, username string, grades [3]float32) (err error) {
	var Hasher = xxh3.New()
	if _, err = Hasher.WriteString(username + courseCode + professorUUID); err != nil {
		return
	}
	hash := Hasher.Sum64()

	if graded, err := d.checkGraded(hash); err != nil {
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

	return execStmtContext(d.conn, d.ctx, stmt, fmt.Sprintf("%d", hash), professorUUID, courseCode, grades[0], grades[1], grades[2], time.Now().UnixNano())
}

// CheckGraded checks if a user graded a course.
// The hash parameter is obtained by hashing
// the concatenation of the username, course code,
// and professor uuid using the xxh3 algorithm.
func (d *DB) checkGraded(hash uint64) (graded bool, err error) {
	var count int

	stmt := "SELECT COUNT(*) FROM Scores WHERE hash = ?"
	if err = d.conn.QueryRowContext(d.ctx, stmt, fmt.Sprintf("%d", hash)).Scan(&count); err != nil {
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
