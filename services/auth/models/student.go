package models

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (Student) TableName() string { return "students" }

type StudentRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
