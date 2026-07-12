package global

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

var (
	WaapiBase  = "https://waapi.app/api/v1/instances"
	WaapiToken = "gtF6Hm2kZbAx2bcWMMqFAJTuESiKn9aKAjWbeyAm8765281a"
	IaURL      = "https://apifreellm.com/api/v1/chat"
	IaKey      = "apf_e9tf41wgk7am0l2499yr2eam"
)
var (
	Containers   = make(map[int]*sqlstore.Container)
	ContainersMu sync.Mutex
	ActiveBots   = make(map[int]bool)
	ActiveMu     sync.Mutex
)

const (
	ADMIN_USERNAME = "admin"
	ADMIN_EMAIL    = "admin@example.com"
	ADMIN_PHONE    = "+1234567890"
	ADMIN_PASS     = "admin123"
)
const (
	MAX_BOTS              = 50
	MAX_CONNECT_RETRIES   = 5
	SUBSCRIPTION_DURATION = 7 * 24 * time.Hour
	MAX_HISTORY           = 10
	SESSION_EXPIRATION    = 1 * time.Hour
	RATE_LIMIT_PER_MINUTE = 10
	MAX_PROMPT_LENGTH     = 5000
)

var (
	ConfigDB      *sql.DB
	RedisClient   *goredis.Client
	SessionMW     fiber.Handler
	EncryptionKey []byte
)
