package main

/**
作者： Kyle Ding
模块：装备消息管理模块
说明：
创建时间：2015-11-12
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func getRoleEquiStr(role *GxStatic.Role, EquiID int) (*string, string) {
	switch GxStatic.EquipmentPos(EquiID) {
	case 1:
		return &role.Wings, "Wings"
	case 2:
		return &role.Weapon, "Weapon"
	case 3:
		return &role.Coat, "Coat"
	case 4:
		return &role.Cloakd, "Cloakd"
	case 5:
		return &role.Trouser, "Trouser"
	case 6:
		return &role.Armguard, "Armguard"
	case 7:
		return &role.Shoes, "Shoes"
	}
	return nil, ""
}

//findEquiment 更新请求返回是否已经穿戴该装备和对应装备信息
func findEquiment(role *GxStatic.Role, client *redis.Client, req *GxProto.StrengthenEquipmentReq) (bool, *GxStatic.Item) {
	if req.GetEid() > 0 {
		//已经穿戴的装备
		str, _ := getRoleEquiStr(role, int(req.GetEid()))
		if *str == "" {
			return true, nil
		}
		j, _ := GxMisc.BufToMsg([]byte(*str))
		item := new(GxStatic.Item)
		GxMisc.JSONToStruct(j, item)
		return true, item
	}
	item := GxStatic.GetRoleItem4Bag(client, role.ID, int(req.GetIndx()))
	if item == nil || item.ID == 0 {
		return false, nil
	}
	return false, item
}

func strengthenEquipment(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage, req *GxProto.StrengthenEquipmentReq) {
	role := runInfo.R
	wear, item := findEquiment(role, client, req)
	if item == nil || !GxStatic.IsEquipment(item.ID) {
		//指定背包位置没有装备
		GxMisc.Warn("equipment is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquiNotExists, nil)
		return
	}
	nextEquiId, _ := GxDict.GetDictInt("Package", item.ID, "Levelup")
	if nextEquiId == -1 || !GxStatic.IsEquipment(nextEquiId) {
		//该装备不能强化
		GxMisc.Warn("equipment cannot strengthen, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquipCannotStrengen, nil)
		return
	}

	//强化消耗的金币
	money, _ := GxDict.GetDictInt("Package", item.ID, "DepGold")
	if role.Money < int64(money) {
		GxMisc.Warn("money is not enough, role: %s, money: %d, msg: %s", role.String(), role.Money, req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMoneyNotEnough, nil)
		return
	}

	p := 0
	//缓存要删除的装备
	consumeCount := len(req.GetConsumeIndex())
	consumeItems := make(map[int]*GxStatic.Item)
	var consumeIndexes []int
	for i := 0; i < consumeCount; i++ {
		index := int(req.GetConsumeIndex()[i])
		if !wear && int(req.GetIndx()) == index {
			//指定背包位置没有装备
			GxMisc.Warn("strengthen equipment, cannot consume self, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
			return
		}
		bagItem := GxStatic.GetRoleItem4Bag(client, role.ID, int(index))
		if bagItem == nil || !GxStatic.IsEquipment(bagItem.ID) {
			//指定位置不是装备
			GxMisc.Warn("bag quipment is not exists, role: %s, id: %d", role.String(), bagItem.ID)
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquiNotExists, nil)
			return
		}
		iVal, _ := GxDict.GetDictInt("Package", bagItem.ID, "LendExp")
		p += iVal
		consumeItems[index] = bagItem
		consumeIndexes = append(consumeIndexes, index)

		GxMisc.Debug("strengthen equipment, role: %s, p: %d, consume-equi[%d %d %d]", //
			role.String(), p, index, bagItem.ID, bagItem.Cnt)
	}
	if p == 0 {
		//没有装备
		GxMisc.Warn("consume equipment is null, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	gold := 0
	//是否使用元宝防降级
	goldColorup := req.GetGoldColorup() == 1
	if goldColorup {
		iVal, _ := GxDict.GetDictInt("Package", item.ID, "NoGradeDownCost")
		gold += iVal
	}

	goldTmp, _ := GxDict.GetDictInt("Package", item.ID, "IncreaseProbabilityCost")
	gold += goldTmp
	if req.GetAddProbability() == 1 {
		if role.Gold < int64(gold) {
			GxMisc.Warn("gold is not enough, role: %s, gold: %d, msg: %s", role.String(), role.Gold, req.String())
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
			return
		}
		iVal, _ := GxDict.GetDictInt("Package", item.ID, "GoldInIncreaseExp")
		p += iVal
		GxMisc.Debug("strengthen equipment, role: %s, p: %d, consume-gold: %d", role.String(), p, gold)
	}
	high, _ := GxDict.GetDictInt("Package", item.ID, "UpExp")
	r := GxMisc.GetRandomInterval(1, high)
	GxMisc.Debug("strengthen equipment, role: %s, p: %d, high: %d, r: %d", role.String(), p, high, r)

	var rsp GxProto.StrengthenEquipmentRsp
	rsp.Info = new(GxProto.RespondInfo)

	//扣除金币
	DelItem(client, runInfo, -1, GxStatic.IDMoney, money, rsp.Info)

	if req.GetAddProbability() == 1 {
		//扣除元宝
		DelItem(client, runInfo, -1, GxStatic.IDGold, gold, rsp.Info)
	}

	if r <= p {
		rsp.Result = proto.Int(0)
		//扣除全部装备
		for k, v := range consumeItems {
			DelItem(client, runInfo, k, v.ID, v.Cnt, rsp.Info)
		}
		//更新装备,新装备
		item.ID = nextEquiId
		item.Cnt = 0

		GxMisc.Debug("strengthen equipment ok, role: %s, new-item: %d", role.String(), nextEquiId)
	} else {
		//扣除部分装备
		downCount, _ := GxDict.GetDictInt("CommonConfig", 1, "FailTriggerDemotion")
		downP, _ := GxDict.GetDictInt("CommonConfig", 1, "DemotionProbability")
		p1, _ := GxDict.GetDictInt("CommonConfig", 1, "StrengthenFailLower")
		p2, _ := GxDict.GetDictInt("CommonConfig", 1, "StrengthenFailLower")
		//获取一个删除概率,计算删除装备数量,随机删除装备
		delP := GxMisc.GetRandomInterval(p1, p2)
		if delP > GxStatic.MaxP {
			delP = GxStatic.MaxP
		}
		delCount := consumeCount * delP / GxStatic.MaxP
		if delCount == 0 {
			delCount = 1
		}
		GxMisc.Debug("strengthen equipment fail, role: %s, item: %v, del-item-cnt: %d", role.String(), item, delCount)
		for i := delCount; i > 0; i-- {
			n := GxMisc.GetRandomInterval(0, len(consumeIndexes)-1)
			index := consumeIndexes[n]
			DelItem(client, runInfo, index, consumeItems[index].ID, consumeItems[index].Cnt, rsp.Info)
			//
			delete(consumeItems, index)
			n1 := n + 1
			consumeIndexes = append(consumeIndexes[:n], consumeIndexes[n1:]...)
		}
		//更新装备,保存失败次数
		item.Cnt++

		//如果失败次数大于指定，降级
		//如果使用元宝防降级，就不降级了
		downEquiId, _ := GxDict.GetDictInt("Package", item.ID, "LevelDown")
		if item.Cnt >= downCount && !goldColorup && GxMisc.GetRandomInterval(1, GxStatic.MaxP) <= downP && downEquiId != -1 {
			//装备降级
			rsp.Result = proto.Int(2)
			item.ID = downEquiId
			item.Cnt = 0
			GxMisc.Debug("strengthen equipment fail, and down equipment, role: %s, down-item: %d", role.String(), downEquiId)
		} else {
			rsp.Result = proto.Int(1)
		}
	}
	if wear {
		str, posName := getRoleEquiStr(role, int(req.GetEid()))
		buf, _ := GxMisc.MsgToBuf(item)
		*str = string(buf)
		role.SetField4Redis(client, posName)
	} else {
		role.UpdateItem(client, int(req.GetIndx()), item)
	}
	rsp.Equip = &GxProto.PbItemInfo{
		Id:  proto.Int(item.ID),
		Cnt: proto.Int(item.Cnt),
	}
	sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	GxMisc.Debug("strengthen equipment over, role: %s, rsp: %s", role.String(), rsp.String())
}

func colorupEquipment(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage, req *GxProto.StrengthenEquipmentReq) {
	role := runInfo.R

	wear, item := findEquiment(role, client, req)
	if item == nil || !GxStatic.IsEquipment(item.ID) {
		//指定背包位置没有装备
		GxMisc.Warn("equipment is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquiNotExists, nil)
		return
	}
	nextEquiId, _ := GxDict.GetDictInt("Package", item.ID, "Promotion")
	if nextEquiId <= 0 || !GxStatic.IsEquipment(nextEquiId) {
		//该装备不能突破
		GxMisc.Warn("equipment cannot colorup, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquipCannotStrengen, nil)
		return
	}

	var rsp GxProto.StrengthenEquipmentRsp
	rsp.Info = new(GxProto.RespondInfo)
	if req.GetGoldColorup() == 1 {
		//使用元宝
		gold, _ := GxDict.GetDictInt("Package", item.ID, "DirectlyPromotionCost")

		if role.Gold < int64(gold) {
			GxMisc.Warn("gold is not enough, role: %s, gold: %d, msg: %s", role.String(), role.Gold, req.String())
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
			return
		}

		DelItem(client, runInfo, -1, GxStatic.IDGold, gold, rsp.Info)
	} else {
		//使用材料
		if len(req.GetConsumeIndex()) != 1 {
			GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
			return
		}

		index := int(req.GetConsumeIndex()[0])
		if !wear && int(req.GetIndx()) == index {
			//指定背包位置没有装备
			GxMisc.Warn("colorup equipment, cannot consume self, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
			return
		}
		bagItem := GxStatic.GetRoleItem4Bag(client, role.ID, int(index))
		if bagItem == nil || bagItem.ID != item.ID {
			//指定位置不是装备
			GxMisc.Warn("bag quipment is been consumed by coloruping, role: %s, id: %d", role.String(), bagItem.ID)
			sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquipColorupErr, nil)
			return
		}

		DelItem(client, runInfo, index, bagItem.ID, bagItem.Cnt, rsp.Info)
	}

	//更新装备
	item.ID = nextEquiId
	item.Cnt = 0
	if wear {
		str, posName := getRoleEquiStr(role, int(req.GetEid()))
		buf, _ := GxMisc.MsgToBuf(item)
		*str = string(buf)
		role.SetField4Redis(client, posName)
	} else {
		role.UpdateItem(client, int(req.GetIndx()), item)
	}
	rsp.Equip = &GxProto.PbItemInfo{
		Id:  proto.Int(item.ID),
		Cnt: proto.Int(item.Cnt),
	}

	sendMessage(runInfo, req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	GxMisc.Debug("colorup equipment over, role: %s, rsp: %s", role.String(), rsp.String())
}

//StrengthenEquipmentCallback 强化装备
func StrengthenEquipmentCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.StrengthenEquipmentReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get StrengthenEquipmentReq request, role: %s, msg: %s", role.String(), req.String())
	if req.GetStrengthen() == 0 {
		strengthenEquipment(runInfo, client, msg, &req)
	} else if req.GetStrengthen() == 1 {
		colorupEquipment(runInfo, client, msg, &req)
	} else {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
}

//ReplaceEquipmentCallback 穿上装备
func ReplaceEquipmentCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ReplaceEquipmentReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ReplaceEquipmentCallback request, role: %s, msg: %s", role.String(), req.String())
	item := GxStatic.GetRoleItem4Bag(client, role.ID, int(req.GetIndx()))
	if item == nil || item.ID == 0 {
		//指定背包位置没有装备
		GxMisc.Warn("equipment is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetEquiNotExists, nil)
		return
	}
	if !GxStatic.IsEquipment(item.ID) {
		//指定位置不是装备
		GxMisc.Warn("item cannot replace, role: %s, msg: %s", role.String(), item.ID)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotReplace, nil)
		return
	}
	// lev := role.GetLev()
	// if PackageDict[item.ID]["UpExp"] > lev {
	// 	Warn("level is not enough, roleID: %d, lev: %d, need-lev: %d", role.ID, lev, PackageDict[item.ID]["UpExp"])
	// 	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetLevNotEnough, nil)
	// 	return
	// }

	str, posName := getRoleEquiStr(role, item.ID)
	var oldItem *GxStatic.Item = nil
	if *str != "" {
		j, _ := GxMisc.BufToMsg([]byte(*str))
		oldItem = new(GxStatic.Item)
		GxMisc.JSONToStruct(j, oldItem)
	}

	buf, _ := GxMisc.MsgToBuf(item)
	*str = string(buf)

	if oldItem == nil {
		role.DelEquipment(client, int(req.GetIndx()))
	} else {
		role.UpdateItem(client, int(req.GetIndx()), oldItem)
	}
	role.SetField4Redis(client, posName)

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}

//UnloadEquipmentCallback 卸下装备
func UnloadEquipmentCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.UnloadEquipmentReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get UnloadEquipmentCallback request, role: %s, msg: %s", role.String(), req.String())
	if req.Eid == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
	eID := int(req.GetEid())

	str, posName := getRoleEquiStr(role, int(req.GetEid()))
	if *str == "" {
		//指定位置没有装备
		GxMisc.Warn("hero equiment is empty, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	j, _ := GxMisc.BufToMsg([]byte(*str))
	item := new(GxStatic.Item)
	GxMisc.JSONToStruct(j, item)
	if item.ID != eID {
		//传进来的装备ID和已经装备的装备ID不一致
		GxMisc.Warn("equipment ID error, role: %S, ID: %d, msg: %s", role.String(), item.ID, req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	index := role.NewEquipment(client, eID, 0)
	if index == -1 {
		//传进来的装备ID和已经装备的装备ID不一致
		GxMisc.Warn("bag is full, role: %s, ID: %d, msg: %s", role.String(), item.ID, req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetBagFull, nil)
		return
	}
	*str = ""
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.UnloadEquipmentRsp{
		Indx: proto.Int(index),
	})
	role.SetField4Redis(client, posName)
}
