package cfg

import (
	"fmt"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"strings"
)

func GetSqldb (connstr string, dbType string) (db *gorm.DB) {
	log.Println("connect:",connstr)
	var err error
	if dbType=="mysql"{
		//db, err = gorm.Open(mysql.Open( connstr))
	}else if strings.HasPrefix( dbType,"sqlite"){
		fname:="./dbdata/data_sql_main.db"
		items:=strings.Split(dbType,":")
		if len(items)>1{
			fname=fmt.Sprintf("./dbdata/data_sql_%s.db",items[1])
		}
		db, err = gorm.Open(sqlite.Open(fname), &gorm.Config{})
	}else{
		log.Fatalln("err dbtype:",dbType)
	}
	if err!=nil{
		fmt.Println(err);
		panic(err)
	}
	fmt.Println("InitGormBD,dbType",dbType,connstr);
	return db
}
var Orm *gorm.DB
func init(){
	Orm=GetSqldb("","sqlite")
	Orm.AutoMigrate(tkrpc.NodeInfo{},tkrpc.UserInfo{},tkrpc.LockHtlcPath{})
}
