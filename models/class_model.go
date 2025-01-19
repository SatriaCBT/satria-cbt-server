package models

import (
	"time"

	"gorm.io/gorm"
)


type Class struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"unique;not null" json:"code"`
	Teachers    []Teachers `gorm:"many2many:class_teachers;" json:"teachers"`
	Students    []Students `gorm:"many2many:class_students;" json:"students"`
	CreatedBy   Admins    `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	CreatedByID uint      `json:"createdByID"`
}


func (c *Class) TableName() string {
	return "class"
}


func (c *Class) BeforeCreate(tx *gorm.DB) (err error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return
}


func (c *Class) BeforeUpdate(tx *gorm.DB) (err error) {
	c.UpdatedAt = time.Now()
	return
}


func MigrationClass(db *gorm.DB) {
	db.AutoMigrate(&Class{})
}
  