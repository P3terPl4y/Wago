package handlers

import (
	"App/src/controllers/delet"
	"App/src/global"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func DeleteBot(c fiber.Ctx) error {
	botID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}
	if err := delet.DeleteBot(botID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al eliminar bot"})
	}
	global.ActiveMu.Lock()
	delete(global.ActiveBots, botID)
	global.ActiveMu.Unlock()
	return c.JSON(fiber.Map{"status": "ok"})
}
