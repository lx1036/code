package models

type AppStarred struct {
	Id uint `gorm:"column:id;primary_key;"`
	//App  *App  `gorm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	//User *User `gorm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
}

type appStarredModel struct{}
