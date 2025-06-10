// Package router contains all routes of the example2 application.
package router

import (
	"github.com/gin-gonic/gin"

	gconfig "github.com/pilinux/gorest/config"
	gcontroller "github.com/pilinux/gorest/controller"
	gdb "github.com/pilinux/gorest/database"
	glib "github.com/pilinux/gorest/lib"
	gmiddleware "github.com/pilinux/gorest/lib/middleware"
	gservice "github.com/pilinux/gorest/service"

	"github.com/pilinux/gorest/example2/internal/handler"
	"github.com/pilinux/gorest/example2/internal/repo"
	"github.com/pilinux/gorest/example2/internal/service"
)

// SetupRouter sets up all the routes.
func SetupRouter(configure *gconfig.Configuration) (*gin.Engine, error) {
	// Set Gin mode
	if gconfig.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.Default() = gin.New() + gin.Logger() + gin.Recovery()
	r := gin.Default()

	// Which proxy to trust:
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return r, err
	}

	// when using Cloudflare's CDN:
	// router.TrustedPlatform = gin.PlatformCloudflare
	//
	// when running on Google App Engine:
	// router.TrustedPlatform = gin.PlatformGoogleAppEngine
	//
	/*
		when using apache or nginx reverse proxy
		without Cloudflare's CDN or Google App Engine

		config for nginx:
		=================
		proxy_set_header X-Real-IP       $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	*/
	// router.TrustedPlatform = "X-Real-Ip"
	//
	// set TrustedPlatform to get the real client IP
	trustedPlatform := configure.Security.TrustedPlatform
	if trustedPlatform == "cf" {
		trustedPlatform = gin.PlatformCloudflare
	}
	if trustedPlatform == "google" {
		trustedPlatform = gin.PlatformGoogleAppEngine
	}
	r.TrustedPlatform = trustedPlatform

	// CORS
	if gconfig.IsCORS() {
		r.Use(gmiddleware.CORS(configure.Security.CORS))
	}

	// Origin
	if gconfig.IsOriginCheck() {
		r.Use(gmiddleware.CheckOrigin(gmiddleware.GetCORS().AllowedOrigins))
	}

	// Sentry.io
	if gconfig.IsSentry() {
		_, err := gmiddleware.InitSentry(
			configure.Logger.SentryDsn,
			configure.Server.ServerEnv,
			configure.Version,
			configure.Logger.PerformanceTracing,
			configure.Logger.TracesSampleRate,
		)
		if err != nil {
			return r, err
		}
		r.Use(gmiddleware.SentryCapture())
	}

	// WAF
	if gconfig.IsWAF() {
		r.Use(gmiddleware.Firewall(
			configure.Security.Firewall.ListType,
			configure.Security.Firewall.IP,
		))
	}

	// Rate Limiter
	if gconfig.IsRateLimit() {
		// create a rate limiter instance
		limiterInstance, err := glib.InitRateLimiter(
			configure.Security.RateLimit,
			trustedPlatform,
		)
		if err != nil {
			return r, err
		}
		r.Use(gmiddleware.RateLimit(limiterInstance))
	}

	// Render HTML
	if gconfig.IsTemplatingEngine() {
		r.Use(gmiddleware.Pongo2(configure.ViewConfig.Directory))
	}

	// API Status

	// API:v1.0
	v1 := r.Group("/api/v1/")
	{
		// RDBMS
		if gconfig.IsRDBMS() {
			// Register - no JWT required
			v1.POST("register", gcontroller.CreateUserAuth)

			// Verify email
			if gconfig.IsEmailVerificationService() {
				if gconfig.IsRedis() {
					v1.POST("verify", gcontroller.VerifyEmail)
					v1.POST("resend-verification-email", gcontroller.CreateVerificationEmail)
					v1.POST("verify-updated-email", gcontroller.VerifyUpdatedEmail)
				}
			}

			// Login - app issues JWT
			// - if cookie management is enabled, save tokens on client browser
			v1.POST("login", gcontroller.Login)

			// Logout
			// - if cookie management is enabled, delete tokens from cookies
			// - if Redis is enabled, save tokens in a blacklist until TTL
			rLogout := v1.Group("logout")
			rLogout.Use(gmiddleware.JWT()).Use(gmiddleware.RefreshJWT()).Use(gservice.JWTBlacklistChecker())
			rLogout.POST("", gcontroller.Logout)

			// Refresh - app issues new JWT
			// - if cookie management is enabled, save tokens on client browser
			rJWT := v1.Group("refresh")
			rJWT.Use(gmiddleware.RefreshJWT()).Use(gservice.JWTBlacklistChecker())
			rJWT.POST("", gcontroller.Refresh)

			// Double authentication
			if gconfig.Is2FA() {
				r2FA := v1.Group("2fa")
				r2FA.Use(gmiddleware.JWT()).Use(gservice.JWTBlacklistChecker())
				r2FA.POST("setup", gcontroller.Setup2FA)
				r2FA.POST("activate", gcontroller.Activate2FA)
				r2FA.POST("validate", gcontroller.Validate2FA)
				r2FA.POST("validate-backup-code", gcontroller.ValidateBackup2FA)

				r2FA.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
				// get 2FA backup codes
				r2FA.POST("create-backup-codes", gcontroller.CreateBackup2FA)
				// disable 2FA
				r2FA.POST("deactivate", gcontroller.Deactivate2FA)
			}

			// Update/reset password
			rPass := v1.Group("password")
			// Reset forgotten password
			if gconfig.IsEmailService() {
				// send password recovery email
				rPass.POST("forgot", gcontroller.PasswordForgot)
				// recover account and set new password
				rPass.POST("reset", gcontroller.PasswordRecover)
			}
			rPass.Use(gmiddleware.JWT()).Use(gservice.JWTBlacklistChecker())
			if gconfig.Is2FA() {
				rPass.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			// change password while logged in
			rPass.POST("edit", gcontroller.PasswordUpdate)

			// Change existing email
			rEmail := v1.Group("email")
			rEmail.Use(gmiddleware.JWT()).Use(gservice.JWTBlacklistChecker())
			if gconfig.Is2FA() {
				rPass.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			// add new email to replace the existing one
			rEmail.POST("update", gcontroller.UpdateEmail)
			// retrieve the email which needs to be verified
			rEmail.GET("unverified", gcontroller.GetUnverifiedEmail)
			// resend verification code to verify the modified email address
			rEmail.POST("resend-verification-email", gcontroller.ResendVerificationCodeToModifyActiveEmail)

			// db client
			db := gdb.GetDB()

			userRepo := repo.NewUserRepo(db)
			postRepo := repo.NewPostRepo(db)

			// User
			userSrv := service.NewUserService(userRepo, postRepo)
			userAPI := handler.NewUserAPI(userSrv)

			rUsers := v1.Group("users")
			rUsers.GET("", userAPI.GetUsers)    // Not protected
			rUsers.GET("/:id", userAPI.GetUser) // Not protected
			rUsers.Use(gmiddleware.JWT())
			rUsers.Use(gservice.JWTBlacklistChecker())
			if gconfig.Is2FA() {
				rUsers.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rUsers.POST("", userAPI.CreateUser)   // Protected
			rUsers.PUT("", userAPI.UpdateUser)    // Protected
			rUsers.DELETE("", userAPI.DeleteUser) // Protected, irreversible

			// Post
			postSrv := service.NewPostService(postRepo, userRepo)
			postAPI := handler.NewPostAPI(postSrv)

			rPosts := v1.Group("posts")
			rPosts.GET("", postAPI.GetPosts)    // Not protected
			rPosts.GET("/:id", postAPI.GetPost) // Not protected
			rPosts.Use(gmiddleware.JWT())
			rPosts.Use(gservice.JWTBlacklistChecker())
			if gconfig.Is2FA() {
				rPosts.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rPosts.POST("", postAPI.CreatePost)                // Protected
			rPosts.PUT("/:id", postAPI.UpdatePost)             // Protected
			rPosts.DELETE("/:id", postAPI.DeletePost)          // Protected
			rPosts.DELETE("all", postAPI.DeleteAllPostsOfUser) // Protected

			// Hobby
			hobbyRepo := repo.NewHobbyRepo(db)
			hobbySrv := service.NewHobbyService(hobbyRepo, userRepo)
			hobbyAPI := handler.NewHobbyAPI(hobbySrv)

			rHobbies := v1.Group("hobbies")
			rHobbies.GET("", hobbyAPI.GetHobbies)   // Not protected
			rHobbies.GET("/:id", hobbyAPI.GetHobby) // Not protected
			rHobbies.Use(gmiddleware.JWT())
			rHobbies.Use(gservice.JWTBlacklistChecker())
			if gconfig.Is2FA() {
				rHobbies.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rHobbies.GET("me", hobbyAPI.GetHobbiesMe)             // Protected
			rHobbies.POST("", hobbyAPI.AddHobbyToUser)            // Protected
			rHobbies.DELETE("/:id", hobbyAPI.DeleteHobbyFromUser) // Protected
		}
	}

	return r, nil
}
