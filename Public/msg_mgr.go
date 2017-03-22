package main

/**
作者： Kyle Ding
模块：角色消息处理模块
说明：
创建时间：2015-11-2
**/

import (
	"github.com/golang/protobuf/proto"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func init() {
	GxNet.RegisterMessageCallback(GxStatic.CmdHeartbeat, GateHeartbeatCallback)
	GxNet.RegisterMessageCallback(GxStatic.CmdRecharge, RechargeCallback)
}

//GateHeartbeatCallback 心跳回调
func GateHeartbeatCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {

}

func sendMessage(conn *GxNet.GxTCPConn, info *GxStatic.LoginInfo, req proto.Message, mask uint16, ID uint32, cmd uint16, seq uint16, ret uint16, rsp proto.Message) {
	GxNet.SendPbMessage(conn, mask, ID, cmd, seq, ret, rsp)

	//保存操作
	if info != nil {
		rdClient := GxMisc.PopRedisClient()
		defer GxMisc.PushRedisClient(rdClient)
		GxStatic.PutRoleOperateLog(info, GxStatic.GetRemote4Gate(rdClient, info.GateID, int(info.ConnID)), req, mask, cmd, seq, ret, rsp)
	}
}
