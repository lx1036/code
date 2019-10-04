package base

type APIController struct {
	LoggedInController

	NamespaceId int64
	AppId       int64
}
