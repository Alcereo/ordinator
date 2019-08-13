package filters

import (
	"github.com/spf13/cast"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSessionCreate(t *testing.T) {
	// Given
	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	cacheProvider := CreateStubCacheProvider()

	cookieName := "test-session"
	handler := CreateSessionFilter("Filter name", cookieName, cacheProvider)
	handler.SetNext(&StubHandler{})

	// When
	handler.Handle(w, req)

	// Then
	result := w.Result()
	assertHasSetCookie(cookieName, "c1", result, t)

	value := nextChainRequest.Context().Value(SessionContextKey)
	if value == nil {
		t.Fatalf("Session must be set in context")
	}
	session := value.(*Session)
	assertSession(session, "i1", "c1", t)
}

func TestExistingSessionGet(t *testing.T) {
	// Given
	cacheProvider := CreateStubCacheProvider()
	cookie := cacheProvider.CreateNewCookie()
	cacheProvider.PutSession(&Session{
		id:     cacheProvider.CreateNewIdentifier(),
		cookie: cookie,
	})

	cookieName := "test-session"
	handler := CreateSessionFilter("Filter name", cookieName, cacheProvider)
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
	assertSession(session, "i1", "c1", t)
}

// Internal

func assertSession(session *Session, id string, cookie string, t *testing.T) {
	if string(session.id) != id {
		t.Fatalf("Expecting Session with id: %s actual:%s", id, session.id)
	}
	if string(session.cookie) != cookie {
		t.Fatalf("Expecting Session with cookie: %s actual:%s", cookie, session.cookie)
	}
}

func assertHasSetCookie(name string, value string, response *http.Response, t *testing.T) {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			if cookie.Value != value {
				t.Fatalf("Set-Cookie header %s require value %s, but found: %s", name, value, cookie.Value)
			} else {
				return
			}
		}
	}
	t.Fatalf("Set-Cookie header %s=%s not found in response", name, value)
}

type StubCacheProvider struct {
	idIncrementer     int
	cookieIncrementer int
	sessionMap        map[SessionCookie]*Session
}

func (provider *StubCacheProvider) CreateNewCookie() SessionCookie {
	provider.cookieIncrementer += 1
	return SessionCookie("c" + cast.ToString(provider.cookieIncrementer))
}

func (provider *StubCacheProvider) GetSession(cookie SessionCookie) *Session {
	return provider.sessionMap[cookie]
}

func (provider *StubCacheProvider) CreateNewIdentifier() SessionId {
	provider.idIncrementer += 1
	return SessionId("i" + cast.ToString(provider.idIncrementer))
}

func (provider *StubCacheProvider) PutSession(session *Session) {
	provider.sessionMap[session.cookie] = session
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
