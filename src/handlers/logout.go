package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func Logout(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if err := sess.Destroy(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al cerrar sesión"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
