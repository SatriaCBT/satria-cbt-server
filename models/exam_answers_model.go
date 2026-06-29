package models

import (
	"time"

	"gorm.io/gorm"
)

type ExamAnswer struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	AttemptID uint   `gorm:"not null;index" json:"attemptId"`
	Attempt   ExamAttempt `gorm:"foreignKey:AttemptID" json:"attempt,omitempty"`
	QuestionID uint  `gorm:"not null;index" json:"questionId"`
	Question  Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Answer    string `gorm:"type:text" json:"answer"`
	IsCorrect *bool  `json:"isCorrect,omitempty"`
	Points    int    `gorm:"default:0" json:"points"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (ea *ExamAnswer) TableName() string {
	return "exam_answers"
}

func MigrationExamAnswers(db *gorm.DB) {
	db.AutoMigrate(&ExamAnswer{})
}
