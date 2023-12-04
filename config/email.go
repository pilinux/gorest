package config

// EmailConfig - for external email services
type EmailConfig struct {
	Activate     string
	Provider     string
	APIToken     string
	AddrFrom     string
	TrackOpens   bool
	TrackLinks   string
	DeliveryType string

	// for templated email
	EmailVerificationTemplateID int64
	PasswordRecoverTemplateID   int64
	EmailUpdateVerifyTemplateID int64
	EmailVerificationCodeUUIDv4 bool
	EmailVerificationCodeLength uint64
	PasswordRecoverCodeUUIDv4   bool
	PasswordRecoverCodeLength   uint64
	EmailVerificationTag        string
	PasswordRecoverTag          string
	HTMLModel                   string
	EmailVerifyValidityPeriod   uint64 // in seconds
	PassRecoverValidityPeriod   uint64 // in seconds
}
