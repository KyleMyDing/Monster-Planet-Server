package main

/**
作者： Kyle Ding
模块：网关内网连接管理模块
说明：
创建时间：2015-10-30
**/

import (
	"sync"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//ServersInfo 服务器连接列表
type ServersInfo []*GxNet.GxTCPConn

var serverRouter *GxNet.GxTCPServer
var mutex *sync.Mutex

//CmdServers 保存所有注册cmd的服务器
var CmdServers map[uint16]ServersInfo

//IDServers 保存所有注册id的服务器
var IDServers map[uint32]*GxNet.GxTCPConn

func init() {
	mutex = new(sync.Mutex)
	CmdServers = make(map[uint16]ServersInfo)
	IDServers = make(map[uint32]*GxNet.GxTCPConn)

	serverRouter = GxNet.NewGxTCPServer(0x01FFFFFF, ServerNewConn, ServerDisConn, ServerRawMessage, false)
	serverRouter.RegisterClientCmd(GxStatic.CmdServerConnectGate, ServerConnectGateCallback)
}

//ServerNewConn 新服务器连接
func ServerNewConn(conn *GxNet.GxTCPConn) {
	conn.M = "Srv"
	GxMisc.Debug(">>==<< new server connnect, remote: %s", conn.Remote)
}

//ServerDisConn 服务器断开连接
func ServerDisConn(conn *GxNet.GxTCPConn) {
	GxMisc.Debug("<<==>> dis server connnect, serverID: %d, remote: %s", conn.ServerID, conn.Remote)

	mutex.Lock()
	if len(conn.Data) > 0 {
		for i := 0; i < len(conn.Data); i++ {
			// delete(CmdServers[conn.Data[i]], conn.Remote)
			for j := 0; j < len(CmdServers[conn.Data[i]]); j++ {
				if conn.ServerID == CmdServers[conn.Data[i]][j].ServerID {
					nextJ := j + 1
					CmdServers[conn.Data[i]] = append(CmdServers[conn.Data[i]][:j], CmdServers[conn.Data[i]][nextJ:]...)
					break
				}
			}
		}
		delete(IDServers, conn.ServerID)
	} else {
		delete(IDServers, conn.ServerID)
	}
	mutex.Unlock()
}

//ServerRawMessage 转发服务器消息
func ServerRawMessage(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	//Debug("new server message, remote: %s %s", conn.Remote, msg.String())
	client := clientRouter.FindConnByID(msg.GetID())
	if client == nil {
		GxMisc.Debug("client connect is not exists, msg: %s", msg.String())
		return nil

	}

	if msg.GetMask(GxMessage.MessageMaskInternal) {
		//根据消息查找公共服务器
		server := GetServerByCmd(msg.GetID(), msg.GetCmd())
		if server != nil {
			server.Send(msg)
			return nil
		}

		if client.ServerID == 0 {
			GxMisc.Debug("connid has not serverid, remote: %s, msg: %s", conn.Remote, msg.String())
			return nil
		}

		//0x01FFFFFF server
		server = GetServerByID(client.ServerID)
		if server != nil {
			server.Send(msg)
			return nil
		}
	} else {
		//0x00FFFFFF client and admin
		client.Send(msg)
		if msg.GetMask(GxMessage.MessageMaskDisconn) {
			client.Conn.Close()
		}
		return nil
	}

	GxMisc.Debug("msg cannot find target, remote: %s, msg: %s", conn.Remote, msg.String())
	return nil
}

//ServerConnectGateCallback 服务器注册请求
func ServerConnectGateCallback(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	//register server
	var req GxProto.ServerConnectGateReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return err
	}

	mutex.Lock()
	if len(req.GetCmds()) > 0 {
		if req.GetId() < GxStatic.ServerIDPublicBegin {
			//公用服务器id错误
			GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
			return err
		}
		//公用服务器
		conn.Data = make([]uint16, len(req.GetCmds()))

		for i := 0; i < len(req.GetCmds()); i++ {
			//保存自己处理的消息cmd
			cmd := uint16(req.GetCmds()[i])
			conn.Data[i] = cmd

			//之前是否用相同id的服务器存在
			find := false
			for j := 0; j < len(CmdServers[cmd]); j++ {
				if CmdServers[cmd][j].ServerID == req.GetId() {
					CmdServers[cmd][j] = conn
					find = true
					break
				}
			}
			if !find {
				CmdServers[cmd] = append(CmdServers[cmd], conn)
			}
		}

	} else {
		//游戏区服务器
		if req.GetId() >= GxStatic.ServerIDPublicBegin {
			//游戏区服务器
			GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
			return err
		}
	}
	IDServers[req.GetId()] = conn
	conn.ServerID = req.GetId()
	GxMisc.Debug("new server register, ID: %d", req.GetId())
	mutex.Unlock()

	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
	return nil
}

//GetServerByCmd 根据cmd返回服务器连接指针
func GetServerByCmd(srvid uint32, cmd uint16) *GxNet.GxTCPConn {
	mutex.Lock()
	defer mutex.Unlock()

	count := len(CmdServers[cmd])
	if count == 0 {
		return nil
	}
	return CmdServers[cmd][srvid%uint32(count)]
}

//GetServerByID 根据id返回服务端连接指针
func GetServerByID(ID uint32) *GxNet.GxTCPConn {
	mutex.Lock()
	defer mutex.Unlock()

	conn, ok := IDServers[ID]
	if ok {
		return conn
	}
	return nil
}
