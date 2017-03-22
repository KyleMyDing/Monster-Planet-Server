package GxNet

/**
作者： Kyle Ding
模块：逻辑服务器模块
说明：
创建时间：2015-10-30
**/

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//MessageCallback 消息回调
type MessageCallback func(*GxTCPConn, int, int, *GxMessage.GxMessage)

//lsCmds 逻辑服务器要处理消息的回调列表
var lsCmds map[uint16]MessageCallback

//gates 网关连接列表
var gates map[int]*GxTCPConn

//lsCounter 消息seq计数器
var lsCounter *GxMisc.Counter

//otherCb 逻辑服务器不经过lsCmds处理消息的回调
var otherCb MessageCallback

func init() {
	gates = make(map[int]*GxTCPConn)
	lsCmds = make(map[uint16]MessageCallback)
	lsCounter = GxMisc.NewCounter()
}

//RegisterMessageCallback 注册需要处理的函数
func RegisterMessageCallback(cmd uint16, cb MessageCallback) {
	_, ok := lsCmds[cmd]
	if ok {
		GxMisc.Error("message callback has been registered, cmd: %d", cmd)
		return
	}
	lsCmds[cmd] = cb
}

//RegisterOtherCallback 注册逻辑服务器不经过lsCmds处理消息的回调
func RegisterOtherCallback(cb MessageCallback) {
	otherCb = cb
}

func gateRun(conn *GxTCPConn, gateID int) {
	t := time.NewTicker(10 * time.Second)
	go func(conn *GxTCPConn, t *time.Ticker) {
		for {
			select {
			case <-t.C:
				if !conn.Connected {
					return
				}
				SendPbMessage(conn, 0, 0, GxStatic.CmdHeartbeat, uint16(lsCounter.Genarate()), 0, nil)
			}
		}
	}(conn, t)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	for {
		newMsg, err := conn.Recv()
		if err != nil {
			GxMisc.Debug("Recv error: %s", err)
			return
		}

		if newMsg.GetCmd() != GxStatic.CmdHeartbeat {
			GxMisc.Debug("<<==== recv pb msg, info: %s", newMsg.String())
		}

		//消息流程结束后，系统会回收消息
		//如果回调里需要保存该消息，就复制一个新信息出来自己使用
		cb, ok := lsCmds[newMsg.GetCmd()]
		if ok {
			go func(conn *GxTCPConn, gateID int, connId int, newMsg *GxMessage.GxMessage) {
				cb(conn, gateID, connId, newMsg)
				GxMessage.FreeMessage(newMsg)
			}(conn, gateID, int(newMsg.GetID()), newMsg)
		} else {
			go func(conn *GxTCPConn, gateID int, connId int, newMsg *GxMessage.GxMessage) {
				otherCb(conn, gateID, connId, newMsg)
				GxMessage.FreeMessage(newMsg)
			}(conn, gateID, int(newMsg.GetID()), newMsg)
		}
	}
}

//ConnectAllGate 连接到所有的网关
func ConnectAllGate(serverID uint32) error {
	t := time.NewTicker(10 * time.Second)
	if len(lsCmds) == 0 {
		return errors.New("no cmd is registered")
	}

	var req GxProto.ServerConnectGateReq
	req.Id = proto.Uint32(serverID)
	if serverID >= GxStatic.ServerIDPublicBegin {
		for k := range lsCmds {
			req.Cmds = append(req.Cmds, uint32(k))
		}
	}

	f := func(pb proto.Message) error {
		rdClient := GxMisc.PopRedisClient()
		defer GxMisc.PushRedisClient(rdClient)

		var gatesinfo []*GxStatic.GateInfo
		err := GxStatic.GetAllGate(rdClient, &gatesinfo)
		if err != nil {
			return err
		}

		for i := 0; i < len(gatesinfo); i++ {
			conn, ok := gates[gatesinfo[i].ID]
			if ok {
				if conn.Connected {
					continue
				}
				delete(gates, gatesinfo[i].ID)
			}

			conn = NewTCPConn()
			err = conn.Connect(gatesinfo[i].Host2 + ":" + strconv.Itoa(int(gatesinfo[i].Port2)))
			if err != nil {
				GxMisc.Debug("connnect gate fail, remote: %s:%d", gatesinfo[i].Host2, gatesinfo[i].Port2)
				continue
			}
			GxMisc.Debug("connnect gate ok, remote: %s:%d", gatesinfo[i].Host2, gatesinfo[i].Port2)

			SendPbMessage(conn, 0, 0, GxStatic.CmdServerConnectGate, uint16(lsCounter.Genarate()), 0, pb)

			msg, err2 := conn.Recv()
			if err2 != nil || msg.GetRet() != GxStatic.RetSucc {
				GxMisc.Debug("connnect gate fail, remote: %s:%d", gatesinfo[i].Host2, gatesinfo[i].Port2)
				continue
			}

			gates[gatesinfo[i].ID] = conn
			go gateRun(conn, gatesinfo[i].ID)
		}
		return nil
	}

	//先连接一次
	err := f(&req)
	if err != nil {
		return err
	}

	//后面10秒检查一次
	for {
		select {
		case <-t.C:
			err = f(&req)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//FindGateByID 根据网关ID返回网关连接指针
func FindGateByID(gateID int) *GxTCPConn {
	conn, ok := gates[gateID]
	if ok {
		if conn.Connected {
			return conn
		}
	}
	return nil
}
