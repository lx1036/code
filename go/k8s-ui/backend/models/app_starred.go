package models

type AppStarred struct {
	Id   int64 `gorm:"auto" json:"id,omitempty"`
	App  *App  `gorm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	User *User `gorm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
}

type appStarredModel struct{}
