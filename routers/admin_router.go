package routers


import (
	"satriacbtserver/controllers"
	"github.com/gofiber/fiber/v2"
)


func NewRoutesAdmins(router fiber.Router, admincontroller *controllers.AdminController,) {
	app := router.Group("/admin")
	app.Post("/register", admincontroller.RegisterAdmin)
	app.Post("/login", admincontroller.LoginAdmin)
	app.Get("/profile", admincontroller.GetSessionProfileAdmin)
	app.Put("/update/:id", admincontroller.UpdateAdmin)
	app.Delete("/delete/:id", admincontroller.DeleteAdmin)
}