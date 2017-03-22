package GxStatic

/**
作者： Kyle Ding
模块：充值模块
说明：
创建时间：2015-11-12
**/

import (
	"gopkg.in/redis.v3"
	"strconv"
	"sync"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//rechargeIDTableName 充值ID表，redis的表名
var rechargeIDTableName = "k_recharge_id"
var rechargeIDMutex sync.Mutex

//RoleRechargeList 角色未处理充值列表，redis的表名
var RoleRechargeList = "l_role_recharge_list:"

//Operate 充值操作信息
type Recharge struct {
	RechargeID   int64
	PayerName    string `index:"1"`   //角色帐号名
	RoleID       int    `index:"2"`   //角色ID
	Ts           int64  `type:"time"` //操作时间
	RechargeType int
	Rmb          int
	Gold         int //充值获得元宝
}

var rechargeChan chan *Recharge

func init() {
	rechargeChan = make(chan *Recharge, 10)
}

//NewRechargeID 生成一个新的充值ID
func NewRechargeID(client *redis.Client) int64 {
	rechargeIDMutex.Lock()
	defer rechargeIDMutex.Unlock()

	if !client.Exists(rechargeIDTableName).Val() {
		client.Set(rechargeIDTableName, "10000000", 0)
	}
	return client.Incr(rechargeIDTableName).Val()
}

//SaveRoleRechargeList 保存角色未处理充值列表，登录时检查，或者其他模块通知
func SaveRoleRechargeList(client *redis.Client, roleID int, recharge *Recharge) {
	buf, err := GxMisc.MsgToBuf(recharge)
	if err != nil {
		return
	}
	client.LPush(RoleRechargeList+strconv.Itoa(roleID), string(buf))
}

//GetRoleRechargeList 获取角色未处理充值列表
func GetRoleRechargeList(client *redis.Client, roleID int) []*Recharge {
	tablename := RoleRechargeList + strconv.Itoa(roleID)
	var recharges []*Recharge
	arr := client.LRange(tablename, 0, -1).Val()
	client.Del(tablename)
	for i := 0; i < len(arr); i++ {
		j, err := GxMisc.BufToMsg([]byte(arr[i]))
		if err == nil {
			recharge := new(Recharge)
			GxMisc.JSONToStruct(j, recharge)
			recharges = append(recharges, recharge)
		}
	}
	return recharges
}

func PutRechargeLog(recharge *Recharge) {
	rechargeChan <- recharge
}

func RunRechargeLog() {
	for {
		select {
		case recharge := <-rechargeChan:
			log := GxMisc.GenerateInsertSQL(recharge, "")

			_, err := GxMisc.MysqlPool.Exec(log)
			if err != nil {
				GxMisc.Error("execute log fail ,error: %s, sql: %s", err, log)
				return
			}
		}
	}

}
