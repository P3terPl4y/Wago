package handlers

import (
	"App/src/global"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"golang.org/x/oauth2"
)

// GoogleLogin inicia flujo OAuth2
func GoogleLogin(c fiber.Ctx) error {
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   false, // pon true en producción con HTTPS
		SameSite: "Lax",
		MaxAge:   600,
	})

	url := global.GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return c.Redirect().To(url)
}

// GoogleCallback maneja el retorno de Google
func GoogleCallback(c fiber.Ctx) error {
	// Validar CSRF
	state := c.Query("state")
	cookieState := c.Cookies("oauth_state")
	if state == "" || state != cookieState {
		return c.Status(400).JSON(fiber.Map{"error": "Estado inválido"})
	}

	code := c.Query("code")
	token, err := global.GoogleOAuthConfig.Exchange(c.ClientHelloInfo().Context(), code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al obtener token: " + err.Error()})
	}

	// Obtener datos del usuario
	client := global.GoogleOAuthConfig.Client(c.ClientHelloInfo().Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al obtener datos: " + err.Error()})
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al decodificar"})
	}

	// Buscar o crear usuario en tabla users (sin modificar estructura)
	var userID int
	var role string
	err = global.ConfigDB.QueryRow(`SELECT id, role FROM users WHERE email = $1`, userInfo.Email).Scan(&userID, &role)
	if err != nil {
		// No existe: lo creamos. Usamos el email como username (debe ser único)
		// Generamos un password_hash dummy (no se usará para login con Google)
		dummyHash := "$2a$10$dummyhash"
		err = global.ConfigDB.QueryRow(`
			INSERT INTO users (username, email, phone, password_hash, role)
			VALUES ($1, $2, $3, $4, 'user')
			RETURNING id, role
		`, userInfo.Email, userInfo.Email, "", dummyHash).Scan(&userID, &role)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error al crear usuario: " + err.Error()})
		}
	}

	// Guardar refresh_token en oauth_tokens
	_, err = global.ConfigDB.Exec(`
		INSERT INTO oauth_tokens (user_id, provider, refresh_token, updated_at)
		VALUES ($1, 'google', $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, provider) DO UPDATE
		SET refresh_token = $2, updated_at = CURRENT_TIMESTAMP
	`, userID, token.RefreshToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al guardar token: " + err.Error()})
	}

	// Establecer sesión (igual que en Login)
	sess := session.FromContext(c)
	sess.Set("user_id", userID)
	sess.Set("role", role)

	// Redirigir al dashboard (o donde quieras)
	return c.Redirect().To("/dashboard")
}
