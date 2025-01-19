package controllers

import (
	"satriacbtserver/configs"
	"satriacbtserver/models"
	"satriacbtserver/res"
	_ "time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type ClassController struct {
	studentController *StudentController
	teacherController *TeacherController
}

func NewClassController(studentController *StudentController, teacherController *TeacherController) *ClassController {
	return &ClassController{
		studentController: studentController,
		teacherController: teacherController,
	}
}

func (k *ClassController) CreateClass(c *fiber.Ctx) error {
	var req models.Class

	adminIDClaim, ok := c.Locals("userID").(jwt.MapClaims)
    if !ok {
        return &fiber.Error{
            Code:    fiber.StatusUnauthorized,
            Message: "Unauthorized: invalid token claims",
        }
    }

    adminID, ok := adminIDClaim["id"].(float64)
    if !ok {
        return &fiber.Error{
            Code:    fiber.StatusUnauthorized,
            Message: "Unauthorized: invalid user ID",
        }
    }
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	if req.Name == "" || req.Code == ""{
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Name, code, and createdByID are required",
		}
	}

	class := models.Class{
		Name:        req.Name,
		Code:        req.Code,
		CreatedByID: uint(adminID),
	}

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		var admin models.Admins
		if err := tx.First(&admin, adminID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &fiber.Error{
					Code:    fiber.StatusBadRequest,
					Message: "Creating admin not found",
				}
			}
			return err
		}

		if err := tx.Create(&class).Error; err != nil {
			return err
		}

		if len(req.Teachers) > 0 {
			var count int64
			teacherIDs := make([]uint, len(req.Teachers))
			for i, t := range req.Teachers {
				teacherIDs[i] = t.ID
			}
			if err := tx.Model(&models.Teachers{}).Where("id IN ?", teacherIDs).Count(&count).Error; err != nil {
				return err
			}
			if int(count) != len(teacherIDs) {
				return &fiber.Error{
					Code:    fiber.StatusBadRequest,
					Message: "One or more teacher IDs are invalid",
				}
			}
			if err := tx.Model(&class).Association("Teachers").Append(req.Teachers); err != nil {
				return err
			}
		}

		if len(req.Students) > 0 {
			var count int64
			studentIDs := make([]uint, len(req.Students))
			for i, s := range req.Students {
				studentIDs[i] = s.ID
			}
			if err := tx.Model(&models.Students{}).Where("id IN ?", studentIDs).Count(&count).Error; err != nil {
				return err
			}
			if int(count) != len(studentIDs) {
				return &fiber.Error{
					Code:    fiber.StatusBadRequest,
					Message: "One or more student IDs are invalid",
				}
			}
			if err := tx.Model(&class).Association("Students").Append(req.Students); err != nil {
				return err
			}
		}
		

		if err := tx.Preload("Teachers.Classes").
			Preload("Students.Classes").
			Preload("CreatedBy").
			First(&class, class.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	teachers := make([]res.TeacherResponse, len(class.Teachers))
	for i, teacher := range class.Teachers {
		teacherClasses := make([]res.ClassSummaryResponse, len(teacher.Classes))
		for j, cls := range teacher.Classes {
			teacherClasses[j] = res.ClassSummaryResponse{
				ID:   cls.ID,
				Name: cls.Name,
				Code: cls.Code,
			}
		}


		teachers[i] = res.TeacherResponse{
			ID:             teacher.ID,
			Name:           teacher.Name,
			Username:       teacher.Username,
			Email:          teacher.Email,
			CreatedAt:      teacher.CreatedAt,
			Classes:        teacherClasses,
		}
	}

	students := make([]res.StudentResponse, len(class.Students))
	for i, student := range class.Students {
		studentClasses := make([]res.ClassSummaryResponse, len(student.Classes))
		for j, cls := range student.Classes {
			studentClasses[j] = res.ClassSummaryResponse{
				ID:   cls.ID,
				Name: cls.Name,
				Code: cls.Code,
			}
		}

		students[i] = res.StudentResponse{
			ID:        student.ID,
			Name:      student.Name,
			Username:  student.Username,
			Email:     student.Email,
			CreatedAt: student.CreatedAt,
			Classes:   studentClasses,
		}
	}

	response := res.ClassResponse{
		ID:        class.ID,
		Name:      class.Name,
		Code:      class.Code,
		Teachers:  teachers,
		Students:  students,
		CreatedBy: res.AdminResponse{
			ID:        class.CreatedBy.ID,
			Name:      class.CreatedBy.Name,
			Username:  class.CreatedBy.Username,
			Email:     class.CreatedBy.Email,
			CreatedAt: class.CreatedBy.CreatedAt,
		},
		CreatedAt: class.CreatedAt,
	}

	if class.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusCreated,
		Message: "Class created successfully",
		Data:    response,
	})
}


