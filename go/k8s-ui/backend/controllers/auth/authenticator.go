package auth

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
	rsakey "k8s-lx1036/k8s-ui/backend/apikey"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"time"
)

var registry = make(map[string]Authenticator)

// Authenticator provides interface to authenticate user credentials.
type Authenticator interface {
	// Authenticate ...
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
	auth.Mapping("Me", auth.CurrentUser)
}

// @router /login/:type/?:name [get,post]
func (auth *AuthController) Login() {
	username := auth.Input().Get("username")
	password := auth.Input().Get("password")
	authType := auth.Ctx.Input.Param(":type")

	fmt.Println(username, password)
	//authName := auth.Ctx.Input.Param(":name")
	//next := auth.Ctx.Input.Query("next")
	if authType == "" || username == "admin" {
		authType = models.AuthTypeDB
	}
	authenticator, ok := registry[authType]
	if !ok {

	}
	authModel := models.AuthModel{
		Username: username,
		Password: password,
	}

	user, err := authenticator.Authenticate(authModel)
	if err != nil {

	}

	now := time.Now()
	exp := beego.AppConfig.DefaultInt64("TokenLifeTime", 86400)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "k8s-ui",
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(exp) * time.Second).Unix(),
		"aud": user.Name,
	})
	signedToken, err := token.SignedString(rsakey.RsaPrivateKey)
	if err != nil {

	}

	loginResult := LoginResult{Token: signedToken}
	auth.Data["json"] = base.Result{Data: loginResult}
	auth.ServeJSON()
}

// @router /logout [get]
func (auth *AuthController) Logout() {
	fmt.Println("Logout")

	auth.Data["json"] = base.Result{Data: "testtest"}
	auth.ServeJSON()
}

// @router /me [get]
func (auth *AuthController) CurrentUser() {
	fmt.Println("test")
}

func Register(name string, authenticator Authenticator) {
	if _, ok := registry[name]; ok {
		log.Infof("authenticator [%s] has been registered", name)
		return
	}

	registry[name] = authenticator
}
