package get

import (
	"App/src/controllers/encrypt"
	"App/src/global"
	"App/src/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func GetExpiration(botID int) (time.Time, error) {
	var expiresAt time.Time
	err := global.ConfigDB.QueryRow(`SELECT expires_at FROM subscriptions WHERE bot_id = $1`, botID).Scan(&expiresAt)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return expiresAt, err
}
func GetPrompt(botID int) (string, error) {
	var prompt string
	err := global.ConfigDB.QueryRow(`SELECT prompt FROM prompts WHERE bot_id = $1`, botID).Scan(&prompt)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return prompt, err
}
func GetBotsByUser(userID int) ([]models.Bot, error) {
	rows, err := global.ConfigDB.Query(`SELECT id, user_id, blocked, session_file, payment_status, created_at FROM bots WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bots []models.Bot
	for rows.Next() {
		var b models.Bot
		if err := rows.Scan(&b.ID, &b.UserID, &b.Blocked, &b.SessionFile, &b.PaymentStatus, &b.CreatedAt); err != nil {
			return nil, err
		}
		bots = append(bots, b)
	}
	return bots, nil
}
func GetContainer(botID int) *sqlstore.Container {
	global.ContainersMu.Lock()
	defer global.ContainersMu.Unlock()
	if c, ok := global.Containers[botID]; ok {
		return c
	}
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "WARN", true)
	dbFile := fmt.Sprintf("whatsapp_bot%d.db", botID)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		f, err := os.OpenFile(dbFile, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_sync=NORMAL&_foreign_keys=on&_busy_timeout=5000", dbFile)
	container, err := sqlstore.New(ctx, "sqlite3", dsn, dbLog)
	if err != nil {
		panic(err)
	}
	global.Containers[botID] = container
	fmt.Printf("✅ Base de datos para Bot %d inicializada\n", botID)
	return container
}
func GetChatHistory(botID int, userJID string, limit int) ([]models.ChatMessage, error) {
	if global.RedisClient != nil {
		key := fmt.Sprintf("chat:%d:%s", botID, userJID)
		ctx := context.Background()
		vals, err := global.RedisClient.LRange(ctx, key, -int64(limit), -1).Result()
		if err == nil && len(vals) > 0 {
			messages := []models.ChatMessage{}
			for _, v := range vals {
				var msg map[string]string
				if err := json.Unmarshal([]byte(v), &msg); err == nil {
					messages = append(messages, models.ChatMessage{Role: msg["role"], Content: msg["content"]})
				}
			}
			for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
				messages[i], messages[j] = messages[j], messages[i]
			}
			return messages, nil
		}
	}
	rows, err := global.ConfigDB.Query(`SELECT role, content FROM chat_history WHERE bot_id = $1 AND user_jid = $2 ORDER BY created_at DESC LIMIT $3`,
		botID, userJID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []models.ChatMessage
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}
		decrypted, err := encrypt.DecryptContent(content)
		if err != nil {
			decrypted = content
		}
		messages = append(messages, models.ChatMessage{Role: role, Content: decrypted})
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	if global.RedisClient != nil && len(messages) > 0 {
		key := fmt.Sprintf("chat:%d:%s", botID, userJID)
		ctx := context.Background()
		global.RedisClient.Del(ctx, key)
		for _, msg := range messages {
			msgBytes, _ := json.Marshal(map[string]string{"role": msg.Role, "content": msg.Content})
			global.RedisClient.RPush(ctx, key, string(msgBytes))
		}
		global.RedisClient.Expire(ctx, key, 1*time.Hour)
	}
	return messages, nil
}
func GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	err := global.ConfigDB.QueryRow(`SELECT id, username, email, phone, password_hash, role, created_at FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func GetUserByID(id int) (*models.User, error) {
	var u models.User
	err := global.ConfigDB.QueryRow(`SELECT id, username, email, phone, password_hash, role, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}
func GetBotByID(botID int) (*models.Bot, error) {
	var b models.Bot
	err := global.ConfigDB.QueryRow(`SELECT id, user_id, blocked, session_file, payment_status, created_at FROM bots WHERE id = $1`, botID).
		Scan(&b.ID, &b.UserID, &b.Blocked, &b.SessionFile, &b.PaymentStatus, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}
func GetAllBots() ([]models.Bot, error) {
	rows, err := global.ConfigDB.Query(`SELECT id, user_id, blocked, session_file, payment_status, created_at FROM bots ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bots []models.Bot
	for rows.Next() {
		var b models.Bot
		if err := rows.Scan(&b.ID, &b.UserID, &b.Blocked, &b.SessionFile, &b.PaymentStatus, &b.CreatedAt); err != nil {
			return nil, err
		}
		bots = append(bots, b)
	}
	return bots, nil
}
