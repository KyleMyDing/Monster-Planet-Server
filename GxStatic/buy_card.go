package GxStatic

/**
作者： Kyle Ding
模块：商店抽卡模块
说明：
创建时间：2015-12-02
**/

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"gopkg.in/redis.v3"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//BuyCardInfoType 抽卡信息
type BuyCardInfoType struct {
	Ts      int64 //上次抽卡时间
	FreeCnt int   //已经使用的免费次数
}

/*
type BuyCard struct {
	RoleID int    `pk:true"`
	Info   string `type : "text"`

	CostGoldNum int
	ExtractCard map[int]*BuyCardInfoType `ignore : "true"`
}

func NewBuyCard(roleID int) *BuyCard {
	return &BuyCard{
		RoleID:      roleID,
		CostGoldNum: 0,
		ExtractCard: make(map[int]*BuyCardInfoType),
	}
}

func (bc *BuyCard) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(bc.RoleID)
	for k, v := range bc.ExtractCard {
		v, err := GxMisc.MsgToBuf(*v)
		if err != nil {
			return errors.New(" MsgToBuf Failed ")
		}
		client.HSet(tableName, fmt.Sprintf("%d", k), string(v))
	}
	return nil
}*/

//BuyCard 总抽卡信息
type BuyCard struct {
	RoleID int    `pk:"true"` //角色ID
	Info   string `type:"text"`

	CostGoldNum   int            `ignore:"true"`
	IsReceiveCard map[int]int    `ignore:"true"` //key为领奖的种类,value = 0 表示还没领取奖励
	ExtractCard   map[int]string `ignore:"true"` //key为抽奖的种类
}

//NewBuyCard 生成一个新BuyCard
func NewBuyCard(roleId int) *BuyCard {
	return &BuyCard{
		RoleID:        roleId,
		CostGoldNum:   0,
		IsReceiveCard: make(map[int]int),
		ExtractCard:   make(map[int]string),
	}
}

//Set4Redis  把BuyCardInfoType struct 保存到redis当中
func (bc *BuyCard) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(bc.RoleID)
	client.HSet(tableName, "CostGoldNum", fmt.Sprintf("%d", bc.CostGoldNum))
	for k, v := range bc.ExtractCard {
		client.HSet(tableName, fmt.Sprintf("%d", k), v)
	}
	for k, v := range bc.IsReceiveCard {

		client.HSet(tableName, fmt.Sprintf("%d", k), strconv.Itoa(v))
	}
	return nil
}

//Get4Redis  从redis中读取BuyCardInfoType struct的信息
func (bc *BuyCard) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(bc.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New(" key not existst ")
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return errors.New("table is nil")
	}

	for k, v := range m {
		if k == "CostGoldNum" {
			bc.CostGoldNum, _ = strconv.Atoi(v)
			continue
		}

		if k != "" {
			flag, _ := strconv.Atoi(v)
			if flag == 0 || flag == 1 {
				cardID, _ := strconv.Atoi(k)
				bc.IsReceiveCard[cardID] = flag
			}

			index, _ := strconv.Atoi(k)
			bc.ExtractCard[index] = v
		}

	}

	return nil
}

//Del4Redis 从redis中删除BuyCardInfoType struct的信息
func (bc *BuyCard) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(bc.RoleID)

	client.Del(tableName)
	return nil
}

