package routers_gin

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"k8s-lx1036/k8s-ui/backend/apikey"
	"k8s-lx1036/k8s-ui/backend/controllers"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/models/response/errors"
	"k8s-lx1036/k8s-ui/backend/util/logs"
	"net/http"
	"strings"
)

var (
	User models.User
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// application's global HTTP middleware stack
	router.Use(cors.AllowAll()) // cors

	router.POST("/login/:type", (&auth.AuthController{}).Login())
	router.GET("/logout", (&auth.AuthController{}).Logout())

	authorizedRouter := router.Group("/")
	authorizedRouter.Use(AuthRequired())
	{
		authorizedRouter.GET(`/me`, (&auth.AuthController{}).CurrentUser())
		apiV1Router := authorizedRouter.Group("/api/v1")
		apiV1Router.GET("/configs/base", (&controllers.BaseConfigController{}).ListBase())
	}

	return router
}

func AuthRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		authorizations := strings.Split(authorization, " ")
		if len(authorizations) != 2 || authorizations[0] != "Bearer" {
			logs.Info("AuthString invalid:", authorization)
			context.AbortWithStatusJSON(http.StatusUnauthorized, base.JsonResponse{
				Errno:  -1,
				Errmsg: "failed: Token Invalid!",
				Data:   nil,
			})
		}

		tokenString := authorizations[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return apikey.RsaPublicKey, nil
		})
		errResult := errors.ErrorResult{}
		switch err.(type) {
		case nil: // no error
			if !token.Valid { // but may still be invalid
				errResult.Code = http.StatusUnauthorized
				errResult.Msg = "token is invalid"
			}
		case *jwt.ValidationError:
			errResult.Code = http.StatusUnauthorized
			errResult.Msg = err.Error()
		default:
			errResult.Code = http.StatusInternalServerError
			errResult.Msg = err.Error()
		}
		if err != nil {
			context.AbortWithStatusJSON(errResult.Code, base.JsonResponse{
				Errno:  errResult.Code,
				Errmsg: fmt.Sprintf("failed: %s", errResult.Msg),
				Data:   nil,
			})
		}

		claim := token.Claims.(jwt.MapClaims)
		aud := claim["aud"].(string)
		user, err := models.GetUserByName(aud)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusInternalServerError, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("failed: %s", err.Error()),
				Data:   nil,
			})
		}

		User = user

		context.Next()
	}
}
