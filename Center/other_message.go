package main

/**
作者： Kyle Ding
模块：其他消息处理模块
说明：
创建时间：2015-11-12
**/

import (
// "github.com/golang/protobuf/proto"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//RechagreCallback 充值到账，由充值服务器发送
func RechagreCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	info, ret := GxStatic.CheckLoginInfo(rdClient, gateID, uint32(connId))
	if ret != GxStatic.RetSucc {
		return
	}

	RefRole(info.RoleID, func(rri *RoleRunInfo) {
		if rri != nil {
			msg := GxMessage.GetGxMessage()
			msg.SetCmd(GxStatic.CmdAdminRecharge)
			rri.NotifyQue <- msg
		} else {
			GxMisc.Trace("get Rechagre notify, role[%d] is not online", info.RoleID)
		}
	})
}
