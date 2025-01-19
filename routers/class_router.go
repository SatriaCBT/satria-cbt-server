package routers


import (
	"satriacbtserver/controllers"
	"satriacbtserver/middleware"

	"github.com/gofiber/fiber/v2"
)


func NewRoutesClass(router fiber.Router, classController *controllers.ClassController){
	app := router.Group("/class")
	app.Post("/create", middleware.AuthenticateToken([]string{"admin"}), classController.CreateClass)
	app.Get("/all", middleware.AuthenticateToken([]string{"admin"}), classController.GetAllClass)
	app.Get("/:id", middleware.AuthenticateToken([]string{"admin"}), classController.GetClassByID)
	app.Get("/code/:code", middleware.AuthenticateToken([]string{"admin"}), classController.GetClassByCode)
	app.Put("/update/:id", middleware.AuthenticateToken([]string{"admin"}), classController.UpdateClass)
	app.Delete("/delete/:id", middleware.AuthenticateToken([]string{"admin"}), classController.DeleteClass)
}