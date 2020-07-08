package TiebaSign

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"net/http"
	_ "net/url"
	url2 "net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

//http get方法
func Get(url string) string {
	res, _ := http.Get(url)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func Post(url string, postData map[string]interface{}) string {
	bytesData, err := jsoniter.Marshal(postData)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{
		Timeout: time.Duration(15 * time.Second),
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	//byte数组直接转成string，优化内存
	str := (*string)(unsafe.Pointer(&respBytes))
	return *str
}

//公共贴吧请求（带cookie）
func Fetch(url string, postData map[string]interface{}, bduss string, stoken string) (string, error) {
	return FetchWithHeaders(url, postData, bduss, stoken, nil)
}

func FetchWithHeaders(url string, postData map[string]interface{}, bduss string, stoken string, headers map[string]string) (string, error) {
	var request *http.Request
	httpClient := &http.Client{}
	if nil == postData {
		request, _ = http.NewRequest("GET", url, nil)
	} else {
		postParams := url2.Values{}
		for key, value := range postData {
			postParams.Set(key, Strval(value))
		}
		postDataStr := postParams.Encode()
		postDataBytes := []byte(postDataStr)
		postBytesReader := bytes.NewReader(postDataBytes)
		request, _ = http.NewRequest("POST", url, postBytesReader)
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if "" != bduss {
		request.AddCookie(&http.Cookie{Name: "BDUSS", Value: bduss})
	}
	if "" != stoken {
		request.AddCookie(&http.Cookie{Name: "STOKEN", Value: stoken})
	}
	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}
	response, fetchError := httpClient.Do(request)
	if fetchError != nil {
		return "", fetchError
	}
	defer response.Body.Close()
	body, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return "", readError
	}
	return string(body), nil
}

//获取tbs
func GetTbs(bduss string) string {
	body, err := Fetch("http://tieba.baidu.com/dc/common/tbs", nil, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	isLogin := jsoniter.Get([]byte(body), "is_login").ToInt()
	if isLogin == 1 {
		return jsoniter.Get([]byte(body), "tbs").ToString()
	}
	return ""
}

//BDUSS有效性检测
func CheckBdussValid(bduss string) bool {
	body, err := Fetch("http://tieba.baidu.com/dc/common/tbs", nil, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	isLogin := jsoniter.Get([]byte(body), "is_login").ToInt()
	if isLogin == 1 {
		return true
	}
	return false
}

//获取用户关注的所有贴吧
func GetLikedTiebas(bduss string, uid string) ([]LikedTieba, error) {
	pn := 0
	if uid == "" {
		uid = "" //获取uid
	}
	likedTiebaList := make([]LikedTieba, 0)
	for {
		pn++
		var postData = map[string]interface{}{
			"_client_version": "6.2.2",
			"is_guest":        "0",
			"page_no":         strconv.Itoa(pn),
			"uid":             uid,
		}
		postData["sign"] = DataSign(postData)
		body, err := Fetch("http://c.tieba.baidu.com/c/f/forum/like", postData, bduss, "")
		if err != nil {
			log.Println("err:", err)
		}
		var likedApiRep LikedApiRep
		if err := jsoniter.Unmarshal([]byte(body), &likedApiRep); err != nil {
			log.Println("err: ", err)
			break
		} else {
			for _, likeTb := range likedApiRep.ForumList.Gconforum {
				likedTiebaList = append(likedTiebaList, likeTb)
			}
			for _, likeTb := range likedApiRep.ForumList.NonGconforum {
				likedTiebaList = append(likedTiebaList, likeTb)
			}
			if likedApiRep.HasMore == "0" {
				break
			}
		}
	}
	return likedTiebaList, nil
}

//签到一个贴吧
func SignOneTieBa(tbName string, fid string, bduss string, tbs string) SignResult {
	start := time.Now().UnixNano() / 1e6
	var postData = map[string]interface{}{
		"_client_id":      "03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36",
		"_client_type":    "4",
		"_client_version": "1.2.1.17",
		"_phone_imei":     "540b43b59d21b7a4824e1fd31b08e9a6",
		"fid":             fid,
		"kw":              tbName,
		"net_type":        "3",
		"tbs":             tbs,
	}
	postData["sign"] = DataSign(postData)
	body, err := Fetch("http://c.tieba.baidu.com/c/c/forum/sign", postData, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	errorCode := jsoniter.Get([]byte(body), "error_code").ToString()
	errorMsg := jsoniter.Get([]byte(body), "error_msg").ToString()
	userInfo := jsoniter.Get([]byte(body), "user_info")
	signResult := SignResult{}
	if errorCode == "0" {
		//签到成功
		if userInfo == nil {
			signResult.SignPoint = "0"
			signResult.CountSignNum = "0"
		} else {
			signResult.SignPoint = userInfo.Get("sign_bonus_point").ToString()
			signResult.CountSignNum = userInfo.Get("cont_sign_num").ToString()
		}

		errorMsg = "签到成功"
	}
	signResult.SignTime = time.Now().UnixNano() / 1e6
	signResult.ErrorCode = errorCode
	signResult.ErrorMsg = errorMsg
	span := (time.Now().UnixNano() / 1e6) - start
	signResult.Timespan = span
	return signResult
}

//文库签到
func WenKuSign(bduss string) string {
	headers := make(map[string]string)
	headers["Host"] = "wenku.baidu.com"
	headers["Referer"] = "https://wenku.baidu.com/task/browse/daily"
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4184.0 Safari/537.36"
	body, err := FetchWithHeaders("https://wenku.baidu.com/task/submit/signin", nil, bduss, "", headers)
	if err != nil {
		log.Println("err: ", err)
	}
	errorNo := jsoniter.Get([]byte(body), "error_no").ToString()
	if body != "" && (errorNo != "0" || errorNo != "1") {
		return "已签到"
	}
	return "未签到"
}

//文库签到
func ZhiDaoSign(bduss string) string {
	stokenBody, err1 := FetchWithHeaders("https://zhidao.baidu.com", nil, bduss, "", map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4184.0 Safari/537.36"})
	if err1 != nil {
		log.Println("err: ", err1)
	}
	stoken := GetBetweenStr(stokenBody, `"stoken":"`, `",`)
	stoken = Substr(stoken, 10, 32)
	time := time.Now().UnixNano() / 1e6
	s := strconv.FormatInt(time, 10)
	var postData = map[string]interface{}{
		"cm":     "100509",
		"stoken": stoken,
		"utdata": "52,52,15,5,9,12,9,52,12,4,15,13,17,12,13,5,13," + s,
	}
	body, err := Fetch("http://zhidao.baidu.com/submit/user", postData, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	errorNo := jsoniter.Get([]byte(body), "errorNo").ToString()
	if body != "" && (errorNo != "0" || errorNo != "2") {
		return "已签到"
	}
	return "未签到"
}

//获取用户基本信息
func GetUserProfile(uid string) string {
	var postData = map[string]interface{}{
		"_client_version": "6.1.2",
		"has_plist":       "2",
		"need_post_count": "1",
		"uid":             uid,
	}
	postData["sign"] = DataSign(postData)
	body, err := Fetch("http://c.tieba.baidu.com/c/u/user/profile", postData, "", "")
	if err != nil {
		log.Println("err: ", err)
	}
	return body
}

//封禁
func Commitprison(bduss string, tbs string, userName string, tbName string, days int, reason string,
	portrait string) string {
	var postData = map[string]interface{}{
		"BDUSS":           bduss,
		"_client_type":    2,
		"_client_version": "11.2.8.1",
		"day":             days,
		"fid":             GetFid(tbName),
		"ntn":             "banid",
		"reason":          reason,
		"tbs":             tbs,
		"timestamp":       time.Now().UnixNano() / 1e6,
		"word":            tbName,
		"z":               "1234",
	}
	if len(portrait) > 0 {
		postData["portrait"] = portrait
		postData["nick_name"] = userName
		postData["un"] = ""
	} else {
		postData["un"] = userName
	}
	postData["sign"] = DataSign(postData)
	body, err := Fetch("http://c.tieba.baidu.com/c/c/bawu/commitprison", postData, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	return body
}

//根据贴吧名称获取fid
func GetFid(tbName string) string {
	fid := ""
	body := Get("http://tieba.baidu.com/f/commit/share/fnameShareApi?ie=utf-8&fname=" + tbName)
	jsonBody := jsoniter.Get([]byte(body))
	if jsonBody.Get("no").ToInt() == 0 {
		fid = jsonBody.Get("data").Get("fid").ToString()
	}
	return fid
}

//贴吧未开放此功能
//名人堂助攻： 已助攻{"no":2280006,"error":"","data":[]}
//名人堂助攻： 助攻成功{"no":0,"error":"","data":[...]}
//未关注此吧{"no":3110004,"error":"","data":[]}
func CelebritySupport(bduss string, tbName string, fid string, tbs string) string {
	if fid == "" && tbName == "" {
		log.Fatal("至少包含贴吧名字、FID中的一个")
	} else if fid == "" && tbName != "" {
		fid = GetFid(tbName)
	}
	if tbs == "" {
		tbs = GetTbs(bduss)
	}
	postData := map[string]interface{}{"forum_id": fid, "tbs": tbs}
	body, err := Fetch("http://tieba.baidu.com/celebrity/submit/getForumSupport", postData, bduss, "")
	if err != nil {
		log.Println("err: ", err)
	}
	npcInfo := jsoniter.Get([]byte(body), "data", 0).Get("npc_info")
	if npcInfo.Size() > 0 {
		npcId := npcInfo.Get("npc_id").ToString()
		postData["npc_id"] = npcId
		suportResult, _ := Fetch("http://tieba.baidu.com/celebrity/submit/support", postData, bduss, "")
		no := jsoniter.Get([]byte(suportResult)).Get("no").ToInt()
		if no == 3110004 {
			return "未关注此吧"
		} else if no == 2280006 {
			return "已助攻"
		} else if no == 0 {
			return "助攻成功"
		}
		return suportResult
	}
	return "该贴吧未开放此功能"
}

//贴吧参数sing MD5签名
func DataSign(postData map[string]interface{}) string {
	var keys []string
	for key, _ := range postData {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	sign_str := ""
	for _, key := range keys {
		sign_str += fmt.Sprintf("%s=%s", key, postData[key])
	}
	sign_str += "tiebaclient!!!"
	MD5 := md5.New()
	MD5.Write([]byte(sign_str))
	MD5Result := MD5.Sum(nil)
	signValue := make([]byte, 32)
	hex.Encode(signValue, MD5Result)
	return strings.ToUpper(string(signValue))
}

//k8中的多线程，控制并发数
type DoWorkPieceFunc func(piece int)

// Parallelize is a very simple framework that allow for parallelizing
// N independent pieces of work.
func Parallelize(workers, pieces int, doWorkPiece DoWorkPieceFunc) {
	toProcess := make(chan int, pieces)
	for i := 0; i < pieces; i++ {
		toProcess <- i
	}
	close(toProcess)

	if pieces < workers {
		workers = pieces
	}

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer utilruntime.HandleCrash()
			defer wg.Done()
			for piece := range toProcess {
				doWorkPiece(piece)
			}
		}()
	}
	wg.Wait()
}
func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}
func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := jsoniter.Marshal(value)
		key = string(newValue)
	}

	return key
}
