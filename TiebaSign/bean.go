package TiebaSign

import (
	"time"
)

type Tieba struct {
	Uid          string    `json:"uid"`
	Name         string    `json:"name"`
	Bduss        string    `json:"bduss"`
	Head_url     string    `json:"head_url"`
	Tb_count     int64     `json:"tb_count"`
	Sign_count   int64     `json:"sign_count"`
	Sign_date    time.Time `json:"sign_date"`
	wz           string    `json:"wz"`
	Timespan     int64     `json:"timespan"`
	Retry_count  int64     `json:"retry_count"`
	Open_id      string    `json:"open_id"`
	Cookie_valid int64     `json:"cookie_valid"`
	Sign_time    int64     `json:"sign_time"`
	Stoken       string    `json:"stoken"`
	Excep_count  int64     `json:"excep_count"`
	Black_count  int64     `json:"black_count"`
	Last_job     string    `json:"last_job"`
	Sign_status  int64     `json:"sign_status"`
	Server_name  string    `json:"server_name"`
	Qq           string    `json:"qq"`
	Name_show    string    `json:"name_show"`
	Notify       int64     `json:"notify"`
	Main         int64     `json:"main"`
	Vip          int64     `json:"vip"`
	Like_url     string    `json:"like_url"`
	Wx_name      string    `json:"wx_name"`
}

type LikedTieba struct {
	Id            string `json:"id,omitempty" gorm:"-"`
	Name          string `json:"name,,omitempty" gorm:"-"`
	Favo_type     string `json:"favo_type" gorm:"-"`
	Level_id      string `json:"level_id"`
	Level_name    string `json:"level_name"`
	Cur_score     string `json:"cur_score"`
	Levelup_score string `json:"levelup_score"`
	Avatar        string `json:"avatar"`
	Slogan        string `json:"slogan"`
}

type LikedApiRep struct {
	ForumList  ForumList `json:"forum_list"`
	HasMore    string    `json:"has_more"`
	ServerTime string    `json:"server_time"`
	Time       int64     `json:"time"`
	Ctime      int       `json:"ctime"`
	Logid      int       `json:"logid"`
	ErrorCore  string    `json:"error_core"`
}

type ForumList struct {
	NonGconforum []LikedTieba `json:"non-gconforum"`
	Gconforum    []LikedTieba `json:"gconforum"`
}

type SignResult struct {
	ErrorCode    string `json:"error_code"`
	ErrorMsg     string `json:"error_msg,omitempty" gorm:"-"`
	SignTime     int64  `json:"signTime" gorm:"column:signTime"`
	SignPoint    string `json:"sign_point" gorm:"-"`
	CountSignNum string `json:"count_sign_num" gorm:"-"`
	Timespan     int64  `json:"timespan" gorm:"-"`
}

type ChanSignResult struct {
	Guid   string `json:"guid"`
	Fid    string `json:"fid"`
	Fname  string `json:"fname"`
	Uid    string `json:"uid"`
	Uname  string `json:"uname"`
	RetMsg string `json:"ret_msg"`
	SignResult
	LikedTieba
}
type TieBaModel struct {
	Level float64 `json:"level"`
	ChanSignResult
}
type BqTieBa struct {
	Bduss     string    `json:"bduss"`
	Fid       int       `json:"fid"`
	Tbname    string    `json:"tbname"`
	Guid      string    `json:"guid"`
	ErrorCode string    `json:"error_code"`
	ErrorMsg  string    `json:"error_msg"`
	Username  string    `json:"username"`
	Isdelete  int       `json:"isdelete"`
	SignDate  time.Time `json:"sign_date"`
}

type Prision struct {
	Id       int    `json:"id"`
	Uid      string `json:"uid"`
	Bduss    string `json:"bduss"`
	Tbs      string `json:"tbs"`
	Uname    string `json:"uname"`
	Tbname   string `json:"tbname"`
	Days     int    `json:"days"`
	Reason   string `json:"reason"`
	Portrait string `json:"portrait"`
	Count    int    `json:"count"`
}

func (ChanSignResult) TableName() string {
	return "tieba"
}
