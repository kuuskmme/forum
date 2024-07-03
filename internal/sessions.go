package internal

import (
	"log"
)

func (session *Session) DeleteByUUID() error {
	_, err := Db.Exec("DELETE FROM sessions WHERE uuid = ?", session.UUID)
	if err != nil {
		log.Printf("ERROR | Could not delete session with UUID %s: %v", session.UUID, err)
		return err
	}
	return nil
}

func ValidateSessionFromDB(sessionUUID string) bool {
	var dbSessionUUID string
	err := Db.QueryRow("SELECT uuid FROM sessions WHERE uuid = $1", sessionUUID).Scan(&dbSessionUUID)
	return err == nil
}

func (session *Session) GetUserIDBySessionUUID() (err error) {
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE uuid = $1", session.UUID).
		Scan(&session.UserID)
	log.Printf("State of User ID in Session data after user fetch: %v", session.UserID)
	return
}

// Get the user from the session
func (session *Session) User() (user User, err error) {
	user = User{}
	Logger.Printf("Session userid: %v", session.UserID)
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE uuid = $1", session.UserID).
		Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		Logger.Printf("User fetch failed...")
		return
	}

	return
}
