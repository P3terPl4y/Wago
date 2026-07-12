package handlers

import (
	"App/src/controllers/get"
	"App/src/global"

	"github.com/gofiber/fiber/v3"
)

func ActiveBots(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	global.ActiveMu.Lock()
	ids := []int{}
	if role == "admin" {
		for id := range global.ActiveBots {
			ids = append(ids, id)
		}
	} else {
		bots, err := get.GetBotsByUser(userID)
		if err == nil {
			for _, b := range bots {
				if _, ok := global.ActiveBots[b.ID]; ok {
					ids = append(ids, b.ID)
				}
			}
		}
	}
	global.ActiveMu.Unlock()
	return c.JSON(fiber.Map{"bots": ids})
}
