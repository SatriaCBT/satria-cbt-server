package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/Satria-CBT/satria-cbt-server/services/exam/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ExamHandler struct {
	db *gorm.DB
}

func NewExamHandler(db *gorm.DB) *ExamHandler {
	return &ExamHandler{db: db}
}

func getUserID(c *fiber.Ctx) uint {
	return parseUint(c.Get("X-User-ID", "0"))
}

func getUserRole(c *fiber.Ctx) string {
	return c.Get("X-User-Role", "student")
}

// ---- Subjects ----

func (h *ExamHandler) CreateSubject(c *fiber.Ctx) error {
	var req models.Subject
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"code": 201, "message": "Subject created", "data": req})
}

func (h *ExamHandler) GetSubjects(c *fiber.Ctx) error {
	var list []models.Subject
	h.db.Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

func (h *ExamHandler) UpdateSubject(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.Subject
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var sub models.Subject
	if err := h.db.First(&sub, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if req.Name != "" {
		sub.Name = req.Name
	}
	if req.Code != "" {
		sub.Code = req.Code
	}
	h.db.Save(&sub)
	return c.JSON(fiber.Map{"code": 200, "data": sub})
}

func (h *ExamHandler) DeleteSubject(c *fiber.Ctx) error {
	h.db.Delete(&models.Subject{}, c.Params("id"))
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

// ---- Questions ----

func (h *ExamHandler) CreateQuestion(c *fiber.Ctx) error {
	var req models.Question
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	req.CreatedByID = getUserID(c)
	if req.Points == 0 {
		req.Points = 1
	}
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"code": 201, "data": req})
}

func (h *ExamHandler) GetQuestions(c *fiber.Ctx) error {
	var list []models.Question
	q := h.db
	if s := c.Query("subjectId"); s != "" {
		q = q.Where("subject_id = ?", s)
	}
	if getUserRole(c) == "teacher" {
		q = q.Where("created_by_id = ?", getUserID(c))
	}
	q.Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

func (h *ExamHandler) UpdateQuestion(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.Question
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var q models.Question
	if err := h.db.First(&q, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if req.Question != "" {
		q.Question = req.Question
	}
	if req.CorrectAnswer != "" {
		q.CorrectAnswer = req.CorrectAnswer
	}
	if req.Points > 0 {
		q.Points = req.Points
	}
	if req.Explanation != "" {
		q.Explanation = req.Explanation
	}
	h.db.Save(&q)
	return c.JSON(fiber.Map{"code": 200, "data": q})
}

func (h *ExamHandler) DeleteQuestion(c *fiber.Ctx) error {
	h.db.Delete(&models.Question{}, c.Params("id"))
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

// ---- Exams ----

func (h *ExamHandler) CreateExam(c *fiber.Ctx) error {
	var req models.Exam
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	req.CreatedByID = getUserID(c)
	req.Status = models.ExamDraft
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"code": 201, "data": req})
}

func (h *ExamHandler) GetExams(c *fiber.Ctx) error {
	var list []models.Exam
	q := h.db
	switch getUserRole(c) {
	case "teacher":
		q = q.Where("created_by_id = ?", getUserID(c))
	case "student":
		q = q.Where("status = ? AND (start_time IS NULL OR start_time <= NOW()) AND (end_time IS NULL OR end_time >= NOW())", models.ExamPublished)
	}
	q.Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

func (h *ExamHandler) GetExam(c *fiber.Ctx) error {
	var exam models.Exam
	if err := h.db.First(&exam, c.Params("id")).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"code": 200, "data": exam})
}

func (h *ExamHandler) UpdateExam(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.Exam
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var exam models.Exam
	if err := h.db.First(&exam, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if exam.Status != models.ExamDraft {
		return fiber.NewError(fiber.StatusBadRequest, "Can only update draft exams")
	}
	if req.Title != "" {
		exam.Title = req.Title
	}
	if req.Duration > 0 {
		exam.Duration = req.Duration
	}
	if req.PassingScore > 0 {
		exam.PassingScore = req.PassingScore
	}
	if req.ClassID > 0 {
		exam.ClassID = req.ClassID
	}
	exam.StartTime = req.StartTime
	exam.EndTime = req.EndTime
	exam.ShuffleQuestions = req.ShuffleQuestions
	exam.ShowResult = req.ShowResult
	h.db.Save(&exam)
	return c.JSON(fiber.Map{"code": 200, "data": exam})
}

func (h *ExamHandler) DeleteExam(c *fiber.Ctx) error {
	h.db.Delete(&models.Exam{}, c.Params("id"))
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

func (h *ExamHandler) PublishExam(c *fiber.Ctx) error {
	var exam models.Exam
	if err := h.db.Preload("Questions").First(&exam, c.Params("id")).Error; err != nil {
		return fiber.ErrNotFound
	}
	if exam.Status != models.ExamDraft {
		return fiber.NewError(fiber.StatusBadRequest, "Not in draft")
	}
	if len(exam.Questions) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No questions")
	}
	h.db.Model(&exam).Updates(map[string]interface{}{
		"status":      models.ExamPublished,
		"total_points": h.calcTotalPoints(exam.ID),
	})
	return c.JSON(fiber.Map{"code": 200, "message": "Published"})
}

func (h *ExamHandler) calcTotalPoints(examID uint) int {
	var questions []models.Question
	h.db.Raw(`SELECT q.* FROM questions q JOIN exam_questions eq ON eq.question_id = q.id WHERE eq.exam_id = ?`, examID).Scan(&questions)
	total := 0
	for _, q := range questions {
		total += q.Points
	}
	return total
}

func (h *ExamHandler) AddQuestions(c *fiber.Ctx) error {
	var req struct {
		QuestionIDs []uint `json:"questionIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	for i, qid := range req.QuestionIDs {
		h.db.Exec("INSERT INTO exam_questions (exam_id, question_id, order_index) VALUES (?, ?, ?) ON CONFLICT DO NOTHING", c.Params("id"), qid, i)
	}
	return c.JSON(fiber.Map{"code": 200, "message": "Questions added"})
}

func (h *ExamHandler) RemoveQuestion(c *fiber.Ctx) error {
	h.db.Exec("DELETE FROM exam_questions WHERE exam_id = ? AND question_id = ?", c.Params("id"), c.Params("questionId"))
	return c.JSON(fiber.Map{"code": 200, "message": "Removed"})
}

// ---- Attempts ----

func (h *ExamHandler) StartAttempt(c *fiber.Ctx) error {
	studentID := getUserID(c)
	examID := c.Params("examId")

	var exam models.Exam
	if err := h.db.First(&exam, examID).Error; err != nil {
		return fiber.ErrNotFound
	}
	if exam.Status != models.ExamPublished {
		return fiber.NewError(fiber.StatusBadRequest, "Exam not available")
	}

	var count int64
	h.db.Model(&models.ExamAttempt{}).Where("exam_id = ? AND student_id = ?", examID, studentID).Count(&count)
	if int(count) >= exam.MaxAttempts {
		return fiber.NewError(fiber.StatusBadRequest, "Max attempts reached")
	}

	attempt := models.ExamAttempt{
		ExamID:    exam.ID,
		StudentID: studentID,
		StartTime: time.Now(),
		Status:    models.AttemptInProgress,
	}
	h.db.Create(&attempt)

	var qids []struct{ QuestionID uint }
	if exam.ShuffleQuestions {
		h.db.Raw("SELECT question_id FROM exam_questions WHERE exam_id = ? ORDER BY RANDOM()", examID).Scan(&qids)
	} else {
		h.db.Raw("SELECT question_id FROM exam_questions WHERE exam_id = ? ORDER BY order_index", examID).Scan(&qids)
	}
	for _, q := range qids {
		h.db.Create(&models.ExamAnswer{AttemptID: attempt.ID, QuestionID: q.QuestionID})
	}

	return c.Status(201).JSON(fiber.Map{"code": 201, "data": attempt})
}

func (h *ExamHandler) SubmitAttempt(c *fiber.Ctx) error {
	attemptID := c.Params("attemptId")
	studentID := getUserID(c)

	var req struct {
		Answers []struct {
			QuestionID uint   `json:"questionId"`
			Answer     string `json:"answer"`
		} `json:"answers"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	var attempt models.ExamAttempt
	if err := h.db.First(&attempt, attemptID).Error; err != nil {
		return fiber.ErrNotFound
	}
	if attempt.StudentID != studentID || attempt.Status != models.AttemptInProgress {
		return fiber.NewError(fiber.StatusForbidden, "Cannot submit")
	}

	correct, wrong, score := 0, 0, 0
	for _, a := range req.Answers {
		var q models.Question
		if err := h.db.First(&q, a.QuestionID).Error; err != nil {
			continue
		}
		isCorrect := (q.Type == models.QCMultipleChoice || q.Type == models.QCTrueFalse) && q.CorrectAnswer == a.Answer
		pts := 0
		if isCorrect {
			pts = q.Points
			correct++
		} else {
			wrong++
		}
		score += pts
		h.db.Model(&models.ExamAnswer{}).Where("attempt_id = ? AND question_id = ?", attemptID, a.QuestionID).Updates(map[string]interface{}{
			"answer": a.Answer, "is_correct": isCorrect, "points": pts,
		})
	}

	now := time.Now()
	h.db.Model(&attempt).Updates(map[string]interface{}{
		"end_time": &now, "score": score, "status": models.AttemptCompleted,
		"total_correct": correct, "total_wrong": wrong,
	})

	return c.JSON(fiber.Map{"code": 200, "message": "Submitted", "data": fiber.Map{"score": score, "correct": correct, "wrong": wrong}})
}

func (h *ExamHandler) SaveProgress(c *fiber.Ctx) error {
	attemptID := c.Params("attemptId")
	var req struct {
		Answers []struct {
			QuestionID uint   `json:"questionId"`
			Answer     string `json:"answer"`
		} `json:"answers"`
		CurrentQuestionIdx int `json:"currentQuestionIdx"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	for _, a := range req.Answers {
		h.db.Model(&models.ExamAnswer{}).Where("attempt_id = ? AND question_id = ?", attemptID, a.QuestionID).Update("answer", a.Answer)
	}
	h.db.Model(&models.ExamAttempt{}).Where("id = ?", attemptID).Update("current_question_idx", req.CurrentQuestionIdx)
	return c.JSON(fiber.Map{"code": 200, "message": "Saved"})
}

func (h *ExamHandler) ResumeAttempt(c *fiber.Ctx) error {
	var attempt models.ExamAttempt
	if err := h.db.Preload("Answers").First(&attempt, c.Params("attemptId")).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"code": 200, "data": attempt})
}

func (h *ExamHandler) ReviewAttempt(c *fiber.Ctx) error {
	var attempt models.ExamAttempt
	if err := h.db.Preload("Answers").First(&attempt, c.Params("attemptId")).Error; err != nil {
		return fiber.ErrNotFound
	}
	showResult := attempt.Status != models.AttemptInProgress
	return c.JSON(fiber.Map{"code": 200, "data": fiber.Map{"attempt": attempt, "showResult": showResult}})
}

func (h *ExamHandler) LogTabSwitch(c *fiber.Ctx) error {
	h.db.Model(&models.ExamAttempt{}).Where("id = ?", c.Params("attemptId")).UpdateColumn("tab_switch_count", gorm.Expr("tab_switch_count + 1"))
	return c.JSON(fiber.Map{"code": 200, "message": "Logged"})
}

func (h *ExamHandler) GradeEssay(c *fiber.Ctx) error {
	var req struct {
		QuestionID uint `json:"questionId"`
		Points     int  `json:"points"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	h.db.Model(&models.ExamAnswer{}).Where("attempt_id = ? AND question_id = ?", c.Params("attemptId"), req.QuestionID).Update("points", req.Points)
	return c.JSON(fiber.Map{"code": 200, "message": "Graded"})
}

func (h *ExamHandler) GetAttemptsByExam(c *fiber.Ctx) error {
	var list []models.ExamAttempt
	h.db.Where("exam_id = ?", c.Params("examId")).Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

func (h *ExamHandler) GetMyAttempts(c *fiber.Ctx) error {
	var list []models.ExamAttempt
	h.db.Where("student_id = ?", getUserID(c)).Order("created_at DESC").Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

// ---- Dashboard ----

func (h *ExamHandler) DashboardStats(c *fiber.Ctx) error {
	var exams, subjects, attempts, ongoing int64
	h.db.Model(&models.Exam{}).Count(&exams)
	h.db.Model(&models.Subject{}).Count(&subjects)
	h.db.Model(&models.ExamAttempt{}).Count(&attempts)
	h.db.Model(&models.ExamAttempt{}).Where("status = ?", models.AttemptInProgress).Count(&ongoing)
	return c.JSON(fiber.Map{"code": 200, "data": fiber.Map{"totalExams": exams, "totalSubjects": subjects, "totalAttempts": attempts, "ongoingAttempts": ongoing}})
}

func (h *ExamHandler) ExamStats(c *fiber.Ctx) error {
	var scores []struct{ Score int }
	h.db.Model(&models.ExamAttempt{}).Select("COALESCE(score, 0) as score").Where("exam_id = ? AND status != ?", c.Params("examId"), models.AttemptInProgress).Scan(&scores)
	if len(scores) == 0 {
		return c.JSON(fiber.Map{"code": 200, "data": fiber.Map{"message": "No data"}})
	}
	highest, lowest, total := 0, scores[0].Score, 0
	for _, s := range scores {
		total += s.Score
		if s.Score > highest {
			highest = s.Score
		}
		if s.Score < lowest {
			lowest = s.Score
		}
	}
	return c.JSON(fiber.Map{"code": 200, "data": fiber.Map{"average": float64(total) / float64(len(scores)), "highest": highest, "lowest": lowest, "total": len(scores)}})
}

// ---- Export ----

func (h *ExamHandler) ExportExamResults(c *fiber.Ctx) error {
	var attempts []models.ExamAttempt
	h.db.Where("exam_id = ? AND status != ?", c.Params("examId"), models.AttemptInProgress).Order("score DESC").Find(&attempts)

	f := excelize.NewFile()
	defer f.Close()
	f.SetCellValue("Sheet1", "A1", "Nama")
	f.SetCellValue("Sheet1", "B1", "Nilai")
	for i, a := range attempts {
		row := i + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), a.ID)
		if a.Score != nil {
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), *a.Score)
		}
	}

	c.Response().Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header.Set("Content-Disposition", "attachment; filename=results.xlsx")
	buf, _ := f.WriteToBuffer()
	return c.SendStream(buf)
}

// ---- WebSocket ----

type WSHub struct {
	mu       sync.RWMutex
	conns    map[uint]map[*websocket.Conn]bool
}

var hub = &WSHub{conns: make(map[uint]map[*websocket.Conn]bool)}

func (h *WSHub) Join(examID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[examID] == nil {
		h.conns[examID] = make(map[*websocket.Conn]bool)
	}
	h.conns[examID][conn] = true
}

func (h *WSHub) Leave(examID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[examID], conn)
}

type WSMsg struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func (h *ExamHandler) HandleWS(conn *websocket.Conn) {
	examID := parseUint(conn.Params("examId"))
	if examID == 0 {
		conn.Close()
		return
	}
	hub.Join(examID, conn)
	defer hub.Leave(examID, conn)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func parseUint(s string) uint {
	var id uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + uint(c-'0')
		}
	}
	return id
}

// ---- Init WS routes ----

func InitWS(app *fiber.App, h *ExamHandler) {
	app.Get("/ws/exams/:examId", websocket.New(h.HandleWS))
}

func (h *ExamHandler) GetClasses(c *fiber.Ctx) error {
	var list []models.Class
	h.db.Find(&list)
	return c.JSON(fiber.Map{"code": 200, "data": list})
}

func (h *ExamHandler) CreateClass(c *fiber.Ctx) error {
	var req models.Class
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	req.CreatedByID = getUserID(c)
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"code": 201, "data": req})
}
