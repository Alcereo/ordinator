package cache

import (
	"github.com/Alcereo/ordinator/pkg/common"
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

func (adapter *goCacheSessionCacheAdapter) PutSession(session *common.Session) error {
	return adapter.cookieCache.Add(string(session.Cookie), session, cache.DefaultExpiration)
}

func (adapter *goCacheSessionCacheAdapter) GetSession(cookie common.SessionCookie) (*common.Session, bool) {
	session, found := adapter.cookieCache.Get(string(cookie))
	if found {
		return session.(*common.Session), true
	} else {
		return nil, false
	}
}

func (adapter *goCacheSessionCacheAdapter) RemoveSession(session *common.Session) {
	adapter.cookieCache.Delete(string(session.Cookie))
}

func (*goCacheSessionCacheAdapter) CreateNewIdentifier() common.SessionId {
	return common.SessionId(uuid.NewV1().String())
}

func (*goCacheSessionCacheAdapter) CreateNewCookie() common.SessionCookie {
	return common.SessionCookie(uuid.NewV1().String())
}

// UserAuthenticationPort implementation

func (adapter *goCacheSessionCacheAdapter) FindUserData(session *common.Session) (*common.UserData, bool) {
	userData, found := adapter.cookieCache.Get(string(session.Id))
	if found {
		return userData.(*common.UserData), true
	} else {
		return nil, false
	}
}

func (adapter *goCacheSessionCacheAdapter) PutUserData(session *common.Session, userData *common.UserData) error {
	return adapter.cookieCache.Add(string(session.Id), userData, cache.DefaultExpiration)
}
