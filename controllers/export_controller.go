package controllers

import (
	"fmt"
	"strconv"

	"github.com/Satria-CBT/satria-cbt-server/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ExportController struct {
	db *gorm.DB
}

func NewExportController(db *gorm.DB) *ExportController {
	return &ExportController{db: db}
}

func (e *ExportController) ExportExamResults(c *fiber.Ctx) error {
	examID := c.Params("examId")

	var exam models.Exam
	if err := e.db.First(&exam, examID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	var attempts []models.ExamAttempt
	if err := e.db.Preload("Student").Where("exam_id = ? AND status != ?", examID, models.AttemptInProgress).Order("score DESC").Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Hasil Ujian"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"No", "Nama Siswa", "Username", "Nilai", "Benar", "Salah", "Status", "Waktu Selesai"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheet, "A1", "H1", style)

	for i, a := range attempts {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.Student.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), a.Student.Username)

		status := "Lulus"
		if a.Score != nil && exam.PassingScore > 0 && *a.Score < exam.PassingScore {
			status = "Tidak Lulus"
		}
		if a.Score == nil {
			status = "-"
		}

		scoreStr := "-"
		if a.Score != nil {
			scoreStr = strconv.Itoa(*a.Score)
		}

		endTimeStr := "-"
		if a.EndTime != nil {
			endTimeStr = a.EndTime.Format("02 Jan 2006 15:04")
		}

		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), scoreStr)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), a.TotalCorrect)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), a.TotalWrong)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), status)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), endTimeStr)
	}

	colWidths := []float64{5, 25, 15, 10, 8, 8, 12, 20}
	for i, w := range colWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	c.Response().Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="hasil_ujian_%s.xlsx"`, exam.Title))

	buf, err := f.WriteToBuffer()
	if err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.SendStream(buf)
}

func (e *ExportController) ExportStudentResults(c *fiber.Ctx) error {
	studentID := c.Params("studentId")

	var student models.Students
	if err := e.db.First(&student, studentID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Student not found"}
	}

	var attempts []models.ExamAttempt
	if err := e.db.Preload("Exam").Where("student_id = ? AND status != ?", studentID, models.AttemptInProgress).Order("created_at DESC").Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Hasil Siswa"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Nama Siswa: "+student.Name)
	f.SetCellValue(sheet, "A2", "Username: "+student.Username)
	f.SetCellValue(sheet, "A3", "Email: "+student.Email)

	headers := []string{"No", "Ujian", "Nilai", "Benar", "Salah", "Status", "Tanggal"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue(sheet, cell, h)
	}

	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A5", fmt.Sprintf("%c5", 64+len(headers)), style)

	for i, a := range attempts {
		row := i + 6
		examTitle := a.Exam.Title
		scoreStr := "-"
		if a.Score != nil {
			scoreStr = strconv.Itoa(*a.Score)
		}

		status := "Lulus"
		if a.Score != nil && a.Exam.PassingScore > 0 && *a.Score < a.Exam.PassingScore {
			status = "Tidak Lulus"
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), examTitle)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), scoreStr)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), a.TotalCorrect)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), a.TotalWrong)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), status)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), a.CreatedAt.Format("02 Jan 2006"))
	}

	c.Response().Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="hasil_siswa_%s.xlsx"`, student.Name))

	buf, err := f.WriteToBuffer()
	if err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.SendStream(buf)
}
