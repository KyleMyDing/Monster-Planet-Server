/**
作者： Kyle Ding
模块：好友信息模块
说明：
创建时间：2015-11-14
**/
package GxStatic

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

//RoleAddFriendList 未入档的添加好友通知列表，redis的表名
var RoleAddFriendList = "l_add_friend_list:"

//RoleDealFriendList 未入档的同意添加好友通知列表，redis的表名
var RoleDealFriendYesList = "l_deal_friend_yes_list:"

//RoleDealFriendList 未入档的拒绝添加好友通知列表，redis的表名
var RoleDealFriendNoList = "l_deal_friend_no_list:"

//RoleDeleteFriendList 未入档的删除好友通知列表，redis的表名
var RoleDeleteFriendList = "l_delete_friend_list:"

//RoleFriendSendList 未入档的赠送物品通知列表，redis的表名
var RoleFriendSendList = "l_friend_send_list:"

const (
	//FriendMaskUnaccept 还未接受或者拒绝的好友请求
	FriendMaskUnaccept = 0
	//FriendMaskAccept 正式好友
	FriendMaskAccept = 1
	//FriendMaskSend 是否赠送过物品
	FriendMaskSend = 2
	//FriendMaskAccept 是否可以领取物品
	FriendMaskRecv = 3
)

//Friend 好友信息
type Friend struct {
	RoleID int   `pk:"true"`   //角色ID
	Ts     int64 `type:"time"` //最近一次赠送或者领取时间
	Cnt    int
	Info   string `type:"text"`

	Friends map[int]int `ignore:"true"`
}

//NewFriend 生成一个新好友信息
func NewFriend(roleID int) *Friend {
	return &Friend{
		RoleID:  roleID,
		Friends: make(map[int]int),
	}
}

//Set4Redis ...
func (friend *Friend) Set4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(friend) + ":" + strconv.Itoa(friend.RoleID)

	client.HSet(tableName, "Ts", fmt.Sprintf("%d", friend.Ts))
	client.HSet(tableName, "Cnt", fmt.Sprintf("%d", friend.Cnt))
	for k, v := range friend.Friends {
		client.HSet(tableName, fmt.Sprintf("%d", k), fmt.Sprintf("%d", v))
	}
	return nil
}

//Get4Redis ...
func (friend *Friend) Get4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(friend) + ":" + strconv.Itoa(friend.RoleID)
	if !client.Exists(tableName).Val() {
		return nil
	}

	r := client.HGetAllMap(tableName)
	m, err := r.Result()
	if err != nil {
		return nil
	}
	for k, v := range m {
		if k == "Ts" {
			friend.Ts, _ = strconv.ParseInt(v, 10, 64)
			continue
		}
		if k == "Cnt" {
			friend.Cnt, _ = strconv.Atoi(v)
			continue
		}
		friendID, _ := strconv.Atoi(k)
		cnt, _ := strconv.Atoi(v)
		friend.Friends[friendID] = cnt
	}
	return nil

}

//Del4Redis ...
func (friend *Friend) Del4Redis(client *redis.Client) error {
	tableName := "h_" + GxMisc.GetTableName(friend) + ":" + strconv.Itoa(friend.RoleID)
	if !client.Exists(tableName).Val() {
		return errors.New("key not existst")
	}
	client.Del(tableName)
	return nil
}

