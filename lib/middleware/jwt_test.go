package middleware_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pilinux/gorest/lib/middleware"
)

var customClaims middleware.MyCustomClaims

func TestGetJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	testCases := []struct {
		Algorithm    string
		PrivKeyECDSA interface{}
		PrivKeyRSA   interface{}
		ExpectedErr  error
	}{
		// HMAC
		{
			Algorithm:   "HS256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS512",
			ExpectedErr: nil,
		},
		// ECDSA
		{
			Algorithm:   "ES256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES512",
			ExpectedErr: nil,
		},
		// RSA
		{
			Algorithm:   "RS256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS512",
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Algorithm, func(t *testing.T) {
			middleware.JWTParams.Algorithm = tc.Algorithm

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" ||
				tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				fileName := "private-key" + tc.Algorithm + ".pem"

				// download a file from a remote location and save it
				fileUrl := strings.TrimSpace(os.Getenv("TEST_KEY_FILE_LOCATION"))
				fileUrl += "/" + fileName
				err := downloadFile(fileName, fileUrl)
				if err != nil {
					t.Error(err)
				}

				privateKeyFile := "./" + fileName
				privateKeyBytes, errThis := os.ReadFile(privateKeyFile)
				if errThis != nil {
					t.Errorf("failed to read %v, error: %v", privateKeyBytes, errThis)
				}

				// remove the downloaded file
				err = os.RemoveAll(fileName)
				if err != nil {
					t.Error(err)
				}

				if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" {
					privateKey, errThis := jwt.ParseECPrivateKeyFromPEM(privateKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read privateKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PrivKeyECDSA = privateKey
				}

				if tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
					privateKey, errThis := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read privateKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PrivKeyRSA = privateKey
				}
			}

			// valid access token
			accessJWT, _, err := middleware.GetJWT(customClaims, "access")
			if err != tc.ExpectedErr {
				t.Errorf("unexpected error: got %v, want %v", err, tc.ExpectedErr)
			}

			// valid refresh token
			refreshJWT, _, err := middleware.GetJWT(customClaims, "refresh")
			if err != tc.ExpectedErr {
				t.Errorf("unexpected error: got %v, want %v", err, tc.ExpectedErr)
			}

			if len(accessJWT) == 0 || len(refreshJWT) == 0 {
				t.Errorf("expected non-empty JWT values, got access: %s, refresh: %s", accessJWT, refreshJWT)
			}

			// invalid token type
			_, _, err = middleware.GetJWT(customClaims, "invalid")
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
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

func TestRefreshJWTPayload(t *testing.T) {
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
			expectedStatus: http.StatusUnauthorized,
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

func TestRefreshJWTAuthHeader(t *testing.T) {
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
			name:           "valid authorization header with access token",
			authorization:  "Bearer " + accessJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired authorization header",
			authorization:  "Bearer " + expiredRefreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid authorization header set to be valid 30 seconds in the future",
			authorization:  "Bearer " + validInFutureRefreshJWT,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid authorization header",
			authorization:  "Bearer " + refreshJWT,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid authorization header with access and refresh tokens",
			authorization:  "Bearer " + accessJWT + " " + refreshJWT,
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
			req, err := http.NewRequest("POST", "/refresh", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request: %v", err)
				return
			}
			req.Header.Set("Authorization", tc.authorization)
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, res.Code)
			}
		})
	}
}

func TestValidateAccessJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	testCases := []struct {
		Algorithm   string
		PubKeyECDSA interface{}
		PubKeyRSA   interface{}

		ExpectedAlg string
		ExpectedKey interface{}
		ExpectedErr error
	}{
		// HMAC
		{
			Algorithm:   "HS256",
			ExpectedAlg: "HS256",
			ExpectedKey: []byte("cryptographic_key_1"),
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS384",
			ExpectedAlg: "HS384",
			ExpectedKey: []byte("cryptographic_key_1"),
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS512",
			ExpectedAlg: "HS512",
			ExpectedKey: []byte("cryptographic_key_1"),
			ExpectedErr: nil,
		},
		// unknown algorithm
		{
			Algorithm:   "unknown",
			ExpectedAlg: "unknown",
			ExpectedKey: nil,
			ExpectedErr: fmt.Errorf("unexpected signing method: %v", "unknown"),
		},
		// ECDSA
		{
			Algorithm:   "ES256",
			ExpectedAlg: "ES256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES384",
			ExpectedAlg: "ES384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES512",
			ExpectedAlg: "ES512",
			ExpectedErr: nil,
		},
		// RSA
		{
			Algorithm:   "RS256",
			ExpectedAlg: "RS256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS384",
			ExpectedAlg: "RS384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS512",
			ExpectedAlg: "RS512",
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Algorithm, func(t *testing.T) {
			middleware.JWTParams.Algorithm = tc.Algorithm
			var token *jwt.Token

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" ||
				tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				fileName := "public-key" + tc.Algorithm + ".pem"

				// download a file from a remote location and save it
				fileUrl := strings.TrimSpace(os.Getenv("TEST_KEY_FILE_LOCATION"))
				fileUrl += "/" + fileName
				err := downloadFile(fileName, fileUrl)
				if err != nil {
					t.Error(err)
				}

				publicKeyFile := "./" + fileName
				publicKeyBytes, errThis := os.ReadFile(publicKeyFile)
				if errThis != nil {
					t.Errorf("failed to read %v, error: %v", publicKeyFile, errThis)
				}

				// remove the downloaded file
				err = os.RemoveAll(fileName)
				if err != nil {
					t.Error(err)
				}

				if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" {
					publicKey, errThis := jwt.ParseECPublicKeyFromPEM(publicKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read publicKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PubKeyECDSA = publicKey
					tc.ExpectedKey = publicKey
				}

				if tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
					publicKey, errThis := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read publicKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PubKeyRSA = publicKey
					tc.ExpectedKey = publicKey
				}
			}

			if tc.Algorithm == "HS256" || tc.Algorithm == "HS384" || tc.Algorithm == "HS512" || tc.Algorithm == "unknown" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodHMAC{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodECDSA{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			if tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodRSA{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			key, err := middleware.ValidateAccessJWT(token)

			if err != nil && tc.ExpectedErr != nil && err.Error() != tc.ExpectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tc.ExpectedErr)
			}

			if err == nil && tc.ExpectedErr != nil {
				t.Errorf("expected error: got nil, want %v", tc.ExpectedErr)
			}

			if token.Header["alg"] != tc.ExpectedAlg {
				t.Errorf("unexpected algorithm: got %v, want %v", token.Header["alg"], tc.ExpectedAlg)
			}

			if tc.Algorithm == "HS256" || tc.Algorithm == "HS384" || tc.Algorithm == "HS512" || tc.Algorithm == "unknown" {
				if key != nil && tc.ExpectedKey != nil {
					if len(key.([]byte)) != len(tc.ExpectedKey.([]byte)) {
						t.Errorf("unexpected key length: got %v, want %v", len(key.([]byte)), len(tc.ExpectedKey.([]byte)))
					} else {
						for i := 0; i < len(key.([]byte)); i++ {
							if key.([]byte)[i] != tc.ExpectedKey.([]byte)[i] {
								t.Errorf("unexpected key value at index %d: got %v, want %v", i, key.([]byte)[i], tc.ExpectedKey.([]byte)[i])
								break
							}
						}
					}
				}
			}

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" ||
				tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				if key != nil && tc.ExpectedKey != nil {
					if key != tc.ExpectedKey {
						t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
					}
				}
			}

			if key == nil && tc.ExpectedKey != nil {
				t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
			}

			if key != nil && tc.ExpectedKey == nil {
				t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
			}
		})
	}
}

