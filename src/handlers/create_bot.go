package handlers

import (
	"App/src/controllers/create"
	"App/src/controllers/get"
	"App/src/global"
	"App/src/global/functions"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func CreateBot(c fiber.Ctx) error {
	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}
	if req.UserID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}
	user, err := get.GetUserByID(req.UserID)
	if err != nil || user == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Usuario no encontrado"})
	}
	count, err := functions.CountBotsByUser(req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al contar bots"})
	}
	if count >= global.MAX_BOTS {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("El usuario ya tiene %d bots (límite %d)", count, global.MAX_BOTS)})
	}
	sessionFile := fmt.Sprintf("whatsapp_bot%d.db", 0)
	botID, err := create.CreateBot(req.UserID, sessionFile, "free")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al crear bot"})
	}
	newSessionFile := fmt.Sprintf("whatsapp_bot%d.db", botID)
	_, err = global.ConfigDB.Exec(`UPDATE bots SET session_file = $1 WHERE id = $2`, newSessionFile, botID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al actualizar archivo de sesión"})
	}
	// Opcional: iniciar automáticamente? Dejamos que el usuario lo inicie.
	return c.JSON(fiber.Map{"status": "ok", "bot_id": botID})
}
