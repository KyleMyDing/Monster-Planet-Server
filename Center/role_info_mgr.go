package main

/**
作者： Kyle Ding
模块：角色信息管理模块
说明：
创建时间：2015-11-2
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//FillSelectRoleRsp 填充SelectRoleRsp消息
func FillSelectRoleRsp(client *redis.Client, roleID int, rsp *GxProto.SelectRoleRsp) (string, error) {
	rsp.Role = new(GxProto.RoleCommonInfo)
	playername, err := FillRoleCommonInfo(client, roleID, rsp.GetRole(), true)
	if err != nil {
		return "", err
	}

	rsp.Bag = new(GxProto.PbBagInfo)
	err = FillBagInfo(client, roleID, rsp.Bag, true)
	if err != nil {
		return "", err
	}

	rsp.CardBag = new(GxProto.PbCardBagInfo)
	err = FillCardBagInfo(client, roleID, rsp.CardBag, true)
	if err != nil {
		return "", err
	}

	rsp.FightCardBag = new(GxProto.PbFightCardBagInfo)
	err = FillFightCardBagInfo(client, roleID, int(rsp.GetRole().GetFightCardDistri()), rsp.FightCardBag, true)
	if err != nil {
		return "", err
	}

	rsp.Other = new(GxProto.RoleOtherInfo)
	chechRedis(client, roleID, rsp.Other)

	return playername, nil
}

//FillCreateRoleRsp 填充CreateRoleRsp消息
func FillCreateRoleRsp(client *redis.Client, roleID int, rsp *GxProto.CreateRoleRsp) error {
	rsp.Role = new(GxProto.RoleCommonInfo)
	_, err := FillRoleCommonInfo(client, roleID, rsp.GetRole(), true)
	if err != nil {
		return err
	}

	rsp.Bag = new(GxProto.PbBagInfo)
	err = FillBagInfo(client, roleID, rsp.Bag, true)
	if err != nil {
		return err
	}

	rsp.CardBag = new(GxProto.PbCardBagInfo)
	err = FillCardBagInfo(client, roleID, rsp.CardBag, true)
	if err != nil {
		return err
	}

	rsp.FightCardBag = new(GxProto.PbFightCardBagInfo)
	err = FillFightCardBagInfo(client, roleID, int(rsp.GetRole().GetFightCardDistri()), rsp.FightCardBag, true)
	if err != nil {
		return err
	}

	rsp.Other = new(GxProto.RoleOtherInfo)
	chechRedis(client, roleID, rsp.Other)

	return nil
}

//FillItem 填充PbItemInfo
func FillItem(roleID int, str string, item **GxProto.PbItemInfo) {
	if str == "" {
		return
	}
	j, err := GxMisc.BufToMsg([]byte(str))
	if err != nil {
		GxMisc.Error("role equi format error, role: %d, str: %s", roleID, str)
		return
	}
	if *item == nil {
		*item = new(GxProto.PbItemInfo)
	}
	info := new(GxStatic.Item)
	GxMisc.JSONToStruct(j, info)
	(*item).Id = proto.Int(info.ID)
	(*item).Cnt = proto.Int(info.Cnt)
}

//FillRoleFightInfo 填充战斗相关信息
func FillRoleFightInfo(client *redis.Client, roleID int, role *GxProto.RoleCommonInfo, save bool) (string, error) {
	r := new(GxStatic.Role)
	r.ID = roleID
	err := r.Get4Redis(client)
	if err != nil {
		err = r.Get4Mysql()
		if err != nil {
			return "", err
		}

		//玩家自己拉去自己信息的时候需要保存到数据库，管理后台的话就不需要
		if save {
			r.Set4Redis(client)
		}
	}

	role.Id = proto.Int(r.ID)
	role.Name = proto.String(r.Name)
	role.Sex = proto.Int(r.Sex)
	role.VocationId = proto.Int(r.VocationID)
	role.Vip = proto.Int(r.Vip)
	role.Expr = proto.Int64(r.Expr)

	FillItem(roleID, r.Wings, &role.Wings)
	FillItem(roleID, r.Weapon, &role.Weapon)
	FillItem(roleID, r.Coat, &role.Coat)
	FillItem(roleID, r.Cloakd, &role.Cloakd)
	FillItem(roleID, r.Trouser, &role.Trouser)
	FillItem(roleID, r.Armguard, &role.Armguard)
	FillItem(roleID, r.Shoes, &role.Shoes)

	return r.PlayerName, nil
}

//FillRoleCommonInfo 填充RoleCommonInfo
func FillRoleCommonInfo(client *redis.Client, roleID int, role *GxProto.RoleCommonInfo, save bool) (string, error) {
	r := new(GxStatic.Role)
	r.ID = roleID
	err := r.Get4Redis(client)
	if err != nil {
		err = r.Get4Mysql()
		if err != nil {
			return "", err
		}

		//玩家自己拉去自己信息的时候需要保存到数据库，管理后台的话就不需要
		if save {
			r.Set4Redis(client)
		}
	}

	role.Id = proto.Int(r.ID)
	role.Name = proto.String(r.Name)
	role.Sex = proto.Int(r.Sex)
	role.VocationId = proto.Int(r.VocationID)
	role.Vip = proto.Int(r.Vip)
	role.Expr = proto.Int64(r.Expr)
	role.Money = proto.Int64(r.Money)
	role.Gold = proto.Int64(r.Gold)
	role.Power = proto.Int64(r.Power)
	role.Powerts = proto.Int64(r.PowerTs)
	role.Honor = proto.Int64(r.Honor)
	role.ArmyPay = proto.Int64(r.ArmyPay)

	FillItem(roleID, r.Wings, &role.Wings)
	FillItem(roleID, r.Weapon, &role.Weapon)
	FillItem(roleID, r.Coat, &role.Coat)
	FillItem(roleID, r.Cloakd, &role.Cloakd)
	FillItem(roleID, r.Trouser, &role.Trouser)
	FillItem(roleID, r.Armguard, &role.Armguard)
	FillItem(roleID, r.Shoes, &role.Shoes)

	role.FightCardDistri = proto.Int(r.FightCardDistri)
	role.BuyPowerCnt = proto.Int(r.BuyPowerCnt)
	role.BuyPowerTs = proto.Int64(r.BuyPowerTs)
	role.GetFreePowerTs = proto.Int64(r.GetFreePowerTs)
	return r.PlayerName, nil
}

//FillBagInfo 填充PbBagInfo,save为true则保存到数据库
func FillBagInfo(client *redis.Client, roleID int, bag *GxProto.PbBagInfo, save bool) error {
	b := GxStatic.NewBag(roleID)
	err := b.Get4Redis(client)
	if err != nil {
		//缓存没有，则从mysql中读取
		err = b.Get4Mysql()
		if err != nil {
			return nil
		}
		if save {
			b.Set4Redis(client)
		}
	}
	bag.BuyCount = proto.Int(b.BuyCount)
	for i := 0; i < 256; i++ {
		if b.Distri[i/64]&(1<<uint32(i%64)) == 0 {
			continue
		}
		info := new(GxProto.PbBagCellInfo)
		info.Indx = proto.Int(i)
		info.Item = new(GxProto.PbItemInfo)
		FillItem(roleID, b.Cells[i], &info.Item)

		bag.Cells = append(bag.Cells, info)
	}
	return nil
}

//FillCardBagInfo 填充PbCardBagInfo,save为true则保存到数据库
func FillCardBagInfo(client *redis.Client, roleID int, cardBag *GxProto.PbCardBagInfo, save bool) error {
	c := GxStatic.NewCard(roleID)
	err := c.Get4Redis(client)
	if err != nil {
		//缓存没有，则从mysql中读取
		err = c.Get4Mysql()
		if err != nil {
			return nil
		}
		if save {
			c.Set4Redis(client)
		}
	}

	for cardID, cnt := range c.Cards {
		cardBag.Cards = append(cardBag.Cards, &GxProto.PbItemInfo{
			Id:  proto.Int(cardID),
			Cnt: proto.Int(cnt),
		})
	}

	return nil
}

//FillFightCardBagInfo 填充PbFightCardBagInfo,save为true则保存到数据库
func FillFightCardBagInfo(client *redis.Client, roleID int, distri int, cardBag *GxProto.PbFightCardBagInfo, save bool) error {

	for i := 0; i < 32; i++ {
		if distri&(1<<uint32(i)) == 0 {
			continue
		}

		c := GxStatic.NewFightCard(roleID, i, "")
		err := c.Get4Redis(client)
		if err != nil {
			//缓存没有，则从mysql中读取
			err = c.Get4Mysql()
			if err != nil {
				continue
			}
			if save {
				c.Set4Redis(client)
			}
		}
		info := new(GxProto.PbCardBagInfo)
		info.Indx = proto.Int(i)
		info.Name = proto.String(c.BagName)
		for cardID, cnt := range c.Cards {
			info.Cards = append(info.Cards, &GxProto.PbItemInfo{
				Id:  proto.Int(cardID),
				Cnt: proto.Int(cnt),
			})
		}
		cardBag.Cards = append(cardBag.Cards, info)
	}
	return nil
}

//chechRedis 登录时候检查是否都已经缓存，没有缓存则从mysql中读取
func chechRedis(client *redis.Client, roleID int, other *GxProto.RoleOtherInfo) {
	//buycard
	b := GxStatic.NewBuyCard(roleID)
	if b.Get4Redis(client) != nil {
		b.Get4Mysql()
		b.Set4Redis(client)
	}

	//chapter
	c := GxStatic.NewChapter(roleID)
	if c.Get4Redis(client) != nil {
		c.Get4Mysql()
		c.Set4Redis(client)
	}
	other.ChapterId = proto.Int(c.NowChapterId)
	other.LevelId = proto.Int(c.NowLevelId)
	other.PointId = proto.Int(c.NowPointId)

	//mail
	count := 0
	ids := GxStatic.GetRoleAllMailIds(client, roleID)
	for i := 0; i < len(ids); i++ {
		mail := &GxStatic.Mail{
			RoleID: roleID,
			ID:     ids[i],
		}
		if mail.Get4Redis(client) == nil {
			continue
		}
		mail.Get4Mysql()
		mail.Set4Redis(client)

		//{}
		if mail.Rd == 0 {
			count++
		}
	}
	other.UnreadMailCmt = proto.Int(count)
}
