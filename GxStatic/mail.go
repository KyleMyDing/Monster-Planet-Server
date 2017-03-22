/**
作者： Kyle Ding
模块：邮件模块
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
	"sync"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//mailIDTableName 邮ID表，redis的表名
var mailIDTableName = "k_mail_id"
var mailIDMutex sync.Mutex

//mailListTableName 角色邮件列表，redis的表名
var mailListTableName = "s_mail_list:"

//RoleUngetMailList 未入档邮件列表，redis的表名
var RoleUngetMailList = "l_unget_mail_list:"

//Mail 邮件信息
type Mail struct {
	RoleID int    `pk:"true"` //角色ID
	ID     int    `pk:"true"` //邮件ID
	Sender string //发送者
	Ts     int64  //邮件生成时间
	Title  string `len:"128"`   //标题
	Text   string `len:"1024"`  //内容
	Info   string `type:"text"` //可领取物品
	Rd     int    //是否读取， 1-读取，0-未读取

	Del int `ignore:"true"` //邮件是否已经删除
}

//NewRoleID 生成一个新的邮件ID
func NewMailID(client *redis.Client) int {
	mailIDMutex.Lock()
	defer mailIDMutex.Unlock()

	if !client.Exists(mailIDTableName).Val() {
		client.Set(mailIDTableName, "1", 0)
	}
	return int(client.Incr(mailIDTableName).Val())
}

//NewMail 生成一个新邮件
func NewMail(client *redis.Client, roleID int, sender string, title string, text string, items []*Item) *Mail {
	j := simplejson.New()
	for i := 0; i < len(items); i++ {
		j.Set(fmt.Sprintf("%d", items[i].ID), items[i].Cnt)
	}
	buf, _ := j.Encode()
	return &Mail{
		RoleID: roleID,
		ID:     NewMailID(client),
		Sender: sender,
		Ts:     time.Now().Unix(),
		Title:  title,
		Text:   text,
		Info:   string(buf),
		Del:    0,
	}
}

//NewMail4String 生成一个新邮件
func NewMail4String(client *redis.Client, roleID int, sender string, title string, text string, items string) *Mail {
	return &Mail{
		RoleID: roleID,
		ID:     NewMailID(client),
		Sender: sender,
		Ts:     time.Now().Unix(),
		Title:  title,
		Text:   text,
		Info:   items,
		Del:    0,
	}
}

//Set4Redis ...
func (mail *Mail) Set4Redis(client *redis.Client) error {
	GxMisc.SaveToRedis(client, mail)
	return nil
}

//Get4Redis ...
func (mail *Mail) Get4Redis(client *redis.Client) error {
	return GxMisc.LoadFromRedis(client, mail)
}

//Del4Redis ...
func (mail *Mail) Del4Redis(client *redis.Client) error {
	return GxMisc.DelFromRedis(client, mail)
}

//Set4Mysql ...
func (mail *Mail) Set4Mysql() error {
	r := new(Mail)
	r.RoleID = mail.RoleID
	r.ID = mail.ID
	var str string
	if r.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(mail, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(mail, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (mail *Mail) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(mail, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&mail.Sender, &mail.Ts, &mail.Title, &mail.Text, &mail.Info, &mail.Rd)
	}
	return errors.New("null")
}

//Del4Mysql ...
func (mail *Mail) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(mail, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return err
}

//GetItems 获取邮件物品列表
func (mail *Mail) GetItems() []*Item {
	var item []*Item
	j, err := simplejson.NewJson([]byte(mail.Info))
	if err != nil {
		return nil
	}
	m, _ := j.Map()
	for k, v := range m {
		id, _ := strconv.Atoi(k)
		cnt, _ := v.(json.Number).Int64()
		item = append(item, &Item{
			ID:  id,
			Cnt: int(cnt),
		})
	}
	return item
}

//GetRoleAllMailIds 获取角色所有邮件id列表
func GetRoleAllMailIds(client *redis.Client, roleID int) []int {
	var ids []int
	ret := client.SMembers(mailListTableName + strconv.Itoa(roleID))
	arr, err := ret.Result()
	if err != nil {
		return nil
	}
	for i := 0; i < len(arr); i++ {
		id, _ := strconv.Atoi(arr[i])
		ids = append(ids, id)
	}
	return ids
}

//GetRoleAllMail 获取角色所有邮件
func GetRoleAllMail(client *redis.Client, roleID int) []*Mail {
	var mails []*Mail

	ids := GetRoleAllMailIds(client, roleID)
	for i := 0; i < len(ids); i++ {
		mail := &Mail{
			RoleID: roleID,
			ID:     ids[i],
		}
		if mail.Get4Redis(client) != nil {
			continue
		}
		mails = append(mails, mail)
	}
	return mails
}

//SetRoleMail 保存邮件
func SetRoleMail(client *redis.Client, roleID int, mail *Mail) {
	mail.Set4Redis(client)
	client.SAdd(mailListTableName+strconv.Itoa(roleID), strconv.Itoa(mail.ID))
}

//GetRoleMail 获取邮件
func GetRoleMail(client *redis.Client, roleID int, mailID int) *Mail {
	mail := &Mail{
		RoleID: roleID,
		ID:     mailID,
	}
	if mail.Get4Redis(client) != nil {
		return nil
	}
	return mail
}

//DelRoleMail 删除邮件
func DelRoleMail(client *redis.Client, roleID int, mailID int) {
	client.SRem(mailListTableName+strconv.Itoa(roleID), strconv.Itoa(mailID))

	mail := &Mail{
		RoleID: roleID,
		ID:     mailID,
		Del:    1,
	}
	GxMisc.SetFieldFromRedis(client, mail, "Del")
}

//SaveRoleUngetMailList 保存角色未入档邮件，登录时检查，或者其他模块通知
func SaveRoleUngetMailList(client *redis.Client, roleID int, mail *Mail) {
	buf, err := GxMisc.MsgToBuf(mail)
	if err != nil {
		return
	}
	client.LPush(RoleUngetMailList+strconv.Itoa(roleID), string(buf))
}

//GetRoleUngetMailList 获取角色未入档邮件列表
func GetRoleUngetMailList(client *redis.Client, roleID int) []*Mail {
	tablename := RoleUngetMailList + strconv.Itoa(roleID)
	var mails []*Mail
	arr := client.LRange(tablename, 0, -1).Val()
	client.Del(tablename)
	for i := 0; i < len(arr); i++ {
		j, err := GxMisc.BufToMsg([]byte(arr[i]))
		if err == nil {
			mail := new(Mail)
			GxMisc.JSONToStruct(j, mail)
			mails = append(mails, mail)
		}
	}
	return mails
}
