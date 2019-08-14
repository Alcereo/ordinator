package auth

import (
	"balancer/balancer"
	"balancer/filters"
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const UserDataContextKey string = "UserDataContextKey"

type UserData struct {
	Username string
}

type UserAuthCachePort interface {
	FindUserData(session *filters.Session) (*UserData, bool)
	PutUserData(session *filters.Session, userData *UserData) error
}

type userAuthenticationFilter struct {
	next          *balancer.RequestHandler
	cacheProvider UserAuthCachePort
	Name          string
}

func NewUserAuthenticationFilter(
	cacheProvider UserAuthCachePort,
	Name string,
) *userAuthenticationFilter {
	return &userAuthenticationFilter{
		next:          nil,
		cacheProvider: cacheProvider,
		Name:          Name,
	}
}

func (filter *userAuthenticationFilter) Handle(writer http.ResponseWriter, request *http.Request) {
	enchantedRequest := filter.updateRequestContext(request)
	if filter.next != nil {
		(*filter.next).Handle(writer, enchantedRequest)
	} else {
		log.Debugf("User authentication filter: %v doesn't have next handler", filter.Name)
	}
}

func (filter *userAuthenticationFilter) updateRequestContext(request *http.Request) *http.Request {
	sessionNillable := request.Context().Value(filters.SessionContextKey)
	if sessionNillable == nil {
		log.Debugf("Session not found in the request context")
		return request
	} else {
		session := sessionNillable.(*filters.Session)
		log.Debugf("Found Session in request context. Id: %v", session.Id)

		userData, found := filter.cacheProvider.FindUserData(session)
		if !found {
			log.Debugf("User data not found in the cache")
			return request
		} else {
			log.Debugf("User data found and put to context. %v", userData)
			newContext := context.WithValue(request.Context(), UserDataContextKey, userData)
			newRequest := request.WithContext(newContext)
			return newRequest
		}
	}
}

func (filter *userAuthenticationFilter) SetNext(handler balancer.RequestHandler) {
	filter.next = &handler
}
