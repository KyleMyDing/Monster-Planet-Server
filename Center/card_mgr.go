/**
作者： Kyle Ding
模块：卡牌消息管理模块
说明：
创建时间：2015-11-14
**/
package main

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func decomposeCard(runInfo *RoleRunInfo, client *redis.Client, color int, info *GxProto.RespondInfo) {
	str, _ := GxDict.GetDictString("Cardsdecomposition", color, "SoulStone")
	decompressItems := strings.Split(str, ",")
	n := len(decompressItems) / 3
	for i := 0; i < n; i++ {
		low, _ := strconv.Atoi(decompressItems[i*3])
		high, _ := strconv.Atoi(decompressItems[i*3+1])
		dropID, _ := strconv.Atoi(decompressItems[i*3+2])
		count := GxMisc.GetRandomInterval(low, high)
		GxMisc.Trace("Decompose Card, low: %d, high: %d, dropID: %d", low, high, dropID)
		for j := 0; j < count; j++ {
			items := GxDict.GetDrop4CardDecompress(dropID)
			for k := 0; k < len(items); k++ {
				GetItem(client, runInfo, items[k].ID, items[k].Cnt, info)
			}
		}
	}
}

//FuseCardCallback 融合卡片
func FuseCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.FuseCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get FuseCardReq request, role: %s, msg: %s", role.String(), req.String())
	if req.Card == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
	if !GxDict.CheckDictId("HerosCard", int(req.GetCard().GetId())) {
		//卡牌不存在
		GxMisc.Warn("card is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardNotExists, nil)
		return
	}
	card := GxStatic.GetRoleCard(client, role.ID, int(req.GetCard().GetId()))
	if card == nil || card.Cnt < int(req.GetCard().GetCnt()) || req.GetCard().GetCnt() < 2 {
		//卡牌不足
		GxMisc.Warn("card is not enough, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardNotEnough, nil)
		return
	}

	color, _ := GxDict.GetDictInt("HerosCard", int(req.GetCard().GetId()), "Color")
	nextCardID, _ := GxDict.GetDictInt("HerosCard", int(req.GetCard().GetId()), "CardsFusion")
	if nextCardID == -1 {
		//已经是最高等级
		GxMisc.Warn("card has been max color, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardMaxLevel, nil)
		return
	}

	//money := (*GxDict.Dicts)["Cardsfusion"].D[color]["CostCoin"].(int)
	money, _ := GxDict.GetDictInt("Cardsfusion", color, "CostCoin")
	if role.Money < int64(money) {
		//金币不足
		GxMisc.Warn("Money is not enough, role: %s, Money: %d, need-money: %d, msg: %s", role.String(), role.Money, money, req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMoneyNotEnough, nil)
		return
	}

	// 基础概率	元宝加成	元宝数	卡牌数+1
	// Probability	GoldBonus	GoldNumber	CardNumber+1
	p, _ := GxDict.GetDictInt("Cardsfusion", color, "Probability")
	goldBonus, _ := GxDict.GetDictInt("Cardsfusion", color, "GoldBonus")
	gold, _ := GxDict.GetDictInt("Cardsfusion", color, "GoldNumber")
	cardNumber, _ := GxDict.GetDictInt("Cardsfusion", color, "CardNumberAdd")
	if req.GetUseGold() == 1 && role.Gold < int64(gold) {
		GxMisc.Warn("gold is not enough, role: %s, gold: %d, msg: %s", role.String(), role.Gold, req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
		return
	}

	//计算概率
	totalProbability := p + (int(req.GetCard().GetCnt())-2)*cardNumber
	if req.GetUseGold() == 1 {
		totalProbability += goldBonus
	}

	var rsp GxProto.FuseCardRsp
	rsp.Info = new(GxProto.RespondInfo)

	//扣钱
	DelItem(client, runInfo, -1, GxStatic.IDMoney, money, rsp.GetInfo())
	if req.GetUseGold() == 1 {
		DelItem(client, runInfo, -1, GxStatic.IDGold, gold, rsp.GetInfo())
	}

	//删卡片
	DelItem(client, runInfo, -1, int(req.GetCard().GetId()), int(req.GetCard().GetCnt()), rsp.GetInfo())

	//新卡牌
	r := GxMisc.GetRandom(10000)
	if r <= totalProbability {
		GxMisc.Info("Fuse card success, role: %s, random: %d, new-card: %d", role.String(), r, nextCardID)
		rsp.Success = proto.Int(1)
		GetItem(client, runInfo, nextCardID, 1, rsp.Info)
	} else {
		rsp.Success = proto.Int(0)

		//分解卡牌
		for i := 0; i < int(req.GetCard().GetCnt()); i++ {
			decomposeCard(runInfo, client, color, rsp.Info)
		}
		GxMisc.Info("Fuse card fail, role: %s, random: %d", role.String(), r)
	}

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//DecomposeCardCallback 分解卡片
func DecomposeCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.DecomposeCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get DecomposeCardReq request, role: %s, msg: %s", role.String(), req.String())
	if len(req.GetCard()) == 0 {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	money := 0
	for i := 0; i < len(req.GetCard()); i++ {
		id := int(req.GetCard()[i].GetId())
		cnt := int(req.GetCard()[i].GetCnt())
		if !GxDict.CheckDictId("HerosCard", id) {
			//卡牌不存在
			GxMisc.Warn("card is not exists, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardNotExists, nil)
			return
		}
		card := GxStatic.GetRoleCard(client, role.ID, id)
		if card == nil || card.Cnt < cnt {
			//卡牌不足
			GxMisc.Warn("card is not enough, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardNotEnough, nil)
			return
		}

		color, _ := GxDict.GetDictInt("HerosCard", id, "Color")
		iVal, _ := GxDict.GetDictInt("Cardsdecomposition", color, "Gold")
		money += iVal * cnt
	}

	if role.Money < int64(money) {
		GxMisc.Warn("Money is not enough, role: %s, Money: %d, need-money: %d, msg: %s", role.String(), role.Money, money, req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMoneyNotEnough, nil)
		return
	}

	var rsp GxProto.DecomposeCardRsp
	rsp.Info = new(GxProto.RespondInfo)

	//扣钱
	DelItem(client, runInfo, -1, GxStatic.IDMoney, money, rsp.GetInfo())

	for i := 0; i < len(req.GetCard()); i++ {
		id := int(req.GetCard()[i].GetId())
		cnt := int(req.GetCard()[i].GetCnt())
		//扣除道具
		DelItem(client, runInfo, -1, id, cnt, rsp.GetInfo())

		color, _ := GxDict.GetDictInt("HerosCard", id, "Color")
		decomposeCard(runInfo, client, color, rsp.GetInfo())
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//NewFightCardBagCallback 新建卡组
func NewFightCardBagCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.NewFightCardBagReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get NewFightCardBagReq request, role: %s, msg: %s", role.String(), req.String())
	if req.Name == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	//新增卡组
	index, ts := role.NewFightCardBag(client, req.GetName())
	if index == -1 {
		//卡包数量已经达到最大限制
		GxMisc.Warn("fight bag count is max, role: %s", role.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagMax, nil)
		return
	}

	//回复
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &GxProto.NewFightCardBagRsp{
		Id: proto.Int(index),
		Ts: proto.Int64(ts),
	})
}

//DelFightCardBagCallback 删除卡组
func DelFightCardBagCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.DelFightCardBagReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get DelFightCardBagReq request, role: %s, msg: %s", role.String(), req.String())

	//删除卡组
	if ok := role.DelFightCardBag(client, int(req.GetId())); !ok {
		//卡包不存在
		GxMisc.Warn("fight bag is not existst, role: %s, req: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagNotExists, nil)
		return
	}

	//回复
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}

//AddFightCardCallback 往卡组添加卡牌
func AddFightCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.AddFightCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get AddFightCardReq request, role: %s, msg: %s", role.String(), req.String())
	if len(req.GetCard()) == 0 {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	if req.GetCardId() >= 32 || role.FightCardDistri&(1<<uint32(req.GetCardId())) == 0 {
		//卡包不存在
		GxMisc.Warn("fight bag is not existst, role: %s, req: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagNotExists, nil)
		return
	}

	fightCardBag := GxStatic.NewFightCard(role.ID, int(req.GetCardId()), "")
	fightCardBag.Get4Redis(client)

	for i := 0; i < len(req.GetCard()); i++ {
		ID := int(req.GetCard()[i].GetId())
		cnt := int(req.GetCard()[i].GetCnt())
		count := cnt + fightCardBag.Cards[ID]

		card := GxStatic.GetRoleCard(client, role.ID, ID)
		if count > card.Cnt {
			//卡片不足
			GxMisc.Warn("card is not enough, role: %s, card: %v", role.String(), card)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagNotExists, nil)
			return
		}
	}

	for i := 0; i < len(req.GetCard()); i++ {
		ID := int(req.GetCard()[i].GetId())
		cnt := int(req.GetCard()[i].GetCnt())
		count := cnt + fightCardBag.Cards[ID]

		GxStatic.SetRoleFightCard(client, role.ID, int(req.GetCardId()), &GxStatic.Item{
			ID:  ID,
			Cnt: count,
		})
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}

//DelFightCardCallback 从卡组删除卡牌
func DelFightCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.DelFightCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get DelFightCardReq request, role: %s, msg: %s", role.String(), req.String())
	if len(req.GetCard()) == 0 {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	if req.GetCardId() >= 32 || role.FightCardDistri&(1<<uint32(req.GetCardId())) == 0 {
		//卡包不存在
		GxMisc.Warn("fight bag is not existst, role: %s, req: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagNotExists, nil)
		return
	}

	fightCardBag := GxStatic.NewFightCard(role.ID, int(req.GetCardId()), "")
	fightCardBag.Get4Redis(client)

	var updateCards []*GxStatic.Item
	var deleteCards []int
	for i := 0; i < len(req.GetCard()); i++ {
		id := int(req.GetCard()[i].GetId())
		cnt := int(req.GetCard()[i].GetCnt())
		if cnt > fightCardBag.Cards[id] {
			//卡片不足
			GxMisc.Warn("card is not enough, role: %s, card: %d:%d", role.String(), id, cnt)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFightBagNotExists, nil)
			return
		}
		fightCardBag.Cards[id] -= cnt
		if fightCardBag.Cards[id] == 0 {
			delete(fightCardBag.Cards, id)
			deleteCards = append(deleteCards, id)
		} else {
			updateCards = append(updateCards, &GxStatic.Item{
				ID:  id,
				Cnt: fightCardBag.Cards[id],
			})
		}
	}

	for i := 0; i < len(updateCards); i++ {
		GxStatic.SetRoleFightCard(client, role.ID, int(req.GetCardId()), updateCards[i])
	}
	for i := 0; i < len(deleteCards); i++ {
		GxStatic.DelRoleFightCard(client, role.ID, int(req.GetCardId()), deleteCards[i])
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, nil)
}
