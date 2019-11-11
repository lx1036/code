package db

import (
	"k8s-lx1036/wayne/backend/controllers/auth"
	"k8s-lx1036/wayne/backend/models"
)

type DBAuth struct{}

func (auth DBAuth) Authenticate(model models.AuthModel) (*models.User, error) {
	user, err := models.UserModel.GetUserByName(model.Username)

	return user, err
}

func init() {
	auth.Register(models.AuthTypeDB, &DBAuth{})
}


