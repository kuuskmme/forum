package internal

import (
	"context"
	"net/http"
)

type ContextKey string

const IsLoggedInCtxKey ContextKey = "IsLoggedIn"

const UserCtxKey ContextKey = "UserCtx"

type UserCtx struct {
	UUID      string `db:"uuid"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

func (user *User) ReloadUserIntoContext(r *http.Request) *http.Request {
	userCtx := UserCtx{user.UUID, user.Name, user.CreatedAt}
	// We make sure to only load in the ctx data
	ctx := context.WithValue(r.Context(), UserCtxKey, userCtx)
	Logger.Printf("Context %v loaded into HTTP Response", ctx.Value(UserCtxKey))
	return r.WithContext(ctx)
}

func (session *Session) LoadUserIntoContext() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT uuid, name, created_at FROM users WHERE uuid = $1", session.UserID).
		Scan(&user.UUID, &user.Name, &user.CreatedAt)
	if err != nil {
		Logger.Printf("User context fetch failed...")
		return
	}

	return
}

// Retrieve
func getUserCtx(r *http.Request) any {
	userCtx := r.Context().Value(UserCtxKey)
	return userCtx
}
