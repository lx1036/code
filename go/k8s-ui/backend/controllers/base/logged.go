package base

import "k8s-lx1036/k8s-ui/backend/models"

type LoggedInController struct {
	ParamBuilderController

	User *models.User
}
