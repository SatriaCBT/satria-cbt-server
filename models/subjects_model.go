package models

import (
	"time"

	"gorm.io/gorm"
)

type Subject struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"unique;not null" json:"code"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *Subject) TableName() string {
	return "subjects"
}

func MigrationSubjects(db *gorm.DB) {
	db.AutoMigrate(&Subject{})
}
