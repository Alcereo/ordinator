package auth

import (
	"balancer/filters"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type googleOAuth2Provider struct {
	cacheProvider      UserAuthCachePort
	successLoginUrl    string
	googleClientId     string
	googleClientSecret string
	domainUrl          string
	grantType          string
	pattern            string
}

func NewGoogleOAuth2Provider(
	cacheProvider UserAuthCachePort,
	successLoginUrl string,
	googleClientId string,
	googleSecretId string,
) *googleOAuth2Provider {
	return &googleOAuth2Provider{
		cacheProvider:      cacheProvider,
		successLoginUrl:    successLoginUrl,
		googleClientId:     googleClientId,
		googleClientSecret: googleSecretId,
		domainUrl:          "http://localhost:8080",
		grantType:          "authorization_code",
		pattern:            "/authentication/google",
	}
}

func (router *googleOAuth2Provider) Handle(writer http.ResponseWriter, request *http.Request) {
	sessionNillable := request.Context().Value(filters.SessionContextKey)
	if sessionNillable == nil {
		logrus.Errorf("Session not found in the request context! Session filter should be executed before authorization")
		writer.WriteHeader(501)
		return
	}

	userData, err := router.getUserData(request)
	if err != nil {
		logrus.Errorf("User data not found. %v", err)
		writer.WriteHeader(403)
		return
	}

	session := sessionNillable.(*filters.Session)
	if err := router.cacheProvider.PutUserData(session, userData); err != nil {
		logrus.Errorf("Saving user data to cache error. %v", err)
		writer.WriteHeader(500)
		return
	}

	logrus.Debugf("User data successful retrieved and stored to cache. %v", userData)
	http.Redirect(writer, request, router.successLoginUrl, 301)
}

func (router *googleOAuth2Provider) getUserData(request *http.Request) (*UserData, error) {

	// Get code from request
	if err := request.URL.Query().Get("error"); err != "" {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Reason: %v", err))
	}
	accessCode := request.URL.Query().Get("code")
	if accessCode == "" {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Reason: 'code' query param not found or empty"))
	}

	response, err := router.retrieveAccessToken(accessCode)
	if err != nil {
		return nil, err
	}

	// Get user data;
	req, err := http.NewRequest(
		"GET",
		"https://www.googleapis.com/oauth2/v3/userinfo",
		nil,
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Building request error. Reason: %v", err))
	}
	req.Header.Set("Authorization", "Bearer "+response.AccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", string(responseBody)))
	}

	var googleAuthTokenResponse GoogleUserInfo
	if err := json.Unmarshal(responseBody, &googleAuthTokenResponse); err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}

	logrus.Debugf("Successful retrieve auth token, %v", googleAuthTokenResponse)
	return &UserData{
		Username: googleAuthTokenResponse.Username,
	}, nil
}

func (router *googleOAuth2Provider) retrieveAccessToken(accessCode string) (*GoogleOAuth2TokenResponse, error) {
	requestPayload := GoogleRequestBuilder{
		Code:         accessCode,
		ClientId:     router.googleClientId,
		ClientSecret: router.googleClientSecret,
		RedirectUri:  router.domainUrl + router.pattern,
		GrantType:    router.grantType,
	}
	req, err := http.NewRequest(
		"POST",
		"https://www.googleapis.com/oauth2/v4/token",
		bytes.NewBuffer([]byte(requestPayload.String())),
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Building request error. Reason: %v", err))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", string(responseBody)))
	}
	var googleAuthTokenResponse GoogleOAuth2TokenResponse
	if err := json.Unmarshal(responseBody, &googleAuthTokenResponse); err != nil {
		return nil, errors.New(fmt.Sprintf("Error while authentication. Get token request error. Reason: %v", err))
	}
	return &googleAuthTokenResponse, nil
}

type GoogleUserInfo struct {
	Identifier string `json:"sub"`
	Username   string `json:"name"`
	Email      string `json:"email"`
}

type GoogleOAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type GoogleRequestBuilder struct {
	Code         string
	ClientId     string
	ClientSecret string
	RedirectUri  string
	GrantType    string
}

func (builder *GoogleRequestBuilder) String() string {
	return fmt.Sprintf("code=%s"+
		"&client_id=%s"+
		"&client_secret=%s"+
		"&redirect_uri=%s"+
		"&grant_type=%s",
		builder.Code,
		builder.ClientId,
		builder.ClientSecret,
		builder.RedirectUri,
		builder.GrantType,
	)
}
