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
	assertHasSetCookie(cookieName, "1", result, t)

	value := nextChainRequest.Context().Value(SessionContextKey)
	if value == nil {
		t.Fatalf("Session must be set in context")
	}
	session := value.(*Session)
	if session.id != "1" {
		t.Fatalf("Session must have apropriate id filed")
	}
}

func assertHasSetCookie(name string, value string, response *http.Response, t *testing.T) {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			if cookie.Value != value {
				t.Fatalf("Cookie %s require value %s, but found: %s", name, value, cookie.Value)
			} else {
				return
			}
		}
	}
	t.Fatalf("Cookie %s=%s not found in response", name, value)
}

// Internal

type StubCacheProvider struct {
	incrementer int
	sessionMap map[string]*Session
}

func (provider *StubCacheProvider) GetSession(identifier string) *Session {
	return provider.sessionMap[identifier]
}

func (provider *StubCacheProvider) CreateNewIdentifier() string {
	provider.incrementer += 1
	return cast.ToString(provider.incrementer)
}

func (provider *StubCacheProvider) PutSession(session *Session) {
	provider.sessionMap[session.id] = session
}

func CreateStubCacheProvider() *StubCacheProvider {
	return &StubCacheProvider{
		incrementer: 0,
		sessionMap: make(map[string]*Session),
	}
}

var nextChainRequest *http.Request

type StubHandler struct {
}

func (handler *StubHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	nextChainRequest = request
}
