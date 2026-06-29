package models

import (
	"time"

	"gorm.io/gorm"
)

type AttemptStatus string

const (
	AttemptInProgress AttemptStatus = "in_progress"
	AttemptCompleted  AttemptStatus = "completed"
	AttemptGraded     AttemptStatus = "graded"
)

type ExamAttempt struct {
	gorm.Model
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	ExamID      uint          `gorm:"not null;index" json:"examId"`
	Exam        Exam          `gorm:"foreignKey:ExamID" json:"exam,omitempty"`
	StudentID   uint          `gorm:"not null;index" json:"studentId"`
	Student     Students      `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	StartTime   time.Time     `gorm:"not null" json:"startTime"`
	EndTime     *time.Time    `json:"endTime,omitempty"`
	Score       *int          `json:"score,omitempty"`
	TotalPoints int           `gorm:"default:0" json:"totalPoints"`
	TotalCorrect       int           `gorm:"default:0" json:"totalCorrect"`
	TotalWrong         int           `gorm:"default:0" json:"totalWrong"`
	CurrentQuestionIdx int           `gorm:"default:0" json:"currentQuestionIdx"`
	TabSwitchCount     int           `gorm:"default:0" json:"tabSwitchCount"`
	Suspicious         bool          `gorm:"default:false" json:"suspicious"`
	SuspiciousReason   string        `gorm:"type:text" json:"suspiciousReason,omitempty"`
	Status             AttemptStatus `gorm:"type:varchar(20);default:'in_progress'" json:"status"`
	Answers     []ExamAnswer  `gorm:"foreignKey:AttemptID" json:"answers,omitempty"`
	CreatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (ea *ExamAttempt) TableName() string {
	return "exam_attempts"
}

func MigrationExamAttempts(db *gorm.DB) {
	db.AutoMigrate(&ExamAttempt{})
}
