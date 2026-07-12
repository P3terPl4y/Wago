package handlers

import (
	"App/src/controllers/create"
	"App/src/controllers/get"
	"App/src/global"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/skip2/go-qrcode"
)

func StartBot(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	bots, err := get.GetBotsByUser(userID)
	if err != nil {
		return c.JSON(fiber.Map{"status": "error", "message": "Error al verificar bots"})
	}
	var botID int
	var paymentStatus string

	if len(bots) > 0 {
		bot := bots[0]
		botID = bot.ID
		paymentStatus = bot.PaymentStatus

		if bot.Blocked {
			return c.JSON(fiber.Map{"status": "error", "message": "El bot está bloqueado. Contacta al administrador."})
		}
		// Verificar estado de pago para usuarios normales
		if role != "admin" {
			if paymentStatus == "pending" {
				return c.JSON(fiber.Map{"status": "pending_payment", "id": botID, "payment_status": "pending", "message": "Pago pendiente. Espera confirmación del administrador."})
			}
			if paymentStatus != "paid" && paymentStatus != "free" {
				return c.JSON(fiber.Map{"status": "error", "message": "Estado de pago inválido. Contacta al administrador."})
			}
		}
		// Si es admin o el pago está OK, iniciar (si no está activo)
		global.ActiveMu.Lock()
		_, active := global.ActiveBots[botID]
		global.ActiveMu.Unlock()
		if active {
			return c.JSON(fiber.Map{"status": "session_exists", "id": botID})
		}
		// Iniciar el bot (si no está activo)
		qrResult := make(chan string, 1)
		go create.InitBot(botID, qrResult)
		select {
		case result := <-qrResult:
			switch result {
			case "SESSION_EXISTS":
				return c.JSON(fiber.Map{"status": "session_exists", "id": botID})
			case "TIMEOUT":
				return c.JSON(fiber.Map{"status": "timeout", "id": botID})
			default:
				png, err := qrcode.Encode(result, qrcode.Medium, 256)
				if err != nil {
					return c.JSON(fiber.Map{"status": "error", "message": "Error generando QR"})
				}
				return c.JSON(fiber.Map{"status": "qr", "qr": base64.StdEncoding.EncodeToString(png), "id": botID})
			}
		case <-time.After(35 * time.Second):
			return c.JSON(fiber.Map{"status": "error", "message": "Tiempo de espera agotado"})
		}
	} else {
		// No tiene bot: crear uno
		// Si es admin, crear gratis y lanzar; si es usuario normal, crear pendiente de pago
		var status string
		if role == "admin" {
			status = "free"
		} else {
			status = "pending"
		}
		sessionFile := fmt.Sprintf("whatsapp_bot%d.db", 0) // temporal
		newID, err := create.CreateBot(userID, sessionFile, status)
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "message": "Error al crear bot"})
		}
		botID = newID
		newSessionFile := fmt.Sprintf("whatsapp_bot%d.db", botID)
		_, err = global.ConfigDB.Exec(`UPDATE bots SET session_file = $1 WHERE id = $2`, newSessionFile, botID)
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "message": "Error al actualizar archivo de sesión"})
		}
		if role == "admin" {
			// Admin: iniciar inmediatamente
			qrResult := make(chan string, 1)
			go create.InitBot(botID, qrResult)
			select {
			case result := <-qrResult:
				switch result {
				case "SESSION_EXISTS":
					return c.JSON(fiber.Map{"status": "session_exists", "id": botID})
				case "TIMEOUT":
					return c.JSON(fiber.Map{"status": "timeout", "id": botID})
				default:
					png, err := qrcode.Encode(result, qrcode.Medium, 256)
					if err != nil {
						return c.JSON(fiber.Map{"status": "error", "message": "Error generando QR"})
					}
					return c.JSON(fiber.Map{"status": "qr", "qr": base64.StdEncoding.EncodeToString(png), "id": botID})
				}
			case <-time.After(35 * time.Second):
				return c.JSON(fiber.Map{"status": "error", "message": "Tiempo de espera agotado"})
			}
		} else {
			// Usuario normal: devolver pendiente
			return c.JSON(fiber.Map{"status": "pending_payment", "id": botID, "payment_status": "pending", "message": "Bot creado. Se requiere confirmación de pago por parte del administrador."})
		}
	}
	return c.JSON(fiber.Map{"status": "error", "message": "No se pudo procesar la solicitud"})
}
