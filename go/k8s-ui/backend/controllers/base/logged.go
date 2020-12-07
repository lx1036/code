package base

import (
	"github.com/dgrijalva/jwt-go"
	"k8s-lx1036/k8s-ui/backend/models"
	"net/http"
	"strings"
)

type LoggedInController struct {
	ParamBuilderController

	User *models.User
}

func (controller *LoggedInController) publishRequestMessage(code int, data interface{}) {

}

func (controller *LoggedInController) Prepare() {
	authString := controller.Ctx.Input.Header("Authorization")
	kv := strings.Split(authString, " ")
	if len(kv) != 2 || kv[0] != "Bearer" {
		logs.Info("AuthString invalid:", authString)
		controller.CustomAbort(http.StatusUnauthorized, "Token invalid.")
	}

	tokenString := kv[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, e error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return rsakey.RsaPublicKey, nil
	})

	switch err.(type) {

	}

	if err != nil {

	}

	claim := token.Claims.(jwt.MapClaims)
	aud := claim["aud"].(string) // aud is `users`.`name`
	controller.User, err = models.UserModel.GetUserDetail(aud)
	if err != nil {
		controller.CustomAbort(http.StatusInternalServerError, err.Error())
	}
}
