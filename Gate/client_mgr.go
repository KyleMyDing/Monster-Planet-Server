package main

/**
作者: guangbo
模块: 网关外网连接管理模块
说明：
创建时间：2015-11-16
**/

import (
	"errors"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

var clientRouter *GxNet.GxTCPServer
var myRedis *redis.Client = nil

var clientMessageDst map[uint32]uint32

func init() {
	clientMessageDst = make(map[uint32]uint32)

	clientRouter = GxNet.NewGxTCPServer(0x00FFFFFF, ClientNewConn, ClientDisConn, ClientRawMessage, false)

	clientRouter.RegisterClientCmd(GxStatic.CmdAdminConnect, AdminConnectCallback)
	// clientRouter.RegisterClientCmd(CmdAdminMessage, AdminMessageCallback)
}

//ClientNewConn 新客户端连接
func ClientNewConn(conn *GxNet.GxTCPConn) {
	conn.M = "Cli"
	GxMisc.Debug(">>==<< new client connnect, ID: %d, remote: %s", conn.ID, conn.Remote)

	if myRedis == nil {
		myRedis = GxMisc.PopRedisClient()
	}
	GxStatic.SetRemote4Gate(myRedis, config.ID, int(conn.ID), conn.Remote)
}

//ClientDisConn 客户端断开连接
func ClientDisConn(conn *GxNet.GxTCPConn) {
	//检查是不是管理后台的连接
	if CheckAdmin(conn) {
		AdminDisconn(conn)
		return
	}

	GxMisc.Debug("<<==>> dis client connnect, ID: %d, remote: %s", conn.ID, conn.Remote)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	//玩家断开连接，通知他已经连接上的服务器
	playerName := GxStatic.GetGateLoginInfo(rdClient, config.ID, conn.ID)
	if playerName != "" {
		info := new(GxStatic.LoginInfo)
		info.PlayerName = playerName
		info.Get4Redis(rdClient)
		if info.ServerID != 0 {
			server := GetServerByID(uint32(info.ServerID))
			if server == nil {
				GxMisc.Debug("Can not find a server, serverID: %d", info.ServerID)
				return
			}
			GxNet.SendPbMessage(server, GxMessage.MessageMaskNotify, conn.ID, GxStatic.CmdClientLogout, 0, 0, nil)
		}
	}
}

//ClientRawMessage 不需要处理的消息,根据服务器id转发
func ClientRawMessage(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	//检查是不是管理后台的连接
	if CheckAdmin(conn) {
		return AdminRawMessage(conn, msg)
	}

	//Debug("new client message, remote: %s %s", conn.Remote, msg.String())
	//保存目的服务器ID
	serverID := msg.GetID()
	msg.SetID(conn.ID)

	//保存连接所属的服务器
	if conn.ServerID != 0 && serverID != conn.ServerID {
		GxMisc.Debug("message dst error, old serverID: %d, serverID: %d, info: %s", conn.Remote, serverID, conn.ServerID, msg.String())
		GxNet.SendPbMessage(conn, 0, conn.ID, msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return errors.New("message dst error")
	}
	conn.ServerID = serverID

	server := GetServerByCmd(serverID, msg.GetCmd())
	if server != nil {
		server.Send(msg)
		return nil
	}

	server = GetServerByID(serverID)
	if server == nil {
		GxMisc.Debug("Can not find a server, serverID: %d, info: %s", serverID, msg.String())
		GxNet.SendPbMessage(conn, 0, conn.ID, msg.GetCmd(), msg.GetSeq(), GxStatic.RetServerMaintain, nil)
		return errors.New("Can not find a server")
	}
	server.Send(msg)
	return nil
}
