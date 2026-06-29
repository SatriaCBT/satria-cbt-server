package models

import (
	"time"

	"gorm.io/gorm"
)

type QuestionType string

const (
	QuestionMultipleChoice QuestionType = "multiple_choice"
	QuestionTrueFalse      QuestionType = "true_false"
	QuestionEssay          QuestionType = "essay"
)

type Question struct {
	gorm.Model
	ID            uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectID     uint         `gorm:"not null;index" json:"subjectId"`
	Subject       Subject      `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Type          QuestionType `gorm:"type:varchar(20);not null;default:'multiple_choice'" json:"type"`
	Question      string       `gorm:"type:text;not null" json:"question"`
	Options       string       `gorm:"type:jsonb" json:"options,omitempty"`
	CorrectAnswer string       `gorm:"type:text;not null" json:"correctAnswer"`
	Points        int          `gorm:"not null;default:1" json:"points"`
	CreatedByID   uint         `gorm:"not null;index" json:"createdById"`
	CreatedBy     Teachers     `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
	CreatedAt     time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (q *Question) TableName() string {
	return "questions"
}

func MigrationQuestions(db *gorm.DB) {
	db.AutoMigrate(&Question{})
}
