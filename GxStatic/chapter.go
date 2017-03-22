/**
作者： Kyle Ding
模块：副本信息模块
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

//ChapterPointInfo 角色战斗点通关信息
type ChapterPointInfo struct {
	PassCnt  int //胜利次数
	FightCnt int //战斗次数
}

//Chapter 副本信息
type Chapter struct {
	RoleID       int    `pk:"true"` //角色ID
	ChapterId    int    //当前战斗章节id
	LevelId      int    //当前战斗关卡id
	PointId      int    //当前战斗战斗点id
	NowChapterId int    //当前进行到的章节
	NowLevelId   int    //当前进行到的关卡
	NowPointId   int    //当前进行到的战斗点
	Info         string `type:"text"`

	Points map[string]string `ignore:"true"` //战斗点通关信息,key-ChapterId:LevelId:PointId value-ChapterPointInfo
}

//NewChapter 生成一个新副本信息
func NewChapter(roleID int) *Chapter {
	return &Chapter{
		RoleID: roleID,
		Points: make(map[string]string),
	}
}

//Set4Redis ...
func (chapter *Chapter) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	client.HSet(tableName, "ChapterId", fmt.Sprintf("%d", chapter.ChapterId))
	client.HSet(tableName, "LevelId", fmt.Sprintf("%d", chapter.LevelId))
	client.HSet(tableName, "PointId", fmt.Sprintf("%d", chapter.PointId))
	client.HSet(tableName, "NowChapterId", fmt.Sprintf("%d", chapter.NowChapterId))
	client.HSet(tableName, "NowLevelId", fmt.Sprintf("%d", chapter.NowLevelId))
	client.HSet(tableName, "NowPointId", fmt.Sprintf("%d", chapter.NowPointId))

	for k, v := range chapter.Points {
		client.HSet(tableName, k, v)
	}
	return nil
}

//Get4Redis ...
func (chapter *Chapter) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return errors.New("table is nil")
	}
	for k, v := range m {
		if k == "ChapterId" {
			chapter.ChapterId, _ = strconv.Atoi(v)
			continue
		} else if k == "LevelId" {
			chapter.LevelId, _ = strconv.Atoi(v)
			continue
		} else if k == "PointId" {
			chapter.PointId, _ = strconv.Atoi(v)
			continue
		} else if k == "NowChapterId" {
			chapter.NowChapterId, _ = strconv.Atoi(v)
			continue
		} else if k == "NowLevelId" {
			chapter.NowLevelId, _ = strconv.Atoi(v)
			continue
		} else if k == "NowPointId" {
			chapter.NowPointId, _ = strconv.Atoi(v)
			continue
		} else {
			chapter.Points[k] = v
		}
	}
	return nil
}

//Del4Redis ...
func (chapter *Chapter) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}

	client.Del(tableName)
	return nil
}

//Set4Mysql ...
func (chapter *Chapter) Set4Mysql() error {
	if chapter.Info == "" {
		j := simplejson.New()
		for k, v := range chapter.Points {
			j.Set(k, v)
		}
		buf, _ := j.MarshalJSON()
		chapter.Info = string(buf)
	}

	b := NewChapter(chapter.RoleID)
	var str string
	if b.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(chapter, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(chapter, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (chapter *Chapter) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(chapter, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&chapter.ChapterId, &chapter.LevelId, &chapter.PointId, &chapter.NowChapterId, &chapter.NowLevelId, &chapter.NowPointId, &chapter.Info)

		if chapter.Info == "" {
			return nil
		}
		j, _ := simplejson.NewJson([]byte(chapter.Info))
		m, _ := j.Map()
		for k, v := range m {
			chapter.Points[k] = v.(string)
		}
		return nil
	}
	return errors.New("null")
}

//Del4Mysql ...
func (chapter *Chapter) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(chapter, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//SetFightChapter 保存当前战斗关卡
func (chapter *Chapter) SetFightChapter(client *redis.Client) {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	client.HSet(tableName, "ChapterId", fmt.Sprintf("%d", chapter.ChapterId))
	client.HSet(tableName, "LevelId", fmt.Sprintf("%d", chapter.LevelId))
	client.HSet(tableName, "PointId", fmt.Sprintf("%d", chapter.PointId))
}

//SaveFightChapter 保存当前战斗关卡
func (chapter *Chapter) GetFightChapter(client *redis.Client) {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	if !client.Exists(tableName).Val() {
		return
	}
	chapterId, _ := client.HGet(tableName, "ChapterId").Int64()
	chapter.ChapterId = int(chapterId)

	levelId, _ := client.HGet(tableName, "LevelId").Int64()
	chapter.LevelId = int(levelId)

	pointId, _ := client.HGet(tableName, "PointId").Int64()
	chapter.PointId = int(pointId)
}

//SetChapter 保存当前进度关卡
func (chapter *Chapter) SetChapter(client *redis.Client) {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	client.HSet(tableName, "NowChapterId", fmt.Sprintf("%d", chapter.NowChapterId))
	client.HSet(tableName, "NowLevelId", fmt.Sprintf("%d", chapter.NowLevelId))
	client.HSet(tableName, "NowPointId", fmt.Sprintf("%d", chapter.NowPointId))
}

//GetChapter 保存当前进度关卡
func (chapter *Chapter) GetChapter(client *redis.Client) {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	if !client.Exists(tableName).Val() {
		return
	}
	chapterId, _ := client.HGet(tableName, "NowChapterId").Int64()
	chapter.NowChapterId = int(chapterId)

	levelId, _ := client.HGet(tableName, "NowLevelId").Int64()
	chapter.NowLevelId = int(levelId)

	pointId, _ := client.HGet(tableName, "NowPointId").Int64()
	chapter.NowPointId = int(pointId)
}

//SetPoint 保存当前战斗点信息
func (chapter *Chapter) SetPoint(client *redis.Client, chapterId int, levelId int, pointId int, point *ChapterPointInfo) {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	buf, err := GxMisc.MsgToBuf(point)
	if err != nil {
		return
	}
	client.HSet(tableName, fmt.Sprintf("%d:%d:%d", chapterId, levelId, pointId), string(buf))
}

//GetPoint 读取战斗点信息
func (chapter *Chapter) GetPoint(client *redis.Client, chapterId int, levelId int, pointId int) *ChapterPointInfo {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	info := client.HGet(tableName, fmt.Sprintf("%d:%d:%d", chapterId, levelId, pointId)).Val()
	j, err := GxMisc.BufToMsg([]byte(info))
	if err != nil {
		return nil
	}
	point := new(ChapterPointInfo)
	GxMisc.JSONToStruct(j, point)
	return point
}

//PointExists 战斗点信息是否存在，读取是否通关时使用
func (chapter *Chapter) PointExists(client *redis.Client, chapterId int, levelId int, pointId int) bool {
	tableName := "h_" + GxMisc.GetTableName(chapter) + ":" + strconv.Itoa(chapter.RoleID)
	return client.Exists(tableName).Val() && client.HExists(tableName, fmt.Sprintf("%d:%d:%d", chapterId, levelId, pointId)).Val()
}
