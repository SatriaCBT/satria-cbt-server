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
