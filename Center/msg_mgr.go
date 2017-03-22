package main

/**
作者： Kyle Ding
模块：角色消息处理模块
说明：
创建时间：2015-11-2
**/

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func init() {
	//管理后台操作
	GxNet.RegisterMessageCallback(GxStatic.CmdAdminMessage, AdminMessageCallback)
	//心跳
	GxNet.RegisterMessageCallback(GxStatic.CmdHeartbeat, GateHeartbeatCallback)
	GxNet.RegisterMessageCallback(GxStatic.CmdRechargeNotify, RechagreCallback)

	//角色相关的消息由RoleCallback处理
	GxNet.RegisterOtherCallback(RoleCallback)
}

//RoleCallback 角色消息总入口
func RoleCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {
	checkRoleRunInfo(conn, gateID, connId)
	roleMsg(gateID, connId, GxMessage.Copy4MessagePool(msg))
}

//GateHeartbeatCallback 心跳回调
func GateHeartbeatCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {

}
