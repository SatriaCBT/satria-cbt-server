package models

import (
	"time"
	"gorm.io/gorm"
)

type Students struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	Classes   []Class   `gorm:"many2many:class_students;" json:"classes"`
}

type StudentsRequest struct {
	Email string `gorm:"email" json:"email"`
	Password string `gorm:"password" json:"password"`
}

func (s *Students) TableName() string {
	return "students"
}


func (s *Students) BeforeCreate(tx *gorm.DB) (err error) {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return
}


func (s *Students) BeforeUpdate(tx *gorm.DB) (err error) {
	s.UpdatedAt = time.Now()
	return
}

func MigrationStudents(db *gorm.DB) {
	db.AutoMigrate(&Students{})
}

  