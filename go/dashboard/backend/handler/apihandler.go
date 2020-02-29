package handler

import (
	authApi "k8s-lx1036/dashboard/backend/auth/api"
	clientapi "k8s-lx1036/dashboard/backend/client/api"
	"net/http"
)

// CreateHTTPAPIHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHTTPAPIHandler(iManager integration.IntegrationManager, cManager clientapi.ClientManager,
	authManager authApi.AuthManager, sManager settingsApi.SettingsManager,
	sbManager systembanner.SystemBannerManager) (http.Handler, error) {

}
