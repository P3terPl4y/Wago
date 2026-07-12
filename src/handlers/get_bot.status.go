package handlers

import (
	"App/src/controllers/get"
	"App/src/global"

	"github.com/gofiber/fiber/v3"
)

func GetBotStatus(c fiber.Ctx) error {
	bots, err := get.GetAllBots()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al obtener bots"})
	}
	result := []map[string]interface{}{}
	for _, b := range bots {
		user, err := get.GetUserByID(b.UserID)
		username := ""
		if err == nil && user != nil {
			username = user.Username
		}
		global.ActiveMu.Lock()
		_, isActive := global.ActiveBots[b.ID]
		global.ActiveMu.Unlock()
		prompt, _ := get.GetPrompt(b.ID)
		result = append(result, map[string]interface{}{
			"id":             b.ID,
			"user_id":        b.UserID,
			"username":       username,
			"blocked":        b.Blocked,
			"active":         isActive,
			"payment_status": b.PaymentStatus,
			"session_file":   b.SessionFile,
			"prompt":         prompt,
			"created_at":     b.CreatedAt,
		})
	}
	return c.JSON(fiber.Map{"bots": result})
}
