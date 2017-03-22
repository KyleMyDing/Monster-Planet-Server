package GxStatic

/**
作者： Kyle Ding
模块：战斗卡组信息模块
说明：
创建时间：2015-11-5
**/

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"gopkg.in/redis.v3"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//FightCard 卡组信息
type FightCard struct {
	RoleID   int    `pk:"true"` //角色ID
	BagID    int    `pk:"true"` //卡组ID
	BagName  string //卡组名
	CreateTs int64  //创建时间
	Info     string `type:"text"` //装备实例ID，道具的话为0

	Cards map[int]int `ignore:"true"`
}

//NewFightCard 生成一个新卡组
func NewFightCard(roleID int, BagID int, bagName string) *FightCard {
	return &FightCard{
		RoleID:  roleID,
		BagID:   BagID,
		BagName: bagName,
		Cards:   make(map[int]int),
	}
}

//Set4Redis ...
func (card *FightCard) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID) + ":" + strconv.Itoa(card.BagID)

	client.HSet(tableName, "BagName", card.BagName)
	for k, v := range card.Cards {
		client.HSet(tableName, fmt.Sprintf("%d", k), fmt.Sprintf("%d", v))
	}
	return nil
}

//Get4Redis ...
func (card *FightCard) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID) + ":" + strconv.Itoa(card.BagID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return nil
	}
	for k, v := range m {
		if k == "BagName" {
			card.BagName = v
			continue
		}
		cardID, _ := strconv.Atoi(k)
		cnt, _ := strconv.Atoi(v)
		card.Cards[cardID] = cnt
	}
	return nil

}

//Del4Redis ...
func (card *FightCard) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID) + ":" + strconv.Itoa(card.BagID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}
	client.Del(tableName)
	return nil
}

//Set4Mysql ...
func (card *FightCard) Set4Mysql() error {
	if card.Info == "" {
		j := simplejson.New()

		for k, v := range card.Cards {
			j.Set(fmt.Sprintf("%d", k), v)
		}
		buf, _ := j.MarshalJSON()
		card.Info = string(buf)
	}

	c := NewFightCard(card.RoleID, card.BagID, "")
	var str string
	if c.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(card, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(card, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (card *FightCard) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(card, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&card.BagName, &card.CreateTs, &card.Info)

		if card.Info == "" {
			return nil
		}
		j, _ := simplejson.NewJson([]byte(card.Info))
		m, _ := j.Map()
		for k, v := range m {
			cardID, _ := strconv.Atoi(k)
			cnt, _ := v.(json.Number).Int64()
			card.Cards[cardID] = int(cnt)
		}

		return nil
	}
	return errors.New("null")
}

//Del4Mysql ...
func (card *FightCard) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(card, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//GetRoleFightCard 返回角色指定卡组指定卡牌
func GetRoleFightCard(client *redis.Client, roleID int, bagID int, cardID int) *Item {
	tableName := "h_fight_card:" + strconv.Itoa(roleID) + ":" + strconv.Itoa(bagID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	i64, _ := client.HGet(tableName, fmt.Sprintf("%d", cardID)).Int64()
	item := &(Item{
		ID:  cardID,
		Cnt: int(i64),
	})

	return item
}

//SetRoleFightCard 保存角色指定卡组指定卡牌
func SetRoleFightCard(client *redis.Client, roleID int, bagID int, item *Item) {
	tableName := "h_fight_card:" + strconv.Itoa(roleID) + ":" + strconv.Itoa(bagID)
	client.HSet(tableName, fmt.Sprintf("%d", item.ID), fmt.Sprintf("%d", item.Cnt))
}

//DelRoleFightCard 删除角色指定卡组指定卡牌
func DelRoleFightCard(client *redis.Client, roleID int, bagID int, cardID int) {
	tableName := "h_fight_card:" + strconv.Itoa(roleID) + ":" + strconv.Itoa(bagID)
	client.HDel(tableName, fmt.Sprintf("%d", cardID))
}
