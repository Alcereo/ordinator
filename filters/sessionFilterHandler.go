package filters

import (
	"balancer/balancer"
	"context"
	"log"
	"net/http"
)

const SessionContextKey string = "SessionContextKey"

type Session struct {
	id string
}

type SessionCacheProvider interface {
	GetSession(identifier string) *Session
	CreateNewIdentifier() string
	PutSession(session *Session)
}

type SessionFilterHandler struct {
	Name              string
	next              *balancer.RequestHandler
	SessionCookieName string
	SessionCache      SessionCacheProvider
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
		return filter.SessionCache.GetSession(cookie.Value)
	} else {
		session := Session{
			id: filter.SessionCache.CreateNewIdentifier(),
		}
		// Add to cache
		filter.SessionCache.PutSession(&session)
		// Add to cookie-set
		newCookie := http.Cookie{
			Name:  filter.SessionCookieName,
			Value: session.id,
		}
		http.SetCookie(writer, &newCookie)
		return &session
	}
}

// Factory

func CreateSessionFilter(Name string, cookieName string, provider SessionCacheProvider) *SessionFilterHandler {
	return &SessionFilterHandler{
		Name:              Name,
		SessionCookieName: cookieName,
		SessionCache:      provider,
		next:              nil,
	}
}
