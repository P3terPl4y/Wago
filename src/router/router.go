package router

import (
	"App/src/global"
	"App/src/handlers"
	"App/src/middleware"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func Router(app *fiber.App) {
	limiterMiddleware := limiter.New(limiter.Config{
		Max:        global.RATE_LIMIT_PER_MINUTE,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Demasiadas peticiones. Intenta de nuevo en un minuto."})
		},
	})
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendFile(`./src/static/index.html`)
	})
	// Rutas públicas
	app.Get("/sing", func(c fiber.Ctx) error {

		return c.SendFile("./src/static/sing.html")
	})
	// Autenticación con Google
	app.Get("/auth/google/login", handlers.GoogleLogin)
	app.Get("/auth/google/callback", handlers.GoogleCallback)
	app.Get("/register", func(c fiber.Ctx) error {

		return c.SendFile("./src/static/register.html")
	})

	app.Post("/register", limiterMiddleware, handlers.Register)

	app.Post("/login", limiterMiddleware, handlers.Login)

	app.Post("/logout", middleware.AuthRequired, handlers.Logout)

	// ============================================================
	// DASHBOARD DE USUARIO (con gestión de prompt y estado)
	// ============================================================
	app.Get("/dashboard", middleware.AuthRequired, handlers.Dashboard)

	// ============================================================
	// INICIAR BOT (POST /start-bot) - CON CONTROL DE PAGO
	// ============================================================
	app.Post("/start-bot", middleware.AuthRequired, limiterMiddleware, handlers.StartBot)

	// ============================================================
	// OBTENER ESTADO DEL BOT (GET /bot/:id/status)
	// ============================================================
	app.Get("/bot/:id/status", middleware.AuthRequired, handlers.GetBotIDStatus)

	// ============================================================
	// ACTUALIZAR PROMPT (PUT /bot/:id/prompt)
	// ============================================================
	app.Put("/bot/:id/prompt", middleware.AuthRequired, handlers.UpdateBotIDPrompt)

	// ============================================================
	// CAMBIAR CONTRASEÑA DEL USUARIO
	// ============================================================
	app.Put("/user/password", middleware.AuthRequired, handlers.UpdatePasswordUser)

	// ============================================================
	// PANEL DE ADMINISTRACIÓN
	// ============================================================
	adminGroup := app.Group("/admin", middleware.AuthRequired, middleware.AdminRequired)

	// Obtener todos los bots con estado y usuario
	adminGroup.Get("/bots/status", handlers.GetBotStatus)

	// Crear un nuevo bot para un usuario específico (admin) - sin pago
	adminGroup.Post("/bots/create", handlers.CreateBot)

	// Confirmar pago de un bot (admin)
	// Confirmar pago de un bot (admin)
	adminGroup.Post("/payments/confirm/:id", handlers.PaymentConfirmId)

	// Interfaz principal de administración (con sección de pagos pendientes)
	adminGroup.Get("/", func(c fiber.Ctx) error {
		return c.SendFile("./src/static/pay_session.html")
	})

	// Rutas existentes de admin (users, bots, block, delete)
	adminGroup.Get("/users", handlers.AdminUser)

	adminGroup.Put("/users/:id/password", handlers.UpdateUserPasswordById)

	adminGroup.Delete("/users/:id", handlers.DeleteUserById)

	adminGroup.Post("/bots/:id/block", handlers.BotsIDBlock)

	adminGroup.Delete("/bots/:id", handlers.DeleteBot)

	// ============================================================
	// RUTA PARA OBTENER BOTS ACTIVOS
	// ============================================================
	app.Get("/active-bots", middleware.AuthRequired, handlers.ActiveBots)
}
