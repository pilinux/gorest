// Package router contains all routes
package router

import (
	"github.com/gin-gonic/gin"

	gconfig "github.com/pilinux/gorest/config"
	gcontroller "github.com/pilinux/gorest/controller"
	gmiddleware "github.com/pilinux/gorest/lib/middleware"

	"github.com/pilinux/gorest/example/controller"
)

// SetupRouter sets up all the routes
func SetupRouter(configure *gconfig.Configuration) (*gin.Engine, error) {
	if configure.Server.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode) // Omit this line to enable debug mode
	}

	// Write log file
	// Console color is not required to write the logs to the file
	//	gin.DisableConsoleColor()

	// Create a log file with start time
	// dt := time.Now()
	// t := dt.Format(time.RFC3339)
	// file, err := os.Create("./logs/start:" + t + ".log")
	// if err != nil {
	//	 return nil, err
	// }
	// gin.DefaultWriter = io.MultiWriter(file)

	// If it is required to write the logs to the file and the console
	// at the same time
	// gin.DefaultWriter = io.MultiWriter(file, os.Stdout)

	// Creates a router without any middleware by default
	// router := gin.New()

	// Logger middleware: gin.DefaultWriter = os.Stdout
	// router.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500
	// if there is one
	// router.Use(gin.Recovery())

	// gin.Default() = gin.New() + gin.Logger() + gin.Recovery()
	r := gin.Default()

	// Which proxy to trust:
	// disable this feature as it still fails
	// to provide the real client IP in
	// different scenarios
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
	if configure.Security.MustCORS == gconfig.Activated {
		r.Use(gmiddleware.CORS(configure.Security.CORS))
	}

	// Sentry.io
	if configure.Logger.Activate == gconfig.Activated {
		r.Use(gmiddleware.SentryCapture(configure.Logger.SentryDsn))
	}

	// WAF
	if configure.Security.MustFW == gconfig.Activated {
		r.Use(gmiddleware.Firewall(
			configure.Security.Firewall.ListType,
			configure.Security.Firewall.IP,
		))
	}

	// Render HTML
	if configure.ViewConfig.Activate == gconfig.Activated {
		r.Use(gmiddleware.Pongo2(configure.ViewConfig.Directory))
	}

	// API Status
	r.GET("", controller.APIStatus)

	// API:v1.0
	v1 := r.Group("/api/v1/")
	{
		// RDBMS
		if configure.Database.RDBMS.Activate == gconfig.Activated {
			// Register - no JWT required
			v1.POST("register", gcontroller.CreateUserAuth)

			// Verify email
			if configure.Security.VerifyEmail {
				if configure.Database.REDIS.Activate == gconfig.Activated {
					v1.POST("verify", gcontroller.VerifyEmail)
					v1.POST("resend-verification-email", gcontroller.CreateVerificationEmail)
				}
			}

			// Login - app issues JWT
			v1.POST("login", gcontroller.Login)

			// Refresh - app issues new JWT
			rJWT := v1.Group("refresh")
			rJWT.Use(gmiddleware.RefreshJWT())
			rJWT.POST("", gcontroller.Refresh)

			// Double authentication
			if configure.Security.Must2FA == gconfig.Activated {
				r2FA := v1.Group("2fa")
				r2FA.Use(gmiddleware.JWT())
				r2FA.POST("setup", gcontroller.Setup2FA)
				r2FA.POST("activate", gcontroller.Activate2FA)
				r2FA.POST("validate", gcontroller.Validate2FA)
				if configure.Security.Must2FA == gconfig.Activated {
					r2FA.Use(gmiddleware.TwoFA(
						configure.Security.TwoFA.Status.On,
						configure.Security.TwoFA.Status.Off,
						configure.Security.TwoFA.Status.Verified,
					))
				}
				// disable 2FA
				r2FA.POST("deactivate", gcontroller.Deactivate2FA)
			}

			// Update/reset password
			rPass := v1.Group("password")
			// Reset forgotten password
			if configure.EmailConf.Activate == gconfig.Activated {
				// send password recovery email
				rPass.POST("forgot", gcontroller.PasswordForgot)
				// recover account and set new password
				rPass.POST("reset", gcontroller.PasswordRecover)
			}
			rPass.Use(gmiddleware.JWT())
			if configure.Security.Must2FA == gconfig.Activated {
				rPass.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			// change password while logged in
			rPass.POST("edit", gcontroller.PasswordUpdate)

			// User
			rUsers := v1.Group("users")
			rUsers.GET("", controller.GetUsers)    // Non-protected
			rUsers.GET("/:id", controller.GetUser) // Non-protected
			rUsers.Use(gmiddleware.JWT())
			if configure.Security.Must2FA == gconfig.Activated {
				rUsers.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rUsers.POST("", controller.CreateUser)      // Protected
			rUsers.PUT("", controller.UpdateUser)       // Protected
			rUsers.PUT("/hobbies", controller.AddHobby) // Protected

			// Post
			rPosts := v1.Group("posts")
			rPosts.GET("", controller.GetPosts)    // Non-protected
			rPosts.GET("/:id", controller.GetPost) // Non-protected
			rPosts.Use(gmiddleware.JWT())
			if configure.Security.Must2FA == gconfig.Activated {
				rPosts.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rPosts.POST("", controller.CreatePost)       // Protected
			rPosts.PUT("/:id", controller.UpdatePost)    // Protected
			rPosts.DELETE("/:id", controller.DeletePost) // Protected

			// Hobby
			rHobbies := v1.Group("hobbies")
			rHobbies.GET("", controller.GetHobbies) // Non-protected

			// Test JWT
			rTestJWT := v1.Group("test-jwt")
			rTestJWT.Use(gmiddleware.JWT())
			if configure.Security.Must2FA == gconfig.Activated {
				rTestJWT.Use(gmiddleware.TwoFA(
					configure.Security.TwoFA.Status.On,
					configure.Security.TwoFA.Status.Off,
					configure.Security.TwoFA.Status.Verified,
				))
			}
			rTestJWT.GET("", controller.AccessResource) // Protected
		}

		// REDIS Playground
		if configure.Database.REDIS.Activate == gconfig.Activated {
			rPlayground := v1.Group("playground")
			rPlayground.GET("/redis_read", controller.RedisRead)        // Non-protected
			rPlayground.POST("/redis_create", controller.RedisCreate)   // Non-protected
			rPlayground.DELETE("/redis_delete", controller.RedisDelete) // Non-protected

			rPlayground.GET("/redis_read_hash", controller.RedisReadHash)        // Non-protected
			rPlayground.POST("/redis_create_hash", controller.RedisCreateHash)   // Non-protected
			rPlayground.DELETE("/redis_delete_hash", controller.RedisDeleteHash) // Non-protected
		}

		// Mongo Playground
		if configure.Database.MongoDB.Activate == gconfig.Activated {
			rPlaygroundMongo := v1.Group("playground-mongo")
			rPlaygroundMongo.POST("/mongo_create_one", controller.MongoCreateOne)                 // Non-protected
			rPlaygroundMongo.GET("/mongo_get_all", controller.MongoGetAll)                        // Non-protected
			rPlaygroundMongo.GET("/mongo_get_by_id/:id", controller.MongoGetByID)                 // Non-protected
			rPlaygroundMongo.POST("/mongo_get_by_filter", controller.MongoGetByFilter)            // Non-protected
			rPlaygroundMongo.PUT("/mongo_update_by_id", controller.MongoUpdateByID)               // Non-protected
			rPlaygroundMongo.DELETE("/mongo_delete_field_by_id", controller.MongoDeleteFieldByID) // Non-protected
			rPlaygroundMongo.DELETE("/mongo_delete_doc_by_id/:id", controller.MongoDeleteByID)    // Non-protected
		}

		// Basic Auth demo
		if configure.Security.MustBasicAuth == gconfig.Activated {
			user := configure.Security.BasicAuth.Username
			pass := configure.Security.BasicAuth.Password
			rBasicAuth := v1.Group("access_resources")
			rBasicAuth.Use(gin.BasicAuth(gin.Accounts{
				user: pass,
			}))
			rBasicAuth.GET("", controller.AccessResource) // Protected
		}
	}

	return r, nil
}
