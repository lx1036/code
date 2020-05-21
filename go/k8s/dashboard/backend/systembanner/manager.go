package systembanner


// SystemBannerManager is a structure containing all system banner manager members.
type SystemBannerManager struct {
	systemBanner api.SystemBanner
}

// NewSystemBannerManager creates new settings manager.
func NewSystemBannerManager(message, severity string) SystemBannerManager {
	return SystemBannerManager{
		systemBanner: api.SystemBanner{
			Message:  message,
			Severity: api.GetSeverity(severity),
		},
	}
}
