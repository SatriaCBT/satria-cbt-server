package controllers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/Satria-CBT/satria-cbt-server/middleware"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ExamUpdate struct {
	ExamID    uint   `json:"examId"`
	AttemptID uint   `json:"attemptId"`
	StudentID uint   `json:"studentId"`
	StudentName string `json:"studentName"`
	Event     string `json:"event"`
	Score     *int   `json:"score,omitempty"`
}

type WSHub struct {
	mu       sync.RWMutex
	examConns map[uint]map[*websocket.Conn]bool
}

var hub = &WSHub{
	examConns: make(map[uint]map[*websocket.Conn]bool),
}

func (h *WSHub) JoinExam(examID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.examConns[examID] == nil {
		h.examConns[examID] = make(map[*websocket.Conn]bool)
	}
	h.examConns[examID][conn] = true
}

func (h *WSHub) LeaveExam(examID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.examConns[examID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.examConns, examID)
		}
	}
}

func (h *WSHub) BroadcastToExam(examID uint, msg ExamUpdate) {
	h.mu.RLock()
	conns := h.examConns[examID]
	h.mu.RUnlock()

	data, _ := json.Marshal(msg)
	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("WS write error: %v", err)
			conn.Close()
			go h.LeaveExam(examID, conn)
		}
	}
}

func InitWSRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/ws/exams/:examId", middleware.AuthenticateToken([]string{"admin", "teacher"}), func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("db", db)
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, websocket.New(func(conn *websocket.Conn) {
		examID := conn.Params("examId")
		examIDUint := parseUint(examID)
		if examIDUint == 0 {
			conn.Close()
			return
		}

		hub.JoinExam(examIDUint, conn)
		defer hub.LeaveExam(examIDUint, conn)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
}

func BroadcastAttemptEvent(attemptID, examID, studentID uint, studentName, event string, score *int) {
	hub.BroadcastToExam(examID, ExamUpdate{
		ExamID:      examID,
		AttemptID:   attemptID,
		StudentID:   studentID,
		StudentName: studentName,
		Event:       event,
		Score:       score,
	})
}

func ExtractUserIDFromWS(c *websocket.Conn) uint {
	claims := c.Locals("userID").(jwt.MapClaims)
	return uint(claims["id"].(float64))
}
