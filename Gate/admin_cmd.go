package main

/**
作者： Kyle Ding
模块：游戏管理后台网关操作管理模块
说明：
创建时间：2015-11-16
**/

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func clientCount(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	var req GxProto.AdminReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxMisc.Warn("message error, msg: %s", msg.String())
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return err
	}

	count := 0
	if clientRouter.ConnectCount() > AdminCount() {
		count = clientRouter.ConnectCount() - AdminCount()
	} else {
		count = 0
	}
	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(fmt.Sprintf("client_count: %d", count)),
	})
	return nil
}
