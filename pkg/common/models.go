package common

import "time"

// User data

const UserDataContextKey string = "UserDataContextKey"

type UserData struct {
	Username   string
	Identifier string
	Email      string
	Picture    string
	Locale     string
}

// Session

const SessionContextKey string = "SessionContextKey"

type Session struct {
	Id      SessionId
	Cookie  SessionCookie
	Expires time.Time
}

type SessionId string
type SessionCookie string
