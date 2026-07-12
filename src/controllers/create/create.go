package create

import (
	"App/src/global"
	"App/src/models"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, email, phone, password string) (*models.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var u models.User
	err = global.ConfigDB.QueryRow(`
		INSERT INTO users (username, email, phone, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, phone, role, created_at`,
		username, email, phone, string(hashed)).
		Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.Role, &u.CreatedAt)
	return &u, err
}

func CreateBot(userID int, sessionFile string, paymentStatus string) (int, error) {
	var id int
	err := global.ConfigDB.QueryRow(`
		INSERT INTO bots (user_id, session_file, payment_status)
		VALUES ($1, $2, $3)
		RETURNING id`, userID, sessionFile, paymentStatus).Scan(&id)
	return id, err
}
