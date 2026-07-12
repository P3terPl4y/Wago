package handlers

import (
	"App/src/controllers/get"
	"App/src/controllers/update"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func PaymentConfirmId(c fiber.Ctx) error {
	botID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}
	bot, err := get.GetBotByID(botID)
	if err != nil || bot == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Bot no encontrado"})
	}
	if bot.PaymentStatus != "pending" {
		return c.Status(400).JSON(fiber.Map{"error": "El bot no está pendiente de pago"})
	}
	// Cambiar estado a paid (sin iniciar automáticamente)
	if err := update.UpdateBotPaymentStatus(botID, "paid"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al actualizar estado de pago"})
	}
	// El usuario deberá iniciar el bot manualmente desde su dashboard para ver el QR
	return c.JSON(fiber.Map{"status": "ok", "message": "Pago confirmado. El usuario puede iniciar el bot desde su dashboard."})
}
