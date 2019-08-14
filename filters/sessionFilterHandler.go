package filters

import (
	"balancer/balancer"
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const SessionContextKey string = "SessionContextKey"

type Session struct {
	Id      SessionId
	Cookie  SessionCookie
	Expires time.Time
}

type SessionId string
type SessionCookie string

type SessionCachePort interface {
	PutSession(session *Session) error
	GetSession(cookie SessionCookie) (*Session, bool)
	RemoveSession(session *Session)
	CreateNewIdentifier() SessionId
	CreateNewCookie() SessionCookie
}

type SessionFilterHandler struct {
	Name                   string
	next                   *balancer.RequestHandler
	SessionCookieName      string
	SessionCache           SessionCachePort
	CookieTTLHours         int
	RenewCookieBeforeHours int
	CookiePath             string
	CookieDomain           string
}

func CreateSessionFilter(
	name string,
	cookieName string,
	provider SessionCachePort,
	cookieTTLHours int,
	renewCookieBeforeHours int,
	cookiePath string,
	cookieDomain string,
) *SessionFilterHandler {
	return &SessionFilterHandler{
		Name:                   name,
		SessionCookieName:      cookieName,
		SessionCache:           provider,
		next:                   nil,
		CookieTTLHours:         cookieTTLHours,
		RenewCookieBeforeHours: renewCookieBeforeHours,
		CookieDomain:           cookieDomain,
		CookiePath:             cookiePath,
	}
}

func (filter *SessionFilterHandler) SetNext(nextHandler balancer.RequestHandler) {
	filter.next = &nextHandler
}

func (filter *SessionFilterHandler) Handle(writer http.ResponseWriter, request *http.Request) {

	session := filter.getOrCreateSession(writer, request)
	// Add to context
	log.Debugf("Retrieved session: %v", session)
	newContext := context.WithValue(request.Context(), SessionContextKey, session)
	newRequest := request.WithContext(newContext)

	if filter.next != nil {
		(*filter.next).Handle(writer, newRequest)
	} else {
		log.Debugf("Log filter error: %v. Next handler is empty", filter.Name)
	}
}

func (filter *SessionFilterHandler) getOrCreateSession(writer http.ResponseWriter, request *http.Request) *Session {
	cookie, err := request.Cookie(filter.SessionCookieName)
	if err == nil && cookie != nil {
		log.Tracef("Found cookie in the request context: %v", cookie.Value)
		session, found := filter.SessionCache.GetSession(SessionCookie(cookie.Value))
		if found {
			log.Tracef("Session found in cache. %v", session)
			if !session.Expires.Before(time.Now().Add(time.Hour * time.Duration(filter.RenewCookieBeforeHours))) {
				log.Tracef("Session is valid")
				return session
			} else {
				log.Tracef("Session not valid. Creating new.")
				newSession := filter.createNewSession(writer, session)
				filter.SessionCache.RemoveSession(session)
				return newSession
			}
		} else {
			log.Warnf("Session was not found in the cache. Creating new session.")
			return filter.createNewSession(writer, nil)
		}
	} else {
		log.Tracef("Cookie was not found in the request context. Creating new Session")
		return filter.createNewSession(writer, nil)
	}
}

func (filter *SessionFilterHandler) createNewSession(writer http.ResponseWriter, oldSession *Session) *Session {
	var id SessionId
	if oldSession == nil {
		id = filter.SessionCache.CreateNewIdentifier()
	} else {
		id = oldSession.Id
	}

	expires := time.Now().Add(time.Hour * time.Duration(filter.CookieTTLHours))
	session := &Session{
		Cookie:  filter.SessionCache.CreateNewCookie(),
		Id:      id,
		Expires: expires,
	}

	// Add to cache
	_ = filter.SessionCache.PutSession(session) // TODO Handle error

	// Add to Cookie-set
	newCookie := http.Cookie{
		Name:    filter.SessionCookieName,
		Value:   string(session.Cookie),
		Expires: expires,
		Path:    filter.CookiePath,
		Domain:  filter.CookieDomain,
	}
	http.SetCookie(writer, &newCookie)

	return session
}
