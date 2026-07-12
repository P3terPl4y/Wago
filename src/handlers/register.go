package handlers

import (
	"App/src/controllers/create"
	"App/src/global"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func Register(c fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}
	if req.Username == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todos los campos son obligatorios"})
	}
	if len(req.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "La contraseña debe tener al menos 6 caracteres"})
	}
	var count int
	err := global.ConfigDB.QueryRow(`SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2 OR phone = $3`,
		req.Username, req.Email, req.Phone).Scan(&count)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error interno"})
	}
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Usuario, email o teléfono ya registrados"})
	}
	_, err = create.CreateUser(req.Username, req.Email, req.Phone, req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al registrar usuario"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