//Set4Mysql ...
func (friend *Friend) Set4Mysql() error {
	if friend.Info == "" {
		j := simplejson.New()

		for k, v := range friend.Friends {
			j.Set(fmt.Sprintf("%d", k), v)
		}
		buf, _ := j.MarshalJSON()
		friend.Info = string(buf)
	}

	f := NewFriend(friend.RoleID)
	var str string
	if f.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(friend, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(friend, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (friend *Friend) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(friend, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&friend.Info)

		if friend.Info == "" {
			return nil
		}
		j, _ := simplejson.NewJson([]byte(friend.Info))
		m, _ := j.Map()
		for k, v := range m {
			friendID, _ := strconv.Atoi(k)
			cnt, _ := v.(json.Number).Int64()
			friend.Friends[friendID] = int(cnt)
		}

		return nil
	}
	return errors.New("null")
}

//Del4Mysql ...
func (friend *Friend) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(friend, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//SetFriend4Redis 保存某个好友
func (friend *Friend) SetFriend4Redis(client *redis.Client, friendId int) {
	tableName := "h_friend:" + strconv.Itoa(friend.RoleID)
	client.HSet(tableName, fmt.Sprintf("%d", friendId), fmt.Sprintf("%d", friend.Friends[friendId]))
}

//DeleteFriend4Redis 删除某个好友
func (friend *Friend) DeleteFriend4Redis(client *redis.Client, friendId int) {
	tableName := "h_friend:" + strconv.Itoa(friend.RoleID)
	client.HDel(tableName, fmt.Sprintf("%d", friendId))
}

//IsUnaccept 指定角色是否是一个还未接受或者拒绝的好友
func (friend *Friend) IsUnaccept(roleId int) bool {
	return (friend.Friends[roleId] & (1 << (uint32(FriendMaskUnaccept)))) != 0
}

//IsAccept 指定角色是否是一个正式好友
func (friend *Friend) IsAccept(roleId int) bool {
	return (friend.Friends[roleId] & (1 << (uint32(FriendMaskAccept)))) != 0
}

//GetSend 是否已经赠送过物品
func (friend *Friend) GetSend(roleId int) bool {
	return (friend.Friends[roleId] & (1 << (uint32(FriendMaskSend)))) != 0
}

//SetSend 设置赠送过物品
func (friend *Friend) SetSend(client *redis.Client, roleId int) {
	if !friend.IsAccept(roleId) {
		return
	}
	friend.Friends[roleId] |= 1 << uint32(FriendMaskSend)

	tableName := "h_" + GxMisc.GetTableName(friend) + ":" + strconv.Itoa(friend.RoleID)
	client.HSet(tableName, fmt.Sprintf("%d", roleId), fmt.Sprintf("%d", friend.Friends[roleId]))
}

//GetRecv 是否可领取赠送的物品
func (friend *Friend) GetRecv(roleId int) bool {
	return (friend.Friends[roleId] & (1 << (uint32(FriendMaskRecv)))) != 0
}

//SetRecv 设置领取过赠送的物品
func (friend *Friend) SetRecv(client *redis.Client, roleId int) {
	if !friend.IsAccept(roleId) {
		return
	}
	friend.Friends[roleId] ^= 1 << uint32(FriendMaskRecv)

	tableName := "h_" + GxMisc.GetTableName(friend) + ":" + strconv.Itoa(friend.RoleID)
	client.HSet(tableName, fmt.Sprintf("%d", roleId), fmt.Sprintf("%d", friend.Friends[roleId]))
}

//IsUnacceptFriend 是否未接受好友
func IsUnacceptFriend(client *redis.Client, roleID int, friendID int) bool {
	tableName := "h_friend:" + strconv.Itoa(roleID)
	v, err := client.HGet(tableName, fmt.Sprintf("%d", friendID)).Int64()
	if err != nil {
		return false
	}
	return (v & (1 << uint32(FriendMaskUnaccept))) != 0
}

//IsFriend 是否好友
func IsFriend(client *redis.Client, roleID int, friendID int) bool {
	tableName := "h_friend:" + strconv.Itoa(roleID)
	v, err := client.HGet(tableName, fmt.Sprintf("%d", friendID)).Int64()
	if err != nil {
		return false
	}
	return (v & ((1 << uint32(FriendMaskUnaccept)) | (1 << uint32(FriendMaskAccept)))) != 0
}

func getRoleFriendListName(roleID int, cmd int) string {
	switch cmd {
	case CmdAdminAddFriend:
		return RoleAddFriendList + strconv.Itoa(roleID)
	case CmdAdminDealFriendYes:
		return RoleDealFriendYesList + strconv.Itoa(roleID)
	case CmdAdminDealFriendNo:
		return RoleDealFriendNoList + strconv.Itoa(roleID)
	case CmdAdminDeleteFriend:
		return RoleDeleteFriendList + strconv.Itoa(roleID)
	case CmdAdminFriendSend:
		return RoleFriendSendList + strconv.Itoa(roleID)
	default:
		GxMisc.Error("error cmd: %d", cmd)
		return ""
	}
}

//SaveRoleUndealFriendLMessageist 保存角色未入档好友通知，登录时检查，或者其他模块通知
func SaveRoleUndealFriendLMessageist(client *redis.Client, roleID int, cmd int, friendId int) {
	tableName := getRoleFriendListName(roleID, cmd)
	if tableName == "" {
		return
	}
	client.LPush(tableName, fmt.Sprintf("%d", friendId))
}

//GetRoleUndealFriendLMessageist 获取角色未入档邮件列表
func GetRoleUndealFriendLMessageist(client *redis.Client, roleID int, cmd int) []int {
	tableName := getRoleFriendListName(roleID, cmd)
	if tableName == "" {
		return nil
	}

	var ids []int
	arr := client.LRange(tableName, 0, -1).Val()
	client.Del(tableName)
	for i := 0; i < len(arr); i++ {
		id, _ := strconv.Atoi(arr[i])
		ids = append(ids, id)
	}
	return ids
}
