package main

import (
	"App/src/database"
	"App/src/global"
	"App/src/middleware"
	"App/src/router"
	"fmt"
	"log"

	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {
	global.GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// ============================================================
// SERVIDOR WEB
// ============================================================
func main() {
	database.InitDB()

	app := fiber.New(fiber.Config{
		TrustProxy: true,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": "Ha ocurrido un error. Inténtelo de nuevo más tarde."})
		},
	})

	app.Use(logger.New(logger.Config{
		Format: "${time} - ${method} ${path} ${status}\n",
	}))
	app.Use(middleware.SecurityHeaders)
	app.Use(global.SessionMW)
	app.Use(static.New("./src/static/"))
	router.Router(app)
	// ============================================================
	// INICIO DEL SERVIDOR
	// ============================================================
	fmt.Printf("🚀 Servidor iniciando en http://localhost:3000\n")
	if err := app.Listen("127.0.0.1:3000"); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
}
