package main

import (
	"balancer/balancer"
	"balancer/filters"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
)

func main() {

	configInit()
	config := loadConfig()

	for _, router := range config.Routers {
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

			rootFilterHandler := BuildFilterHandlers(router.Filters, &handler)
			http.HandleFunc(router.Pattern, rootFilterHandler.Handle)

		default:
			panic(fmt.Errorf("Undefined router type: %v.\n", router.Type))
		}
	}

	port := viper.GetInt("port")
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func loadConfig() *ProxyConfiguration {
	var config ProxyConfiguration
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return &config
}

func configInit() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".") // optionally look for config in the working directory

	// Defaults
	viper.SetDefault("port", 8080)

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func BuildFilterHandlers(filters []Filter, mainHandler balancer.RequestHandler) (rootHandler balancer.RequestHandler) {
	if filters == nil {
		return mainHandler
	}

	currentHandler := mainHandler

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]

		handler := buildFilterHandler(filter)

		if handler == nil {
			continue
		}

		handler.SetNext(currentHandler)
		currentHandler = handler
	}

	return currentHandler
}

func buildFilterHandler(filter Filter) balancer.RequestChainedHandler {
	switch filter.Type {
	case LogFilter:
		log.Printf("Adding Log filter. Name: %s", filter.Name)
		return filters.CreateLogFilter(filter.Name, filter.Template, nil)
	//case SessionFilter:
	//	log.Printf("Adding session filter. Name: %s", filter.Name)
	//	sessionCacheProvider := cache.CreateStubCacheProvider()
	//	return filters.CreateSessionFilter(filter.SessionCookie, sessionCacheProvider)
	default:
		panic(fmt.Errorf("Undefined filter type: %v.\n", filter.Type))
	}
}

type RouterType string

const (
	ReverseProxy RouterType = "ReverseProxy"
)

type FilterType string

const (
	LogFilter     FilterType = "LogFilter"
	SessionFilter FilterType = "SessionFilter"
)

type Filter struct {
	Type          FilterType
	Name          string
	Template      string
	SessionCookie string `mapstructure:"session-cookie"`
}

type Router struct {
	TargetUrl string `mapstructure:"target-url"`
	Type      RouterType
	Pattern   string
	Filters   []Filter
}

type ProxyConfiguration struct {
	Routers []Router
}
