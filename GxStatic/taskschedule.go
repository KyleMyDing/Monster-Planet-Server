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
	"github.com/bitly/go-simplejson"
	"gopkg.in/redis.v3"
	"strconv"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

type TaskSchedule struct {
	RoleID int    `pk:"true"` //角色ID
	Info   string `type:"text"`

	IsCompleteTaskAndReward map[int]int `ignore:"true"`
}

func NewTaskSchedule(roleID int) *TaskSchedule {
	return &TaskSchedule{
		RoleID:                  roleID,
		IsCompleteTaskAndReward: make(map[int]int),
	}
}

func (tSchedule *TaskSchedule) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(tSchedule) + ":" + strconv.Itoa(tSchedule.RoleID)
	for k, v := range tSchedule.IsCompleteTaskAndReward {
		client.HSet(tableName, fmt.Sprintf("%d", k), strconv.Itoa(v))
	}
	return nil
}

func (tSchedule *TaskSchedule) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(tSchedule) + ":" + strconv.Itoa(tSchedule.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New(" key not existst ")
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return errors.New("table is nil")
	}
	for k, v := range m {
		flag, _ := strconv.Atoi(v)
		taskType, _ := strconv.Atoi(k)
		tSchedule.IsCompleteTaskAndReward[taskType] = flag

	}
	return nil
}

func (tSchedule *TaskSchedule) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(tSchedule) + ":" + strconv.Itoa(tSchedule.RoleID)

	client.Del(tableName)
	return nil
}

func (tSchedule *TaskSchedule) Set4Mysql() error {
	if tSchedule.Info == "" {
		j := simplejson.New()
		for k, v := range tSchedule.IsCompleteTaskAndReward {
			j.Set(fmt.Sprintf("%d", k), v)
		}
		buf, _ := j.MarshalJSON()
		tSchedule.Info = string(buf)

	}

	b := NewTaskSchedule(tSchedule.RoleID)
	var str string
	if b.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(tSchedule, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(tSchedule, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

func (tSchedule *TaskSchedule) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(tSchedule, strconv.Itoa(ServerID))
	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&tSchedule.Info)
		if tSchedule.Info == "" {
			return nil
		}

		isCompleteAndRewardJson, _ := simplejson.NewJson([]byte(tSchedule.Info))
		ptrDict, err := GxDict.GetDict("Tasks")
		if err != nil {
			GxMisc.Warn("msg  GetDict error, get4RedisError: %s", err)
			return err
		}
		for k := range *ptrDict {
			if !GxDict.CheckDictId("Tasks", k) {
				tSchedule.IsCompleteTaskAndReward[k], _ = isCompleteAndRewardJson.Get(strconv.Itoa(k)).Int()
			}
		}

		return nil
	}
	return errors.New("null")
}

func (tSchedule *TaskSchedule) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(tSchedule, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

func (tSchedule *TaskSchedule) SetIsCompleteTaskAndReward4Redis(client *redis.Client, key, flag int) error {
	tableName := "h_" + GxMisc.GetTableName(tSchedule) + ":" + strconv.Itoa(tSchedule.RoleID)
	client.HSet(tableName, fmt.Sprintf("%d", key), strconv.Itoa(flag))
	return nil
}

func (tSchedule *TaskSchedule) GetIsCompleteTaskAndReward4Redis(client *redis.Client, key int) (int, error) {
	tableName := "h_" + GxMisc.GetTableName(tSchedule) + ":" + strconv.Itoa(tSchedule.RoleID)
	if !client.Exists(tableName).Val() {
		return -1, errors.New("key not exists")
	}

	info := client.HGet(tableName, fmt.Sprintf("%d", key)).Val()
	if info == "" {
		GxMisc.Debug("Info is nil")
		return 0, nil
	}
	flag, err := strconv.Atoi(info)
	if err != nil {
		return 0, errors.New("Atoi error")
	}
	return flag, nil
}
