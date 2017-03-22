/**
作者： Kyle Ding
模块：帐号信息模块
说明：
创建时间：2015-10-30
**/
package GxStatic

import (
	"crypto/md5"
	"errors"
	"fmt"
	"gopkg.in/redis.v3"
	"io"
	"sync"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

var salt1 = "f4g*h(j"
var salt2 = "1^2&4*(d)"

//PlayerIDTableName 帐号ID信息 redis表名
var PlayerIDTableName = "k_player_id"

//PlayerTokenTableName 帐号ID信息 redis表名
var PlayerTokenTableName = "k_player_token:"

//PlayerNameListTableName 帐号ID信息 redis表名
var PlayerNameListTableName = "s_player_name:"

var IDMutex *sync.Mutex

//Player 帐号信息
type Player struct {
	ID       uint32 //帐号ID
	Username string `pk:"true"` //帐号名
	Password string //帐号密码
	CreateTs uint64 `type:"time"` //创建帐号时间
	Platform uint32 //所属平台
}

func init() {
	IDMutex = new(sync.Mutex)
}

//newPalayerID 生成一个帐号ID
func newPalayerID(client *redis.Client) uint32 {
	IDMutex.Lock()
	defer IDMutex.Unlock()

	if !client.Exists(PlayerIDTableName).Val() {
		client.Set(PlayerIDTableName, "100000", 0)

	}
	return uint32(client.Incr(PlayerIDTableName).Val())
}

//NewPlayer 生成一个帐号ID
func NewPlayer(client *redis.Client, username string, password string, platform uint32) *Player {
	player := &Player{
		ID:       newPalayerID(client),
		Username: username,
		Password: generatePassward(username, password),
		CreateTs: uint64(time.Now().Unix()),
		Platform: platform,
	}
	return player
}

//Set4Redis ...
func (player *Player) Set4Redis(client *redis.Client) error {
	GxMisc.SaveToRedis(client, player)
	return nil
}

//Get4Redis ...
func (player *Player) Get4Redis(client *redis.Client) error {
	return GxMisc.LoadFromRedis(client, player)
}

//Del4Redis ...
func (player *Player) Del4Redis(client *redis.Client) error {
	return GxMisc.DelFromRedis(client, player)
}

//Set4Mysql ...
func (player *Player) Set4Mysql() error {
	r := new(Player)
	r.ID = player.ID
	var str string
	if r.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(player, "")
	} else {
		str = GxMisc.GenerateInsertSQL(player, "")
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (player *Player) Get4Mysql() error {
	str := GxMisc.GenerateSelectSQL(player, "")

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&player.ID, &player.Password, &player.CreateTs, &player.Platform)

		fmt.Println("get one")
		return nil
	}
	fmt.Println("get null")
	return errors.New("null")
}

//Del4Mysql ...
func (player *Player) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(player, "")
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return nil
}

//生成密码函数
func generatePassward(username string, password string) string {
	h := md5.New()

	io.WriteString(h, salt1)
	io.WriteString(h, username)
	io.WriteString(h, salt2)
	io.WriteString(h, password)

	return fmt.Sprintf("%x", h.Sum(nil))
}

//VerifyPassword 验证帐号密码
func VerifyPassword(player *Player, password string) bool {
	return player.Password == generatePassward(player.Username, password)
}

//SaveToken 生成token函数
func (player *Player) SaveToken(client *redis.Client) string {
	h := md5.New()
	io.WriteString(h, player.Username)
	io.WriteString(h, fmt.Sprintf("%ld", time.Now().Unix()))
	io.WriteString(h, "2@#RR#R@")
	token := fmt.Sprintf("%x", h.Sum(nil))

	client.Set(PlayerTokenTableName+token, player.Username, time.Hour*1)
	return token
}

//CheckToken 验证token
func CheckToken(client *redis.Client, token string) string {
	key := PlayerTokenTableName + token
	if !client.Exists(key).Val() {
		return ""
	}
	return client.Get(key).Val()
}

//CheckPlayerNameConflict 创建角色时候检查角色名是否冲突
func CheckPlayerNameConflict(client *redis.Client, playerName string) bool {
	return client.SIsMember(PlayerNameListTableName, playerName).Val()
}
