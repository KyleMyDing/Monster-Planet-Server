package main

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//Players 帐号列表
type Players map[string]*GxStatic.Player

var players Players
var playersMutex *sync.Mutex

var sqlList *list.List
var sqlMutex *sync.Mutex

func init() {
	sqlList = list.New()
	sqlMutex = new(sync.Mutex)

	players = make(Players)
	playersMutex = new(sync.Mutex)
}

//PushPlayerSQL 保存一个帐号操作sql
func PushPlayerSQL(sql string) {
	sqlMutex.Lock()
	defer sqlMutex.Unlock()

	sqlList.PushBack(sql)
}

//PopPlayerSQL 取出一个帐号操作sql
func PopPlayerSQL() string {
	sqlMutex.Lock()
	defer sqlMutex.Unlock()

	if sqlList.Len() == 0 {
		return ""
	}

	sql := sqlList.Front().Value.(string)
	sqlList.Remove(sqlList.Front())
	return sql
}

//LoadPlayer 加载所有帐号
func LoadPlayer() error {
	str := GxMisc.GenerateSelectAllSQL(&GxStatic.Player{}, "")
	rows, err := GxMisc.MysqlPool.Query(str)
	defer rows.Close()
	if err != nil {
		return err
	}

	n := 0
	for rows.Next() {
		player := new(GxStatic.Player)
		err = rows.Scan(&player.ID, &player.Username, &player.Password, &player.CreateTs, &player.Platform)
		players[player.Username] = player
		n++
	}
	GxMisc.Debug("load player, count: %d", n)

	go func() {
		for {
			sql := PopPlayerSQL()
			if sql == "" {
				time.Sleep(time.Second * 1)
				continue
			}
			_, err := GxMisc.MysqlPool.Exec(sql)
			if err != nil {
				fmt.Println(err, ",sql:", sql)
			}
		}

	}()

	return nil
}

//FindPlayer 根据帐号名返回帐号信息
func FindPlayer(name string) *GxStatic.Player {
	playersMutex.Lock()
	defer playersMutex.Unlock()

	player, ok := players[name]
	if ok {
		return player
	}

	return nil
}

//AddPlayer 添加一个新帐号
func AddPlayer(player *GxStatic.Player) error {
	playersMutex.Lock()
	defer playersMutex.Unlock()

	_, ok := players[player.Username]
	if ok {
		return errors.New("player name is exists")
	}

	players[player.Username] = player
	PushPlayerSQL(GxMisc.GenerateInsertSQL(player, ""))
	return nil
}
