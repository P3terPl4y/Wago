package handlers

import (
	"App/src/global"
	"time"

	"github.com/gofiber/fiber/v3"
)

func AdminUser(c fiber.Ctx) error {
	rows, err := global.ConfigDB.Query(`SELECT id, username, email, phone, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al obtener usuarios"})
	}
	defer rows.Close()
	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email, phone, role string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &email, &phone, &role, &createdAt); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id": id, "username": username, "email": email, "phone": phone,
			"role": role, "created_at": createdAt,
		})
	}
	return c.JSON(fiber.Map{"users": users})
}
