package config

var Routes = []RouteConfig{
	{
		Prefix:        "/users",
		ServiceUrl:    getConfig("USER_SERVICE_URL"),
		RequireAuth:   false,
		RateLimit:     10,
		MaxUploadSize: MaxUploadSize,
	},
	// {
	// 	Prefix:        "/test",
	// 	ServiceUrl:    getConfig("TEST_SERVICE_URL"),
	// 	RequireAuth:   true,
	// 	RateLimit:     10,
	// 	MaxUploadSize: MaxUploadSize,
	// },
	// {
	// 	Prefix:        "/test-2",
	// 	ServiceUrl:    getConfig("TEST_SERVICE_URL_2"),
	// 	RequireAuth:   true,
	// 	RateLimit:     10,
	// 	MaxUploadSize: MaxUploadSize,
	// },
}
