package auth

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	rsakey "k8s-lx1036/k8s-ui/backend/apikey"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/models/response/errors"
	"k8s-lx1036/k8s-ui/backend/util/logs"
	"net/http"
	"strings"
	"time"
)

var registry = make(map[string]Authenticator)

// Authenticator provides interface to authenticate user credentials.
type Authenticator interface {
	Authenticate(model models.AuthModel) (*models.User, error)
}

type LoginResult struct {
	Token string `json:"token"`
}

type AuthController struct {
}

func (controller *AuthController) Init() {

}

func (controller *AuthController) Login() gin.HandlerFunc {
	return func(context *gin.Context) {
		username := context.PostForm("username")
		password := context.PostForm("password")
		authType := context.Query("type")
		//authName := context.Query("name")
		if authType == "" || username == "admin" {
			authType = models.AuthTypeDB
		}
		authenticator, ok := registry[authType]
		if !ok {
			context.JSON(http.StatusBadRequest, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("failed: auth type[%s] is not supported", authType),
				Data:   nil,
			})
			return
		}

		authModel := models.AuthModel{
			Username: username,
			Password: password,
		}
		if authType == models.AuthTypeOAuth2 { // login with oauth2

		}
		user, err := authenticator.Authenticate(authModel)
		if err != nil {
			context.JSON(http.StatusBadRequest, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("try to login in with user [%s] error: %v", authModel.Username, err),
				Data:   nil,
			})
			return
		}

		now := time.Now()
		exp := beego.AppConfig.DefaultInt64("TokenLifeTime", 86400)
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": beego.AppConfig.DefaultString("appname", "k8s-ui"), // 签发者
			"iat": now.Unix(),                                         // 签发时间
			"exp": now.Add(time.Duration(exp) * time.Second).Unix(),   // 过期时间
			"aud": user.Name,
		})
		signedToken, err := token.SignedString(rsakey.RsaPrivateKey)
		if err != nil {
			context.JSON(http.StatusInternalServerError, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("try to create token, error: %v", err),
				Data:   nil,
			})
			return
		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   LoginResult{Token: signedToken},
		})
	}
}

func (controller *AuthController) Logout() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func (controller *AuthController) CurrentUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		authorizations := strings.Split(authorization, " ")
		if len(authorizations) != 2 || authorizations[0] != "Bearer" {
			logs.Info("AuthString invalid:", authorization)
			context.JSON(http.StatusUnauthorized, base.JsonResponse{
				Errno:  -1,
				Errmsg: "failed: Token Invalid!",
				Data:   nil,
			})
			return
		}

		tokenString := authorizations[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return rsakey.RsaPublicKey, nil
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
			context.JSON(errResult.Code, base.JsonResponse{
				Errno:  errResult.Code,
				Errmsg: fmt.Sprintf("failed: %s", errResult.Msg),
				Data:   nil,
			})
			return
		}

		claim := token.Claims.(jwt.MapClaims)
		aud := claim["aud"].(string)
		user, err := models.UserModel.GetUserDetail(aud)
		if err != nil {
			context.JSON(http.StatusInternalServerError, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("failed: %s", err.Error()),
				Data:   nil,
			})
			return
		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   user,
		})
	}
}

func Register(name string, authenticator Authenticator) {
	if _, ok := registry[name]; ok {
		log.Infof("authenticator [%s] has been registered", name)
		return
	}

	registry[name] = authenticator
}
