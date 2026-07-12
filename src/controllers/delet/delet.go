package delet

import "App/src/global"

func DeleteBot(botID int) error {
	_, err := global.ConfigDB.Exec(`DELETE FROM bots WHERE id = $1`, botID)
	return err
}
