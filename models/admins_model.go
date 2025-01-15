package models

import (
	"time"

	"gorm.io/gorm"
)


type Admins struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

type AdminsRequest struct {
	Email string `gorm:"email" json:"email"`
	Password string `gorm:"password" json:"password"`
}


func(a *Admins) TableName() string {
	return "admins"
}

func MigrationAdmin(db *gorm.DB) {
	db.AutoMigrate(&Admins{})
}
  