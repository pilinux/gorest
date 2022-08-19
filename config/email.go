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
}
