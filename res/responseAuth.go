package res

import (
	"time"
	"satriacbtserver/models"

)

type AdminResponse struct {
	ID        uint    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type AdminLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Token     string    `json:"token"`
}


type TeacherResponse struct {
    ID            uint    `json:"id"`
    Name          string    `json:"name"`
    Username      string    `json:"username"`
    Email         string    `json:"email"`
    CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
    Classes       []models.Class   `json:"classes"`  
    CreatedClasses []models.Class  `json:"createdClasses"`
}

type TeacherLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Token     string    `json:"token"`
}

type StudentResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Classes   []models.Class   `json:"classes"`  
}

type StudentLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Token     string    `json:"token"`
}


type ClassResponse struct {
    ID          uint    `json:"id"`
    Name        string    `json:"name"`
    Code        string    `json:"code"`
    Teachers    []models.Teachers`json:"teachers"`  
    Students    []models.Students `json:"students"` 
    CreatedBy   models.Teachers   `json:"createdBy"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
