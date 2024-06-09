package db

// DB is the database interface.
type DB interface {
	Close() error
	AddCourse(course *Course) error
	AddCourseMany([]*Course) error
	AddProfessor(string) error
	AddProfessorMany(names []string) error
	AddCourseProfessor(professorUUID, courseCode string) error
	AddCourseProfessorMany(professorUUIDS, courseCodes []string) error
	RemoveCourse(string, bool) error
	RemoveProfessor(string, bool) error
	GetLastCourses() ([]*Course, error)
	GetLastProfessors() ([]*Professor, error)
	GetLastScores() ([]*Score, error)
	GetCoursesByProfessorUUID(string) ([]*Course, error)
	GetProfessorsByCourseCode(string) ([]*Professor, error)
	GetProfessorUUIDByName(string) (string, error)
	GetScoresByProfessorUUID(string) ([]*Score, error)
	GetScoresByProfessorName(string) ([]*Score, error)
	GetScoresByProfessorNameLike(string) ([]*Score, error)
	GetScoresByCourseName(string) ([]*Score, error)
	GetScoresByCourseNameLike(string) ([]*Score, error)
	GetScoresByCourseCode(string) ([]*Score, error)
	GetScoresByCourseCodeLike(string) ([]*Score, error)
	GradeCourseProfessor(string, string, string, [3]float32) error
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
	ScoreCourseWork float32 `json:"scoreCoursework"` // Score related to the homeworks, quizzes, and exams given by the professor
	ScoreLearning   float32 `json:"scoreLearning"`   // Score related to the learning outcomes of the course
	ScoreAverage    float32 `json:"scoreAverage"`    // Average score of the teaching, coursework, and learning scores
	Count           int     `json:"count"`           // Numbero of students who graded this course
}
