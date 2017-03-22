/**
作者： Kyle Ding
模块：任务模块
说明：
创建时间：2015-12-22
**/
package GxStatic

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"reflect"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
)

type Tasks struct {
	RoleID int `pk:"true"`

	PassChapterTs       int64 `type:"time"` //完成副本任务时间
	PassChapterCnt      int   //通过副本的次数
	BlueHerosCnt        int   //蓝色武将个数
	PurpleHerosCnt      int   //紫色武将个数
	OrangeHerosCnt      int   //橙色武将个数
	ReceiveFreePowerCnt int   //领取体力次数
	BuyPowerCnt         int   //购买体力次数
	SignInDaily         int   //每日签到
	SignInMonthCnt      int   //每月签到次数
	MoneyPointCnt       int   //点金次数
	CompetitiveDailyCnt int   //每日竞技场次数
	CompetitiveTotalCnt int   //竞技场总次数
	PlayerLevel         int   //玩家的等级
	IsPassTaskChapter   int   //任务指定的关卡
}

func NewTasks(roleId int) *Tasks {
	role := new(Role)
	return &Tasks{
		RoleID: roleId,

		PassChapterTs:       0,
		PassChapterCnt:      0,
		BlueHerosCnt:        0,
		PurpleHerosCnt:      0,
		OrangeHerosCnt:      0,
		ReceiveFreePowerCnt: 0,
		BuyPowerCnt:         0,
		SignInDaily:         0,
		SignInMonthCnt:      0,
		MoneyPointCnt:       0,
		CompetitiveDailyCnt: 0,
		CompetitiveTotalCnt: 0,
		PlayerLevel:         role.GetLev(),
		IsPassTaskChapter:   0,
	}
}

func (tasks *Tasks) Set4Redis(client *redis.Client) error {
	GxMisc.SaveToRedis(client, tasks)
	return nil
}

func (tasks *Tasks) Get4Redis(client *redis.Client) error {
	GxMisc.LoadFromRedis(client, tasks)
	return nil
}

func (tasks *Tasks) Del4Redis(client *redis.Client) error {
	GxMisc.DelFromRedis(client, tasks)
	return nil
}

func (tasks *Tasks) Set4Mysql() error {
	t := NewTasks(tasks.RoleID)
	var str string
	if t.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(tasks, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(tasks, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

func (tasks *Tasks) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(tasks, strconv.Itoa(ServerID))
	//fmt.Println(str)
	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&tasks.PassChapterTs, &tasks.PassChapterCnt,
			&tasks.BlueHerosCnt, &tasks.PurpleHerosCnt, &tasks.OrangeHerosCnt,
			&tasks.ReceiveFreePowerCnt, &tasks.BuyPowerCnt,
			&tasks.SignInDaily, &tasks.SignInMonthCnt, &tasks.MoneyPointCnt,
			&tasks.CompetitiveDailyCnt, &tasks.CompetitiveTotalCnt,
			&tasks.PlayerLevel, &tasks.IsPassTaskChapter,
		)
	}
	return errors.New("null")
}

func (tasks *Tasks) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(tasks, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return err
}

func (tasks *Tasks) SetBlueHerosCnt4Redis(client *redis.Client, newBuleHerosCardNum int) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	client.HIncrBy(tableName, "BlueHerosCnt", int64(newBuleHerosCardNum))
}

func (tasks *Tasks) GetBuleHerosCnt4Redis(client *redis.Client) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New(" key not existst ")
	}
	buleHeroCardCnt, _ := client.HGet(tableName, "BlueHerosCnt").Int64()
	return int(buleHeroCardCnt), nil
}

func (tasks *Tasks) SetPurpleHerosCnt4Redis(client *redis.Client, newPurpleHerosNum int) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	client.HIncrBy(tableName, "PurpleHerosCnt", int64(newPurpleHerosNum))
}

func (tasks *Tasks) GetPurpleHerosCnt4Redis(client *redis.Client) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New(" key not existst ")
	}
	purpleHerosCnt, _ := client.HGet(tableName, "PurpleHerosCnt").Int64()
	return int(purpleHerosCnt), nil
}

func (tasks *Tasks) SetOrangeHerosCnt4Redis(client *redis.Client, newOrangeHerosNum int) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	client.HIncrBy(tableName, "OrangeHerosCnt", int64(newOrangeHerosNum))
}

func (tasks *Tasks) GetOrangeHerosCnt4Redis(client *redis.Client) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New(" key not existst ")
	}
	orangeHerosCnt, _ := client.HGet(tableName, "OrangeHerosCnt").Int64()
	return int(orangeHerosCnt), nil
}

func (tasks *Tasks) SetHerosCnt4Redis(client *redis.Client, newHerosNum int, field string) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	client.HIncrBy(tableName, field, int64(newHerosNum))
}

func (tasks *Tasks) GetHerosCnt4Redis(client *redis.Client, field string) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(tasks) + ":" + strconv.Itoa(tasks.RoleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New(" key not existst ")
	}
	herosCnt, _ := client.HGet(tableName, field).Int64()
	return int(herosCnt), nil
}

func GetInfo(rsp proto.Message) *GxProto.RespondInfo {
	var info *GxProto.RespondInfo

	dataStruct := reflect.Indirect(reflect.ValueOf(rsp))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)

		if fieldType.Name == "Info" {
			info = fieldValue.Interface().(*GxProto.RespondInfo)
			/*fmt.Println("===1===")
			if info != nil {
				fmt.Printf("===2===%s\n", info.String())
			}*/
			break
		}
	}
	return info
}
