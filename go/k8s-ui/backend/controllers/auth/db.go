package auth

import (
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/models"
)

type DBAuth struct{}

func (auth DBAuth) Authenticate(model models.AuthModel) (*models.User, error) {
	var user models.User
	err := lorm.DB.Where("name=?", model.Username).First(&user).Error

	if err != nil {
		return nil, err
	}
	//user, err := models.UserModel.GetUserByName(model.Username)

	return &user, nil
}

func init() {
	Register(models.AuthTypeDB, &DBAuth{})
}
