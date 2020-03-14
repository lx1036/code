package authenticator

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthenticationController struct {

}




func (controller *AuthenticationController) Login() gin.HandlerFunc {
	return func(context *gin.Context) {
		var spec LoginSpec
		if err := context.ShouldBindJSON(&spec); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"errno":  -1,
				"errmsg": err.Error(),
				"data":   nil,
			})
			return
		}

		authenticator := getAuthenticator(spec)
		authInfo := authenticator.GetAuthInfo()

		err = CheckPolicy(authInfo)

		token := tokenManager.Generate(authInfo)

		context.JSON(http.StatusOK, gin.H{
			"errno": 0,
			"errmsg": nil,
			"data": AuthResponse{JweToken: token},
		})
	}
}


func (controller *AuthenticationController) GetLoginModes() gin.HandlerFunc {
	return func(context *gin.Context) {
		modes := AuthenticationModes()
		context.JSON(http.StatusOK, gin.H{
			"errno": 0,
			"errmsg": "success",
			"data": LoginModesResponse{Modes:modes},
		})
	}
}


func getAuthenticator(spec LoginSpec) Authenticator {
	switch {
	case len(spec.Username) >0 && len(spec.Password) > 0 && authenticationModes.IsEnabled(Basic):
		return NewBasicAuthenticator(spec)
	case len(spec.Token) > 0 && authenticationModes.IsEnabled(Token):
		return NewTokenAuthenticator(spec)
	case len(spec.KubeConfig) > 0:
		return NewKubeConfigAuthenticator(spec)
	}
}

