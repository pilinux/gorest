package config

// IsProd returns true when app is running in production mode
func IsProd() bool {
	return GetConfig().Server.ServerEnv == "production"
}

// IsSentry returns true when sentry logger is enabled in .env
func IsSentry() bool {
	return GetConfig().Logger.Activate == Activated
}

// IsBasicAuth returns true when basic auth is enabled in .env
func IsBasicAuth() bool {
	return GetConfig().Security.MustBasicAuth == Activated
}

// IsJWT returns true when JWT is enabled in .env
func IsJWT() bool {
	return GetConfig().Security.MustJWT == Activated
}

// InvalidateJWT returns true when this feature is enabled in .env
func InvalidateJWT() bool {
	return GetConfig().Security.InvalidateJWT == Activated
}

// IsAuthCookie returns true when auth cookie is enabled in .env
func IsAuthCookie() bool {
	return GetConfig().Security.AuthCookieActivate
}

// IsHashPass returns true when password hashing is enabled in .env
func IsHashPass() bool {
	return GetConfig().Security.MustHash == Activated
}

// IsCipher returns true when encryption at rest is enabled in .env
func IsCipher() bool {
	return GetConfig().Security.MustCipher
}

// Is2FA returns true when two-factor authentication is enabled in .env
func Is2FA() bool {
	return GetConfig().Security.Must2FA == Activated
}

// Is2FADoubleHash returns true when double hashing is enabled in .env
func Is2FADoubleHash() bool {
	return GetConfig().Security.TwoFA.DoubleHash
}

// IsWAF returns true when app firewall is enabled in .env
func IsWAF() bool {
	return GetConfig().Security.MustFW == Activated
}

// IsCORS returns true when CORS is enabled in .env
func IsCORS() bool {
	return GetConfig().Security.MustCORS == Activated
}

// IsOriginCheck returns true when origin check is enabled in .env
func IsOriginCheck() bool {
	return GetConfig().Security.CheckOrigin == Activated
}

// IsRateLimit returns true when rate limiting is enabled in .env
func IsRateLimit() bool {
	return GetConfig().Security.RateLimit != ""
}

// IsTemplatingEngine returns true when serving HTML is enabled in .env
func IsTemplatingEngine() bool {
	return GetConfig().ViewConfig.Activate == Activated
}

// IsRDBMS returns true when RDBMS is enabled in .env
func IsRDBMS() bool {
	return GetConfig().Database.RDBMS.Activate == Activated
}

// IsRedis returns true when Redis is enabled in .env
func IsRedis() bool {
	return GetConfig().Database.REDIS.Activate == Activated
}

// IsMongo returns true when Mongo is enabled in .env
func IsMongo() bool {
	return GetConfig().Database.MongoDB.Activate == Activated
}

// IsEmailService returns true when email service is enabled in .env
func IsEmailService() bool {
	return GetConfig().EmailConf.Activate == Activated
}

// IsEmailVerificationService returns true when it is enabled in .env
func IsEmailVerificationService() bool {
	return GetConfig().Security.VerifyEmail
}

// IsPassRecoveryService returns true when it is enabled in .env
func IsPassRecoveryService() bool {
	return GetConfig().Security.RecoverPass
}

// IsEmailVerificationCodeUUIDv4 returns true when it is enabled in .env
func IsEmailVerificationCodeUUIDv4() bool {
	return GetConfig().EmailConf.EmailVerificationCodeUUIDv4
}

// IsPasswordRecoverCodeUUIDv4 returns true when it is enabled in .env
func IsPasswordRecoverCodeUUIDv4() bool {
	return GetConfig().EmailConf.PasswordRecoverCodeUUIDv4
}
