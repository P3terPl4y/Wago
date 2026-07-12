package database

import (
	"App/src/global"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	_ "github.com/lib/pq" // PostgreSQL
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func InitDB() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ La variable de entorno DATABASE_URL es obligatoria")
	}
	global.ConfigDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	if err = global.ConfigDB.Ping(); err != nil {
		log.Fatalf("❌ No se puede conectar a PostgreSQL: %v", err)
	}
	global.ConfigDB.SetMaxOpenConns(10)

	// Crear tablas (con nueva columna payment_status)
	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS bots (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		blocked BOOLEAN DEFAULT FALSE,
		session_file TEXT,
		payment_status TEXT DEFAULT 'free',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS prompts (
		bot_id INTEGER PRIMARY KEY REFERENCES bots(id) ON DELETE CASCADE,
		prompt TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS subscriptions (
		bot_id INTEGER PRIMARY KEY REFERENCES bots(id) ON DELETE CASCADE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS chat_history (
		id SERIAL PRIMARY KEY,
		bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
		user_jid TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = global.ConfigDB.Exec(createTables)
	if err != nil {
		log.Fatalf("❌ Error creando tablas: %v", err)
	}
	// Migrar bots existentes (añadir columna si no existe)
	_, _ = global.ConfigDB.Exec(`ALTER TABLE bots ADD COLUMN IF NOT EXISTS payment_status TEXT DEFAULT 'free'`)
	// Crear admin por defecto
	var count int
	err = global.ConfigDB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	if err != nil {
		log.Fatalf("❌ Error consultando admin: %v", err)
	}
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(global.ADMIN_PASS), bcrypt.DefaultCost)
		_, err = global.ConfigDB.Exec(`INSERT INTO users (username, email, phone, password_hash, role) VALUES ($1, $2, $3, $4, 'admin')`,
			global.ADMIN_USERNAME, global.ADMIN_EMAIL, global.ADMIN_PHONE, string(hashed))
		if err != nil {
			log.Fatalf("❌ Error creando admin: %v", err)
		}
		fmt.Printf("✅ Admin creado: %s / %s\n", global.ADMIN_USERNAME, global.ADMIN_PASS)
	}

	// Clave de cifrado
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		log.Fatal("❌ La variable de entorno ENCRYPTION_KEY es obligatoria. Debe ser una cadena hexadecimal de 64 caracteres (32 bytes).")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		log.Fatal("❌ ENCRYPTION_KEY inválida: debe ser hexadecimal de 64 caracteres (32 bytes).")
	}
	global.EncryptionKey = key

	// Redis
	InitRedis()
	var storage fiber.Storage
	secureCookie := os.Getenv("COOKIE_SECURE") == "true"
	redisURL := os.Getenv("REDIS_URL")

	if global.RedisClient != nil && redisURL != "" {
		redisStore := redis.New(redis.Config{URL: redisURL})
		storage = redisStore
		fmt.Println("✅ Sesiones almacenadas en Redis")
	} else {
		log.Println("⚠️ REDIS_URL no configurada o no disponible, sesiones en memoria.")
		storage = nil
	}

	global.SessionMW = session.New(session.Config{
		CookieSecure:   secureCookie,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		IdleTimeout:    global.SESSION_EXPIRATION,
		Storage:        storage,
	})

	fmt.Println("✅ Base de datos PostgreSQL y sesiones inicializadas.")
}
