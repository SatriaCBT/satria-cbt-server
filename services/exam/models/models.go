package models

import (
	"time"

	"gorm.io/gorm"
)

type Subject struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"unique;not null" json:"code"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Subject) TableName() string { return "subjects" }

type QuestionType string

const (
	QCMultipleChoice QuestionType = "multiple_choice"
	QCTrueFalse      QuestionType = "true_false"
	QCEssay          QuestionType = "essay"
)

type Question struct {
	ID            uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectID     uint         `gorm:"not null;index" json:"subjectId"`
	Type          QuestionType `gorm:"type:varchar(20);default:'multiple_choice'" json:"type"`
	Question      string       `gorm:"type:text;not null" json:"question"`
	Options       string       `gorm:"type:jsonb" json:"options,omitempty"`
	CorrectAnswer string       `gorm:"type:text;not null" json:"correctAnswer"`
	Points        int          `gorm:"default:1" json:"points"`
	Explanation   string       `gorm:"type:text" json:"explanation,omitempty"`
	CreatedByID   uint         `gorm:"not null;index" json:"createdById"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

func (Question) TableName() string { return "questions" }

type ExamStatus string

const (
	ExamDraft     ExamStatus = "draft"
	ExamPublished ExamStatus = "published"
	ExamCompleted ExamStatus = "completed"
)

type Exam struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title            string     `gorm:"type:varchar(200);not null" json:"title"`
	Description      string     `gorm:"type:text" json:"description,omitempty"`
	SubjectID        uint       `gorm:"not null;index" json:"subjectId"`
	Duration         int        `gorm:"not null" json:"duration"`
	StartTime        *time.Time `json:"startTime,omitempty"`
	EndTime          *time.Time `json:"endTime,omitempty"`
	TotalPoints      int        `gorm:"default:0" json:"totalPoints"`
	PassingScore     int        `gorm:"default:0" json:"passingScore"`
	ShuffleQuestions bool       `gorm:"default:false" json:"shuffleQuestions"`
	ShowResult       bool       `gorm:"default:true" json:"showResult"`
	MaxAttempts      int        `gorm:"default:1" json:"maxAttempts"`
	Status           ExamStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CreatedByID      uint       `gorm:"not null;index" json:"createdById"`
	ClassID          uint       `gorm:"index" json:"classId"`
	Questions        []Question `gorm:"-" json:"questions,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (Exam) TableName() string { return "exams" }

func (e *Exam) AfterFind(tx *gorm.DB) error {
	return tx.Raw(`SELECT q.* FROM questions q JOIN exam_questions eq ON eq.question_id = q.id WHERE eq.exam_id = ? ORDER BY eq.order_index`, e.ID).Scan(&e.Questions).Error
}

type ExamQuestion struct {
	ExamID     uint `gorm:"primaryKey"`
	QuestionID uint `gorm:"primaryKey"`
	OrderIndex int  `gorm:"default:0"`
}

func (ExamQuestion) TableName() string { return "exam_questions" }

type AttemptStatus string

const (
	AttemptInProgress AttemptStatus = "in_progress"
	AttemptCompleted  AttemptStatus = "completed"
	AttemptGraded     AttemptStatus = "graded"
)

type ExamAttempt struct {
	ID                 uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	ExamID             uint          `gorm:"not null;index" json:"examId"`
	StudentID          uint          `gorm:"not null;index" json:"studentId"`
	StartTime          time.Time     `gorm:"not null" json:"startTime"`
	EndTime            *time.Time    `json:"endTime,omitempty"`
	Score              *int          `json:"score,omitempty"`
	TotalPoints        int           `gorm:"default:0" json:"totalPoints"`
	TotalCorrect       int           `gorm:"default:0" json:"totalCorrect"`
	TotalWrong         int           `gorm:"default:0" json:"totalWrong"`
	CurrentQuestionIdx int           `gorm:"default:0" json:"currentQuestionIdx"`
	TabSwitchCount     int           `gorm:"default:0" json:"tabSwitchCount"`
	Suspicious         bool          `gorm:"default:false" json:"suspicious"`
	SuspiciousReason   string        `gorm:"type:text" json:"suspiciousReason,omitempty"`
	Status             AttemptStatus `gorm:"type:varchar(20);default:'in_progress'" json:"status"`
	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}

func (ExamAttempt) TableName() string { return "exam_attempts" }

type ExamAnswer struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	AttemptID  uint   `gorm:"not null;index" json:"attemptId"`
	QuestionID uint   `gorm:"not null;index" json:"questionId"`
	Answer     string `gorm:"type:text" json:"answer"`
	IsCorrect  *bool  `json:"isCorrect,omitempty"`
	Points     int    `gorm:"default:0" json:"points"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (ExamAnswer) TableName() string { return "exam_answers" }

type Class struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"unique;not null" json:"code"`
	CreatedByID uint      `json:"createdByID"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Class) TableName() string { return "class" }
