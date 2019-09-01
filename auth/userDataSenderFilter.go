package auth

import (
	"balancer/common"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type userDataSenderFilter struct {
	next               *common.RequestHandler
	cacheProvider      UserAuthCachePort
	Name               string
	userDataSerializer UserDataSerializer
}

func NewUserDataSenderFilter(
	cacheProvider UserAuthCachePort,
	name string,
	userDataSerializer UserDataSerializer,
) *userDataSenderFilter {
	return &userDataSenderFilter{
		next:               nil,
		cacheProvider:      cacheProvider,
		Name:               name,
		userDataSerializer: userDataSerializer,
	}
}

func (filter *userDataSenderFilter) SetNext(handler common.RequestHandler) {
	filter.next = &handler
}

func (filter *userDataSenderFilter) Handle(writer http.ResponseWriter, request *http.Request) {
	enchantedRequest := filter.updateRequest(request)
	if filter.next != nil {
		(*filter.next).Handle(writer, enchantedRequest)
	} else {
		log.Debugf("User authentication filter: %v doesn't have next handler", filter.Name)
	}
}

func (filter *userDataSenderFilter) updateRequest(request *http.Request) *http.Request {
	sessionNillable := request.Context().Value(common.SessionContextKey)
	if sessionNillable == nil {
		log.Warnf("Session not found in the request context. Skip User data sending.")
		return request
	}

	session := sessionNillable.(*common.Session)
	userData, found := filter.cacheProvider.FindUserData(session)

	if !found {
		log.Warnf("User data not found in the request context. Skip user data sending.")
		return request
	}

	jwtToken, err := filter.userDataSerializer.Serialize(userData)
	if err != nil {
		log.Errorf("User data serializing error. Skip user data sending. %+v", err)
		return request
	}

	request.Header.Add("X-USER-DATA", jwtToken)
	return request
}

type UserDataSerializer interface {
	Serialize(*common.UserData) (string, error)
}
