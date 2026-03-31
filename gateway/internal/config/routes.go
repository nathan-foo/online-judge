package config

var Routes = []RouteConfig{
	{
		Prefix:        "/test",
		ServiceUrl:    getConfig("TEST_SERVICE_URL"),
		RequireAuth:   true,
		RateLimit:     10,
		MaxUploadSize: MaxUploadSize,
	},
}
