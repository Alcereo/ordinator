package filters

import (
	"github.com/spf13/cast"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSessionCreate(t *testing.T) {
	// Given
	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	cacheProvider := CreateStubCacheProvider()

	cookieName := "test-session"
	handler := CreateSessionFilter("Filter name", cookieName, cacheProvider, 3, 0)
	handler.SetNext(&StubHandler{})

	// When
	handler.Handle(w, req)

	// Then
	result := w.Result()
	assertHasSetCookie(cookieName, "c1", time.Now().Add(time.Hour*4), result, t)

	value := nextChainRequest.Context().Value(SessionContextKey)
	if value == nil {
		t.Fatalf("Session must be set in context")
	}
	session := value.(*Session)
	assertSession(session, "i1", "c1", time.Now().Add(time.Hour*3), t)
}

func TestExistingSessionGet(t *testing.T) {
	// Given
	cacheProvider := CreateStubCacheProvider()
	cookie := cacheProvider.CreateNewCookie()
	_ = cacheProvider.PutSession(&Session{
		Id:      cacheProvider.CreateNewIdentifier(),
		Cookie:  cookie,
		Expires: time.Now().Add(time.Hour * 12),
	})

	cookieName := "test-session"
	handler := CreateSessionFilter("Filter name", cookieName, cacheProvider, 3, 0)
	handler.SetNext(&StubHandler{})

	req := httptest.NewRequest("GET", "/foo", nil)
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: string(cookie),
	})
	w := httptest.NewRecorder()

	// When
	handler.Handle(w, req)

	// Then
	value := nextChainRequest.Context().Value(SessionContextKey)
	if value == nil {
		t.Fatalf("Session must be set in context")
	}
	session := value.(*Session)
	assertSession(session, "i1", "c1", time.Now().Add(time.Hour*12), t)
}

func TestRenewCookie(t *testing.T) {
	// Given
	cacheProvider := CreateStubCacheProvider()
	cookie := cacheProvider.CreateNewCookie()
	_ = cacheProvider.PutSession(&Session{
		Id:      cacheProvider.CreateNewIdentifier(),
		Cookie:  cookie,
		Expires: time.Now().Add(time.Hour * 2),
	})

	cookieName := "test-session"
	handler := CreateSessionFilter("Filter name", cookieName, cacheProvider, 5, 3)
	handler.SetNext(&StubHandler{})

	req := httptest.NewRequest("GET", "/foo", nil)
	req.AddCookie(&http.Cookie{
		Name:    cookieName,
		Value:   string(cookie),
		Expires: time.Now().Add(time.Hour * 2),
	})
	w := httptest.NewRecorder()

	// When
	handler.Handle(w, req)

	// Then
	result := w.Result()
	assertHasSetCookie(cookieName, "c2", time.Now().Add(time.Hour*5), result, t)

	value := nextChainRequest.Context().Value(SessionContextKey)
	if value == nil {
		t.Fatalf("Session must be set in context")
	}
	session := value.(*Session)
	assertSession(session, "i1", "c2", time.Now().Add(time.Hour*5), t)
}

// Internal

func assertSession(session *Session, id string, cookie string, expiresBefore time.Time, t *testing.T) {
	if string(session.Id) != id {
		t.Fatalf("Expecting Session with id: %s actual:%s", id, session.Id)
	}
	if string(session.Cookie) != cookie {
		t.Fatalf("Expecting Session with Cookie: %s actual:%s", cookie, session.Cookie)
	}
	if !session.Expires.Before(expiresBefore) {
		t.Fatalf("Expecting Session expires: %v is before: %v", session.Expires, expiresBefore)
	}
}

func assertHasSetCookie(name string, value string, expiresBefore time.Time, response *http.Response, t *testing.T) {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			if cookie.Value != value {
				t.Fatalf("Set-Cookie header %s require value %s, but found: %s", name, value, cookie.Value)
			}

			if !cookie.Expires.Before(expiresBefore) {
				t.Fatalf("Set-Cookie %s require value expures before %v", cookie.String(), expiresBefore)
			}

			return
		}
	}
	t.Fatalf("Set-Cookie header %s=%s not found in response", name, value)
}

type StubCacheProvider struct {
	idIncrementer     int
	cookieIncrementer int
	sessionMap        map[SessionCookie]*Session
}

func (provider *StubCacheProvider) RemoveSession(session *Session) {
	provider.sessionMap[session.Cookie] = nil
}

func (provider *StubCacheProvider) CreateNewCookie() SessionCookie {
	provider.cookieIncrementer += 1
	return SessionCookie("c" + cast.ToString(provider.cookieIncrementer))
}

func (provider *StubCacheProvider) GetSession(cookie SessionCookie) (*Session, bool) {
	return provider.sessionMap[cookie], true // TODO Handle not found case
}

func (provider *StubCacheProvider) CreateNewIdentifier() SessionId {
	provider.idIncrementer += 1
	return SessionId("i" + cast.ToString(provider.idIncrementer))
}

func (provider *StubCacheProvider) PutSession(session *Session) error {
	provider.sessionMap[session.Cookie] = session
	return nil // TODO Handle error case
}

func CreateStubCacheProvider() *StubCacheProvider {
	return &StubCacheProvider{
		idIncrementer:     0,
		cookieIncrementer: 0,
		sessionMap:        make(map[SessionCookie]*Session),
	}
}

var nextChainRequest *http.Request

type StubHandler struct {
}

func (handler *StubHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	nextChainRequest = request
}
