package auth

import (
	"context"
	"errors"
	"github.com/Alcereo/ordinator/pkg/common"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type UserAuthCachePort interface {
	FindUserData(session *common.Session) (*common.UserData, bool)
	PutUserData(session *common.Session, userData *common.UserData) error
}

type userAuthenticationFilter struct {
	next             *common.RequestHandler
	cacheProvider    UserAuthCachePort
	Name             string
	userDataRequired bool
}

func NewUserAuthenticationFilter(
	cacheProvider UserAuthCachePort,
	Name string,
	userDataRequired bool,
) *userAuthenticationFilter {
	return &userAuthenticationFilter{
		next:             nil,
		cacheProvider:    cacheProvider,
		Name:             Name,
		userDataRequired: userDataRequired,
	}
}

func (filter *userAuthenticationFilter) Handle(log *log.Entry, writer http.ResponseWriter, request *http.Request) {
	log = log.WithField("filterName", filter.Name)
	enchantedRequest, err := filter.updateRequestContext(log, request)
	if err != nil {
		if filter.userDataRequired {
			log.Debugf("Getting user data for session error. Reason: %v", err.Error())
			writer.WriteHeader(401)
			return
		} else {
			log.Debugf("Getting user data for session error. Reason: %v", err.Error())
		}
	}
	if filter.next != nil {
		(*filter.next).Handle(log, writer, enchantedRequest)
	} else {
		log.Debugf("User authentication filter: %v doesn't have next handler", filter.Name)
	}
}

func (filter *userAuthenticationFilter) updateRequestContext(log *log.Entry, request *http.Request) (*http.Request, error) {
	sessionNillable := request.Context().Value(common.SessionContextKey)
	if sessionNillable == nil {
		return request, errors.New("session not found in the request context")
	} else {
		session := sessionNillable.(*common.Session)
		log.Debugf("Found Session in request context. Id: %v", session.Id)

		userData, found := filter.cacheProvider.FindUserData(session)
		if !found {
			return request, errors.New("user data not found in the cache")
		} else {
			log.Debugf("User data found and put to context. %v", userData)
			newContext := context.WithValue(request.Context(), common.UserDataContextKey, userData)
			newRequest := request.WithContext(newContext)
			return newRequest, nil
		}
	}
}

func (filter *userAuthenticationFilter) SetNext(handler common.RequestHandler) {
	filter.next = &handler
}
