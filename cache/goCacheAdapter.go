package cache

import (
	"balancer/filters"
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"time"
)

type goCacheSessionCacheAdapter struct {
	cookieCache *cache.Cache
}

func NewGoCacheSessionCacheProvider(expirationTimeHours int, evictScheduleTimeHours int) *goCacheSessionCacheAdapter {
	cookieCache := cache.New(
		time.Hour*time.Duration(expirationTimeHours),
		time.Hour*time.Duration(evictScheduleTimeHours),
	)
	return &goCacheSessionCacheAdapter{
		cookieCache: cookieCache,
	}
}

func (adapter *goCacheSessionCacheAdapter) PutSession(session *filters.Session) error {
	return adapter.cookieCache.Add(string(session.Cookie), session, cache.DefaultExpiration)
}

func (adapter *goCacheSessionCacheAdapter) GetSession(cookie filters.SessionCookie) (*filters.Session, bool) {
	session, found := adapter.cookieCache.Get(string(cookie))
	if found {
		return session.(*filters.Session), found
	} else {
		return nil, false
	}
}

func (adapter *goCacheSessionCacheAdapter) RemoveSession(session *filters.Session) {
	adapter.cookieCache.Delete(string(session.Cookie))
}

func (*goCacheSessionCacheAdapter) CreateNewIdentifier() filters.SessionId {
	return filters.SessionId(uuid.NewV1().String())
}

func (*goCacheSessionCacheAdapter) CreateNewCookie() filters.SessionCookie {
	return filters.SessionCookie(uuid.NewV1().String())
}
