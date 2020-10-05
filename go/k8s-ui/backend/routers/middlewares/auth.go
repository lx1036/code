package middlewares

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s-ui/backend/common/rsa"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/models/response/errors"
	"net/http"
	"strings"
)

var (
	User models.User
)

func AuthRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		authorizations := strings.Split(authorization, " ")
		if len(authorizations) != 2 || authorizations[0] != "Bearer" {
			log.Info("AuthString invalid:", authorization)
			context.AbortWithStatusJSON(http.StatusUnauthorized, base.JsonResponse{
				Errno:  -1,
				Errmsg: "failed: need authorization token!",
				Data:   nil,
			})
		}

		tokenString := authorizations[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return rsa.RsaPublicKey, nil
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
