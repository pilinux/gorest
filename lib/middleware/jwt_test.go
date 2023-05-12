package middleware_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

var customClaims middleware.MyCustomClaims

func TestGetJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	// valid access token
	accessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT: %v", err)
	}

	// valid refresh token
	refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT: %v", err)
	}

	if len(accessJWT) == 0 || len(refreshJWT) == 0 {
		t.Errorf("expected non-empty JWT values, got access: %s, refresh: %s", accessJWT, refreshJWT)
	}
}

func TestJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	// valid access token
	accessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT: %v", err)
	}

	// valid refresh token
	refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT: %v", err)
	}

	// access token to be valid 30 seconds in the future
	middleware.JWTParams.AccNbf = 30
	validInFutureAccessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT to be valid 30 seconds in the future: %v", err)
	}
	middleware.JWTParams.AccNbf = 0 // reset

	// access token set to expire immediately
	middleware.JWTParams.AccessKeyTTL = -1
	expiredAccessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating expired access JWT: %v", err)
	}

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
	}{
		{
			name:           "no authorization header",
			authorization:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty Bearer token",
			authorization:  "Bearer",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization header (first test)",
			authorization:  "Bearer invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization header (second test)",
			authorization:  "Bearer invalid token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid authorization header with refresh token",
			authorization:  "Bearer " + refreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired authorization header",
			authorization:  "Bearer " + expiredAccessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid authorization header set to be valid 30 seconds in the future",
			authorization:  "Bearer " + validInFutureAccessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid authorization header",
			authorization:  "Bearer " + accessJWT,
			expectedStatus: http.StatusOK,
		},
	}

	// set up a gin router and handler
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	err = router.SetTrustedProxies(nil)
	if err != nil {
		t.Errorf("failed to set trusted proxies to nil")
	}
	router.TrustedPlatform = "X-Real-Ip"

	router.Use(middleware.JWT())

	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request: %v", err)
				return
			}
			req.Header.Set("Authorization", test.authorization)
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != test.expectedStatus {
				t.Errorf("expected status code %d, got %d", test.expectedStatus, res.Code)
			}
		})
	}
}

func TestJWTAuthCookie(t *testing.T) {
	// set JWT params
	setParamsJWT()

	// valid access token
	accessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT: %v", err)
	}

	// valid refresh token
	refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT: %v", err)
	}

	// access token to be valid 30 seconds in the future
	middleware.JWTParams.AccNbf = 30
	validInFutureAccessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT to be valid 30 seconds in the future: %v", err)
	}
	middleware.JWTParams.AccNbf = 0 // reset

	// access token set to expire immediately
	middleware.JWTParams.AccessKeyTTL = -1
	expiredAccessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating expired access JWT: %v", err)
	}

	tests := []struct {
		name           string
		accessJWT      string
		expectedStatus int
	}{
		{
			name:           "empty cookie",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid access token",
			accessJWT:      "access token invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid refresh token",
			accessJWT:      refreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired access token",
			accessJWT:      expiredAccessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid access token set to be valid 30 seconds in the future",
			accessJWT:      validInFutureAccessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid access token",
			accessJWT:      accessJWT,
			expectedStatus: http.StatusOK,
		},
	}

	// set up a gin router and handler
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	err = router.SetTrustedProxies(nil)
	if err != nil {
		t.Errorf("failed to set trusted proxies to nil")
	}
	router.TrustedPlatform = "X-Real-Ip"

	router.Use(middleware.JWT())

	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request: %v", err)
				return
			}

			// set the HTTP-only cookie in the request
			req.AddCookie(&http.Cookie{
				Name:     "accessJWT",
				Value:    test.accessJWT,
				HttpOnly: true,
			})

			// create a new response recorder
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != test.expectedStatus {
				t.Errorf("expected status code %d, got %d", test.expectedStatus, res.Code)
			}
		})
	}
}

func TestRefreshJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	// valid access token
	accessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT: %v", err)
	}

	// valid refresh token
	refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT: %v", err)
	}

	// refresh token to be valid 30 seconds in the future
	middleware.JWTParams.RefNbf = 30 // set refresh token to be valid 30 seconds in the future
	validInFutureRefreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT to be valid 30 seconds in the future: %v", err)
	}
	middleware.JWTParams.RefNbf = 0 // reset

	// refresh token set to expire immediately
	middleware.JWTParams.RefreshKeyTTL = -1
	expiredRefreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating expired refresh JWT: %v", err)
	}

	testCases := []struct {
		name           string
		payload        string
		expectedStatus int
	}{
		{
			name:           "wrong payload type",
			payload:        `{"refreshJWT": 1}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid access JWT",
			payload:        fmt.Sprintf(`{"refreshJWT": "%s"}`, accessJWT),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid refresh JWT",
			payload:        fmt.Sprintf(`{"refreshJWT": "%s"}`, refreshJWT),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "refresh JWT to be valid in the future",
			payload:        fmt.Sprintf(`{"refreshJWT": "%s"}`, validInFutureRefreshJWT),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired refresh JWT",
			payload:        fmt.Sprintf(`{"refreshJWT": "%s"}`, expiredRefreshJWT),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	// set up a gin router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	err = router.SetTrustedProxies(nil)
	if err != nil {
		t.Errorf("failed to set trusted proxies to nil")
	}
	router.TrustedPlatform = "X-Real-Ip"

	// define the route and attach the RefreshJWT middleware
	router.POST("/refresh", middleware.RefreshJWT(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte(tc.payload)))
			if err != nil {
				t.Errorf("failed to create an HTTP request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, res.Code)
			}
		})
	}
}

func TestRefreshJWTAuthCookie(t *testing.T) {
	// set JWT params
	setParamsJWT()

	// valid access token
	accessJWT, _, err := middleware.GetJWT(customClaims, "access")
	if err != nil {
		t.Errorf("error creating access JWT: %v", err)
	}

	// valid refresh token
	refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT: %v", err)
	}

	// refresh token to be valid 30 seconds in the future
	middleware.JWTParams.RefNbf = 30 // set refresh token to be valid 30 seconds in the future
	validInFutureRefreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating refresh JWT to be valid 30 seconds in the future: %v", err)
	}
	middleware.JWTParams.RefNbf = 0 // reset

	// refresh token set to expire immediately
	middleware.JWTParams.RefreshKeyTTL = -1
	expiredRefreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
	if err != nil {
		t.Errorf("error creating expired refresh JWT: %v", err)
	}

	testCases := []struct {
		name           string
		refreshJWT     string
		expectedStatus int
	}{
		{
			name:           "empty cookie",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid refresh token",
			refreshJWT:     "refresh token invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid access JWT",
			refreshJWT:     accessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "refresh JWT to be valid in the future",
			refreshJWT:     validInFutureRefreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired refresh JWT",
			refreshJWT:     expiredRefreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid refresh JWT",
			refreshJWT:     refreshJWT,
			expectedStatus: http.StatusOK,
		},
	}

	// set up a gin router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	err = router.SetTrustedProxies(nil)
	if err != nil {
		t.Errorf("failed to set trusted proxies to nil")
	}
	router.TrustedPlatform = "X-Real-Ip"

	// define the route and attach the RefreshJWT middleware
	router.POST("/refresh", middleware.RefreshJWT(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte("")))
			if err != nil {
				t.Errorf("failed to create an HTTP request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// set the HTTP-only cookie in the request
			req.AddCookie(&http.Cookie{
				Name:     "refreshJWT",
				Value:    tc.refreshJWT,
				HttpOnly: true,
			})

			// create a new response recorder
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)
			if res.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, res.Code)
			}
		})
	}
}

// set params
func setParamsJWT() {
	middleware.JWTParams.AccessKey = []byte("cryptographic_key_1")
	middleware.JWTParams.AccessKeyTTL = 5
	middleware.JWTParams.RefreshKey = []byte("cryptographic_key_2")
	middleware.JWTParams.RefreshKeyTTL = 60

	middleware.JWTParams.Audience = "audience"
	middleware.JWTParams.Issuer = "gorest"
	middleware.JWTParams.AccNbf = 0
	middleware.JWTParams.RefNbf = 0
	middleware.JWTParams.Subject = "subject"

	customClaims.AuthID = 123
	customClaims.Email = "test@example.com"
	customClaims.Role = "admin"
	customClaims.Scope = "full_access"
	customClaims.TwoFA = "on"
	customClaims.SiteLan = "en"
	customClaims.Custom1 = "custom value 1"
	customClaims.Custom2 = "custom value 2"
}
