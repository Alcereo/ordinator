package context

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
	UserDataSenderFilter     FilterType = "UserDataSenderFilter"
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

type UserDataSerializerType string

const (
	JwtUserDataSerializer UserDataSerializerType = "JwtUserDataSerializer"
)

type UserDataSerializer struct {
	Type   UserDataSerializerType
	Secret string
}

type Filter struct {
	Type                   FilterType
	Name                   string
	Template               string
	CacheAdapterIdentifier string             `mapstructure:"cache-adapter-identifier"`
	CookieDomain           string             `mapstructure:"cookie-domain"`
	CookiePath             string             `mapstructure:"cookie-path"`
	CookieName             string             `mapstructure:"cookie-name"`
	CookieTTLHours         int                `mapstructure:"cookie-ttl-hours"`
	CookieRenewBeforeHours int                `mapstructure:"cookie-renew-before-hours"`
	UserDataTypeSerializer UserDataSerializer `mapstructure:"user-data-serializer"`
	UserDataHeader         string             `mapstructure:"user-data-header"`
}

type Router struct {
	TargetUrl              string `mapstructure:"target-url"`
	Type                   RouterType
	Pattern                string
	Filters                []Filter
	CacheAdapterIdentifier string `mapstructure:"cache-adapter-identifier"`
	SuccessLoginUrl        string `mapstructure:"success-login-url"`
	AccessTokenRequestUrl  string `mapstructure:"access-toke-request-url"`
	UserInfoRequestUrl     string `mapstructure:"user-info-request-url"`
}

type LogLevel string

const (
	Debug LogLevel = "debug"
	Trace LogLevel = "trace"
	Info  LogLevel = "info"
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
