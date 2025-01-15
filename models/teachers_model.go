package models

import (
	"time"
	"gorm.io/gorm"
)

type Teachers struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	Username      string    `gorm:"unique;not null" json:"username"`
	Email         string    `gorm:"unique;not null" json:"email"`
	Password      string    `gorm:"not null" json:"password"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	Classes       []Class   `gorm:"many2many:class_teachers;" json:"classes"`
	CreatedClasses []Class  `gorm:"foreignkey:CreatedByID" json:"createdClasses"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (t *Teachers) TableName() string {
	return "teachers"
}

func (t *Teachers) BeforeCreate(tx *gorm.DB) (err error) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return
}

func (t *Teachers) BeforeUpdate(tx *gorm.DB) (err error) {
	t.UpdatedAt = time.Now()
	return
}

func MigrationTeachers(db *gorm.DB) {
	db.AutoMigrate(&Teachers{})
}
  