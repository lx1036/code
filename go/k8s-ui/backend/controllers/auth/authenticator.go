package auth

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
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
	beego.Controller
}

func (auth *AuthController) URLMapping() {
	auth.Mapping("Login", auth.Login)
	auth.Mapping("Logout", auth.Logout)
	auth.Mapping("CurrentUser", auth.CurrentUser)
}

// @router /login/:type/?:name [get,post]
func (controller *AuthController) Login() {
	username := controller.Input().Get("username")
	password := controller.Input().Get("password")
	authType := controller.Ctx.Input.Param(":type")
	//oauthName := controller.Ctx.Input.Param(":name")

	logrus.WithFields(logrus.Fields{
		"username": username,
		"password": password,
	}).Info("login")

	//authName := auth.Ctx.Input.Param(":name")
	//next := auth.Ctx.Input.Query("next")
	if authType == "" || username == "admin" {
		authType = models.AuthTypeDB
	}

	logs.Info("auth type is ", authType)

	authenticator, ok := registry[authType]
	if !ok {
		controller.Ctx.Output.SetStatus(http.StatusBadRequest)
		controller.Data["json"] = base.Result{Data: fmt.Sprintf("auth type[%s] is not supported", authType)}
		controller.ServeJSON()
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
		controller.Ctx.Output.SetStatus(http.StatusBadRequest)
		controller.Data["json"] = base.Result{Data: fmt.Sprintf("try to login in with user [%s] error: %v", authModel.Username, err)}
		controller.ServeJSON()
		return
	}

	now := time.Now()
	exp := beego.AppConfig.DefaultInt64("TokenLifeTime", 86400)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "k8s-ui", // 签发者
		"iat": now.Unix(), // 签发时间
		"exp": now.Add(time.Duration(exp) * time.Second).Unix(), // 过期时间
		"aud": user.Name,
	})
	signedToken, err := token.SignedString(rsakey.RsaPrivateKey)
	if err != nil {
		controller.Ctx.Output.SetStatus(http.StatusInternalServerError)
		controller.Data["json"] = base.Result{Data: fmt.Sprintf("try to create token, error: %v", err)}
		controller.ServeJSON()
		return
	}

	loginResult := LoginResult{Token: signedToken}
	controller.Data["json"] = base.Result{Data: loginResult}
	controller.ServeJSON()
}

// @router /logout [get]
func (controller *AuthController) Logout() {
	fmt.Println("Logout")

	controller.Data["json"] = base.Result{Data: "testtest"}
	controller.ServeJSON()
}

// @router /currentuser [get]
func (controller *AuthController) CurrentUser() {
	controller.Controller.Prepare()
	authString := controller.Ctx.Input.Header("Authorization")
	kv := strings.Split(authString, " ")
	if len(kv) != 2 || kv[0] != "Bearer" {
		logs.Info("AuthString invalid:", authString)
		controller.CustomAbort(http.StatusUnauthorized, "Token Invalid ! ")
	}

	tokenString := kv[1]
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
		controller.CustomAbort(errResult.Code, errResult.Msg)
	}

	claim := token.Claims.(jwt.MapClaims)
	aud := claim["aud"].(string)
	user, err := models.UserModel.GetUserDetail(aud)
	if err != nil {
		controller.CustomAbort(http.StatusInternalServerError, err.Error())
	}

	controller.Data["json"] = base.Result{Data: user}
	controller.ServeJSON()
}

func Register(name string, authenticator Authenticator) {
	if _, ok := registry[name]; ok {
		log.Infof("authenticator [%s] has been registered", name)
		return
	}

	registry[name] = authenticator
}
