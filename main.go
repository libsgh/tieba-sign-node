package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"tieba-sign-node/TiebaSign"
)

var version string = "1.3.4"

func main() {
	//gin.SetMode(gin.ReleaseMode)
	port := os.Getenv("PORT")
	if port == "" {
		//log.Fatal("必须设置 $PORT")
		port = "8080"
	}
	//首先先生成一个gin实例
	r := gin.New()
	r.Use(gin.Logger())
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "static")
	r.StaticFile("/favicon.ico", "./static/img/favicon.ico")
	//声明一个路由
	r.GET("/", info)
	r.GET("/info", info)
	r.POST("/tbs", tbs)
	r.GET("/tbs", tbs)
	TiebaSign.InitJobs() //初始化定时任务
	r.Run(":" + port)    // 监听并在 0.0.0.0:8080 上启动服务

}

/**
首页
*/
func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{"name": "Libs"})
}

/**
返回程序信息
*/
func info(c *gin.Context) {
	serverName := os.Getenv("NODE_NAME")
	//currentTime := time.Now()
	/*c.JSON(http.StatusOK, gin.H{
		"tb_count":         36630,
		"blcaklist":        41,
		"node_name":        "签到节点-01",
		"cookie_valid":     1,
		"re_sign":          0,
		"wait_sign":        0,
		"version":          "v1.2.9",
		"request_time":     currentTime.Format("2006-01-02 15:04:05"),
		"main_program":     "http://noki.tk/tieba",
		"signed":           36098,
		"signed_user":      79,
		"signed_execption": 491,
	})*/
	c.String(http.StatusOK, TiebaSign.Get("https://noki.top/signnode/info?servername="+serverName+"&version="+version))
}

//查询签到从详情
func tbs(c *gin.Context) {
	uid := c.Request.FormValue("uid")
	fName := c.Request.FormValue("fname")
	currPage, _ := strconv.Atoi(c.Request.FormValue("currPage"))
	pageSize, _ := strconv.Atoi(c.Request.FormValue("pageSize"))
	status, _ := strconv.Atoi(c.Request.FormValue("status"))
	c.JSON(http.StatusOK, TiebaSign.SignDetailInfo(uid, fName, currPage, pageSize, status))
}
