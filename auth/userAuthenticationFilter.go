package auth

import (
	"balancer/common"
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type UserAuthCachePort interface {
	FindUserData(session *common.Session) (*common.UserData, bool)
	PutUserData(session *common.Session, userData *common.UserData) error
}

type userAuthenticationFilter struct {
	next          *common.RequestHandler
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
	sessionNillable := request.Context().Value(common.SessionContextKey)
	if sessionNillable == nil {
		log.Debugf("Session not found in the request context")
		return request
	} else {
		session := sessionNillable.(*common.Session)
		log.Debugf("Found Session in request context. Id: %v", session.Id)

		userData, found := filter.cacheProvider.FindUserData(session)
		if !found {
			log.Debugf("User data not found in the cache")
			return request
		} else {
			log.Debugf("User data found and put to context. %v", userData)
			newContext := context.WithValue(request.Context(), common.UserDataContextKey, userData)
			newRequest := request.WithContext(newContext)
			return newRequest
		}
	}
}

func (filter *userAuthenticationFilter) SetNext(handler common.RequestHandler) {
	filter.next = &handler
}
