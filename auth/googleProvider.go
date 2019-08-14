package auth

import (
	"balancer/filters"
	"github.com/sirupsen/logrus"
	"net/http"
)

type googleOAuth2Provider struct {
	cacheProvider   UserAuthCachePort
	successLoginUrl string
}

func NewGoogleOAuth2Provider(
	cacheProvider UserAuthCachePort,
	successLoginUrl string,
) *googleOAuth2Provider {
	return &googleOAuth2Provider{
		cacheProvider:   cacheProvider,
		successLoginUrl: successLoginUrl,
	}
}

func (router *googleOAuth2Provider) Handle(writer http.ResponseWriter, request *http.Request) {
	sessionNillable := request.Context().Value(filters.SessionContextKey)
	if sessionNillable == nil {
		logrus.Warnf("Session not found in the request context! Session filter should be executed before authorization")
		writer.WriteHeader(501)
	}
	userData, found := router.getUserData(request)
	if !found {
		logrus.Debugf("User data not found")
		writer.WriteHeader(403)
	} else {
		session := sessionNillable.(*filters.Session)
		err := router.cacheProvider.PutUserData(session, userData)
		if err != nil {
			logrus.Errorf("Saving user data to cache error. %v", err)
			writer.WriteHeader(500)
		} else {
			logrus.Debugf("User data successful retrieved and stored to cache. %v", userData)
			http.Redirect(writer, request, router.successLoginUrl, 301)
		}
	}
}

func (router *googleOAuth2Provider) getUserData(request *http.Request) (*UserData, bool) {
	username := request.URL.Query().Get("user")
	if username != "" {
		return &UserData{
			Username: username,
		}, true
	} else {
		return nil, false
	}
}
