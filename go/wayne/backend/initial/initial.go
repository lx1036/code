package initial

import (
    "database/sql"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/logs"
    "github.com/astaxie/beego/orm"
    "k8s-lx1036/wayne/backend/bus"
    "k8s-lx1036/wayne/backend/client"
    "k8s.io/apimachinery/pkg/util/wait"
    "strings"
    "time"
    "github.com/go-sql-driver/mysql"
    _ "github.com/go-sql-driver/mysql"
)

const DbDriverName = "mysql"


// MySQL
func InitDb() {
    _ = orm.RegisterDriver(DbDriverName, orm.DRMySQL)

    // ensure database exist
    err := ensureDatabase()
    if err != nil {
        panic(err)
    }
    db, err := orm.GetDB()
    if err != nil {
        panic(err)
    }

    ttl := beego.AppConfig.DefaultInt("DBConnTTL", 30)
    db.SetConnMaxLifetime(time.Duration(ttl) * time.Second)
    orm.Debug = beego.AppConfig.DefaultBool("ShowSql", false)
}

func InitClient() {
	// 定期更新client
	go wait.Forever(client.BuildApiserverClient, 5*time.Second)
}

// bus
func InitBus() {
    var err error
    bus.DefaultBus, err = bus.NewBus(beego.AppConfig.String("BusRabbitMQURL"))
    if err != nil {
        panic(err)
    }
}

func InitRsaKey() {

}

func InitKubeLabel() {

}

func ensureDatabase() error  {
    needInit := false
    dbName := beego.AppConfig.String("DBName")
    dbURL := fmt.Sprintf("%s:%s@%s/", beego.AppConfig.String("DBUser"),
        beego.AppConfig.String("DBPasswd"), beego.AppConfig.String("DBTns"))
    db, err := sql.Open(DbDriverName, fmt.Sprintf("%s%s", dbURL, dbName))
    if err != nil {
        return err
    }
    defer db.Close()
    err = db.Ping()
    if err != nil {
        switch e := err.(type) {
        case *mysql.MySQLError:
            // MySQL error unkonw database;
            // refer https://dev.mysql.com/doc/refman/5.6/en/error-messages-server.html
            if e.Number == 1049 {
                needInit = true
                dbForCreateDatabase, err := sql.Open(DbDriverName, addLocation(dbURL))
                if err != nil {
                    return err
                }
                defer dbForCreateDatabase.Close()
                _, err = dbForCreateDatabase.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci;", dbName))
                if err != nil {
                    return err
                }

            } else {
                return err
            }
        default:
            return err
        }
    }

    logs.Debug("Initialize database connection: %s", strings.Replace(dbURL, beego.AppConfig.String("DBPasswd"), "****", 1))
    err = orm.RegisterDataBase("default", "mysql", addLocation(fmt.Sprintf("%s%s", dbURL, dbName)))
    if err != nil {
        return err
    }
    if needInit {
        err = orm.RunSyncdb("default", false, true)
        if err != nil {
            return err
        }
        for _, insertSql := range initial.InitialData {
            _, err = orm.NewOrm().Raw(insertSql).Exec()
            if err != nil {
                return err
            }
        }

    }
    return nil

}
