package TiebaSign

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var signFlag bool = false

var bqFlag bool = false

var prisionFlag bool = false

func InitJobs() {
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		log.Fatal("必须设置 $SERVER_NAME")
	}
	c := cron.New()
	c.AddFunc("0 0/1 * * * ?", func() {
		if signFlag == false {
			signFlag = true
			log.Println("查询待签到贴吧信息，并加入队列...")
			body := Get("https://toolsbox.herokuapp.com/sign/query?servername=" + serverName)
			val := []byte(body)
			var tiebas []Tieba
			if err := jsoniter.Unmarshal(val, &tiebas); err != nil {
				log.Println("error: ", err)
			}
			Parallelize(5, len(tiebas), func(piece int) {
				//1. 同步签到状态
				Get("https://toolsbox.herokuapp.com/sign/syncStatus?uid=" + tiebas[piece].Uid)
				isLogin := CheckBdussValid(tiebas[piece].Bduss)
				if !isLogin {
					postData := make(map[string]interface{})
					postData["cookie_valid"] = 0
					postData["sign_status"] = 2
					Post("https://toolsbox.herokuapp.com/sign/update?flag=1&uid="+tiebas[piece].Uid, postData)
					log.Println(tiebas[piece].Name + "------BDUSS失效")
				}
				OneBtnToSign(tiebas[piece])
			})
			log.Println("签到任务执行完毕")
			signFlag = false
		}
	})
	c.AddFunc("0 0/1 * * * ?", func() {
		if bqFlag == false {
			bqFlag = true
			log.Println("补签任务开始执行...")
			body := Get("https://toolsbox.herokuapp.com/sign/bq/query?servername=" + serverName)
			val := []byte(body)
			var bqTieBas []BqTieBa
			if err := jsoniter.Unmarshal(val, &bqTieBas); err != nil {
				log.Println("error: ", err)
			}
			bc := 0
			for _, bq := range bqTieBas {
				tbs := GetTbs(bq.Bduss)
				signResult := SignOneTieBa(bq.Tbname, strconv.Itoa(bq.Fid), bq.Bduss, tbs)
				bc++
				SqliteDb.
					Table("tieba").
					Where("uname = ? and fid = ?", bq.Username, bq.Fid).
					Updates(map[string]interface{}{"error_code": signResult.ErrorCode,
						"ret_msg": signResult.ErrorMsg, "signTime": time.Now().UnixNano() / 1e6})
				signResultJson, _ := jsoniter.MarshalToString(signResult)
				log.Println("补签>>>>>>" + strconv.Itoa(bc) + "\t" + bq.Username + "\t" + signResultJson)
				if signResult.ErrorCode == "0" || signResult.ErrorCode == "160002" || signResult.ErrorCode == "199901" {
					//签到成功、已签到、封禁
					Post("https://toolsbox.herokuapp.com/sign/bq/update?guid="+bq.Guid,
						map[string]interface{}{"isdelete": 1})
				} else if signResult.ErrorCode == "340006" || signResult.ErrorCode == "300004" {
					//贴吧目录出问题、贴吧数据信息加载失败
					Post("https://toolsbox.herokuapp.com/sign/bq/excep?ce=1&guid="+bq.Guid, map[string]interface{}{})
				} else if signResult.ErrorCode == "340008" {
					//黑名单
					Post("https://toolsbox.herokuapp.com/sign/bq/excep?bl=1&guid="+bq.Guid, map[string]interface{}{})
				}
				time.Sleep(time.Duration(5) * time.Second)
			}
			log.Println("补签任务执行完毕")
			bqFlag = false
		}
	})
	c.AddFunc("0 0/1 * * * ?", func() {
		if prisionFlag == false {
			prisionFlag = true
			log.Println("封禁签任务开始执行...")
			body := Get("https://toolsbox.herokuapp.com/prision/taskList?servername=" + serverName)
			val := []byte(body)
			var prisions []Prision
			if err := jsoniter.Unmarshal(val, &prisions); err != nil {
				log.Println("error: ", err)
			}
			pc := 0
			for _, p := range prisions {
				pc++
				pResult := Commitprison(p.Bduss, p.Tbs, p.Uname, p.Tbname, p.Days, p.Reason, p.Portrait)
				pJsonResult := jsoniter.Get([]byte(pResult))
				if pJsonResult.Get("error_code").ToString() == "0" {
					log.Println(strconv.Itoa(pc) + "\t" + p.Uname + "\t封禁成功\t" + pResult)
					profile := GetUserProfile(p.Uid)
					headUrl := "http://tb.himg.baidu.com/sys/portrait/item/" +
						jsoniter.Get([]byte(profile), "user").Get("portrait").ToString()
					nameShow := jsoniter.Get([]byte(profile), "user").Get("name_show").ToString()
					Post("https://toolsbox.herokuapp.com/prision/update", map[string]interface{}{
						"prision_time": pJsonResult.Get("time").ToInt() * 1000,
						"head_url":     headUrl,
						"id":           p.Id,
						"name_show":    nameShow,
						"count":        p.Count + 1,
					})

				}
			}
			log.Println("封禁签任务执行完毕...")
			prisionFlag = false
		}
	})
	c.Start()
}
func OneBtnToSign(tieba Tieba) {
	likedTiebaList, err := GetLikedTiebas(tieba.Bduss, tieba.Uid)
	log.Println("【" + tieba.Name + "】关注贴吧：" + strconv.Itoa(len(likedTiebaList)))
	if err != nil {
		log.Println("err: ", err)
	}
	tbs := GetTbs(tieba.Bduss)
	chs := make(chan ChanSignResult, 5000)
	var ops uint64 = 0
	Parallelize(5, len(likedTiebaList), func(piece int) {
		baInfo := likedTiebaList[piece]
		//签到一个贴吧
		sstb := SyncSignTieBa(baInfo, tieba, tbs, chs)
		atomic.AddUint64(&ops, 1)
		opsFinal := atomic.LoadUint64(&ops)
		log.Println(strconv.Itoa(int(opsFinal)) + "/" + strconv.Itoa(len(likedTiebaList)) + "\t" + tieba.Name + "--" + sstb.Fname + "\t" + sstb.ErrorCode + "\t" + sstb.ErrorMsg)
		//名人堂助攻
		CelebritySupport(tieba.Bduss, "", baInfo.Id, tbs)
	})
	close(chs)
	//将签到结果写入到本地sqlite
	SqliteDb.Where("uid = ?", tieba.Uid).Delete(ChanSignResult{})
	totalCount := len(likedTiebaList)
	cookieValidCount := 0
	excepCount := 0
	blackCount := 0
	signCount := 0
	retryCount := 0
	var timespan int64
	for ch := range chs {
		timespan += ch.Timespan
		SqliteDb.Create(ch)
		if ch.ErrorCode == "1" {
			cookieValidCount++
		} else if ch.ErrorCode == "340006" || ch.ErrorCode == "300004" {
			//贴吧目录出问题，加载数据失败2
			excepCount++
		} else if ch.ErrorCode == "340008" {
			//黑名单
			blackCount++
		} else if ch.ErrorCode == "0" || ch.ErrorCode == "160002" || ch.ErrorCode == "199901" {
			//签到成功、已经签到、账号封禁，签到不涨经验
			signCount++
		} else if ch.ErrorCode == "2280007" || ch.ErrorCode == "340011" || ch.ErrorCode == "1989004" {
			//签到服务忙、签到过快、数据加载失败1
			//三种情况需要重签
			postData := make(map[string]interface{})
			postData["tbname"] = ch.Fname
			postData["username"] = tieba.Name
			postData["bduss"] = tieba.Bduss
			postData["fid"], err = strconv.Atoi(ch.Fid)
			if err != nil {
				fmt.Println("贴吧fid转换出错: ", err)
			}
			postData["guid"] = uuid.NewV4().String()
			postData["error_code"] = ch.ErrorCode
			postData["error_msg"] = ch.ErrorMsg
			go Post("https://toolsbox.herokuapp.com/sign/bq/insert", postData)
		}
	}
	signData := make(map[string]interface{})
	wkSignResult := WenKuSign(tieba.Bduss)
	zdSignResult := ZhiDaoSign(tieba.Bduss)
	if (totalCount != 0 && totalCount == cookieValidCount) || (totalCount == 0 && !CheckBdussValid(tieba.Bduss)) {
		//BDUSS失效
		signData["cookie_valid"] = 0
		Post("https://toolsbox.herokuapp.com/sign/update?flag=1&uid="+tieba.Uid, signData)
	} else {
		infoJson := GetUserProfile(tieba.Uid)
		headUrl := "http://tb.himg.baidu.com/sys/portrait/item/" +
			jsoniter.Get([]byte(infoJson), "user").Get("portrait").ToString()
		nameShow := jsoniter.Get([]byte(infoJson), "user").Get("name_show").ToString()
		signData["sign_count"] = signCount
		signData["head_url"] = headUrl
		signData["name_show"] = nameShow
		signData["retry_count"] = retryCount
		signData["excep_count"] = excepCount
		signData["black_count"] = blackCount
		signData["tb_count"] = totalCount
		signData["timespan"] = timespan
		signData["sign_date"] = time.Now()
		signData["sign_time"] = time.Now().UnixNano() / 1e6
		signData["wz"] = "文库：" + wkSignResult + ";" + "知道：" + zdSignResult
		Post("https://toolsbox.herokuapp.com/sign/update?flag=0&uid="+tieba.Uid, signData)
	}
	//最后设置签到状态
	Post("https://toolsbox.herokuapp.com/sign/update?flag=1&uid="+tieba.Uid, map[string]interface{}{
		"sign_status": 2,
	})
}
func SyncSignTieBa(baInfo LikedTieba, tieba Tieba, tbs string, chs chan ChanSignResult) ChanSignResult {
	signResult := SignOneTieBa(baInfo.Name, baInfo.Id, tieba.Bduss, tbs)
	guid := uuid.NewV4().String()
	csr := ChanSignResult{guid, baInfo.Id, baInfo.Name, tieba.Uid,
		tieba.Name, signResult.ErrorMsg, signResult, baInfo}
	chs <- csr
	return csr
}
