package handlers

import (
	"App/src/controllers/update"
	"App/src/global"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func BotsIDBlock(c fiber.Ctx) error {
	botID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}
	var req struct {
		Blocked bool `json:"blocked"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}
	if err := update.UpdateBotBlocked(botID, req.Blocked); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al actualizar estado"})
	}
	if req.Blocked {
		global.ActiveMu.Lock()
		delete(global.ActiveBots, botID)
		global.ActiveMu.Unlock()
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
