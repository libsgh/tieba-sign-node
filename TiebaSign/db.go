package TiebaSign

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	gorm_logrus "github.com/onrik/gorm-logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"os"
)

var Db *gorm.DB

func init() {
	var err error
	dialector := postgres.Open("port=5432 host=localhost user=xx password=xxxa dbname=xxx")
	if os.Getenv("DATABASE_URL") != "" {
		db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
		dialector = postgres.New(postgres.Config{
			Conn: db,
		})
	}
	Db, err = gorm.Open(dialector, &gorm.Config{
		Logger: gorm_logrus.New(),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	//SqliteDb, err = gorm.Open("sqlite3", "/mnt/data/sign-node.db")
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		fmt.Println("Sqlite数据库连接成功")
	}
	data, err := ioutil.ReadFile("postgres_init.sql")
	if err != nil {
		fmt.Println("read file err:", err.Error())
		return
	}
	//打印sql语句
	//SqliteDb.LogMode(true)
	//初始化sqlite数据库
	Db.Exec(string(data))
}

func SignDetailInfo(uid string, fName string, currPage, pageSize, status int) map[string]interface{} {
	result := make(map[string]interface{})
	start := GetPageStart(currPage, pageSize)
	list := []TieBaModel{}
	var totalCount int64
	con := " "
	if status == 1 {
		con = " and error_code!='0' and  error_code!='160002' and error_code is not null"
	}
	if len(fName) > 0 {
		Db.Model(&ChanSignResult{}).Where("uid = ? and fname like ?"+con, uid, "%"+fName+"%").Count(&totalCount)
		Db.Raw("select *, round(CAST(cur_score as numeric)/CAST(levelup_score as numeric),2) AS level from tieba where uid=? and fname like ?"+con+" limit ?,?", uid, "%"+fName+"%", start, pageSize).Find(&list)
	} else {
		Db.Model(&ChanSignResult{}).Where("uid = ?"+con, uid).Count(&totalCount)
		Db.Raw("select *, round(CAST(cur_score as numeric)/CAST(levelup_score as numeric),2) AS level from tieba where uid=? "+con+" limit ?,?", uid, start, pageSize).Find(&list)
	}
	result["list"] = list
	result["totalCount"] = totalCount
	result["currPage"] = currPage
	result["pageSize"] = pageSize
	result["pages"] = GetTotalPage(int(totalCount), pageSize)
	return result
}

//获取查询游标start
func GetPageStart(pageNo, pageSize int) int {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 0
	}
	return (pageNo - 1) * pageSize
}

//获取总页数
func GetTotalPage(totalCount int, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	if totalCount%pageSize == 0 {
		return totalCount / pageSize
	} else {
		return totalCount/pageSize + 1
	}
}