func (k *ClassController) GetAllClass(c *fiber.Ctx) error {
    var classes []models.Class

    err := configs.Database().Transaction(func(tx *gorm.DB) error {
        if err := tx.Preload("Teachers").Preload("Students").Preload("CreatedBy").Find(&classes).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        return &fiber.Error{
            Code:    fiber.StatusInternalServerError,
            Message: err.Error(),
        }
    }

    var response []res.ClassResponse
    for _, class := range classes {
        teachers := make([]res.TeacherResponse, len(class.Teachers))
        for i, teacher := range class.Teachers {
            teachers[i] = res.TeacherResponse{
                ID:        teacher.ID,
                Name:      teacher.Name,
                Username:  teacher.Username,
                Email:     teacher.Email,
				Classes:   []res.ClassSummaryResponse{
					{
						ID:   class.ID,
						Name: class.Name,
						Code: class.Code,
					},
				},
                CreatedAt: teacher.CreatedAt,
            }
        }

        students := make([]res.StudentResponse, len(class.Students))
        for i, student := range class.Students {
            students[i] = res.StudentResponse{
                ID:        student.ID,
                Name:      student.Name,
                Username:  student.Username,
                Email:     student.Email,
				Classes:   []res.ClassSummaryResponse{
					{
						ID:   class.ID,
						Name: class.Name,
						Code: class.Code,
					},
				},
                CreatedAt: student.CreatedAt,

            }
        }

        classResponse := res.ClassResponse{
            ID:   class.ID,
            Name: class.Name,
            Code: class.Code,
            Teachers: teachers,
            Students: students,
			CreatedBy: res.AdminResponse{
				ID:        class.CreatedBy.ID,
				Name:      class.CreatedBy.Name,
				Username:  class.CreatedBy.Username,
				Email:     class.CreatedBy.Email,
			},
            CreatedAt: class.CreatedAt,
        }

        if !class.UpdatedAt.IsZero() {
            classResponse.UpdatedAt = &class.UpdatedAt
        }

        response = append(response, classResponse)
    }

    return c.JSON(res.ResponseCode{
        Code:    fiber.StatusOK,
        Message: "Classes retrieved successfully",
        Data:    response,
    })
}


func (k *ClassController) GetClassByID(c *fiber.Ctx) error {
	classID := c.Params("id")

	var class models.Class
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Teachers").Preload("Students").Preload("CreatedBy").First(&class, classID).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	teachers := make([]res.TeacherResponse, len(class.Teachers))
	for i, teacher := range class.Teachers {
		teachers[i] = res.TeacherResponse{
			ID:        teacher.ID,
			Name:      teacher.Name,			
			Username:  teacher.Username,
			Email:     teacher.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: teacher.CreatedAt,
		}
	}	

	students := make([]res.StudentResponse, len(class.Students))
	for i, student := range class.Students {
		students[i] = res.StudentResponse{
			ID:        student.ID,
			Name:      student.Name,			
			Username:  student.Username,
			Email:     student.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: student.CreatedAt,
		}
	}

	classResponse := res.ClassResponse{
		ID:   class.ID,
		Name: class.Name,
		Code: class.Code,
		Teachers: teachers,
		Students: students,
		CreatedBy: res.AdminResponse{
			ID:        class.CreatedBy.ID,
			Name:      class.CreatedBy.Name,
			Username:  class.CreatedBy.Username,
			Email:     class.CreatedBy.Email,
		},
		CreatedAt: class.CreatedAt,
	}

	if !class.UpdatedAt.IsZero() {
		classResponse.UpdatedAt = &class.UpdatedAt
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Class retrieved successfully",
		Data:    classResponse,
	})
}


func (k *ClassController) GetClassByCode(c *fiber.Ctx) error {
	classCode := c.Params("code")

	var class models.Class
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Teachers").Preload("Students").Preload("CreatedBy").First(&class, "code = ?", classCode).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	teachers := make([]res.TeacherResponse, len(class.Teachers))
	for i, teacher := range class.Teachers {
		teachers[i] = res.TeacherResponse{
			ID:        teacher.ID,
			Name:      teacher.Name,			
			Username:  teacher.Username,
			Email:     teacher.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: teacher.CreatedAt,
		}
	}	

	students := make([]res.StudentResponse, len(class.Students))
	for i, student := range class.Students {
		students[i] = res.StudentResponse{
			ID:        student.ID,
			Name:      student.Name,			
			Username:  student.Username,
			Email:     student.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: student.CreatedAt,
		}
	}

	classResponse := res.ClassResponse{
		ID:   class.ID,
		Name: class.Name,		
		Code: class.Code,
		Teachers: teachers,
		Students: students,
		CreatedBy: res.AdminResponse{
			ID:        class.CreatedBy.ID,
			Name:      class.CreatedBy.Name,
			Username:  class.CreatedBy.Username,
			Email:     class.CreatedBy.Email,
		},
		CreatedAt: class.CreatedAt,
	}

	if class.UpdatedAt.IsZero() {
		classResponse.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Class retrieved successfully",
		Data:    classResponse,
	})
}


