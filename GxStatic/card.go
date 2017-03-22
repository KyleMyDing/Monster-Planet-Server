package GxStatic

/**
作者： Kyle Ding
模块：卡包信息模块
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

//Card 卡包信息
type Card struct {
	RoleID int    `pk:"true"`   //角色ID
	Info   string `type:"text"` //

	Cards map[int]int `ignore:"true"`
}

//NewCard 生成一个新卡包
func NewCard(roleID int) *Card {
	return &Card{
		RoleID: roleID,
		Cards:  make(map[int]int),
	}
}

//Set4Redis ...
func (card *Card) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID)

	for k, v := range card.Cards {
		client.HSet(tableName, fmt.Sprintf("%d", k), fmt.Sprintf("%d", v))
	}
	return nil
}

//Get4Redis ...
func (card *Card) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return nil
	}
	for k, v := range m {
		cardID, _ := strconv.Atoi(k)
		cnt, _ := strconv.Atoi(v)
		card.Cards[cardID] = cnt
	}
	return nil

}

//Del4Redis ...
func (card *Card) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(card) + ":" + strconv.Itoa(card.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}
	client.Del(tableName)
	return nil
}

//Set4Mysql ...
func (card *Card) Set4Mysql() error {
	if card.Info == "" {
		j := simplejson.New()

		for k, v := range card.Cards {
			j.Set(fmt.Sprintf("%d", k), v)
		}
		buf, _ := j.MarshalJSON()
		card.Info = string(buf)
	}

	c := NewCard(card.RoleID)
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
func (card *Card) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(card, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&card.Info)

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
func (card *Card) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(card, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//GetRoleCard 返回角色指定卡片数量
func GetRoleCard(client *redis.Client, roleID int, cardID int) *Item {
	tableName := "h_card:" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	i64, _ := client.HGet(tableName, fmt.Sprintf("%d", cardID)).Int64()
	return &(Item{
		ID:  cardID,
		Cnt: int(i64),
	})
}

//SetRoleCard 保存角色指定卡牌信息
func SetRoleCard(client *redis.Client, roleID int, item *Item) {
	tableName := "h_card:" + strconv.Itoa(roleID)
	client.HSet(tableName, fmt.Sprintf("%d", item.ID), fmt.Sprintf("%d", item.Cnt))
}

//DelRoleCard 删除角色指定卡片
func DelRoleCard(client *redis.Client, roleID int, cardID int) {
	tableName := "h_card:" + strconv.Itoa(roleID)
	client.HDel(tableName, fmt.Sprintf("%d", cardID))
}
