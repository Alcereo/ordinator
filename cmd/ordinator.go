package main

import (
	"fmt"
	ctx "github.com/Alcereo/ordinator/pkg/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func main() {

	configInit()
	config := loadConfig()
	setupLogging(config.LogLevel)

	bytes, _ := yaml.Marshal(config)
	log.Tracef("Resolved config:\n%+v", string(bytes))

	context := ctx.NewContext()
	context.SetupCache(config.CacheAdapters)
	context.SetupRouters(config.Routers, config.GoogleSecret)

	port := viper.GetInt("port")
	log.Printf("Server starting on port %v", port)
	log.Fatal(context.BuildServer(port).ListenAndServe())
}

func setupLogging(logLevel ctx.LogLevel) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	switch logLevel {
	case ctx.Info:
		log.SetLevel(log.InfoLevel)
	case ctx.Debug:
		log.SetLevel(log.DebugLevel)
	case ctx.Trace:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func loadConfig() *ctx.ProxyConfiguration {
	var config ctx.ProxyConfiguration
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.SetEnvPrefix("")
	_ = viper.BindEnv("GOOGLE_CLIENT_ID")
	config.GoogleSecret.ClientId = viper.GetString("GOOGLE_CLIENT_ID")

	_ = viper.BindEnv("GOOGLE_CLIENT_SECRET")
	config.GoogleSecret.ClientSecret = viper.GetString("GOOGLE_CLIENT_SECRET")
	return &config
}

func configInit() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./cmd")

	// Defaults
	viper.SetDefault("port", 8080)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
