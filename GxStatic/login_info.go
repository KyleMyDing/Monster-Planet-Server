package GxStatic

/**
作者： Kyle Ding
模块：玩家登录信息模块
说明：
创建时间：2015-10-30
**/

import (
	"database/sql"
	"fmt"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//LoginInfo 帐号录信息
type LoginInfo struct {
	PlayerName string `pk:"true"` //帐号名
	GateID     int    //当前网关ID
	ConnID     uint32 //当前网关连接ID
	BeginTs    int64  //登录时间
	EndTs      int64  //登出时间
	ServerID   int    //当前连接的游戏服务器
	RoleID     int    //当前登录的角色ID
}

//GateLoginInfoTableName 指定网关和网关的客户连接ID的连接用户信息，redis的表名，GateID:ConnID=PlayerName
var GateLoginInfoTableName = "k_gate_login_info:"

//PlayerLastServerTableName 帐号最近登录服务器ID信息，redis的表名
var PlayerLastServerTableName = "k_player_last_server:"

//GateRemoteTableName 网关连接地址列表
var GateRemoteTableName = "h_gate_remote"

//Set4Redis ...
func (info *LoginInfo) Set4Redis(client *redis.Client) error {
	GxMisc.SaveToRedis(client, info)
	return nil
}

//Get4Redis ...
func (info *LoginInfo) Get4Redis(client *redis.Client) error {
	return GxMisc.LoadFromRedis(client, info)
}

//Del4Redis ...
func (info *LoginInfo) Del4Redis(client *redis.Client) error {
	return GxMisc.DelFromRedis(client, info)
}

//Set4Mysql ...
func (info *LoginInfo) Set4Mysql(db *sql.DB) error {
	return nil
}

//Get4Mysql ...
func (info *LoginInfo) Get4Mysql(db *sql.DB) error {
	return nil
}

//Del4Mysql ...
func (info *LoginInfo) Del4Mysql(db *sql.DB) error {
	return nil
}

func CheckLoginInfo(client *redis.Client, gateID int, connID uint32) (*LoginInfo, uint16) {
	playerName := GetGateLoginInfo(client, gateID, connID)
	if playerName == "" {
		return nil, RetNotLogin
	}
	info := &LoginInfo{
		PlayerName: playerName,
	}
	info.Get4Redis(client)

	if info.RoleID == 0 {
		return nil, RetNotLogin
	}

	return info, RetSucc
}

//SaveGateLoginInfo 保存指定网关和网关的客户连接ID的连接用户信息
func SaveGateLoginInfo(client *redis.Client, gateID int, connID uint32, palyerName string) {
	key := GateLoginInfoTableName + strconv.Itoa(gateID) + ":" + strconv.Itoa(int(connID))

	client.Set(key, palyerName, 0)
}

//GetGateLoginInfo 获取指定网关和网关的客户连接ID的连接用户信息
func GetGateLoginInfo(client *redis.Client, gateID int, connID uint32) string {
	key := GateLoginInfoTableName + strconv.Itoa(gateID) + ":" + strconv.Itoa(int(connID))
	return client.Get(key).Val()
}

//DelGateLoginInfo 删除指定网关和网关的客户连接ID的连接用户信息
func DelGateLoginInfo(client *redis.Client, gateID int, connID uint32) {
	key := GateLoginInfoTableName + strconv.Itoa(gateID) + ":" + strconv.Itoa(int(connID))

	client.Del(key)
}

//DisconnLogin 掉线重连
func DisconnLogin(client *redis.Client, token string, info *LoginInfo) uint16 {
	playerName := CheckToken(client, token)
	if playerName == "" {
		return RetTokenError
	}
	oldInfo := new(LoginInfo)
	oldInfo.PlayerName = playerName
	oldInfo.Get4Redis(client)

	if info.GateID != oldInfo.GateID || info.ConnID != oldInfo.ConnID {
		info.PlayerName = oldInfo.PlayerName
		info.ServerID = oldInfo.ServerID
		info.RoleID = oldInfo.RoleID
		info.BeginTs = oldInfo.BeginTs
		DelGateLoginInfo(client, oldInfo.GateID, oldInfo.ConnID)
		SaveGateLoginInfo(client, info.GateID, info.ConnID, playerName)
		info.Set4Redis(client)
	}
	return RetSucc
}

//SavePlayerLastServer 保存帐号最近登录服务器ID信息
func SavePlayerLastServer(client *redis.Client, playerName string, serverID int, ts int64) {
	client.Set(PlayerLastServerTableName+playerName, fmt.Sprintf("%d:%d", ServerID, ts), 0)
}

//GetPlayerLastServer 获取帐号最近登录服务器ID信息
func GetPlayerLastServer(client *redis.Client, playerName string) (int, int64) {
	if !client.Exists(PlayerLastServerTableName + playerName).Val() {
		return 0, 0
	}
	arr := strings.Split(client.Get(PlayerLastServerTableName+playerName).Val(), ":")
	if len(arr) != 2 {
		return 0, 0
	}
	serverID, _ := strconv.Atoi(arr[0])
	ts, _ := strconv.ParseInt(arr[1], 10, 64)
	return serverID, ts
}

//SetRemote4Gate 保存网关客户端连接地址
func SetRemote4Gate(client *redis.Client, gateId int, connId int, remote string) {
	client.HSet(GateRemoteTableName, fmt.Sprintf("%d:%d", gateId, connId), remote)
}

//GetRemote4Gate 获取网关客户端连接地址
func GetRemote4Gate(client *redis.Client, gateId int, connId int) string {
	return client.HGet(GateRemoteTableName, fmt.Sprintf("%d:%d", gateId, connId)).Val()
}
