package main

/**
作者： Kyle Ding
模块：角色登录验证处理模块
说明：
创建时间：2015-11-2
**/

import (
	// "github.com/golang/protobuf/proto"
	// "gopkg.in/redis.v3"
	// "strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//RechargeCallback 充值请求
func RechargeCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	var req GxProto.RechargeReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(conn, nil, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	info, ret := GxStatic.CheckLoginInfo(rdClient, gateID, uint32(connId))
	if ret != GxStatic.RetSucc {
		sendMessage(conn, nil, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), ret, nil)
		return
	}

	GxMisc.Trace("get RechargeReq request, role: %d, msg: %s", info.RoleID, req.String())

	sendMessage(conn, info, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	GxStatic.SaveRoleRechargeList(rdClient, info.RoleID, &GxStatic.Recharge{
		RechargeID:   GxStatic.NewRechargeID(rdClient),
		PayerName:    info.PlayerName,
		RoleID:       info.RoleID,
		Ts:           time.Now().Unix(),
		RechargeType: int(req.GetRechargeType()),
		Rmb:          10,
		Gold:         100,
	})
	sendMessage(conn, nil, nil, GxMessage.MessageMaskInternal, msg.GetID(), GxStatic.CmdRechargeNotify, msg.GetSeq(), GxStatic.RetSucc, nil)
}
