package handlers

import (
	"App/src/controllers/update"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func UpdatePasswordUser(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	var req struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}
	if len(req.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "La contraseña debe tener al menos 6 caracteres"})
	}

	if err := update.UpdateUserPassword(userID, req.Password); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al actualizar contraseña"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
