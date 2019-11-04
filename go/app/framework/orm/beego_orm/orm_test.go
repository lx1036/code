package beego_orm

import (
    "fmt"
    "testing"
    "time"

    //"time"

    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql" // import your used driver
)

type Team struct {
    Id uint32 `orm:"column(id);size(8)"` // mediumint(8) unsigned // bigint unsigned
    OrganizationId uint16 `orm:"column(organization_id);size(5)"` // smallint unsigned
    Name string `orm:"column(name);size(255)"` // varchar(255)
    Description string `orm:"description;type(text);default('');null"` // text
    CreatedAt time.Time `orm:"created_at;type(datetime);default(CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)"`
    UpdatedAt time.Time `orm:"updated_at;type(datetime);default(CURRENT_TIMESTAMP)"`
}

func init() {
    address := "testing:testing@tcp(127.0.0.1:3306)"
    database := "orm"
    charset := "utf8mb4"
    _ = orm.RegisterDataBase("default", "mysql",
        address + "/" + database + "?charset=" + charset, 30)
    orm.RegisterModel(new(Team))
    _ = orm.RunSyncdb("default", true, true)
}

func TestOrm(test *testing.T) {
    manager := orm.NewOrm()
    user := Team{
        Id:             1,
        OrganizationId: 0,
        Name:           "test",
        Description:    "description",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    id, err := manager.Insert(&user)
    fmt.Printf("ID: %d, ERR: %v\n", id, err)
    teamRead := Team{
        Id: user.Id,
    }
    err = manager.Read(&teamRead)
    fmt.Printf("ERR: %v, team %v\n", err, teamRead)
}

func TestQueryBuilder(test *testing.T) {
    queryBuilder, _ := orm.NewQueryBuilder("mysql")
    queryBuilder.Select("name").From("team").Limit(10).Offset(0)
    sql := queryBuilder.String()
    var names []string
    _, _ = (orm.NewOrm()).Raw(sql).QueryRows(&names)
    fmt.Println(names)
}
