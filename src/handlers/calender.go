package handlers

import (
	"App/src/global"
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func CreateAppointment(c fiber.Ctx) error {
	// Obtener user_id de la sesión (middleware AuthRequired ya lo puso en locals)
	userID, ok := c.Locals("user_id").(int)
	if !ok || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "No autenticado"})
	}

	// Obtener refresh_token
	var refreshToken string
	err := global.ConfigDB.QueryRow(`
		SELECT refresh_token FROM oauth_tokens
		WHERE user_id = $1 AND provider = 'google'
	`, userID).Scan(&refreshToken)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "El usuario no ha vinculado su calendario. Inicia sesión con Google."})
	}

	// Parsear request
	var req struct {
		ClientEmail   string `json:"clientEmail"`
		Title         string `json:"title"`
		StartDateTime string `json:"startDateTime"`
		EndDateTime   string `json:"endDateTime"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
	}

	// Crear cliente Calendar
	tokenSource := global.GoogleOAuthConfig.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	})
	httpClient := oauth2.NewClient(context.Background(), tokenSource)
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al crear servicio: " + err.Error()})
	}

	// Construir evento
	event := &calendar.Event{
		Summary:     req.Title,
		Description: "Agendado por " + req.ClientEmail,
		Start: &calendar.EventDateTime{
			DateTime: req.StartDateTime,
			TimeZone: "America/Mexico_City",
		},
		End: &calendar.EventDateTime{
			DateTime: req.EndDateTime,
			TimeZone: "America/Mexico_City",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: req.ClientEmail},
		},
	}

	created, err := srv.Events.Insert("primary", event).SendUpdates("all").Do()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al crear evento: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"eventId":     created.Id,
		"htmlLink":    created.HtmlLink,
		"hangoutLink": created.HangoutLink,
		"message":     "Evento creado",
	})
}
