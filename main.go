package main

import (
	"balancer/auth"
	"balancer/balancer"
	"balancer/cache"
	"balancer/filters"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
)

type Context struct {
	sessionCacheAdapters  map[string]filters.SessionCachePort
	userAuthCacheAdapters map[string]auth.UserAuthCachePort
}

func main() {

	configInit()
	config := loadConfig()
	setupLogging(config)

	bytes, _ := yaml.Marshal(config)
	log.Tracef("Resolved config:\n%+v", string(bytes))

	var PrimaryContext = &Context{
		sessionCacheAdapters:  make(map[string]filters.SessionCachePort),
		userAuthCacheAdapters: make(map[string]auth.UserAuthCachePort),
	}

	setupCacheAdapters(config.CacheAdapters, PrimaryContext)
	setupRouters(config.Routers, config.GoogleSecret, PrimaryContext)

	port := viper.GetInt("port")
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func setupLogging(config *ProxyConfiguration) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	switch config.LogLevel {
	case info:
		log.SetLevel(log.InfoLevel)
	case debug:
		log.SetLevel(log.DebugLevel)
	case trace:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func setupRouters(routers []Router, googleSecret GoogleSecret, context *Context) {
	for _, router := range routers {
		switch router.Type {
		case ReverseProxy:
			log.Printf(
				"Adding Reverse proxy router. Pattern: %s; Target: %s",
				router.Pattern,
				router.TargetUrl,
			)

			targetUrl, _ := new(url.URL).Parse(router.TargetUrl)
			handler := balancer.ReversiveProxyHandler{
				TargetAddress: *targetUrl,
			}

			rootFilterHandler := BuildFilterHandlers(router.Filters, &handler, context)
			http.HandleFunc(router.Pattern, rootFilterHandler.Handle)
		case GoogleOauth2Authorization:
			log.Printf(
				"Adding Google Oauth2 authorization endpoint. Pattern: %s;",
				router.Pattern,
			)
			cacheAdapter := context.userAuthCacheAdapters[router.CacheAdapterIdentifier]
			if cacheAdapter == nil {
				panic(fmt.Errorf("User cache adapter with identifier '%v' not found.\n", router.CacheAdapterIdentifier))
			}
			handler := auth.NewGoogleOAuth2Provider(
				cacheAdapter,
				router.SuccessLoginUrl,
				googleSecret.ClientId,
				googleSecret.ClientSecret,
			)

			rootFilterHandler := BuildFilterHandlers(router.Filters, handler, context)
			http.HandleFunc(router.Pattern, rootFilterHandler.Handle)
		default:
			panic(fmt.Errorf("Undefined router type: %v.\n", router.Type))
		}
	}
}

func setupCacheAdapters(adapters []CacheAdapter, context *Context) {
	for _, adapter := range adapters {
		switch adapter.Type {
		case GoCache:
			provider := cache.NewGoCacheSessionCacheProvider(
				adapter.ExpirationTimeHours,
				adapter.EvictScheduleTimeHours,
			)
			// GoCache can be both
			context.sessionCacheAdapters[adapter.Identifier] = provider
			context.userAuthCacheAdapters[adapter.Identifier] = provider
		default:
			panic(fmt.Errorf("Undefined session filter cache adapter type: %v.\n", adapter.Type))
		}
	}
}

func loadConfig() *ProxyConfiguration {
	var config ProxyConfiguration
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.SetEnvPrefix("gw")
	_ = viper.BindEnv("GOOGLE_CLIENT_ID")
	config.GoogleSecret.ClientId = viper.GetString("GOOGLE_CLIENT_ID")

	_ = viper.BindEnv("GOOGLE_CLIENT_SECRET")
	config.GoogleSecret.ClientSecret = viper.GetString("GOOGLE_CLIENT_SECRET")
	return &config
}

func configInit() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// Defaults
	viper.SetDefault("port", 8080)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func BuildFilterHandlers(filters []Filter, mainHandler balancer.RequestHandler, context *Context) (rootHandler balancer.RequestHandler) {
	if filters == nil {
		return mainHandler
	}

	currentHandler := mainHandler

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]

		handler := buildFilterHandler(filter, context)

		if handler == nil {
			continue
		}

		handler.SetNext(currentHandler)
		currentHandler = handler
	}

	return currentHandler
}

func buildFilterHandler(filter Filter, context *Context) balancer.RequestChainedHandler {
	switch filter.Type {
	case LogFilter:
		log.Printf("Adding Log filter. Name: %s", filter.Name)
		return filters.CreateLogFilter(filter.Name, filter.Template, nil)
	case SessionFilter:
		log.Printf("Adding session filter. Name: %s", filter.Name)
		cacheAdapter := context.sessionCacheAdapters[filter.CacheAdapterIdentifier]
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
		log.Printf("Adding user authentication filter. Name: %s", filter.Name)
		cacheAdapter := context.userAuthCacheAdapters[filter.CacheAdapterIdentifier]
		if cacheAdapter == nil {
			panic(fmt.Errorf("User cache adapter with identifier '%v' not found.\n", filter.CacheAdapterIdentifier))
		}
		return auth.NewUserAuthenticationFilter(
			cacheAdapter,
			filter.Name,
		)
	default:
		panic(fmt.Errorf("Undefined filter type: %v.\n", filter.Type))
	}
}

type RouterType string

const (
	ReverseProxy              RouterType = "ReverseProxy"
	GoogleOauth2Authorization RouterType = "GoogleOauth2Authorization"
)

type FilterType string

const (
	LogFilter                FilterType = "LogFilter"
	SessionFilter            FilterType = "SessionFilter"
	UserAuthenticationFilter FilterType = "UserAuthenticationFilter"
)

type CacheAdapterType string

const (
	GoCache CacheAdapterType = "GoCache"
)

type CacheAdapter struct {
	Identifier             string
	Type                   CacheAdapterType
	ExpirationTimeHours    int `mapstructure:"evict-time-hours"`
	EvictScheduleTimeHours int `mapstructure:"evict-schedule-time-hours"`
}

type Filter struct {
	Type                   FilterType
	Name                   string
	Template               string
	CacheAdapterIdentifier string `mapstructure:"cache-adapter-identifier"`
	CookieDomain           string `mapstructure:"cookie-domain"`
	CookiePath             string `mapstructure:"cookie-path"`
	CookieName             string `mapstructure:"cookie-name"`
	CookieTTLHours         int    `mapstructure:"cookie-ttl-hours"`
	CookieRenewBeforeHours int    `mapstructure:"cookie-renew-before-hours"`
}

type Router struct {
	TargetUrl              string `mapstructure:"target-url"`
	Type                   RouterType
	Pattern                string
	Filters                []Filter
	CacheAdapterIdentifier string `mapstructure:"cache-adapter-identifier"`
	SuccessLoginUrl        string `mapstructure:"success-login-url"`
}

type LogLevel string

const (
	debug LogLevel = "debug"
	trace LogLevel = "trace"
	info  LogLevel = "info"
)

type GoogleSecret struct {
	ClientId     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
}

type ProxyConfiguration struct {
	GoogleSecret  GoogleSecret `mapstructure:"google-secret"`
	LogLevel      LogLevel     `mapstructure:"log-level"`
	Routers       []Router
	CacheAdapters []CacheAdapter `mapstructure:"cache-adapters"`
}
