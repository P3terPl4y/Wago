package update

import (
	"App/src/global"

	"golang.org/x/crypto/bcrypt"
)

func UpdateBotBlocked(botID int, blocked bool) error {
	_, err := global.ConfigDB.Exec(`UPDATE bots SET blocked = $1 WHERE id = $2`, blocked, botID)
	return err
}

func UpdateBotPaymentStatus(botID int, status string) error {
	_, err := global.ConfigDB.Exec(`UPDATE bots SET payment_status = $1 WHERE id = $2`, status, botID)
	return err
}
func UpdateUserPassword(userID int, newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = global.ConfigDB.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, string(hashed), userID)
	return err
}
