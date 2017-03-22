package main

/**
作者： Kyle Ding
模块：背包消息管理模块
说明：
创建时间：2015-11-12
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

//GetFriendCallback 获取好友列表
func GetFriendCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var rsp GxProto.GetFriendRsp

	friend := GxStatic.NewFriend(role.ID)
	friend.Get4Redis(client)

	rsp.Ts = proto.Int(int(friend.Ts))
	rsp.Cnt = proto.Int(friend.Cnt)
	for k, v := range friend.Friends {
		if !GxStatic.RoleExists(client, k) {
			//角色不存在
			GxMisc.Error("friend role is not exitst, role: %s, friend: [%d:%d]", role.String(), k, v)
			continue
		}

		newFriend := &GxStatic.Role{
			ID: k,
		}
		newFriend.Get4Redis(client)

		online := 0
		if RoleIsOnline(k) {
			online = 1
		}

		send := 0
		if friend.GetSend(k) {
			send = 1
		}
		recv := 0
		if friend.GetRecv(k) {
			recv = 1
		}

		info := &GxProto.RoleCommonInfo{
			Id:         proto.Int(newFriend.ID),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
			FightValue: proto.Int(0),
			Online:     proto.Int(online),
			Send:       proto.Int(send),
			Recv:       proto.Int(recv),
		}

		if friend.IsUnaccept(k) {
			rsp.Unaccepts = append(rsp.Unaccepts, info)
		} else if friend.IsAccept(k) {
			rsp.Accepts = append(rsp.Accepts, info)
		} else {
			GxMisc.Error("error friend, role: %s, friend: [%d:%d]", role.String(), k, v)
		}
	}

	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//GetReferrerCallback 获取推荐好友列表
func GetReferrerCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var rsp GxProto.GetReferrerRsp

	friend := GxStatic.NewFriend(role.ID)
	friend.Get4Redis(client)

	//暂时随机获取
	ids := GxStatic.GetRoleIds4Server(client)
	for {
		if len(ids) == 0 {
			break
		}
		index := GxMisc.GetRandomInterval(0, len(ids)-1)
		id := ids[index]

		index1 := index + 1
		ids = append(ids[:index], ids[index1:]...)

		if !GxStatic.RoleExists(client, id) || friend.Friends[id] != 0 || id == role.ID {
			//角色不存在或者已经在好友列表
			continue
		}

		newFriend := &GxStatic.Role{
			ID: id,
		}
		newFriend.Get4Redis(client)

		online := 0
		if RoleIsOnline(id) {
			online = 1
		}

		//返回id,name,vocationId,expr,FightValue,online等字段
		rsp.Roles = append(rsp.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(id),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
			FightValue: proto.Int(0),
			Online:     proto.Int(online),
		})

		//最多10个推荐好友
		if len(rsp.Roles) >= 10 {
			break
		}
	}

	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//AddFriendCallback 添加好友列表
func AddFriendCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.AddFriendReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get AddFriendReq request, role: %s, msg: %s", role.String(), req.String())

	var ids []int
	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])
		if !GxStatic.RoleExists(client, id) {
			//该角色不存在
			GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
			return
		}

		if GxStatic.IsFriend(client, role.ID, id) {
			//自己好友或者待确认好友
			GxMisc.Warn("role has been friend, role: %s, friend-id: %d", role.String(), id)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetHasBeenFriend, nil)
			return
		}
		ids = append(ids, id)
	}

	for i := 0; i < len(req.GetRoleName()); i++ {
		id := GxStatic.GetRoleNameId(client, req.GetRoleName()[i])
		if id == 0 || !GxStatic.RoleExists(client, id) {
			//该角色不存在
			GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
			return
		}

		if GxStatic.IsFriend(client, role.ID, id) {
			//自己好友或者待确认好友
			GxMisc.Warn("role has been friend, role: %s, friend-id: %d", role.String(), id)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetHasBeenFriend, nil)
			return
		}
		ids = append(ids, id)
	}

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	for i := 0; i < len(ids); i++ {
		id := int(ids[i])

		//往目标玩家发送添加好友通知
		GxStatic.SaveRoleUndealFriendLMessageist(client, id, GxStatic.CmdAdminAddFriend, role.ID)
		RefRole(id, func(rri *RoleRunInfo) {
			if rri != nil {
				msg := GxMessage.GetGxMessage()
				msg.SetCmd(GxStatic.CmdAdminAddFriend)
				rri.NotifyQue <- msg
			} else {
				GxMisc.Trace("role[%d] is not online", id)
			}
		})
	}
}

//DealFriendCallback 同意或者拒绝添加好友
func DealFriendCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.DealFriendReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get DealFriendReq request, role: %s, msg: %s", role.String(), req.String())

	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])
		if !GxStatic.RoleExists(client, id) {
			//该角色不存在
			GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
			return
		}

		if !GxStatic.IsUnacceptFriend(client, role.ID, id) {
			//不是待确认好友
			GxMisc.Warn("role is not unaccept friend, role: %s, friend-id: %d", role.String(), id)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotUnacceptFriend, nil)
			return
		}
	}

	friend := GxStatic.NewFriend(role.ID)

	//添加好友
	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])
		if req.GetAgree() == 0 {
			//拒绝
			friend.DeleteFriend4Redis(client, id)
		} else {
			//同意
			friend.Friends[id] = 1 << uint32(GxStatic.FriendMaskAccept)
			friend.SetFriend4Redis(client, id)
		}
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])

		//往目标玩家发送好友添加处理通知
		cmd := GxStatic.CmdAdminDealFriendYes
		if req.GetAgree() == 0 {
			cmd = GxStatic.CmdAdminDealFriendNo
		}

		GxStatic.SaveRoleUndealFriendLMessageist(client, id, cmd, role.ID)

		RefRole(id, func(rri *RoleRunInfo) {
			if rri != nil {
				msg := GxMessage.GetGxMessage()
				msg.SetCmd(uint16(cmd))
				rri.NotifyQue <- msg
			} else {
				GxMisc.Trace("role[%d] is not online", id)
			}
		})
	}
}

//DelelteFriendCallback 删除好友
func DelelteFriendCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.DealFriendReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get DealFriendReq request, role: %s, msg: %s", role.String(), req.String())

	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])
		if !GxStatic.RoleExists(client, id) {
			//该角色不存在
			GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
			return
		}

		if !GxStatic.IsFriend(client, role.ID, id) {
			//不是正式好友
			GxMisc.Warn("role is not friend, role: %s, friend-id: %d", role.String(), id)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotFriend, nil)
			return
		}
	}

	//删除好友
	friend := GxStatic.NewFriend(role.ID)
	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])
		friend.DeleteFriend4Redis(client, id)
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	for i := 0; i < len(req.GetRoleId()); i++ {
		id := int(req.GetRoleId()[i])

		//往目标玩家发送添加好友通知
		GxStatic.SaveRoleUndealFriendLMessageist(client, id, GxStatic.CmdAdminDeleteFriend, role.ID)
		RefRole(id, func(rri *RoleRunInfo) {
			if rri != nil {
				msg := GxMessage.GetGxMessage()
				msg.SetCmd(GxStatic.CmdAdminDeleteFriend)
				rri.NotifyQue <- msg
			} else {
				GxMisc.Trace("role[%d] is not online", id)
			}
		})
	}
}

//FriendSendCallback 赠送物品好友
func FriendSendCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.FriendSendReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get FriendSendReq request, role: %s, msg: %s", role.String(), req.String())

	id := int(req.GetRoleId())
	if !GxStatic.RoleExists(client, id) {
		//该角色不存在
		GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
		return
	}

	if !GxStatic.IsFriend(client, role.ID, id) {
		//不是正式好友
		GxMisc.Warn("role is not friend, role: %s, friend-id: %d", role.String(), id)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotFriend, nil)
		return
	}

	friend := GxStatic.NewFriend(role.ID)
	friend.Get4Redis(client)
	if friend.GetSend(id) {
		//今天已经赠送过该好友
		GxMisc.Warn("today has been sent to friend, role: %s, friend-id: %d", role.String(), id)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetHasSendFriend, nil)
		return
	}

	friend.SetSend(client, id)
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)

	//往目标玩家发送赠送通知
	GxStatic.SaveRoleUndealFriendLMessageist(client, id, GxStatic.CmdAdminFriendSend, role.ID)
	RefRole(id, func(rri *RoleRunInfo) {
		if rri != nil {
			msg := GxMessage.GetGxMessage()
			msg.SetCmd(GxStatic.CmdAdminFriendSend)
			rri.NotifyQue <- msg
		} else {
			GxMisc.Trace("role[%d] is not online", id)
		}
	})
}

//FriendRecvCallback 领取赠送物品好友
func FriendRecvCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.FriendRecvReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get FriendRecvReq request, role: %s, msg: %s", role.String(), req.String())

	id := int(req.GetRoleId())
	if !GxStatic.RoleExists(client, id) {
		//该角色不存在
		GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
		return
	}

	if !GxStatic.IsFriend(client, role.ID, id) {
		//不是正式好友
		GxMisc.Warn("role is not friend, role: %s, friend-id: %d", role.String(), id)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotFriend, nil)
		return
	}

	friend := GxStatic.NewFriend(role.ID)
	friend.Get4Redis(client)
	if !friend.GetRecv(id) {
		//今天已经领取
		GxMisc.Warn("today has been sent to friend, role: %s, friend-id: %d", role.String(), id)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetHasRecvFriend, nil)
		return
	}

	if friend.Ts < (GxMisc.NextTime(0, 0, 0) - GxStatic.Day) {
		friend.Ts = time.Now().Unix()
		friend.Cnt = 0
	}
	max := 10
	if friend.Cnt >= max {
		//已经达到当天最大领取次数
		GxMisc.Warn("today it is max count, role: %s, count: %d", role.String(), friend.Cnt)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRecvFriendMax, nil)
		return
	}
	friend.Cnt++
	GxMisc.SetFieldFromRedis(client, friend, "Ts")
	GxMisc.SetFieldFromRedis(client, friend, "Cnt")

	var rsp GxProto.FriendRecvRsp
	rsp.Info = new(GxProto.RespondInfo)

	power := 5
	GetItem(client, runInfo, GxStatic.IDPower, power, rsp.GetInfo())
	friend.SetRecv(client, id)
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}
