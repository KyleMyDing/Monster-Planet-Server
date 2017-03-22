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

//UseItemCallback 使用物品
func UseItemCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.UseItemReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get UseItemReq request, role: %s, msg: %s", role.String(), req.String())
	if req.Cnt == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
	item := GxStatic.GetRoleItem4Bag(client, role.ID, int(req.GetIndx()))
	if item == nil || item.ID == 0 {
		//指定背包位置没有装备
		GxMisc.Warn("item is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotExists, nil)
		return
	}
	costCount := int(req.GetCnt())
	needCount, _ := GxDict.GetDictInt("Items", item.ID, "RequireNum")

	if item.Cnt < costCount || costCount < needCount {
		//道具不足
		GxMisc.Warn("item is not enough, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotEnough, nil)
		return
	}
	if !GxDict.CheckDictId("Items", item.ID) {
		//道具不存在
		GxMisc.Warn("item is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotExists, nil)
		return
	}

	dropTable := GxDict.GetItemDict4Drop(item.ID)
	count := len(dropTable)
	if count == 0 {
		//道具不能打开
		GxMisc.Warn("item can not use, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemCanotOpen, nil)
		return
	}
	if role.BagEmptyCount(client) < count {
		//背包空间不足
		GxMisc.Warn("bag zone is not enough, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetBagZoneIsNotEnough, nil)
		return
	}
	var rsp GxProto.UseItemRsp
	rsp.Info = new(GxProto.RespondInfo)

	//删除一个物品
	n := costCount / needCount
	actualCount := n * needCount
	DelItem(client, runInfo, int(req.GetIndx()), item.ID, actualCount, rsp.GetInfo())

	//增加一组物品
	for i := 0; i < n; i++ {
		items := GxDict.GetDrop4DropTable(dropTable)
		for i := 0; i < len(items); i++ {
			GetItem(client, runInfo, items[i].ID, items[i].Cnt, rsp.GetInfo())
		}
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//SellItemCallback 出售物品
func SellItemCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.SellItemReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get SellItemReq request, role: %s, msg: %s", role.String(), req.String())
	if req.Indx == nil || req.Cnt == nil {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}
	item := GxStatic.GetRoleItem4Bag(client, role.ID, int(req.GetIndx()))
	if item == nil || item.ID == 0 {
		//指定背包位置没有装备
		GxMisc.Warn("item is not exists, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotExists, nil)
		return
	}
	var price int
	if GxStatic.IsEquipment(item.ID) {
		price1, _ := GxDict.GetDictInt("Package", item.ID, "SellPrice")
		price = price1
		if err != nil {

		}
		if price == 0 {
			//物品不能出售
			GxMisc.Warn("equipment cannot been sell, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemCanotSell, nil)
			return
		}
		GxStatic.DelRoleItem4Bag(client, role.ID, int(req.GetIndx()))
	} else if GxStatic.IsItem(item.ID) {
		price, _ = GxDict.GetDictInt("Package", item.ID, "SellPrice")
		if price == 0 {
			//物品不能出售
			GxMisc.Warn("item cannot been sell, role: %s, msg: %s", role.String(), req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemCanotSell, nil)
			return
		}
		if item.Cnt < int(req.GetCnt()) {
			//物品不足
			GxMisc.Warn("item is not enough, role: %s, cnt: %d, msg: %s", role.String(), item.Cnt, req.String())
			sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetItemNotEnough, nil)
			return
		}

	} else {
		//物品不能出售
		GxMisc.Warn("item cannot been sell, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	var rsp GxProto.SellItemRsp
	rsp.Info = new(GxProto.RespondInfo)

	DelItem(client, runInfo, int(req.GetIndx()), item.ID, int(req.GetCnt()), rsp.GetInfo())
	GetItem(client, runInfo, GxStatic.IDMoney, price*int(req.GetCnt()), rsp.GetInfo())
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//OrderBagCallback 整理背包
func OrderBagCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var rsp GxProto.OrderBagRsp

	bag := GxStatic.NewBag(role.ID)
	bag.Get4Redis(client)

	swapCell := func(bag *GxStatic.Bag, i int, j int) {
		bi := (bag.Distri[i/64] & (1 << uint32(i%64))) != 0
		bj := (bag.Distri[j/64] & (1 << uint32(j%64))) != 0

		if bj != bi {
			bag.Distri[i/64] ^= 1 << uint32(i%64)
			bag.Distri[j/64] ^= 1 << uint32(j%64)
		}
		bag.Cells[i], bag.Cells[j] = bag.Cells[j], bag.Cells[i]
	}

	//a比b优先级高返回true
	compareCell := func(a, b string) bool {
		if a == "" || b == "" {
			return b == ""
		}

		js1, _ := GxMisc.BufToMsg([]byte(a))
		item1 := new(GxStatic.Item)
		GxMisc.JSONToStruct(js1, item1)

		js2, _ := GxMisc.BufToMsg([]byte(b))
		item2 := new(GxStatic.Item)
		GxMisc.JSONToStruct(js2, item2)
		//装备在道具前面
		aIsEqui := GxStatic.IsEquipment(item1.ID)
		bIsEqui := GxStatic.IsEquipment(item2.ID)
		if aIsEqui != bIsEqui {
			return aIsEqui
		}

		//品质
		dictName := "Items"
		if aIsEqui {
			dictName = "Package"
		}
		aColor, _ := GxDict.GetDictInt(dictName, item1.ID, "Color")

		bColor, _ := GxDict.GetDictInt("dictName", item2.ID, "Color")

		if aColor != bColor {
			return aColor >= bColor
		}

		//ID小的优先级高
		if item1.ID > item2.ID {
			return false
		}
		return true
	}

	//插入排序
	count := role.GetBagCellsCount(client)
	GxMisc.Trace("role[%s] order bag, count: %d", role.String(), count)
	for i := 1; i < count; i++ {
		for j := i; j > 0; j-- {
			if compareCell(bag.Cells[j-1], bag.Cells[j]) {
				break
			}
			swapCell(bag, j-1, j)
		}
	}

	//更新内存以及初始化响应内容
	rsp.Bag = new(GxProto.PbBagInfo)
	for i := 0; i < count; i++ {
		if (bag.Distri[i/64] & (1 << uint32(i%64))) == 0 {
			break
		}
		js, _ := GxMisc.BufToMsg([]byte(bag.Cells[i]))
		item := new(GxStatic.Item)
		GxMisc.JSONToStruct(js, item)

		info := new(GxProto.PbBagCellInfo)
		info.Indx = proto.Int(i)
		info.Item = new(GxProto.PbItemInfo)
		info.GetItem().Id = proto.Int(item.ID)
		info.GetItem().Cnt = proto.Int(item.Cnt)
		rsp.GetBag().Cells = append(rsp.GetBag().Cells, info)
	}
	bag.Del4Redis(client)
	bag.Set4Redis(client)

	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//ExpandBagCallback 扩展背包
func ExpandBagCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R

	bagNumUpper, _ := GxDict.GetDictInt("Playergrowup", role.GetLev(), "BagNumUpper")

	addGrid, _ := GxDict.GetDictInt("CommonConfig", role.GetLev(), "AddGrid")
	buyGridCost, _ := GxDict.GetDictString("CommonConfig", role.GetLev(), "BuyGridCost")

	//当前格子数
	count := GxStatic.GetRoleBuyCount4Bag(client, role.ID)
	//使用元宝购买次数
	BuyCount := GxStatic.GetRoleBuyCount4Bag(client, role.ID) / addGrid
	//消耗元宝列表
	buyGridCostArr := strings.Split(buyGridCost, ",")
	if (count+addGrid) > bagNumUpper || BuyCount >= len(buyGridCostArr) {
		//格子数限制
		GxMisc.Warn("bag cannot expand, role: %s", role.String())
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetBagCannotExpand, nil)
		return
	}
	gold, _ := strconv.Atoi(buyGridCostArr[BuyCount])
	if role.Gold < int64(gold) {
		//元宝不足
		GxMisc.Warn("gold is not enough, role: %s, gold: %d, need-gold: %d", role.String(), role.Gold, gold)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
		return
	}

	var rsp GxProto.ExpandBagRsp
	rsp.Info = new(GxProto.RespondInfo)

	//扣钱
	DelItem(client, runInfo, -1, GxStatic.IDGold, gold, rsp.GetInfo())

	count += addGrid
	GxStatic.SetRoleBuyCount4Bag(client, role.ID, count)

	rsp.BuyCount = proto.Int(count)
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}
