package filters

import (
	"balancer/balancer"
	"context"
	"log"
	"net/http"
	"time"
)

const SessionContextKey string = "SessionContextKey"

type Session struct {
	id      SessionId
	cookie  SessionCookie
	expires time.Time
}

type SessionId string
type SessionCookie string

type SessionCacheProvider interface {
	PutSession(session *Session)
	GetSession(cookie SessionCookie) *Session
	RemoveSession(session *Session)
	CreateNewIdentifier() SessionId
	CreateNewCookie() SessionCookie
}

type SessionFilterHandler struct {
	Name                   string
	next                   *balancer.RequestHandler
	SessionCookieName      string
	SessionCache           SessionCacheProvider
	CookieTTLHours         int
	RenewCookieBeforeHours int
}

func CreateSessionFilter(
	name string,
	cookieName string,
	provider SessionCacheProvider,
	cookieTTLHours int,
	renewCookieBeforeHours int,
) *SessionFilterHandler {
	return &SessionFilterHandler{
		Name:                   name,
		SessionCookieName:      cookieName,
		SessionCache:           provider,
		next:                   nil,
		CookieTTLHours:         cookieTTLHours,
		RenewCookieBeforeHours: renewCookieBeforeHours,
	}
}

func (filter *SessionFilterHandler) SetNext(nextHandler balancer.RequestHandler) {
	filter.next = &nextHandler
}

func (filter *SessionFilterHandler) Handle(writer http.ResponseWriter, request *http.Request) {

	session := filter.getOrCreateSession(writer, request)
	// Add to context
	newContext := context.WithValue(request.Context(), SessionContextKey, session)
	newRequest := request.WithContext(newContext)

	if filter.next != nil {
		(*filter.next).Handle(writer, newRequest)
	} else {
		log.Printf("Log filter error: %v. Next handler is empty", filter.Name)
	}
}

func (filter *SessionFilterHandler) getOrCreateSession(writer http.ResponseWriter, request *http.Request) *Session {
	cookie, err := request.Cookie(filter.SessionCookieName)
	if err == nil && cookie != nil {
		session := filter.SessionCache.GetSession(SessionCookie(cookie.Value))
		if session.expires.Before(time.Now().Add(time.Hour * time.Duration(filter.RenewCookieBeforeHours))) {
			newSession := filter.createNewSession(writer, session)
			filter.SessionCache.RemoveSession(session)
			session = newSession
		}
		return session
	} else {
		session := filter.createNewSession(writer, nil)
		return session
	}
}

func (filter *SessionFilterHandler) createNewSession(writer http.ResponseWriter, oldSession *Session) *Session {
	var id SessionId
	if oldSession == nil {
		id = filter.SessionCache.CreateNewIdentifier()
	} else {
		id = oldSession.id
	}

	expires := time.Now().Add(time.Hour * time.Duration(filter.CookieTTLHours))
	session := &Session{
		cookie:  filter.SessionCache.CreateNewCookie(),
		id:      id,
		expires: expires,
	}

	// Add to cache
	filter.SessionCache.PutSession(session)

	// Add to cookie-set
	newCookie := http.Cookie{
		Name:    filter.SessionCookieName,
		Value:   string(session.cookie),
		Expires: expires,
	}
	http.SetCookie(writer, &newCookie)

	return session
}
