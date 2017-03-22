package main

/**
作者:""uangbo
模块：角色初始化管理模块
说明：
创建时间：2015-11-2
**/

import (
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//GetRoleFightInfoCallback 获取玩家战斗信息
func GetRoleFightInfoCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.GetRoleFightInfoReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get GetRoleFightInfoReq request, role: %s, msg: %s", role.String(), req.String())

	roleID := int(req.GetRoleId())
	if !GxStatic.RoleExists(client, roleID) {
		//角色不存在
		GxMisc.Warn("role is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetRoleNotOnline, nil)
	}

	var rsp GxProto.GetRoleFightInfoRsp
	rsp.Role = new(GxProto.RoleCommonInfo)
	rsp.FightCardBag = new(GxProto.PbFightCardBagInfo)

	FillRoleFightInfo(client, roleID, rsp.GetRole(), true)
	FillFightCardBagInfo(client, roleID, int(rsp.GetRole().GetFightCardDistri()), rsp.FightCardBag, true)

	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//SelectVocationCallback 选择职业
func SelectVocationCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.SelectVocationReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	vocationID := int(req.GetVocationId())
	if !GxDict.CheckDictId("Player", vocationID) {
		GxMisc.Warn("vocation is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetVocationNotExists, nil)
		return
	}

	if role.VocationID != 0 {
		GxMisc.Warn("vocation is selected, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetVocationSelected, nil)
		return
	}

	role.VocationID = vocationID
	role.SetField4Redis(client, "VocationID")
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}

//CreateRole 创建一个角色
func CreateRole(client *redis.Client, playerName string, roleName string, sex int) int {
	ID := initRoleComm(client, playerName, roleName, sex)

	return ID
}

func initRoleComm(client *redis.Client, playerName string, roleName string, sex int) int {
	gold, _ := GxDict.GetDictInt("CommonConfig", 1, "InitialGold")
	power, _ := GxDict.GetDictInt("CommonConfig", 1, "InitialVigor")
	role := &GxStatic.Role{
		ID:         GxStatic.NewRoleID(client),
		PlayerName: playerName,
		Name:       roleName,
		Sex:        sex,
		VocationID: 0,
		Vip:        0,
		Expr:       0,
		Money:      0,
		Gold:       int64(gold),
		Power:      int64(power),
		PowerTs:    time.Now().Unix(),
		Honor:      0,
		ArmyPay:    0,

		Wings:    "",
		Weapon:   "",
		Coat:     "",
		Cloakd:   "",
		Trouser:  "",
		Armguard: "",
		Shoes:    "",

		FightCardDistri: 0,
		BuyPowerCnt:     0,
		BuyPowerTs:      0,
		GetFreePowerTs:  0,
	}
	role.ServerID = GxStatic.ServerID
	role.Set4Redis(client)

	//初始化背包
	bag := GxStatic.NewBag(role.ID)
	bag.Set4Redis(client)

	/*初始化抽卡模块 by jie*/
	buyCard := GxStatic.NewBuyCard(role.ID)
	buyCard.Set4Redis(client)

	//初始化背包物品
	initialPackageStr, _ := GxDict.GetDictString("CommonConfig", 1, "InitialPackage")
	initialPackage := strings.Split(initialPackageStr, ";")
	for i := 0; i < len(initialPackage); i++ {
		itemInfo := strings.Split(initialPackage[i], ",")
		id, _ := strconv.Atoi(itemInfo[0])
		count, _ := strconv.Atoi(itemInfo[1])
		if GxStatic.IsEquipment(id) {
			for j := 0; j < count; j++ {
				role.NewEquipment(client, id, 0)
			}
		} else if GxStatic.IsItem(id) {
			role.NewItem(client, id, count)
		} else {
			GxMisc.Warn("error item id: %d", id)
		}
	}

	//初始化副本
	chapter := GxStatic.NewChapter(role.ID)
	chapter.Set4Redis(client)

	//c初始化卡组卡包
	fightCardBagName, _ := GxDict.GetDictString("CommonConfig", 1, "CardGroupName")
	index, _ := role.NewFightCardBag(client, fightCardBagName)

	initialCardGroupStr, _ := GxDict.GetDictString("CommonConfig", 1, "InitialCardGroup")
	cards := strings.Split(initialCardGroupStr, ",")
	for i := 0; i < len(cards); i++ {
		id, _ := strconv.Atoi(cards[i])
		role.NewCard(client, id, 1)

		item := GxStatic.GetRoleFightCard(client, role.ID, index, id)
		if item == nil {
			item = &GxStatic.Item{
				ID:  id,
				Cnt: 0,
			}
		}
		item.Cnt += 1

		GxStatic.SetRoleFightCard(client, role.ID, index, item)
	}

	//将角色保存创建列表
	GxStatic.SaveCreateRoleId4Server(client, role.ID)

	return role.ID
}
