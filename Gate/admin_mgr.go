package main

/**
作者： Kyle Ding
模块：游戏管理后台连接管理模块
说明：
创建时间：2015-11-16
**/

import (
	"errors"
	"strings"
	"sync"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//trusts 信任ip列表
var trusts = []string{"127.0.0.1", "192.168.1.114"}

//IDGenerater 管理后台连接id生成器
var IDGenerater *GxMisc.Counter

var adminMutex sync.Mutex
var adminIDs map[uint32]*GxNet.GxTCPConn

var adminCb map[string]func(*GxNet.GxTCPConn, *GxMessage.GxMessage) error

func init() {
	adminIDs = make(map[uint32]*GxNet.GxTCPConn)
	adminCb = make(map[string]func(*GxNet.GxTCPConn, *GxMessage.GxMessage) error)

	adminCb["client_count"] = clientCount
}

//AdminConnectCallback 新管理后台连接
func AdminConnectCallback(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	//简单验证ip
	remote := conn.Remote
	ok := false
	for i := 0; i < len(trusts); i++ {
		if strings.Index(remote, trusts[i]) == 0 {
			ok = true
			break
		}
	}
	if !ok {
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetAdminDistrust, nil)
		return errors.New("admin is distrustful")
	}

	adminMutex.Lock()
	adminIDs[conn.ID] = conn
	adminMutex.Unlock()
	conn.M = "Adm"

	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
	return nil
}

//AdminMessageCallback 网关的管理后台命令总入口
func AdminMessageCallback(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) bool {
	var req GxProto.AdminReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxMisc.Warn("message error, msg: %s", msg.String())
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return true
	}

	f, ok := adminCb[req.GetCmd()]
	if ok {
		f(conn, msg)
		return true
	}
	return false
}

//CheckAdmin 检查连接是否属于管理后台
func CheckAdmin(conn *GxNet.GxTCPConn) bool {
	return conn.M == "Adm"
}

//FindAdminByID 根据id获取管理后台连接指针
func FindAdminByID(ID uint32) *GxNet.GxTCPConn {
	adminMutex.Lock()
	defer adminMutex.Unlock()
	return adminIDs[ID]
}

//AdminRawMessage 管理后台需要转发的消息
func AdminRawMessage(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	//Debug("new client message, remote: %s %s", conn.Remote, msg.String())
	//保存目的服务器ID
	serverID := msg.GetID()
	if AdminMessageCallback(conn, msg) {
		return nil
	}

	conn.ServerID = serverID

	msg.SetID(conn.ID)
	var server *GxNet.GxTCPConn
	if serverID < GxStatic.ServerIDPublicBegin {
		server = GetServerByID(serverID)
		if server == nil {
			GxMisc.Debug("Can not find a server, info: %s", msg.String())
			GxNet.SendPbMessage(conn, 0, conn.ID, msg.GetCmd(), msg.GetSeq(), GxStatic.RetServerNotExists, nil)
			return nil
		}
		server.Send(msg)
	} else {
		//根据admin请求转发到公共服务器
		// server = GetServerByID(serverID)
		// if server == nil {
		// 	GxMisc.Debug("Can not find a server, info: %s", msg.String())
		// 	GxNet.SendPbMessage(conn, 0, conn.ID, msg.GetCmd(), msg.GetSeq(), GxStatic.RetServerNotExists, nil)
		// 	return nil
		// }
		return nil
	}

	return nil
}

//AdminDisconn 管理后台断开连接
func AdminDisconn(conn *GxNet.GxTCPConn) {
	adminMutex.Lock()
	defer adminMutex.Unlock()

	delete(adminIDs, conn.ServerID)
}

//AdminCount 当前管理后台连接数
func AdminCount() int {
	adminMutex.Lock()
	defer adminMutex.Unlock()

	return len(adminIDs)
}
