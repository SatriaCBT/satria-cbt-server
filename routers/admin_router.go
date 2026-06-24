package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)


func NewRoutesAdmins(router fiber.Router, admincontroller *controllers.AdminController,) {
	app := router.Group("/admin")
	app.Post("/register", admincontroller.RegisterAdmin)
	app.Post("/login", admincontroller.LoginAdmin)
	app.Get("/profile", middleware.AuthenticateToken([]string{"admin"}), admincontroller.GetSessionProfileAdmin)
	app.Put("/update/:id", middleware.AuthenticateToken([]string{"admin"}), admincontroller.UpdateAdmin)
	app.Delete("/delete/:id", middleware.AuthenticateToken([]string{"admin"}), admincontroller.DeleteAdmin)
}