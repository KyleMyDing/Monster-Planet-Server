/**
作者： Kyle Ding
模块：背包信息模块
说明：
创建时间：2015-11-5
**/
package GxStatic

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"gopkg.in/redis.v3"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//Item 道具信息
type Item struct {
	ID  int //类型ID
	Cnt int //装备-等级，道具-数量
}

//Bag 背包信息
type Bag struct {
	RoleID int    `pk:"true"`   //角色ID
	Info   string `type:"text"` //装备实例ID，道具的话为0

	Distri   [4]int64       `ignore:"true"` //最大256个格子
	BuyCount int            `ignore:"true"`
	Cells    map[int]string `ignore:"true"`
}

//NewBag 生成一个新背包
func NewBag(roleID int) *Bag {
	return &Bag{
		RoleID:   roleID,
		Cells:    make(map[int]string),
		BuyCount: 0,
	}
}

//Set4Redis ...
func (bag *Bag) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bag) + ":" + strconv.Itoa(bag.RoleID)
	client.HSet(tableName, "BuyCount", fmt.Sprintf("%d", bag.BuyCount))

	for i := 0; i < len(bag.Distri); i++ {
		client.HSet(tableName, fmt.Sprintf("Distri%d", i), fmt.Sprintf("%d", bag.Distri[i]))
	}

	for k, v := range bag.Cells {
		if v == "" {
			continue
		}
		client.HSet(tableName, fmt.Sprintf("%d", k), v)
	}
	return nil
}

//Get4Redis ...
func (bag *Bag) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bag) + ":" + strconv.Itoa(bag.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}
	n, _ := client.HGet(tableName, "BuyCount").Int64()
	bag.BuyCount = int(n)
	for i := 0; i < len(bag.Distri); i++ {
		bag.Distri[i], _ = client.HGet(tableName, fmt.Sprintf("Distri%d", i)).Int64()
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return errors.New("table is nil")
	}

	// for k, v := range m {
	// 	if k == "BuyCount" {
	// 		bag.BuyCount, _ = strconv.Atoi(v)
	// 		continue
	// 	}
	// 	for i := 0; i < len(bag.Distri); i++ {
	// 		if k == fmt.Sprintf("Distri%d", i) {
	// 			bag.Distri[i], _ = strconv.ParseInt(v, 10, 64)
	// 			break
	// 		}
	// 	}
	// }
	for k, v := range m {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if bag.Distri[index/64]&(1<<uint32(index%64)) == 0 {
			continue
		}
		bag.Cells[index] = v
	}
	return nil
}

//Del4Redis ...
func (bag *Bag) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bag) + ":" + strconv.Itoa(bag.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}

	client.Del(tableName)
	return nil
}

//Set4Mysql ...
func (bag *Bag) Set4Mysql() error {
	if bag.Info == "" {
		j := simplejson.New()
		j.Set("BuyCount", bag.BuyCount)
		for i := 0; i < len(bag.Distri); i++ {
			j.Set(fmt.Sprintf("Distri%d", i), bag.Distri[i])
		}
		for k, v := range bag.Cells {
			j1, _ := simplejson.NewJson([]byte(v))
			j.Set(fmt.Sprintf("%d", k), j1)
		}
		buf, _ := j.MarshalJSON()
		bag.Info = string(buf)
	}

	b := NewBag(bag.RoleID)
	var str string
	if b.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(bag, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(bag, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (bag *Bag) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(bag, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&bag.Info)

		if bag.Info == "" {
			return nil
		}
		j, _ := simplejson.NewJson([]byte(bag.Info))
		bag.BuyCount, _ = j.Get("BuyCount").Int()
		for i := 0; i < len(bag.Distri); i++ {
			bag.Distri[i], _ = j.Get(fmt.Sprintf("Distri%d", i)).Int64()
		}
		for i := 0; i < 256; i++ {
			if bag.Distri[i/64]&(1<<uint32(i%64)) == 0 {
				continue
			}
			buf, _ := j.Get(fmt.Sprintf("%d", i)).Encode()
			bag.Cells[i] = string(buf)
		}
		return nil
	}
	return errors.New("null")
}

//Del4Mysql ...
func (bag *Bag) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(bag, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//GetRoleBuyCount4Bag 返回角色当前已经购买的格子数
func GetRoleBuyCount4Bag(client *redis.Client, roleID int) int {
	tableName := "h_bag:" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return 0
	}
	buyCount, _ := client.HGet(tableName, "BuyCount").Int64()
	return int(buyCount)
}

//SetRoleBuyCount4Bag 保存角色当前格子数
func SetRoleBuyCount4Bag(client *redis.Client, roleID int, buyCount int) {
	tableName := "h_bag:" + strconv.Itoa(roleID)
	client.HSet(tableName, "BuyCount", fmt.Sprintf("%d", buyCount))
}

//GetRoleDistri4Bag 返回当前角色背包分布
func GetRoleDistri4Bag(client *redis.Client, roleID int) []int64 {
	tableName := "h_bag:" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	var distri [4]int64
	for i := 0; i < len(distri); i++ {
		distri[i], _ = client.HGet(tableName, fmt.Sprintf("Distri%d", i)).Int64()
	}

	return distri[:]
}

//SetRoleDistri4Bag 保存当前角色背包分布
func SetRoleDistri4Bag(client *redis.Client, roleID int, distri []int64) {
	tableName := "h_bag:" + strconv.Itoa(roleID)

	for i := 0; i < len(distri); i++ {
		client.HSet(tableName, fmt.Sprintf("Distri%d", i), fmt.Sprintf("%d", distri[i]))
	}

	return
}

//GetRoleItem4Bag 返回当前角色背包指定位置的物品
func GetRoleItem4Bag(client *redis.Client, roleID int, indx int) *Item {
	tableName := "h_bag:" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	info := client.HGet(tableName, fmt.Sprintf("%d", indx)).Val()
	j, err := GxMisc.BufToMsg([]byte(info))
	if err != nil {
		return nil
	}
	item := new(Item)
	GxMisc.JSONToStruct(j, item)
	return item
}

//SetRoleItem4Bag 保存指定物品到角色背包指定位置
func SetRoleItem4Bag(client *redis.Client, roleID int, indx int, item *Item) {
	tableName := "h_bag:" + strconv.Itoa(roleID)

	buf, err := GxMisc.MsgToBuf(item)
	if err != nil {
		return
	}

	client.HSet(tableName, fmt.Sprintf("%d", indx), string(buf))
}

//DelRoleItem4Bag 删除角色背包指定格子的物品，不管多少数量
func DelRoleItem4Bag(client *redis.Client, roleID int, index int) {
	tableName := "h_bag:" + strconv.Itoa(roleID)

	distri := GetRoleDistri4Bag(client, roleID)
	if distri == nil {
		return
	}
	distri[index/64] &= ^(1 << uint32(index%64))

	client.HSet(tableName, "Distri"+fmt.Sprintf("%d", index/64), fmt.Sprintf("%d", distri[index/64]))
	client.HDel(tableName, fmt.Sprintf("%d", index))
}
