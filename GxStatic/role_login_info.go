package GxStatic

/**
作者： Kyle Ding
模块：保存角色最近登录信息模块
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

//RoleLoginTableName 指定角色缓存状态，redis的表名
var RoleLoginTableName = "h_role_Login:"

//RoleLoginInfo 角色缓存状态
type RoleLoginInfo struct {
	RoleID int   //角色ID
	Del    int   //是否已经清除了缓存
	Ts     int64 `type:"time"` //登录时间
}

//SaveRoleLogin 保存指定服务器指定角色缓存状态
func SaveRoleLogin(client *redis.Client, serverID uint32, info *RoleLoginInfo) error {
	buf, err := GxMisc.MsgToBuf(info)
	if err != nil {
		return err
	}

	client.HSet(RoleLoginTableName+strconv.Itoa(int(serverID)), strconv.Itoa(int(info.RoleID)), string(buf))

	return nil
}

//GetAllRoleLogin 获取指定服务器所有角色缓存状态
func GetAllRoleLogin(client *redis.Client, serverID uint32, infos *[]*RoleLoginInfo) error {
	m := client.HGetAllMap(RoleLoginTableName + strconv.Itoa(int(serverID)))
	r, err := m.Result()
	if err != nil {
		return err
	}

	for _, v := range r {
		j, err2 := GxMisc.BufToMsg([]byte(v))
		if err2 != nil {
			return err2
		}
		info := new(RoleLoginInfo)
		GxMisc.JSONToStruct(j, info)
		*infos = append(*infos, info)
	}
	return nil
}
