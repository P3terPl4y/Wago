package save

import (
	"App/src/controllers/encrypt"
	"App/src/global"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func SavePrompt(botID int, prompt string) error {
	_, err := global.ConfigDB.Exec(`INSERT INTO prompts (bot_id, prompt) VALUES ($1, $2) ON CONFLICT (bot_id) DO UPDATE SET prompt = $2`, botID, prompt)
	return err
}

func SaveExpiration(botID int, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	_, err := global.ConfigDB.Exec(`INSERT INTO subscriptions (bot_id, expires_at) VALUES ($1, $2) ON CONFLICT (bot_id) DO UPDATE SET expires_at = $2`, botID, expiresAt)
	return err
}
func SaveChatMessage(botID int, userJID string, role string, content string) error {
	encrypted, err := encrypt.EncryptContent(content)
	if err != nil {
		return err
	}
	_, err = global.ConfigDB.Exec(`INSERT INTO chat_history (bot_id, user_jid, role, content) VALUES ($1, $2, $3, $4)`,
		botID, userJID, role, encrypted)
	if err != nil {
		return err
	}
	if global.RedisClient != nil {
		key := fmt.Sprintf("chat:%d:%s", botID, userJID)
		ctx := context.Background()
		msgBytes, _ := json.Marshal(map[string]string{"role": role, "content": content})
		err = global.RedisClient.RPush(ctx, key, string(msgBytes)).Err()
		if err == nil {
			global.RedisClient.LTrim(ctx, key, -global.MAX_HISTORY, -1)
			global.RedisClient.Expire(ctx, key, 1*time.Hour)
		} else {
			log.Printf("⚠️ Redis: %v", err)
		}
	}
	return nil
}
