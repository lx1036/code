package handler

import (
	"github.com/emicklei/go-restful"
	authApi "k8s-lx1036/dashboard/backend/auth/api"
	clientapi "k8s-lx1036/dashboard/backend/client/api"
	"k8s-lx1036/dashboard/backend/integration"
	"net/http"
)

// CreateHTTPAPIHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHTTPAPIHandler(iManager integration.IntegrationManager, cManager clientapi.ClientManager,
	authManager authApi.AuthManager, sManager settingsApi.SettingsManager,
	sbManager systembanner.SystemBannerManager) (http.Handler, error) {

}


func (apiHandler *APIHandler) handleDeploy(request *restful.Request, response *restful.Response) {
	k8sClient, err := apiHandler.cManager.Client(request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	appDeploymentSpec := new(deployment.AppDeploymentSpec)
	if err := request.ReadEntity(appDeploymentSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := deployment.DeployApp(appDeploymentSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, appDeploymentSpec)
}
