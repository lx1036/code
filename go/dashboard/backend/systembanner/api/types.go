package api

// SystemBannerSeverity represents severity of system banner.
type SystemBannerSeverity string

// SystemBanner represents system banner.
type SystemBanner struct {
	Message  string               `json:"message"`
	Severity SystemBannerSeverity `json:"severity"`
}
