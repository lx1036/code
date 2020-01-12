package models

type AppStarred struct {
	Id   int64 `orm:"auto" json:"id,omitempty"`
	App  *App  `orm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	User *User `orm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
}

type appStarredModel struct{}
