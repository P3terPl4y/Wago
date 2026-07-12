package handlers

import (
	"App/src/global"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func DeleteUserById(c fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}
	var role string
	err = global.ConfigDB.QueryRow(`SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Usuario no encontrado"})
	}
	if role == "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "No se puede eliminar al administrador"})
	}
	_, err = global.ConfigDB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al eliminar usuario"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
