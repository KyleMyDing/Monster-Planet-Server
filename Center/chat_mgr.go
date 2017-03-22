package main

/**
作者： Kyle Ding
模块：聊天消息管理模块
说明：
创建时间：2015-11-26
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//PubChat 世界聊天
func PubChat(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage, req *GxProto.SendChatReq) {
	role := runInfo.R

	//直接返回结果
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	//发送聊天通知给所有在线玩家
	OnlineRoleSendMsg(role, GxStatic.CmdNewChat, 0, GxStatic.RetSucc, &GxProto.NewChatNotify{
		Type: req.Type,
		Ts:   proto.Int64(time.Now().Unix()),
		Text: req.Text,
		Role: &GxProto.RoleCommonInfo{
			Id:         proto.Int(role.ID),
			Name:       proto.String(role.Name),
			Vip:        proto.Int(role.Vip),
			VocationId: proto.Int(role.VocationID),
		},
	})
}

//ArmyChat 军团聊天
func ArmyChat(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage, req *GxProto.SendChatReq) {

}

//PriChat 私聊
func PriChat(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage, req *GxProto.SendChatReq) {
	role := runInfo.R

	RefRole(int(req.GetRoleId()), func(rri *RoleRunInfo) {
		if rri == nil {
			sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotOnline, nil)
		} else {
			sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
			RoleSendMsg(rri, GxStatic.CmdNewChat, 0, GxStatic.RetSucc, &GxProto.NewChatNotify{
				Type: req.Type,
				Ts:   proto.Int64(time.Now().Unix()),
				Text: req.Text,
				Role: &GxProto.RoleCommonInfo{
					Id:         proto.Int(role.ID),
					Name:       proto.String(role.Name),
					Vip:        proto.Int(role.Vip),
					VocationId: proto.Int(role.VocationID),
				},
			})
		}
	})
}

//SendChatCallback 发送聊天消息
func SendChatCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.SendChatReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get SendChatCallback request, role: %s, msg: %s", role.String(), req.String())
	if req.Type == nil || (req.GetType() == 3 && req.RoleId == nil) || req.Text == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
	switch req.GetType() {
	case 1:
		PubChat(runInfo, client, msg, &req)
		break
	case 2:
		ArmyChat(runInfo, client, msg, &req)
		break
	case 3:
		PriChat(runInfo, client, msg, &req)
		break
	default:
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
}