//Set4Mysql 把BuyCardInfoType struct 保存到MySql当中
func (bc *BuyCard) Set4Mysql() error {
	if bc.Info == "" {
		j := simplejson.New()
		j.Set("CostGoldNum", bc.CostGoldNum)
		for k, v := range bc.ExtractCard {
			j1, _ := simplejson.NewJson([]byte(v))
			j.Set(fmt.Sprintf("%d", k), j1)
		}
		for k, v := range bc.IsReceiveCard {
			j.Set(fmt.Sprintf("%d", k), v)
		}
		buf, _ := j.MarshalJSON()
		bc.Info = string(buf)
	}

	b := NewBuyCard(bc.RoleID)
	var str string
	if b.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(bc, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(bc, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql 从MySql中读取BuyCardInfoType struct的信息
func (bc *BuyCard) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(bc, strconv.Itoa(ServerID))
	//fmt.Println(str)
	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&bc.Info)
		if bc.Info == "" {
			return nil
		}
		j, _ := simplejson.NewJson([]byte(bc.Info))
		bc.CostGoldNum, _ = j.Get("CostGoldNum").Int()

		//for k := range (*GxDict.Dicts)["BuyCardCost"].D {
		ptrDict, err := GxDict.GetDict("BuyCardCost")
		if err != nil {
			GxMisc.Warn("msg  GetDict error, get4RedisError: %s", err)
			return err
		}
		for k := range *ptrDict {
			//if (*GxDict.Dicts)["BuyCardCost"].D[k] != nil {
			if !GxDict.CheckDictId("BuyCardCost", k) {
				arr, _ := j.Get(strconv.Itoa(k)).Encode()
				bc.ExtractCard[k] = string(arr)
			}

		}

		//for k := range (*GxDict.Dicts)["BuyCardSpecial"].D {
		ptrDict1, err := GxDict.GetDict("BuyCard")
		if err != nil {
			GxMisc.Warn("msg  GetDict error, get4RedisError: %s", err)
			return err
		}
		for k := range *ptrDict1 {
			//if (*GxDict.Dicts)["BuyCardSpecial"].D[k] != nil {
			if !GxDict.CheckDictId("BuyCardSpecial", k) {
				bc.IsReceiveCard[k], _ = j.Get(strconv.Itoa(k)).Int()
			}
		}
		return nil
	}
	return errors.New("null")
}

//Del4Mysql 从MySql中删除BuyCardInfoType struct的信息
func (bc *BuyCard) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(bc, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//SetRoleBuyCardCostGoldNum 在redis上设置玩家抽奖花费的元宝数
func (bc *BuyCard) SetRoleBuyCardCostGoldNum(client *redis.Client, roleID int, constGoldNum int) {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(roleID)
	//client.HSet(tableName, "CostGoldNum", fmt.Sprintf("%d", constGoldNum))
	client.HIncrBy(tableName, "CostGoldNum", int64(constGoldNum))
	// client.IncrBy(key, value)
}

//GetRoleBuyCardConstGoldNum 从redis上获得玩家抽奖花费的元宝书
func (bc *BuyCard) GetRoleBuyCardConstGoldNum(client *redis.Client, roleID int) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New(" key not existst ")
	}
	CostGoldNum, _ := client.HGet(tableName, "CostGoldNum").Int64()
	return int(CostGoldNum), nil
}

//SetBuyCardFreeCnt4Redis  在redis上设置玩家抽卡的信息
func (bc *BuyCard) SetBuyCardInfoType4Redis(client *redis.Client, key int, buyCardInfo *BuyCardInfoType) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(bc.RoleID)

	v, err := GxMisc.MsgToBuf(buyCardInfo)
	if err != nil {
		return errors.New(" MsgToBuf Failed ")
	}
	client.HSet(tableName, fmt.Sprintf("%d", key), string(v))
	return nil
}

//GetBuyCardFreeCnt4Redis  从redis上获取玩家抽卡的信息
func (bc *BuyCard) GetBuyCardInfoType4Redis(client *redis.Client, key, roleID int) (string, error) {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return "", errors.New("key not exists")
	}

	Info := client.HGet(tableName, fmt.Sprintf("%d", key)).Val()
	return Info, nil
}

func (bc *BuyCard) SetIsReceiveCard4Redis(client *redis.Client, key, flag, roleID int) error {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(roleID)
	client.HSet(tableName, fmt.Sprintf("%d", key), strconv.Itoa(flag))
	return nil
}

func (bc *BuyCard) GetReceiveCard4Redis(client *redis.Client, key, roleID int) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(bc) + ":" + strconv.Itoa(roleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New("key not exists")
	}

	Info := client.HGet(tableName, fmt.Sprintf("%d", key)).Val()
	if Info == "" {
		return 0, nil
	}
	flag, _ := strconv.Atoi(Info)
	return flag, nil
}
