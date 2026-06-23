package res

import (
	"time"

)

type ClassSummaryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type AdminResponse struct {
	ID        uint    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type AdminLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
    Token     string    `json:"token"`
}


type TeacherResponse struct {
    ID            uint    `json:"id"`
    Name          string    `json:"name"`
    Username      string    `json:"username"`
    Email         string    `json:"email"`
    CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
    Classes       []ClassSummaryResponse   `json:"classes"`  
}

type TeacherLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
    Token     string    `json:"token"`
}

type StudentResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
    Classes   []ClassSummaryResponse  `json:"classes"`  
}

type StudentLoginResponse struct {
    ID        uint    `json:"id"`
    Name      string    `json:"name"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Classes []ClassSummaryResponse `json:"classes"`
    CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
    Token     string    `json:"token"`
}


type ClassResponse struct {
    ID        uint              `json:"id"`
    Name      string           `json:"name"`
    Code      string           `json:"code"`
    Teachers  []TeacherResponse `json:"teachers"`
    Students  []StudentResponse `json:"students"`
    CreatedBy AdminResponse   `json:"createdBy"`
    CreatedAt time.Time        `json:"createdAt"`
    UpdatedAt *time.Time       `json:"updatedAt,omitempty"`
}
