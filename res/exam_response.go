package res

import "time"

type SubjectResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type QuestionResponse struct {
	ID            uint      `json:"id"`
	SubjectID     uint      `json:"subjectId"`
	Type          string    `json:"type"`
	Question      string    `json:"question"`
	Options       string    `json:"options,omitempty"`
	Points        int       `json:"points"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type QuestionResponseWithAnswer struct {
	ID            uint      `json:"id"`
	SubjectID     uint      `json:"subjectId"`
	Type          string    `json:"type"`
	Question      string    `json:"question"`
	Options       string    `json:"options,omitempty"`
	CorrectAnswer string    `json:"correctAnswer"`
	Points        int       `json:"points"`
	Explanation   string    `json:"explanation,omitempty"`
	CreatedByID   uint      `json:"createdById"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ExamResponse struct {
	ID               uint              `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description,omitempty"`
	SubjectID        uint              `json:"subjectId"`
	Subject          *SubjectResponse  `json:"subject,omitempty"`
	Duration         int               `json:"duration"`
	StartTime        *time.Time        `json:"startTime,omitempty"`
	EndTime          *time.Time        `json:"endTime,omitempty"`
	TotalPoints      int               `json:"totalPoints"`
	PassingScore     int               `json:"passingScore"`
	ShuffleQuestions bool              `json:"shuffleQuestions"`
	ShowResult       bool              `json:"showResult"`
	MaxAttempts      int               `json:"maxAttempts"`
	Status           string            `json:"status"`
	CreatedByID      uint              `json:"createdById"`
	ClassID          uint              `json:"classId"`
	QuestionCount    int               `json:"questionCount"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

type ExamDetailResponse struct {
	ExamResponse
	Questions []QuestionResponseWithAnswer `json:"questions,omitempty"`
}

type ExamAttemptResponse struct {
	ID           uint                   `json:"id"`
	ExamID       uint                   `json:"examId"`
	StudentID    uint                   `json:"studentId"`
	StartTime    time.Time              `json:"startTime"`
	EndTime      *time.Time             `json:"endTime,omitempty"`
	Score        *int                   `json:"score,omitempty"`
	TotalPoints  int                    `json:"totalPoints"`
	TotalCorrect int                    `json:"totalCorrect"`
	TotalWrong   int                    `json:"totalWrong"`
	Status       string                 `json:"status"`
	Answers      []ExamAnswerResponse   `json:"answers,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

type ExamAnswerResponse struct {
	ID         uint   `json:"id"`
	AttemptID  uint   `json:"attemptId"`
	QuestionID uint   `json:"questionId"`
	Answer     string `json:"answer"`
	IsCorrect  *bool  `json:"isCorrect,omitempty"`
	Points     int    `json:"points"`
}

type ExamAnswerWithQuestionResponse struct {
	ExamAnswerResponse
	Question QuestionResponse `json:"question"`
}

type DashboardStatsResponse struct {
	TotalExams      int64 `json:"totalExams"`
	TotalSubjects   int64 `json:"totalSubjects"`
	TotalStudents   int64 `json:"totalStudents"`
	TotalTeachers   int64 `json:"totalTeachers"`
	TotalAttempts   int64 `json:"totalAttempts"`
	OngoingAttempts int64 `json:"ongoingAttempts"`
}

type ExamStatResponse struct {
	ExamID        uint    `json:"examId"`
	ExamTitle     string  `json:"examTitle"`
	TotalAttempts int     `json:"totalAttempts"`
	AverageScore  float64 `json:"averageScore"`
	HighestScore  int     `json:"highestScore"`
	LowestScore   int     `json:"lowestScore"`
	PassCount     int     `json:"passCount"`
	FailCount     int     `json:"failCount"`
	PassRate      float64 `json:"passRate"`
}

type StudentPerformanceResponse struct {
	StudentID    uint   `json:"studentId"`
	StudentName  string `json:"studentName"`
	TotalExams   int    `json:"totalExams"`
	AverageScore float64 `json:"averageScore"`
	HighestScore int    `json:"highestScore"`
	LowestScore  int    `json:"lowestScore"`
	TotalPass    int    `json:"totalPass"`
	TotalFail    int    `json:"totalFail"`
}

type ClassPerformanceResponse struct {
	ClassID      uint    `json:"classId"`
	ClassName    string  `json:"className"`
	TotalStudents int    `json:"totalStudents"`
	TotalAttempts int    `json:"totalAttempts"`
	AverageScore float64 `json:"averageScore"`
	PassCount    int     `json:"passCount"`
	FailCount    int     `json:"failCount"`
}

type RecentAttemptResponse struct {
	AttemptID    uint      `json:"attemptId"`
	ExamTitle    string    `json:"examTitle"`
	StudentName  string    `json:"studentName"`
	Score        *int      `json:"score,omitempty"`
	Status       string    `json:"status"`
	SubmittedAt  *time.Time `json:"submittedAt,omitempty"`
}
