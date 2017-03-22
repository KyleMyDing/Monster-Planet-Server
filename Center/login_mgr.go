package main

/**
作者： Kyle Ding
模块：角色登录验证处理模块
说明：
创建时间：2015-11-2
**/

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//GetRoleListCallback 获取角色列表，同时需要验证token
func GetRoleListCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	info := runInfo.Info

	var req GxProto.GetRoleListReq
	var rsp GxProto.GetRoleListRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	if req.GetInfo() == nil || req.GetInfo().Token == nil {
		GxMisc.Warn("msg format error, msg: %s", req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	playerName := GxStatic.CheckToken(client, req.GetInfo().GetToken())
	if playerName == "" {
		GxMisc.Warn("token[%s] is not existst", req.GetInfo().GetToken())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetTokenError, nil)
		return
	}

	oldInfo := new(GxStatic.LoginInfo)
	oldInfo.PlayerName = playerName
	err = oldInfo.Get4Redis(client)
	if err == nil && (oldInfo.GateID != info.GateID || oldInfo.ConnID != info.ConnID) {
		GxMisc.Warn("player login conflict, playname: %s, info: %v", playerName, *oldInfo)
		sendMessage(runInfo, &req, GxMessage.MessageMaskDisconn, oldInfo.ConnID, GxStatic.CmdLoginConflict, 0, GxStatic.RetLoginConflict, nil)
	}

	IDs := GxStatic.GetRoleIDs4RoleList(client, playerName, config.ID)
	for i := 0; i < len(IDs); i++ {
		ID, _ := strconv.Atoi(IDs[i])

		role := new(GxStatic.Role)
		role.ID = ID
		err = role.Get4Redis(client)
		if err != nil {
			GxMisc.Warn("role %d is not existst", ID)
			continue
		}
		rsp.Roles = append(rsp.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(role.ID),
			Name:       proto.String(role.Name),
			Expr:       proto.Int64(role.Expr),
			VocationId: proto.Int(role.VocationID),
		})
	}

	//保存角色登录状态
	info.PlayerName = playerName
	info.BeginTs = time.Now().Unix()
	info.ServerID = config.ID
	GxStatic.SaveGateLoginInfo(client, info.GateID, info.ConnID, playerName)
	info.Set4Redis(client)

	//保存玩家最近登录服务器
	GxStatic.SavePlayerLastServer(client, playerName, config.ID, info.BeginTs)

	//更新roleGoroutine状态
	runInfo.Status = 1

	rsp.ServerTs = proto.Int(int(time.Now().Unix()))
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//SelectRoleCallback 选择角色进入游戏
func SelectRoleCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	info := runInfo.Info

	var req GxProto.SelectRoleReq
	var rsp GxProto.SelectRoleRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	if req.GetRoleId() == 0 {
		GxMisc.Warn("msg format error, msg: %s", req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	if runInfo.Status == 0 {
		GxMisc.Warn("role[%d] has not been logined", req.GetRoleId())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotLogin, nil)
		return
	}

	playerName, err1 := FillSelectRoleRsp(client, int(req.GetRoleId()), &rsp)
	if err1 != nil {
		GxMisc.Warn("role %d is not existst", req.GetRoleId())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
		return
	}

	//登陆角色不属于当前账号
	if info.PlayerName != playerName {
		GxMisc.Warn("role %d is not player: %s 's role", req.GetRoleId(), info.PlayerName)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)

	//保存角色登录状态
	info.RoleID = int(req.GetRoleId())
	info.Set4Redis(client)

	//更新roleGoroutine状态
	runInfo.R = new(GxStatic.Role)
	runInfo.R.ID = info.RoleID
	runInfo.R.Get4Redis(client)
	runInfo.Status = 2
	rolesGateInfo[info.RoleID] = fmt.Sprintf("%d:%d", runInfo.Info.GateID, runInfo.Info.ConnID)
	runInfo.Init <- 1

	//保存玩家最近登录时间,以便判断删除缓存时间
	GxStatic.SaveRoleLogin(client, uint32(config.ID), &GxStatic.RoleLoginInfo{
		RoleID: info.RoleID,
		Del:    0,
		Ts:     time.Now().Unix(),
	})
}

//CreateRoleCallback 创建角色进入游戏
func CreateRoleCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	info := runInfo.Info

	var req GxProto.CreateRoleReq
	var rsp GxProto.CreateRoleRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	if req.GetName() == "" || req.GetSex() == 0 {
		GxMisc.Warn("msg format error, msg: %s", req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	if runInfo.Status == 0 {
		GxMisc.Warn("conn[%d:%d] has not been logined", info.GateID, info.ConnID)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotLogin, nil)
		return
	}

	IDs := GxStatic.GetRoleIDs4RoleList(client, info.PlayerName, info.ServerID)
	if len(IDs) > 0 {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleExists, nil)
		return
	}

	//检查角色名冲突
	if !GxStatic.SaveRoleName(client, info.ServerID, req.GetName()) {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNameConflict, nil)
		return
	}

	//保存角色登录状态
	info.RoleID = CreateRole(client, info.PlayerName, req.GetName(), int(req.GetSex()))
	info.Set4Redis(client)

	//保存玩家角色列表
	GxStatic.SaveRoleID4RoleList(client, info.PlayerName, config.ID, info.RoleID)
	//
	err = FillCreateRoleRsp(client, info.RoleID, &rsp)
	if err != nil {
		GxMisc.Warn("role %d load fail, error: %s", info.RoleID, err)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotExists, nil)
		return
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)

	//更新roleGoroutine状态
	runInfo.R = new(GxStatic.Role)
	runInfo.R.ID = info.RoleID
	runInfo.R.Get4Redis(client)
	runInfo.Status = 2
	rolesGateInfo[info.RoleID] = fmt.Sprintf("%d:%d", runInfo.Info.GateID, runInfo.Info.ConnID)
	runInfo.Init <- 1

	//保存玩家最近登录时间,以便判断删除缓存时间
	GxStatic.SaveRoleLogin(client, uint32(config.ID), &GxStatic.RoleLoginInfo{
		RoleID: info.RoleID,
		Del:    0,
		Ts:     time.Now().Unix(),
	})

	GxStatic.SetRoleNameId(client, req.GetName(), info.RoleID)
}

//ClientLogout 角色登出通知
func ClientLogout(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	runInfo.Quit <- 1
}
