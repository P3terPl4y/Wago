package handlers

import (
	"App/src/controllers/get"
	"App/src/global"
	"App/src/global/functions"
	"App/src/models"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"golang.org/x/crypto/bcrypt"
)

func Login(c fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}
	if (req.Username == "" && req.Email == "") || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Usuario/email y contraseña requeridos"})
	}
	var user *models.User
	if req.Username != "" {
		user, _ = get.GetUserByUsername(req.Username)
	} else {
		user, _ = get.GetUserByUsername(req.Email)
		if user == nil {
			var u models.User
			err := global.ConfigDB.QueryRow(`SELECT id, username, email, phone, password_hash, role, created_at FROM users WHERE email = $1`, req.Email).
				Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
			if err == nil {
				user = &u
			}
		}
	}
	if user == nil {
		functions.LogFailedLogin(c.IP(), "usuario no existe")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		functions.LogFailedLogin(c.IP(), "contraseña incorrecta")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
	}
	sess := session.FromContext(c)
	sess.Set("user_id", user.ID)
	sess.Set("role", user.Role)
	return c.JSON(fiber.Map{"status": "ok"})
}
