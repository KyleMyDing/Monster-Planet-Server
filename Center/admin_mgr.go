package main

/**
作者： Kyle Ding
模块：管理后台消息管理模块
说明：
创建时间：2015-11-12
**/

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"strconv"
	// "gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//adminCb 管理后台接口列表
var adminCb map[string]func(*GxNet.GxTCPConn, *GxMessage.GxMessage)

func init() {
	adminCb = make(map[string]func(*GxNet.GxTCPConn, *GxMessage.GxMessage))
	adminCb["get_role_info"] = adminGetRoleInfo
	adminCb["reload_dict"] = adminReloadDict

	adminCb["get_role_bag_info"] = adminGetRoleBagInfo
	adminCb["get_role_card_info"] = adminGetRoleCardBagInfo
	adminCb["get_role_fight_card_info"] = adminGetRoleFightCardBagInfo
	adminCb["add_item"] = adminAddItem
	adminCb["online_roles"] = adminOnlineRoles

	adminCb["new_mail"] = adminNewMail
	adminCb["new_all_mail"] = adminNewMail4AllRole
}

//AdminMessageCallback 管理后台消息总入口函数
func AdminMessageCallback(conn *GxNet.GxTCPConn, gateID int, connId int, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxMisc.Warn("message error, msg: %s", msg.String())
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	if adminCb[req.GetCmd()] == nil {
		GxMisc.Warn("admin cmd is not support, cmd: %s", req.GetCmd())
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMessageNotSupport, nil)
		return
	}
	adminCb[req.GetCmd()](conn, msg)
}

//adminReloadDict 重新加载字典
func adminReloadDict(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	err := GxDict.LoadAllDict(config.DictDir)
	if err != nil {
		GxMisc.Debug("load dict fail, err: %s", err)
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
			Info: proto.String(fmt.Sprintf("load dict fail, err: %s", err)),
		})
		return
	}
	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(GxStatic.RetString[GxStatic.RetSucc]),
	})
}

//adminGetRoleInfo 获取角色常用信息
func adminGetRoleInfo(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	var rsp GxProto.AdminRsp
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	rsp.Role = new(GxProto.RoleCommonInfo)
	roleID, _ := strconv.Atoi(req.GetParameter()[0])
	_, err := FillRoleCommonInfo(rdClient, roleID, rsp.GetRole(), false)
	if err != nil {
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, &GxProto.AdminRsp{
			Info: proto.String(GxStatic.RetString[GxStatic.RetRoleNotExists]),
		})
		return
	}

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//adminGetRoleBagInfo 获取角色背包信息
func adminGetRoleBagInfo(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	var rsp GxProto.AdminRsp
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	rsp.Bag = new(GxProto.PbBagInfo)
	roleID, _ := strconv.Atoi(req.GetParameter()[0])
	err := FillBagInfo(rdClient, roleID, rsp.GetBag(), false)
	if err != nil {
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, &GxProto.AdminRsp{
			Info: proto.String(GxStatic.RetString[GxStatic.RetRoleNotExists]),
		})
		return
	}

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//adminGetRoleCardBagInfo 获取角色卡包信息
func adminGetRoleCardBagInfo(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	var rsp GxProto.AdminRsp
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	rsp.CardBag = new(GxProto.PbCardBagInfo)
	roleID, _ := strconv.Atoi(req.GetParameter()[0])
	err := FillCardBagInfo(rdClient, roleID, rsp.GetCardBag(), false)
	if err != nil {
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, &GxProto.AdminRsp{
			Info: proto.String(GxStatic.RetString[GxStatic.RetRoleNotExists]),
		})
		return
	}

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//adminGetRoleFightCardBagInfo 获取角色卡组信息
func adminGetRoleFightCardBagInfo(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	var rsp GxProto.AdminRsp
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	rsp.FightCardBag = new(GxProto.PbFightCardBagInfo)
	roleID, _ := strconv.Atoi(req.GetParameter()[0])

	role := &GxStatic.Role{
		ID: roleID,
	}
	role.GetField4Redis(rdClient, "FightCardDistri")

	err := FillFightCardBagInfo(rdClient, roleID, role.FightCardDistri, rsp.GetFightCardBag(), false)
	if err != nil {
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, &GxProto.AdminRsp{
			Info: proto.String(GxStatic.RetString[GxStatic.RetRoleNotExists]),
		})
		return
	}

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//adminAddItem 给角色发放物品
func adminAddItem(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	roleID, _ := strconv.Atoi(req.GetParameter()[0])
	if !GxStatic.RoleExists(rdClient, roleID) {
		GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, &GxProto.AdminRsp{
			Info: proto.String(GxStatic.RetString[GxStatic.RetRoleNotExists]),
		})
		return
	}

	count := (len(req.GetParameter()) - 1) / 2

	for i := 0; i < count; i++ {
		ID, _ := strconv.Atoi(req.GetParameter()[i*2+1])
		cnt, _ := strconv.Atoi(req.GetParameter()[i*2+2])

		GxStatic.SaveRoleUngetItemList(rdClient, roleID, &GxStatic.Item{
			ID:  ID,
			Cnt: cnt,
		})
	}

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(GxStatic.RetString[GxStatic.RetSucc]),
	})

	RefRole(roleID, func(rri *RoleRunInfo) {
		if rri != nil {
			msg := GxMessage.GetGxMessage()
			msg.SetCmd(GxStatic.CmdAdminInfoChange)
			rri.NotifyQue <- msg
		} else {
			GxMisc.Trace("role[%d] is not online", roleID)
		}
	})
}

//adminOnlineRoles 返回在线角色列表
func adminOnlineRoles(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	msg.UnpackagePbmsg(&req)

	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()
	count := 0
	temp := ""
	for _, runInfo := range rolesRunInfo {
		if runInfo.R == nil {
			continue
		}
		temp += runInfo.R.String() + "\r\n"
	}
	str := fmt.Sprintf("total online role count: %d\r\n", count) + temp
	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(str),
	})
}

//adminNewMail 新邮件
func adminNewMail(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	roleID, _ := strconv.Atoi(req.GetParameter()[0])
	sender := req.GetParameter()[1]
	title := req.GetParameter()[2]
	text := req.GetParameter()[3]
	items := req.GetParameter()[4] //json string, key-id value-count

	//todo 检查邮件内容的格式
	GxMisc.Trace("new mail to %d, title: %s, items: %s", roleID, title, items)

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(GxStatic.RetString[GxStatic.RetSucc]),
	})

	SendMailToRole(rdClient, roleID, GxStatic.NewMail4String(rdClient, roleID, sender, title, text, items))
}

//adminNewMail4AllRole 全服发放新邮件
func adminNewMail4AllRole(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) {
	var req GxProto.AdminReq
	msg.UnpackagePbmsg(&req)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	sender := req.GetParameter()[0]
	title := req.GetParameter()[1]
	text := req.GetParameter()[2]
	items := req.GetParameter()[3] //json string, key-id value-count

	//todo 检查邮件内容的格式
	GxMisc.Trace("new mail, title: %s, items: %s", title, items)

	GxNet.SendPbMessage(conn, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.AdminRsp{
		Info: proto.String(GxStatic.RetString[GxStatic.RetSucc]),
	})

	//go
	ids := GxStatic.GetRoleIds4Server(rdClient)
	for i := 0; i < len(ids); i++ {
		SendMailToRole(rdClient, ids[i], GxStatic.NewMail4String(rdClient, ids[i], sender, title, text, items))
	}
}