func TestValidateRefreshJWT(t *testing.T) {
	// set JWT params
	setParamsJWT()

	testCases := []struct {
		Algorithm   string
		PubKeyECDSA interface{}
		PubKeyRSA   interface{}

		ExpectedAlg string
		ExpectedKey interface{}
		ExpectedErr error
	}{
		// HMAC
		{
			Algorithm:   "HS256",
			ExpectedAlg: "HS256",
			ExpectedKey: []byte("cryptographic_key_2"),
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS384",
			ExpectedAlg: "HS384",
			ExpectedKey: []byte("cryptographic_key_2"),
			ExpectedErr: nil,
		},
		{
			Algorithm:   "HS512",
			ExpectedAlg: "HS512",
			ExpectedKey: []byte("cryptographic_key_2"),
			ExpectedErr: nil,
		},
		// unknown algorithm
		{
			Algorithm:   "unknown",
			ExpectedAlg: "unknown",
			ExpectedKey: nil,
			ExpectedErr: fmt.Errorf("unexpected signing method: %v", "unknown"),
		},
		// ECDSA
		{
			Algorithm:   "ES256",
			ExpectedAlg: "ES256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES384",
			ExpectedAlg: "ES384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "ES512",
			ExpectedAlg: "ES512",
			ExpectedErr: nil,
		},
		// RSA
		{
			Algorithm:   "RS256",
			ExpectedAlg: "RS256",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS384",
			ExpectedAlg: "RS384",
			ExpectedErr: nil,
		},
		{
			Algorithm:   "RS512",
			ExpectedAlg: "RS512",
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Algorithm, func(t *testing.T) {
			middleware.JWTParams.Algorithm = tc.Algorithm
			var token *jwt.Token

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" ||
				tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				fileName := "public-key" + tc.Algorithm + ".pem"

				// download a file from a remote location and save it
				fileUrl := strings.TrimSpace(os.Getenv("TEST_KEY_FILE_LOCATION"))
				fileUrl += "/" + fileName
				err := downloadFile(fileName, fileUrl)
				if err != nil {
					t.Error(err)
				}

				publicKeyFile := "./" + fileName
				publicKeyBytes, errThis := os.ReadFile(publicKeyFile)
				if errThis != nil {
					t.Errorf("failed to read %v, error: %v", publicKeyFile, errThis)
				}

				// remove the downloaded file
				err = os.RemoveAll(fileName)
				if err != nil {
					t.Error(err)
				}

				if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" {
					publicKey, errThis := jwt.ParseECPublicKeyFromPEM(publicKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read publicKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PubKeyECDSA = publicKey
					tc.ExpectedKey = publicKey
				}

				if tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
					publicKey, errThis := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
					if errThis != nil {
						t.Errorf("failed to read publicKeyBytes, error: %v", errThis)
					}
					middleware.JWTParams.PubKeyRSA = publicKey
					tc.ExpectedKey = publicKey
				}
			}

			if tc.Algorithm == "HS256" || tc.Algorithm == "HS384" || tc.Algorithm == "HS512" || tc.Algorithm == "unknown" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodHMAC{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodECDSA{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			if tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				token = &jwt.Token{
					Method: &jwt.SigningMethodRSA{
						Name: tc.Algorithm,
					},
					Header: map[string]interface{}{
						"alg": tc.Algorithm,
					},
				}
			}

			key, err := middleware.ValidateRefreshJWT(token)

			if err != nil && tc.ExpectedErr != nil && err.Error() != tc.ExpectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tc.ExpectedErr)
			}

			if err == nil && tc.ExpectedErr != nil {
				t.Errorf("expected error: got nil, want %v", tc.ExpectedErr)
			}

			if token.Header["alg"] != tc.ExpectedAlg {
				t.Errorf("unexpected algorithm: got %v, want %v", token.Header["alg"], tc.ExpectedAlg)
			}

			if tc.Algorithm == "HS256" || tc.Algorithm == "HS384" || tc.Algorithm == "HS512" || tc.Algorithm == "unknown" {
				if key != nil && tc.ExpectedKey != nil {
					if len(key.([]byte)) != len(tc.ExpectedKey.([]byte)) {
						t.Errorf("unexpected key length: got %v, want %v", len(key.([]byte)), len(tc.ExpectedKey.([]byte)))
					} else {
						for i := 0; i < len(key.([]byte)); i++ {
							if key.([]byte)[i] != tc.ExpectedKey.([]byte)[i] {
								t.Errorf("unexpected key value at index %d: got %v, want %v", i, key.([]byte)[i], tc.ExpectedKey.([]byte)[i])
								break
							}
						}
					}
				}
			}

			if tc.Algorithm == "ES256" || tc.Algorithm == "ES384" || tc.Algorithm == "ES512" ||
				tc.Algorithm == "RS256" || tc.Algorithm == "RS384" || tc.Algorithm == "RS512" {
				if key != nil && tc.ExpectedKey != nil {
					if key != tc.ExpectedKey {
						t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
					}
				}
			}

			if key == nil && tc.ExpectedKey != nil {
				t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
			}

			if key != nil && tc.ExpectedKey == nil {
				t.Errorf("unexpected key: got %v, want %v", key, tc.ExpectedKey)
			}
		})
	}
}

func TestValidateFailure(t *testing.T) {
	// mock a JWT token with an unsupported signing method (e.g., "PS256" - RSASSA-PSS using SHA-256)
	token := jwt.New(jwt.GetSigningMethod("PS256"))

	testCases := []struct {
		testName  string
		validator func(*jwt.Token) (interface{}, error)
	}{
		{
			testName:  "HMAC-Access",
			validator: middleware.ValidateHMACAccess,
		},
		{
			testName:  "HMAC-Refresh",
			validator: middleware.ValidateHMACRefresh,
		},
		{
			testName:  "ECDSA",
			validator: middleware.ValidateECDSA,
		},
		{
			testName:  "RSA",
			validator: middleware.ValidateRSA,
		},
	}

	// loop through the test cases
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// call the respective validator function based on the test case
			_, err := tc.validator(token)

			// the test should fail since the token has an unsupported signing method
			if err == nil {
				t.Errorf("expected an error, but got nil")
			}

			// check the specific error message
			expectedErrorMessage := "unexpected signing method: PS256"
			if err.Error() != expectedErrorMessage {
				t.Errorf("Expected error message: '%s', but got: '%s'", expectedErrorMessage, err.Error())
			}
		})
	}
}

// set params
func setParamsJWT() {
	middleware.JWTParams.Algorithm = "HS256"
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
