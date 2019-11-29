package base

import "k8s-lx1036/wayne/backend/models"

type LoggedInController struct {
	ParamBuilderController

	User *models.User
}
