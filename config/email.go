package config

// EmailConfig - for external email services
type EmailConfig struct {
	Activate   string
	Provider   string
	APIToken   string
	TemplateID int64
	AddrFrom   string
	Tag        string
	TrackOpens bool
	TrackLinks string
	MsgType    string
}
