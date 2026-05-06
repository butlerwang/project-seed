package handler_test

import "github.com/butlerwang/project-seed/backend/internal/config"

func testConfig() config.Config {
	return config.Config{
		JWTSecret:      "test-secret-at-least-32-chars-xxxx",
		AppURL:         "http://localhost:8080",
		FrontendURL:    "http://localhost:3000",
		CORSOrigins:    []string{"http://localhost:3000"},
		RateLimitRPS:   100,
		RateLimitBurst: 200,
		StorageBucket:  "test-bucket",
	}
}
