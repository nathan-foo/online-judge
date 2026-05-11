package config

var Routes = []RouteConfig{
	{
		Prefix:        "/users",
		ServiceUrl:    getConfig("USER_SERVICE_URL"),
		RequireAuth:   true,
		RateLimit:     10,
		MaxUploadSize: MaxUploadSize,
	},
	{
		Prefix:        "/webhooks/clerk",
		ServiceUrl:    getConfig("USER_SERVICE_URL"),
		RequireAuth:   false,
		RateLimit:     10,
		MaxUploadSize: MaxUploadSize,
	},
	{
		Prefix:        "/quizzes",
		ServiceUrl:    getConfig("QUIZ_SERVICE_URL"),
		RequireAuth:   true,
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
}
