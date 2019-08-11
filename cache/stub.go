package cache

import "balancer/filters"

type StubCacheProvider struct {
}

func (StubCacheProvider) GetSession(identifier string) *filters.Session {
	panic("implement me")
}

func (StubCacheProvider) CreateNewIdentifier() string {
	panic("implement me")
}

func (StubCacheProvider) PutSession(session *filters.Session) {
	panic("implement me")
}

// Factory


