package context

import (
	"fmt"
	"github.com/Alcereo/ordinator/pkg/auth"
	"github.com/Alcereo/ordinator/pkg/cache"
	"github.com/Alcereo/ordinator/pkg/common"
	"github.com/Alcereo/ordinator/pkg/filters"
	"github.com/Alcereo/ordinator/pkg/proxy"
	"github.com/Alcereo/ordinator/pkg/serializers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type context struct {
	sessionCacheAdapters  map[string]filters.SessionCachePort
	userAuthCacheAdapters map[string]auth.UserAuthCachePort
	serverMultiplexer     *http.ServeMux
}

func NewContext() *context {
	return &context{
		sessionCacheAdapters:  make(map[string]filters.SessionCachePort),
		userAuthCacheAdapters: make(map[string]auth.UserAuthCachePort),
		serverMultiplexer:     http.NewServeMux(),
	}
}

func (ctx *context) SetupCache(adapters []CacheAdapter) {
	for _, adapter := range adapters {
		switch adapter.Type {
		case GoCache:
			provider := cache.NewGoCacheSessionCacheProvider(
				adapter.ExpirationTimeHours,
				adapter.EvictScheduleTimeHours,
			)
			// GoCache can be both
			ctx.sessionCacheAdapters[adapter.Identifier] = provider
			ctx.userAuthCacheAdapters[adapter.Identifier] = provider
		default:
			panic(fmt.Errorf("Undefined session filter cache adapter type: %v.\n", adapter.Type))
		}
	}
}

func (ctx *context) SetupRouters(routers []Router, secret GoogleSecret) {
	for _, router := range routers {
		switch router.Type {
		case ReverseProxy:
			log.Debugf(
				"Adding Reverse proxy router. Pattern: %s; Target: %s",
				router.Pattern,
				router.TargetUrl,
			)

			targetUrl, _ := new(url.URL).Parse(router.TargetUrl)
			handler := proxy.ReverseProxyHandler{
				TargetAddress: *targetUrl,
			}

			rootFilterHandler := ctx.BuildFilterHandlers(router.Filters, &handler)
			ctx.serverMultiplexer.HandleFunc(router.Pattern, rootFilterHandler.Handle)
		case GoogleOauth2Authorization:
			log.Debugf(
				"Adding Google Oauth2 authorization endpoint. Pattern: %s;",
				router.Pattern,
			)
			cacheAdapter := ctx.userAuthCacheAdapters[router.CacheAdapterIdentifier]
			if cacheAdapter == nil {
				panic(fmt.Errorf("User cache adapter with identifier '%v' not found.\n", router.CacheAdapterIdentifier))
			}
			handler := auth.NewGoogleOAuth2Provider(
				cacheAdapter,
				router.SuccessLoginUrl,
				secret.ClientId,
				secret.ClientSecret,
				router.AccessTokenRequestUrl,
				router.UserInfoRequestUrl,
			)

			rootFilterHandler := ctx.BuildFilterHandlers(router.Filters, handler)
			ctx.serverMultiplexer.HandleFunc(router.Pattern, rootFilterHandler.Handle)
		default:
			panic(fmt.Errorf("Undefined router type: %v.\n", router.Type))
		}
	}
}

func (ctx *context) BuildFilterHandlers(filters []Filter, mainHandler common.RequestHandler) (rootHandler common.RequestHandler) {
	if filters == nil {
		return mainHandler
	}

	currentHandler := mainHandler

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]

		handler := ctx.BuildFilterHandler(filter)

		if handler == nil {
			continue
		}

		handler.SetNext(currentHandler)
		currentHandler = handler
	}

	return currentHandler
}

func (ctx *context) BuildFilterHandler(filter Filter) common.RequestChainedHandler {
	switch filter.Type {
	case LogFilter:
		log.Debugf("Adding Log filter. Name: %s", filter.Name)
		return filters.CreateLogFilter(filter.Name, filter.Template, nil)
	case SessionFilter:
		log.Debugf("Adding session filter. Name: %s", filter.Name)
		cacheAdapter := ctx.sessionCacheAdapters[filter.CacheAdapterIdentifier]
		if cacheAdapter == nil {
			panic(fmt.Errorf("Session cache adapter with identifier '%v' not found.\n", filter.CacheAdapterIdentifier))
		}
		return filters.CreateSessionFilter(
			filter.Name,
			filter.CookieName,
			cacheAdapter,
			filter.CookieTTLHours,
			filter.CookieRenewBeforeHours,
			filter.CookiePath,
			filter.CookieDomain,
		)
	case UserAuthenticationFilter:
		log.Debugf("Adding user authentication filter. Name: %s", filter.Name)
		cacheAdapter := ctx.userAuthCacheAdapters[filter.CacheAdapterIdentifier]
		if cacheAdapter == nil {
			panic(fmt.Errorf("User cache adapter with identifier '%v' not found.\n", filter.CacheAdapterIdentifier))
		}
		return auth.NewUserAuthenticationFilter(
			cacheAdapter,
			filter.Name,
		)
	case UserDataSenderFilter:
		log.Debugf("Adding user data sending filter. Name: %s", filter.Name)
		cacheAdapter := ctx.userAuthCacheAdapters[filter.CacheAdapterIdentifier]
		if cacheAdapter == nil {
			panic(fmt.Errorf("User cache adapter with identifier '%v' not found.\n", filter.CacheAdapterIdentifier))
		}
		serializer := buildUserDataSerializer(&filter)
		return auth.NewUserDataSenderFilter(
			cacheAdapter,
			filter.Name,
			serializer,
			filter.UserDataHeader,
		)
	default:
		panic(fmt.Errorf("Undefined filter type: %v.\n", filter.Type))
	}
}

func buildUserDataSerializer(filter *Filter) auth.UserDataSerializer {
	switch filter.UserDataTypeSerializer.Type {
	case JwtUserDataSerializer:
		return serializers.NewJwtUserDataSerializer(
			filter.UserDataTypeSerializer.Secret,
		)
	default:
		panic(fmt.Errorf("Undefined user data serializer type: %v.\n", filter.UserDataTypeSerializer.Type))
	}
}

func (ctx *context) BuildServer(port int) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: ctx.serverMultiplexer,
	}
}
