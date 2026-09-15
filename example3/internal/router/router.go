// Package router wires together all routes of the example3 application.
package router

import (
	"github.com/gin-gonic/gin"

	gconfig "github.com/pilinux/gorest/config"
	gdb "github.com/pilinux/gorest/database"
	glib "github.com/pilinux/gorest/lib"
	gmiddleware "github.com/pilinux/gorest/lib/middleware"

	"github.com/pilinux/gorest/example3/internal/handler"
	"github.com/pilinux/gorest/example3/internal/repo"
	"github.com/pilinux/gorest/example3/internal/service"
)

// SetupRouter builds the gin engine for example3.
//
// The master key must already be loaded before any request is served.
func SetupRouter(
	configure *gconfig.Configuration,
	keyManager *service.KeyManager,
	encryptedFilesDir string,
	maxUploadSize int64,
) (*gin.Engine, error) {
	if gconfig.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.Default() = gin.New() + gin.Logger() + gin.Recovery()
	r := gin.Default()

	if err := r.SetTrustedProxies(nil); err != nil {
		return r, err
	}

	// CORS
	if gconfig.IsCORS() {
		r.Use(gmiddleware.CORS(configure.Security.CORS))
	}

	// Rate Limiter
	if gconfig.IsRateLimit() {
		limiterInstance, err := glib.InitRateLimiter(
			configure.Security.RateLimit,
			configure.Security.TrustedPlatform,
		)
		if err != nil {
			return r, err
		}
		r.Use(gmiddleware.RateLimit(limiterInstance))
	}

	// Health
	r.GET("", handler.APIStatus)
	r.GET("health", handler.APIStatus)

	// example3 only makes sense with MongoDB (the master key lives there)
	if !gconfig.IsMongo() {
		return r, gdb.ErrMongoNotInitialized
	}
	conn := gdb.GetMongo()
	if conn == nil {
		return r, gdb.ErrMongoNotInitialized
	}

	// build the file metadata repository and the two crypto services
	fileRecordRepo := repo.NewFileRecordRepo(conn)

	textSrv := service.NewTextCryptService(keyManager)
	textAPI := handler.NewTextCryptAPI(textSrv)

	fileSrv := service.NewFileCryptService(keyManager, fileRecordRepo, encryptedFilesDir, maxUploadSize)
	fileAPI := handler.NewFileCryptAPI(fileSrv)

	v1 := r.Group("/api/v1/")
	{
		cryptoGroup := v1.Group("crypto")

		// text
		cryptoGroup.POST("text/encrypt", textAPI.EncryptText)
		cryptoGroup.POST("text/decrypt", textAPI.DecryptText)

		// number
		cryptoGroup.POST("number/encrypt", textAPI.EncryptNumber)
		cryptoGroup.POST("number/decrypt", textAPI.DecryptNumber)

		// files. Both encrypt routes store into the same place, so one
		// decrypt and one delete serve either format.
		cryptoGroup.POST("files/encrypt", fileAPI.EncryptFile)
		cryptoGroup.POST("files/encrypt/unpadded", fileAPI.EncryptFileUnpadded)
		cryptoGroup.GET("files/:id/decrypt", fileAPI.DecryptFile)
		cryptoGroup.DELETE("files/:id", fileAPI.DeleteFile)
	}

	return r, nil
}
