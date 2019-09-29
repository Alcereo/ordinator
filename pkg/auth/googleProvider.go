package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Alcereo/ordinator/pkg/common"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type googleOAuth2Provider struct {
	cacheProvider         UserAuthCachePort
	successLoginUrl       string
	googleClientId        string
	googleClientSecret    string
	domainUrl             string
	grantType             string
	authenticationUrl     string
	accessTokenRequestUrl string
	userInfoRequestUrl    string
}

func NewGoogleOAuth2Provider(
	cacheProvider UserAuthCachePort,
	successLoginUrl string,
	googleClientId string,
	googleSecretId string,
	accessTokenRequestUrl string,
	userInfoRequestUrl string,
) *googleOAuth2Provider {
	return &googleOAuth2Provider{
		cacheProvider:         cacheProvider,
		successLoginUrl:       successLoginUrl,
		googleClientId:        googleClientId,
		googleClientSecret:    googleSecretId,
		domainUrl:             "http://localhost:8080",
		grantType:             "authorization_code",
		authenticationUrl:     "/authentication/google",
		accessTokenRequestUrl: accessTokenRequestUrl,
		userInfoRequestUrl:    userInfoRequestUrl,
	}
}

func (router *googleOAuth2Provider) Handle(writer http.ResponseWriter, request *http.Request) {
	const stage = "Performing google authorisation error. Reason: %v"

	sessionNillable := request.Context().Value(common.SessionContextKey)
	if sessionNillable == nil {
		logrus.Errorf(stage, "Session not found in the request context.")
		writer.WriteHeader(501)
		return
	}

	session := sessionNillable.(*common.Session)
	_, found := router.cacheProvider.FindUserData(session)
	if found {
		logrus.Debugf("User data for session already exist. Skip authentication.")
		http.Redirect(writer, request, router.successLoginUrl, 302)
		return
	}

	accessCode, err := getAccessCode(request)
	if err != nil {
		logrus.Errorf(stage, err)
		writer.WriteHeader(403)
		return
	}

	userData, err := router.getUserData(accessCode)
	if err != nil {
		logrus.Errorf(stage, err)
		writer.WriteHeader(403)
		return
	}

	if err := router.cacheProvider.PutUserData(session, userData); err != nil {
		logrus.Errorf(stage, err)
		writer.WriteHeader(500)
		return
	}

	logrus.Debugf("User data successful retrieved and stored to cache. %v", userData)
	http.Redirect(writer, request, router.successLoginUrl, 302)
}

func (router *googleOAuth2Provider) getUserData(accessCode *string) (*common.UserData, error) {
	const stage = "Getting user data error."

	token, err := router.retrieveAccessToken(*accessCode)
	if err != nil {
		return nil, newErr(stage, err)
	}

	googleUserInfo, err := router.getUserInfo(token.AccessToken)
	if err != nil {
		return nil, newErr(stage, err)
	}

	logrus.Debugf("Authentication successful. %+v", googleUserInfo)
	return &common.UserData{
		Identifier: googleUserInfo.Identifier,
		Username:   googleUserInfo.Username,
		Email:      googleUserInfo.Email,
		Picture:    googleUserInfo.Picture,
		Locale:     googleUserInfo.Locale,
	}, nil
}

func getAccessCode(request *http.Request) (*string, error) {
	const stage = "Getting access code error."

	if err := request.URL.Query().Get("error"); err != "" {
		return nil, newErr(stage, err)
	}
	accessCode := request.URL.Query().Get("code")
	if accessCode == "" {
		return nil, newErr(stage, "'code' query param not found or empty.")
	}
	return &accessCode, nil
}

func (router *googleOAuth2Provider) retrieveAccessToken(accessCode string) (*GoogleOAuth2Token, error) {
	const stage = "Retrieving access token error."

	requestPayload := GoogleRequestBuilder{
		Code:         accessCode,
		ClientId:     router.googleClientId,
		ClientSecret: router.googleClientSecret,
		RedirectUri:  router.domainUrl + router.authenticationUrl,
		GrantType:    router.grantType,
	}

	req, err := router.buildAccessTokenRequest(requestPayload)
	if err != nil {
		return nil, newErr(stage, err)
	}

	responseBody, err := performRequest(req)
	if err != nil {
		return nil, newErr(stage, err)
	}

	var googleAuthTokenResponse GoogleOAuth2Token
	if err := json.Unmarshal(*responseBody, &googleAuthTokenResponse); err != nil {
		return nil, newErr(stage, err)
	}
	return &googleAuthTokenResponse, nil
}

func (router *googleOAuth2Provider) buildAccessTokenRequest(requestPayload GoogleRequestBuilder) (*http.Request, error) {
	req, err := http.NewRequest(
		"POST",
		router.accessTokenRequestUrl,
		bytes.NewBuffer([]byte(requestPayload.String())),
	)
	if err != nil {
		return nil, newErr("Building request error.", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (router *googleOAuth2Provider) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	const stage = "Getting user info error."

	req, err := router.buildRequestToGoogleApi(accessToken)
	if err != nil {
		return nil, newErr(stage, err)
	}

	responseBody, err := performRequest(req)
	if err != nil {
		return nil, newErr(stage, err)
	}

	var googleUserInfo GoogleUserInfo
	if err := json.Unmarshal(*responseBody, &googleUserInfo); err != nil {
		return nil, newErr(stage, err)
	}
	return &googleUserInfo, nil
}

func (router *googleOAuth2Provider) buildRequestToGoogleApi(accessToken string) (*http.Request, error) {
	req, err := http.NewRequest(
		"GET",
		router.userInfoRequestUrl,
		nil,
	)
	if err != nil {
		return nil, newErr("Building request error.", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return req, nil
}

func performRequest(req *http.Request) (*[]byte, error) {
	const stage = "Performing request error."

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, newErr(stage, err)
	}
	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newErr(stage, err)
	}
	responseBody := string(responseBodyBytes)
	if resp.StatusCode != 200 {
		return nil, newErr(stage, responseBody)
	} else {
		logrus.Tracef("Got response body: %+v", responseBody)
	}
	return &responseBodyBytes, nil
}

func newErr(stage string, reason interface{}) error {
	return fmt.Errorf("%v Reason: %v", stage, reason)
}

type GoogleUserInfo struct {
	Identifier string `json:"sub"`
	Username   string `json:"name"`
	Picture    string `json:"picture"`
	Email      string `json:"email"`
	Locale     string `json:"locale"`
}

type GoogleOAuth2Token struct {
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
	return fmt.Sprintf(
		"code=%s"+
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