func (k *ClassController) UpdateClass(c *fiber.Ctx) error {
	var req map[string]interface{}

	adminIDClaim, ok := c.Locals("userID").(jwt.MapClaims)
	if !ok {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Unauthorized: invalid token claims",
		}
	}

	adminID, ok := adminIDClaim["id"].(float64)
	if !ok {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Unauthorized: invalid user ID",
		}
	}

	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	allowedFields := map[string]bool{
		"name":     true,
		"code":     true,
		"teachers": true,
		"students": true,
	}

	for key := range req {
		if !allowedFields[key] {
			delete(req, key)
		}
	}

	if req["name"] == "" || req["code"] == "" {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Name and code are required",
		}
	}

	classID := c.Params("id")
	var class models.Class
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Teachers").Preload("Students").Preload("CreatedBy").First(&class, classID).Error; err != nil {
			return err
		}

		if class.CreatedBy.ID != uint(adminID) {
			return &fiber.Error{
				Code:    fiber.StatusUnauthorized,
				Message: "Unauthorized: you are not the creator of this class",
			}
		}

		if name, ok := req["name"]; ok {
			class.Name = name.(string)
		}
		if code, ok := req["code"]; ok {
			class.Code = code.(string)
		}

		if teachers, ok := req["teachers"]; ok {
			teacherIDs := teachers.([]interface{})
			var count int64
			teacherIDsParsed := make([]uint, len(teacherIDs))
			for i, t := range teacherIDs {
				teacherIDsParsed[i] = uint(t.(float64))
			}
	
			if err := tx.Model(&models.Teachers{}).Where("id IN ?", teacherIDsParsed).Count(&count).Error; err != nil {
				return err
			}
	
			if int(count) != len(teacherIDsParsed) {
				return &fiber.Error{
					Code:    fiber.StatusBadRequest,
					Message: "One or more teacher IDs are invalid",
				}
			}
	
			if err := tx.Model(&class).Association("Teachers").Append(teacherIDsParsed); err != nil {
				return err
			}
		}
	
		if students, ok := req["students"]; ok {
			studentIDs := students.([]interface{})
			var count int64
			studentIDsParsed := make([]uint, len(studentIDs))
			for i, s := range studentIDs {
				studentIDsParsed[i] = uint(s.(float64))
			}
	
			if err := tx.Model(&models.Students{}).Where("id IN ?", studentIDsParsed).Count(&count).Error; err != nil {
					return err
			}
			if int(count) != len(studentIDsParsed) {
				return &fiber.Error{
					Code:    fiber.StatusBadRequest,
					Message: "One or more student IDs are invalid",
				}
			}
	
			if err := tx.Model(&class).Association("Students").Append(studentIDsParsed); err != nil {
				return err
			}
		}
		if err := tx.Save(&class).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	teachers := make([]res.TeacherResponse, len(class.Teachers))
	for i, teacher := range class.Teachers {
		teachers[i] = res.TeacherResponse{
			ID:        teacher.ID,
			Name:      teacher.Name,			
			Username:  teacher.Username,
			Email:     teacher.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: teacher.CreatedAt,
		}
	}	

	students := make([]res.StudentResponse, len(class.Students))
	for i, student := range class.Students {
		students[i] = res.StudentResponse{
			ID:        student.ID,
			Name:      student.Name,			
			Username:  student.Username,
			Email:     student.Email,
			Classes:   []res.ClassSummaryResponse{
				{
					ID:   class.ID,
					Name: class.Name,
					Code: class.Code,
				},
			},
			CreatedAt: student.CreatedAt,
		}
	}

	response := res.ClassResponse{
		ID:        class.ID,
		Name:      class.Name,
		Code:      class.Code,
		Teachers:  teachers,
		Students:  students,
		CreatedBy: res.AdminResponse{
			ID:        class.CreatedBy.ID,
			Name:      class.CreatedBy.Name,
			Username:  class.CreatedBy.Username,
			Email:     class.CreatedBy.Email,
			CreatedAt: class.CreatedBy.CreatedAt,
		},
		CreatedAt: class.CreatedAt,
	}

	if class.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	} else {
		response.UpdatedAt = &class.UpdatedAt
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Class updated successfully",
		Data:    response,
	})
}


func (k *ClassController) DeleteClass(c *fiber.Ctx) error {
	classID := c.Params("id")

	var class models.Class

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Teachers").Preload("Students").Preload("CreatedBy").First(&class, classID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&class).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Class deleted successfully",
	})
}

