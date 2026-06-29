package models

import (
	"time"

	"gorm.io/gorm"
)

type ExamStatus string

const (
	ExamDraft     ExamStatus = "draft"
	ExamPublished ExamStatus = "published"
	ExamOngoing   ExamStatus = "ongoing"
	ExamCompleted ExamStatus = "completed"
)

type Exam struct {
	gorm.Model
	ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title             string     `gorm:"type:varchar(200);not null" json:"title"`
	Description       string     `gorm:"type:text" json:"description,omitempty"`
	SubjectID         uint       `gorm:"not null;index" json:"subjectId"`
	Subject           Subject    `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Duration          int        `gorm:"not null" json:"duration"`
	StartTime         *time.Time `json:"startTime,omitempty"`
	EndTime           *time.Time `json:"endTime,omitempty"`
	TotalPoints       int        `gorm:"default:0" json:"totalPoints"`
	PassingScore      int        `gorm:"default:0" json:"passingScore"`
	ShuffleQuestions  bool       `gorm:"default:false" json:"shuffleQuestions"`
	ShowResult        bool       `gorm:"default:true" json:"showResult"`
	MaxAttempts       int        `gorm:"default:1" json:"maxAttempts"`
	Status            ExamStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CreatedByID       uint       `gorm:"not null;index" json:"createdById"`
	CreatedBy         Teachers   `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
	ClassID           uint       `gorm:"index" json:"classId"`
	Class             Class      `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Questions         []Question `gorm:"many2many:exam_questions;" json:"questions,omitempty"`
	CreatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

type ExamQuestion struct {
	ExamID     uint `gorm:"primaryKey"`
	QuestionID uint `gorm:"primaryKey"`
	OrderIndex int  `gorm:"default:0"`
}

func (e *Exam) TableName() string {
	return "exams"
}

func (ExamQuestion) TableName() string {
	return "exam_questions"
}

func MigrationExams(db *gorm.DB) {
	db.AutoMigrate(&Exam{}, &ExamQuestion{})
}
