package config

// EmailConfig - for external email services
type EmailConfig struct {
	Activate   string
	Provider   string
	APIToken   string
	AddrFrom   string
	TrackOpens bool
	TrackLinks string
	MsgType    string

	// for templated email
	EmailVerificationTemplateID int64
	PasswordResetTemplateID     int64
	EmailVerificationCodeLength uint64
	EmailVerificationTag        string
	PasswordResetTag            string
	HTMLModel                   string
	EmailValidityPeriod         uint64 // in seconds
}
