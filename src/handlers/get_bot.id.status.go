package handlers

import (
	"App/src/controllers/get"
	"App/src/global"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func GetBotIDStatus(c fiber.Ctx) error {
	botID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "ID inválido"})
	}
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	bot, err := get.GetBotByID(botID)
	if err != nil || bot == nil {
		return c.JSON(fiber.Map{"status": "error", "message": "Bot no encontrado"})
	}
	if role != "admin" && bot.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"status": "error", "message": "No autorizado"})
	}
	global.ActiveMu.Lock()
	_, active := global.ActiveBots[botID]
	global.ActiveMu.Unlock()
	if active {
		return c.JSON(fiber.Map{"status": "active"})
	}
	// Verificar estado de pago
	if bot.PaymentStatus == "pending" {
		return c.JSON(fiber.Map{"status": "pending_payment", "id": botID})
	}
	if bot.SessionFile != "" {
		return c.JSON(fiber.Map{"status": "inactive"})
	}
	return c.JSON(fiber.Map{"status": "inactive"})
}
