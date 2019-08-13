package cache

import (
	"balancer/filters"
	"testing"
	"time"
)

func TestPutNew(t *testing.T) {
	adapter := NewGoCacheSessionCacheProvider(1, 1)

	session := &filters.Session{
		Id:      "i1",
		Cookie:  "c1",
		Expires: time.Now(),
	}
	err := adapter.PutSession(session)
	if err != nil {
		t.Errorf("Saving session error: %v", err)
	}

	cachedSession, found := adapter.GetSession("c1")
	if !found {
		t.Fatalf("Session not found")
	}
	if cachedSession == nil {
		t.Fatalf("Session is nil")
	}
	if cachedSession != session {
		t.Fatalf("Cached session not equal getted")
	}
}

func TestPutNotUnique(t *testing.T) {
	adapter := NewGoCacheSessionCacheProvider(1, 1)

	sessionFirst := &filters.Session{
		Id:      "i1",
		Cookie:  "c1",
		Expires: time.Now(),
	}
	err := adapter.PutSession(sessionFirst)
	if err != nil {
		t.Errorf("Saving session error: %v", err)
	}

	sessionSecond := &filters.Session{
		Id:      "i2",
		Cookie:  "c1",
		Expires: time.Now(),
	}
	err = adapter.PutSession(sessionSecond)
	if err == nil {
		t.Errorf("Expect saving error")
	}
}

func TestGetNotFound(t *testing.T) {
	adapter := NewGoCacheSessionCacheProvider(1, 1)

	sessionFirst := &filters.Session{
		Id:      "i1",
		Cookie:  "c1",
		Expires: time.Now(),
	}
	err := adapter.PutSession(sessionFirst)
	if err != nil {
		t.Errorf("Saving session error: %v", err)
	}

	session, found := adapter.GetSession("c2")
	if found {
		t.Errorf("Expect not found")
	}
	if session != nil {
		t.Errorf("Expect nil session")
	}
}

func TestRemoveSession(t *testing.T) {
	adapter := NewGoCacheSessionCacheProvider(1, 1)

	session := &filters.Session{
		Id:      "i1",
		Cookie:  "c1",
		Expires: time.Now(),
	}
	err := adapter.PutSession(session)
	if err != nil {
		t.Errorf("Saving session error: %v", err)
	}

	adapter.RemoveSession(session)

	cachedSession, found := adapter.GetSession("c1")
	if found {
		t.Errorf("Expect not found")
	}
	if cachedSession != nil {
		t.Errorf("Expect nil session")
	}
}

func TestIdentifiersCreating(t *testing.T) {
	adapter := NewGoCacheSessionCacheProvider(1, 1)

	identifier := adapter.CreateNewIdentifier()
	if identifier == "" {
		t.Errorf("Expect valid identifier")
	}

	cookie := adapter.CreateNewCookie()
	if cookie == "" {
		t.Errorf("Expect valid identifier")
	}
}
