package TiebaSign

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
)

var SqliteDb *gorm.DB

func init() {
	var err error
	//Db, err = gorm.Open("postgres", "port=5432 host=ec2-54-235-193-0.compute-1.amazonaws.com user=uorhkrqhdtctcq password=356bb77282e543f42fb06e1859a9e0ad639feb1d83a4aa0abfc4c589b3e80b2e dbname=dcqt89gh79quvf")
	SqliteDb, err = gorm.Open("sqlite3", "/mnt/data/sign-node.db")
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		fmt.Println("Sqlite数据库连接成功")
	}
	SqliteDb.SingularTable(true)
	data, err := ioutil.ReadFile("sqlite_init.sql")
	if err != nil {
		fmt.Println("read file err:", err.Error())
		return
	}
	//打印sql语句
	//SqliteDb.LogMode(true)
	//初始化sqlite数据库
	SqliteDb.Exec(string(data))
}

func SignDetailInfo(uid string, fName string, currPage, pageSize, status int) map[string]interface{} {
	result := make(map[string]interface{})
	start := GetPageStart(currPage, pageSize)
	list := []TieBaModel{}
	var totalCount int
	con := " "
	if status == 1 {
		con = " and error_code!='0' and  error_code!='160002' and error_code is not null"
	}
	if len(fName) > 0 {
		SqliteDb.Model(&ChanSignResult{}).Where("uid = ? and fname like ?"+con, uid, "%"+fName+"%").Count(&totalCount)
		SqliteDb.Raw("select *, (CAST(cur_score as FLOAT)/CAST(levelup_score as FLOAT)) AS level from tieba where uid=? and fname like ?"+con+" limit ?,?", uid, "%"+fName+"%", start, pageSize).Find(&list)
	} else {
		SqliteDb.Model(&ChanSignResult{}).Where("uid = ?"+con, uid).Count(&totalCount)
		SqliteDb.Raw("select *, (CAST(cur_score as FLOAT)/CAST(levelup_score as FLOAT)) AS level from tieba where uid=? "+con+" limit ?,?", uid, start, pageSize).Find(&list)
	}
	result["list"] = list
	result["totalCount"] = totalCount
	result["currPage"] = currPage
	result["pageSize"] = pageSize
	result["pages"] = GetTotalPage(totalCount, pageSize)
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
func GetTotalPage(totalCount, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	if totalCount%pageSize == 0 {
		return totalCount / pageSize
	} else {
		return totalCount/pageSize + 1
	}
}
