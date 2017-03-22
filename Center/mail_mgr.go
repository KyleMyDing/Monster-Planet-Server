package main

/**
作者： Kyle Ding
模块：邮件消息模块
说明：
创建时间：2015-11-14
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func mailsToPbMails(mails *[]*GxStatic.Mail, pbmails *[]*GxProto.PbMailInfo) {
	for i := 0; i < len((*mails)); i++ {
		if (*mails)[i].Del == 1 {
			continue
		}
		mail := new(GxProto.PbMailInfo)
		mail.Id = proto.Int((*mails)[i].ID)
		mail.Sender = proto.String((*mails)[i].Sender)
		mail.Ts = proto.Int(int((*mails)[i].Ts))
		mail.Title = proto.String((*mails)[i].Title)
		mail.Text = proto.String((*mails)[i].Text)
		mail.Read = proto.Int(int((*mails)[i].Rd))

		items := (*mails)[i].GetItems()
		for j := 0; j < len(items); j++ {
			mail.Items = append(mail.Items, &GxProto.PbItemInfo{
				Id:  proto.Int(items[j].ID),
				Cnt: proto.Int(items[j].Cnt),
			})
		}

		(*pbmails) = append((*pbmails), mail)
	}
}

//GetMailListCallback 获取邮件列表
func GetMailListCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R

	GxMisc.Trace("get GetMailListReq request, role: %s", role.String())

	var rsp GxProto.GetMailListRsp
	mails := GxStatic.GetRoleAllMail(client, role.ID)
	mailsToPbMails(&mails, &rsp.Mails)
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//GetMailItemCallback 领取(删除)邮件
func GetMailItemCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.GetMailItemReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get GetMailItemReq request, role: %s, msg: %s", role.String(), req.String())

	if len(req.GetId()) == 0 {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	var mails []*GxStatic.Mail
	for i := 0; i < len(req.GetId()); i++ {
		id := int(req.GetId()[i])
		mail := GxStatic.GetRoleMail(client, role.ID, id)
		if mail == nil {
			//邮件不存在
			GxMisc.Warn("mail is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMailNotExists, nil)
			return
		}
		mails = append(mails, mail)
	}

	var rsp GxProto.GetMailItemRsp
	rsp.Info = new(GxProto.RespondInfo)

	//领取物品
	for i := 0; i < len(mails); i++ {
		items := mails[i].GetItems()
		for j := 0; j < len(items); j++ {
			GetItem(client, runInfo, items[j].ID, items[j].Cnt, rsp.GetInfo())
		}

		//删除邮件
		GxStatic.DelRoleMail(client, role.ID, mails[i].ID)
	}

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//SendMailToRole 发放新邮件
func SendMailToRole(client *redis.Client, roleID int, mail *GxStatic.Mail) {
	GxStatic.SaveRoleUngetMailList(client, roleID, mail)

	RefRole(roleID, func(rri *RoleRunInfo) {
		if rri != nil {
			msg := GxMessage.GetGxMessage()
			msg.SetCmd(GxStatic.CmdAdminNewMail)
			rri.NotifyQue <- msg
		} else {
			GxMisc.Trace("role[%d] is not online", roleID)
		}
	})
}

//ReadMailCallback 阅读邮件
func ReadMailCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ReadMailReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ReadMailReq request, role: %s, msg: %s", role.String(), req.String())

	id := int(req.GetId())
	mail := GxStatic.GetRoleMail(client, role.ID, id)
	if mail == nil {
		//邮件不存在
		GxMisc.Warn("mail is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMailNotExists, nil)
		return
	}

	mail.Rd = 1
	GxMisc.SetFieldFromRedis(client, mail, "Read")

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}
