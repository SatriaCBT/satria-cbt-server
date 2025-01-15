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
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdminLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Token     string    `json:"token"`
}


type TeacherResponse struct {
    ID            uint    `json:"id"`
    Name          string    `json:"name"`
    Username      string    `json:"username"`
    Email         string    `json:"email"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
    Classes       []models.Class   `json:"classes"`  
    CreatedClasses []models.Class  `json:"createdClasses"`
}

type StudentResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Classes   []models.Class   `json:"classes"`  
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
