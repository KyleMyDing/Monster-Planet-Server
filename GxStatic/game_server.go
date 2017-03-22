package GxStatic

/**
作者： Kyle Ding
模块：游戏服务器信息模块
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

//GameServerTableName 服务器列表，redis的表名
var GameServerTableName = "h_game_server"

//GameServer 服务器信息
type GameServer struct {
	ID     int    //服务器ID
	Name   string //服务器名称
	Status uint32 //服务器状态
	OpenTs int64  //服务器开放时间
}

//SaveGameServer 保存指定服务器信息
func SaveGameServer(client *redis.Client, server *GameServer) error {
	buf, err := GxMisc.MsgToBuf(server)
	if err != nil {
		return err
	}

	client.HSet(GameServerTableName, strconv.Itoa(server.ID), string(buf))

	return nil
}

//GetAllGameServer 读取所有服务器信息
func GetAllGameServer(client *redis.Client, servers *[]*GameServer) error {
	m := client.HGetAllMap(GameServerTableName)
	r, err := m.Result()
	if err != nil {
		return err
	}

	for _, v := range r {
		j, err2 := GxMisc.BufToMsg([]byte(v))
		if err2 != nil {
			return err2
		}
		server := new(GameServer)
		GxMisc.JSONToStruct(j, server)
		*servers = append(*servers, server)
	}
	return nil
}
