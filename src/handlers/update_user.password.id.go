package handlers

import (
	"App/src/controllers/update"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func UpdateUserPasswordById(c fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}
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
