package main

/**
作者： Kyle Ding
模块：抽卡信息管理模块
说明：处理玩家抽卡行为的请求
创建时间：2015-12-7
**/
import (
	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
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

//ExtractCardCallback 抽取英雄卡牌回调函数
func ExtractCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ExtractCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ExtractCardReq request, role: %s, msg: %s", role.String(), req.String())

	if !GxDict.CheckDictId("BuyCardCost", int(req.GetExtractCardType())) {
		//判断抽卡请求的类型是否在合理的范围内
		GxMisc.Warn("ExtractCardType is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetExtractCardType, nil)
		return
	}

	// 免费的抽卡倒计时,如果等于-1，表示不免费,
	freeInterval, _ := GxDict.GetDictInt("BuyCardCost", int(req.GetExtractCardType()), "FreeInterval")

	//now time
	now := time.Now().Unix()
	//每天免费抽奖的上限,如果等于-1，表示不免费,
	freeUpper, _ := GxDict.GetDictInt("BuyCardCost", int(req.GetExtractCardType()), "FreeUpper")

	//从Redis上获取当前的BuyCardInfoType结构
	buyCard := GxStatic.NewBuyCard(role.ID)
	ec, _ := buyCard.GetBuyCardInfoType4Redis(client, int(req.GetExtractCardType()), role.ID)

	if ec == "" {
		buf, _ := GxMisc.MsgToBuf(GxStatic.BuyCardInfoType{time.Now().Unix() - int64(500000), 0})
		ec = string(buf)
	}
	j, _ := simplejson.NewJson([]byte(ec))
	item := new(GxStatic.BuyCardInfoType)
	GxMisc.JSONToStruct(j, item)

	var rsp GxProto.ExtractCardRsp
	rsp.Info = new(GxProto.RespondInfo)

	//获取一天的秒数
	todayBeginTime := GxMisc.NextTime(0, 0, 0) - 24*60*60
	if item.Ts < todayBeginTime {
		item.FreeCnt = 0
	}

	if freeInterval == -1 || int(now) < freeInterval+int(item.Ts) || (freeUpper != -1 && item.FreeCnt >= freeUpper) {
		//判断需要花费金币或者元宝
		//如果抽卡类型不免费,还没有到达免费抽卡的时间,免费抽卡次数到达当天的上限[根据配置表，如果前面的要求都没有达到，要对免费的天数是否为-1进行 判断]
		str, _ := GxDict.GetDictString("BuyCardCost", int(req.GetExtractCardType()), "Cost")
		arr := strings.Split(str, ",")

		costID, _ := strconv.Atoi(arr[0])
		costCnt, _ := strconv.Atoi(arr[1])

		if costID == GxStatic.IDGold && role.Gold >= int64(costCnt) {
			//判断金币或者元宝是否足够
			// 扣除相应的元宝

			DelItem(client, runInfo, -1, GxStatic.IDGold, costCnt, rsp.GetInfo())
			//增加花费的元宝数量
			buyCard.SetRoleBuyCardCostGoldNum(client, role.ID, costCnt)
			//随机获取相应数量卡牌
			getCard(client, runInfo, rsp.GetInfo(), &req)

		} else if costID == GxStatic.IDMoney && role.Money >= int64(costCnt) {
			// 扣除相应的金币
			DelItem(client, runInfo, -1, GxStatic.IDMoney, costCnt, rsp.GetInfo())

			//随机获取相应数量卡牌
			getCard(client, runInfo, rsp.GetInfo(), &req)
		} else if costID == GxStatic.IDGold {
			//提示元宝不够
			GxMisc.Warn("Gold is not enough, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
			return
		} else {
			//提示金币不够
			GxMisc.Warn("Money is not enough, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMoneyNotEnough, nil)
			return
		}

	} else { //进行免费抽奖
		getCard(client, runInfo, rsp.GetInfo(), &req)
		//freeCnt := item.FreeCnt + 1
		(item.FreeCnt)++
		buyCard.SetBuyCardInfoType4Redis(client, int(*req.ExtractCardType), &GxStatic.BuyCardInfoType{now, item.FreeCnt})
	}
	//判断抽卡请求的类型是否需要进行时间检测,判断是否还有免费的次数,判断是否到达免费抽奖的时间=> 随机获取一张卡牌,更新抽奖的信息,更新卡组的信息,通知客户端
	//判断金币或者元宝是否足够=> 扣除相应的金币或者元宝=> 随机获取相应数量卡牌,更新抽奖的信息,更新卡组的信息,通知客户端
	//提示金币或者元宝不够
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	//GxMisc.Debug("rsp: %s", rsp.String())
}

//getCard   获取相应数量英雄卡牌
func getCard(client *redis.Client, runInfo *RoleRunInfo, rspInfo *GxProto.RespondInfo, req *GxProto.ExtractCardReq) {
	str, _ := GxDict.GetDictString("BuyCardCost", int(req.GetExtractCardType()), "CommonBuySign")

	ExtractCardTimes, _ := GxDict.GetDictInt("BuyCardCost", int(req.GetExtractCardType()), "CommonNum")

	for i := 0; i < ExtractCardTimes; i++ {
		key := randExtractCard(str)
		if key == -1 {
			continue
		}
		GxMisc.Debug("key: %d", key)
		heroCard := GxDict.Random4BuyCardArray(key, runInfo.R.GetLev())
		GetItem(client, runInfo, heroCard, 1, rspInfo)
	}
	//必得的卡牌
	mustNum, _ := GxDict.GetDictInt("BuyCardCost", int(req.GetExtractCardType()), "MustNum")

	if mustNum != -1 {
		ExtractCardMustGet, _ := GxDict.GetDictString("BuyCardCost", int(req.GetExtractCardType()), "MustGetSign")

		for i := 0; i < mustNum; i++ {
			key := randExtractCard(ExtractCardMustGet)
			if key == -1 {
				continue
			}
			heroCard := GxDict.Random4BuyCardArray(key, runInfo.R.GetLev())
			GetItem(client, runInfo, heroCard, 1, rspInfo)
		}
	}
}

//randExtractCard   随机获取某个类型的英雄卡片
func randExtractCard(str string) int {
	rand := GxMisc.GetRandomInterval(1, 10000)
	ptrDict, _ := GxDict.GetDict("BuyCard")
	for k, v := range *ptrDict {
		n := v[str].(int)
		if n >= rand {
			return k
		}
		rand -= n
	}
	return -1
}

//ReceiveCardCallback  领取抽卡奖励回调函数
func ReceiveCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ReceiveCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ExtractCardReq request, role: %s, msg: %s", role.String(), req.String())

	var rsp GxProto.ExtractCardRsp
	rsp.Info = new(GxProto.RespondInfo)

	buyCard := GxStatic.NewBuyCard(role.ID)
	flag, getReceiveCardErr := buyCard.GetReceiveCard4Redis(client, int(req.GetReceiveCardType()), role.ID)
	if getReceiveCardErr != nil {
		GxMisc.Warn("GetReceiveCard4Redis is error", getReceiveCardErr)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	var costGoldNum int
	if !GxDict.CheckDictId("BuyCardSpecial", int(req.GetReceiveCardType())) {
		costGoldNum, _ = GxDict.GetDictInt("BuyCardSpecial", int(req.GetReceiveCardType()), "CostNum")
	}

	//判断花费的元宝数量是否足够 并且 没有拿过该类型的奖励
	if buyCard.CostGoldNum >= costGoldNum && flag == 0 {
		cardID, _ := GxDict.GetDictInt("BuyCardSpecail", int(req.GetReceiveCardType()), "CardID")
		cardCnt, _ := GxDict.GetDictInt("BuyCardSpecial", int(req.GetReceiveCardType()), "Num")

		GetItem(client, runInfo, cardID, cardCnt, rsp.GetInfo())
		flag = 1
		err := buyCard.SetIsReceiveCard4Redis(client, int(req.GetReceiveCardType()), flag, role.ID)
		if err != nil {
			GxMisc.Warn("SetIsReceiveCard4Redis error, err: %s", err)
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
			return
		}
	} else if buyCard.CostGoldNum >= costGoldNum { //花费的元宝数量足够,但没有拿到奖励,说明已经拿过奖励了
		GxMisc.Warn("RetReceiveCard is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
		return
	} else { //花费的元宝数量不够
		GxMisc.Warn("costGold is not enough, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetCardNotExists, nil)
		return
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//ExtractSkillCardCallback  抽取技能卡牌回调函数
func ExtractSkillCardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ExtractSkillCardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ExtractSkillCardReq request, role: %s, msg: %s", role.String(), req.String())

	if !GxDict.CheckDictId("BuySkillCardCost", int(req.GetExtractSkillCardType())) {
		//判断抽卡请求的类型是否在合理的范围内
		GxMisc.Warn("ExtractSkillCardType is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetExtractSkillCardType, nil)
		return
	}

	// 免费的抽卡倒计时,如果等于-1，表示不免费,
	freeInterval, _ := GxDict.GetDictInt("BuySkillCardCost", int(req.GetExtractSkillCardType()), "FreeInterval")

	GxMisc.Debug("freeInterval: %d", freeInterval)
	//now time
	now := time.Now().Unix()
	//每天免费抽奖的上限,如果等于-1，表示不免费,
	freeUpper, _ := GxDict.GetDictInt("BuySkillCardCost", int(req.GetExtractSkillCardType()), "FreeUpper")

	//从Redis上获取当前的BuyCardInfoType结构

	buyCard := GxStatic.NewBuyCard(role.ID)
	ec, getErr := buyCard.GetBuyCardInfoType4Redis(client, int(req.GetExtractSkillCardType()), role.ID)
	if getErr != nil {
		GxMisc.Warn("msg getBuyCardInfoErr error, get4RedisError: %s", getErr)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	if ec == "" {
		buf, _ := GxMisc.MsgToBuf(GxStatic.BuyCardInfoType{time.Now().Unix() - int64(500000), 0})
		ec = string(buf)
	}
	j, _ := simplejson.NewJson([]byte(ec))
	item := new(GxStatic.BuyCardInfoType)
	GxMisc.JSONToStruct(j, item)

	var rsp GxProto.ExtractSkillCardRsp
	rsp.Info = new(GxProto.RespondInfo)

	//获取一天的秒数
	todayBeginTime := GxMisc.NextTime(0, 0, 0) - 24*60*60
	if item.Ts < todayBeginTime {
		item.FreeCnt = 0
	}

	if freeInterval == -1 || int(now) < freeInterval+int(item.Ts) || (freeUpper != -1 && item.FreeCnt >= freeUpper) {

		//判断需要花费金币或者元宝
		//如果抽卡类型不免费,还没有到达免费抽卡的时间,免费抽卡次数到达当天的上限[根据配置表，如果前面的要求都没有达到，要对免费的天数是否为-1进行 判断]
		strTmp, _ := GxDict.GetDictString("BuySkillCardCost", int(req.GetExtractSkillCardType()), "Cost")

		arr := strings.Split(strTmp, ",")

		costID, _ := strconv.Atoi(arr[0])
		costCnt, _ := strconv.Atoi(arr[1])

		if costID == GxStatic.IDGold && role.Gold >= int64(costCnt) {
			//判断金币或者元宝是否足够
			// 扣除相应的元宝
			DelItem(client, runInfo, -1, GxStatic.IDGold, costCnt, rsp.GetInfo())

			//随机获取相应数量卡牌
			getSkillCard(client, runInfo, rsp.GetInfo(), &req)

		} else if costID == GxStatic.IDMoney && role.Money >= int64(costCnt) {
			// 扣除相应的金币
			DelItem(client, runInfo, -1, GxStatic.IDMoney, costCnt, rsp.GetInfo())

			//随机获取相应数量卡牌
			getSkillCard(client, runInfo, rsp.GetInfo(), &req)
		} else if costID == GxStatic.IDGold {
			//提示元宝不够
			GxMisc.Warn("Gold is not enough, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
			return
		} else {
			//提示金币不够
			GxMisc.Warn("Money is not enough, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMoneyNotEnough, nil)
			return
		}

	} else { //进行免费抽奖
		getSkillCard(client, runInfo, rsp.GetInfo(), &req)
		(item.FreeCnt)++
		buyCard.SetBuyCardInfoType4Redis(client, int(req.GetExtractSkillCardType()), &GxStatic.BuyCardInfoType{now, item.FreeCnt})
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	// GxMisc.Debug("rsp: %s", rsp.String())
}

//getSkillCard   获取相应数量技能卡牌
func getSkillCard(client *redis.Client, runInfo *RoleRunInfo, rspInfo *GxProto.RespondInfo, req *GxProto.ExtractSkillCardReq) {
	//str := (*GxDict.Dicts)["BuySkillCardCost"].D[int(req.GetExtractSkillCardType())]["CommonBuySign"].(string)
	str, _ := GxDict.GetDictString("BuySkillCardCost", int(req.GetExtractSkillCardType()), "CommonBuySign")

	//ExtractCardTimes := (*GxDict.Dicts)["BuySkillCardCost"].D[int(req.GetExtractSkillCardType())]["CommonNum"].(int)
	ExtractCardTimes, _ := GxDict.GetDictInt("BuySkillCardCost", int(req.GetExtractSkillCardType()), "CommonNum")

	for i := 0; i < ExtractCardTimes; i++ {
		key := randExtractSkillCard(str)
		if key == -1 {
			continue
		}
		GxMisc.Debug("key: %d", key)
		skillCard := GxDict.Random4BuySkillCardArray(key)
		GetItem(client, runInfo, skillCard, 1, rspInfo)
	}
	//必得的卡牌
	mustNum, _ := GxDict.GetDictInt("BuySkillCardCost", int(req.GetExtractSkillCardType()), "MustNum")

	if mustNum != -1 {
		ExtractCardMustGet, _ := GxDict.GetDictString("BuySkillCardCost", int(req.GetExtractSkillCardType()), "MustGetSign")

		for i := 0; i < mustNum; i++ {
			key := randExtractSkillCard(ExtractCardMustGet)
			if key == -1 {
				continue
			}
			skillCard := GxDict.Random4BuySkillCardArray(key)
			GetItem(client, runInfo, skillCard, 1, rspInfo)
		}
	}
}

//randExtractSkillCard 随机获取某个类型的技能卡片
func randExtractSkillCard(str string) int {
	rand := GxMisc.GetRandomInterval(1, 10000)

	ptrDict, _ := GxDict.GetDict("BuySkillCard")
	for k, v := range *ptrDict {
		n := v[str].(int)
		if n >= rand {
			return k
		}
		rand -= n
	}
	return -1
}

//GetBuyCardInfoCallback   获取抽卡保存的信息，包括上次抽取的时间和抽武将花费的元宝数
func GetBuyCardInfoCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var rsp GxProto.GetBuyCardCardInfoRsp

	buyCard := GxStatic.NewBuyCard(role.ID)
	err := buyCard.Get4Redis(client)
	if err != nil {
		GxMisc.Warn("msg get4Redis error, get4RedisError: %s", err)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	rsp.CostGoldNum = proto.Int(buyCard.CostGoldNum)

	for k, v := range buyCard.ExtractCard {

		j, _ := simplejson.NewJson([]byte(v))
		item := new(GxStatic.BuyCardInfoType)
		GxMisc.JSONToStruct(j, item)
		rsp.BuyCardInfo = append(rsp.BuyCardInfo, &GxProto.PbBuyCardInfo{
			ExtractCardType: proto.Int(k),
			FreeCnt:         proto.Int(item.FreeCnt),
			Ts:              proto.Int64(item.Ts),
		})

	}

	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}
