package GxStatic

/**
作者： Kyle Ding
模块：网关信息模块
说明：
创建时间：2015-10-30
**/

import (
	"gopkg.in/redis.v3"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//GateInfoTableName 网关列表，redis的表名
var GateInfoTableName = "h_gate_info"

//GateInfo 网关信息
type GateInfo struct {
	ID    int    //网关ID
	Host1 string //网关外网ip
	Port1 int    //网关外网端口
	Host2 string //网关内网ip
	Port2 int    //网关内网端口
	Count int    //当前连接数
	Ts    int64  //信息更新时间
}

//SaveGate 保存指定网关信息
func SaveGate(client *redis.Client, gate *GateInfo) error {
	buf, err := GxMisc.MsgToBuf(gate)
	if err != nil {
		return err
	}

	client.HSet(GateInfoTableName, strconv.Itoa(int(gate.ID)), string(buf))
	return nil
}

//GetAllGate 获取所有网关信息
func GetAllGate(client *redis.Client, gates *[]*GateInfo) error {
	m := client.HGetAllMap(GateInfoTableName)
	r, err := m.Result()
	if err != nil {
		return err
	}
	for _, v := range r {
		j, err2 := GxMisc.BufToMsg([]byte(v))
		if err2 != nil {
			return err2
		}
		gate := new(GateInfo)
		GxMisc.JSONToStruct(j, gate)
		*gates = append(*gates, gate)
	}
	return nil
}
