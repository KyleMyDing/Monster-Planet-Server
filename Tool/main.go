package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	// "github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"os"
	// "reflect"
	"strconv"
	//"time"
)

import (
	//"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	// "git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

var db *sql.DB
var rdClient *redis.Client

func connectRedis() bool {
	rdHost, _ := GxMisc.Config.Get("redis").Get("host").String()
	rdPort, _ := GxMisc.Config.Get("redis").Get("port").Int()
	rdDb, _ := GxMisc.Config.Get("redis").Get("db").Int64()
	rdClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", rdHost, rdPort),
		Password: "",   // no password set
		DB:       rdDb, // use default DB
	})
	return rdClient != nil
}

func connectDb() {
	mysqlHost, _ := GxMisc.Config.Get("db").Get("host").String()
	mysqlPort, _ := GxMisc.Config.Get("db").Get("port").Int()
	mysqlUser, _ := GxMisc.Config.Get("db").Get("user").String()
	mysqlPwd, _ := GxMisc.Config.Get("db").Get("pwd").String()
	mysqlDb, _ := GxMisc.Config.Get("db").Get("db").String()
	mysqlCharset, _ := GxMisc.Config.Get("db").Get("charset").String()

	connInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", mysqlUser, mysqlPwd, mysqlHost, mysqlPort, mysqlDb, mysqlCharset)

	fmt.Println("connect to", connInfo)

	var err error
	db, err = sql.Open("mysql", connInfo)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	// defer db.Close()
}

func systemTable() {
	connectDb()

	str := GxMisc.GenerateCreateSQL(&GxStatic.Player{}, "")

	_, err := db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("new system table ok")
}

func serverTable() {
	if len(os.Args) != 3 {
		fmt.Println("paramter error")
		fmt.Println("GxTool server-table <server-ID>")
		return
	}

	connectDb()

	//role
	str := GxMisc.GenerateCreateSQL(&GxStatic.Role{}, os.Args[2])
	fmt.Println(str)
	_, err := db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Role{}, os.Args[2], "1")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	//operate log
	str = GxMisc.GenerateCreateSQL(&GxStatic.Operate{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Operate{}, os.Args[2], "1")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	//recharge
	str = GxMisc.GenerateCreateSQL(&GxStatic.Recharge{}, "")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Recharge{}, "", "1")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Recharge{}, "", "2")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	str = GxMisc.GenerateCreateSQL(&GxStatic.Bag{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	str = GxMisc.GenerateCreateSQL(&GxStatic.Card{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	str = GxMisc.GenerateCreateSQL(&GxStatic.FightCard{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	/*Jie*/
	//ExtractCard
	str = GxMisc.GenerateCreateSQL(&GxStatic.BuyCard{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	/*Written by Jie 2015-12-24*/
	//Tasks
	str = GxMisc.GenerateCreateSQL(&GxStatic.Tasks{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	/*Written by Jie 2015-12-26*/
	//TaskSchedule
	str = GxMisc.GenerateCreateSQL(&GxStatic.TaskSchedule{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	//chapter
	str = GxMisc.GenerateCreateSQL(&GxStatic.Chapter{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	//mail
	str = GxMisc.GenerateCreateSQL(&GxStatic.Mail{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	//friend
	str = GxMisc.GenerateCreateSQL(&GxStatic.Friend{}, os.Args[2])
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("new server table ok")
}

func newServer() {
	if len(os.Args) != 6 {
		fmt.Println("paramter error")
		fmt.Println("GxTool new-server <server-ID> <server-name> <server-status> <open-time>")
		return
	}

	if !connectRedis() {
		fmt.Println("connect redis fail")
		return
	}

	ID, _ := strconv.Atoi(os.Args[2])
	status, _ := strconv.Atoi(os.Args[4])
	opents := GxMisc.StrToTime(os.Args[5])
	if opents == 0 {
		fmt.Println("open time format error")
		return
	}

	server := &GxStatic.GameServer{
		ID:     ID,
		Name:   os.Args[3],
		Status: uint32(status),
		OpenTs: opents,
	}

	err := GxStatic.SaveGameServer(rdClient, server)
	if err != nil {
		fmt.Println("SaveGameServer 1 error: %s", err)
		return
	}

	fmt.Println("new server ok, info:", server)
}

func test() {
	connectDb()

	str := GxMisc.GenerateCreateSQL(&GxStatic.Recharge{}, "")
	fmt.Println(str)
	_, err := db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Recharge{}, "", "1")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
	str = GxMisc.GenerateIndexSQL(&GxStatic.Recharge{}, "", "2")
	fmt.Println(str)
	_, err = db.Exec(str)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("paramter error")
		fmt.Println("GxTool new-server <server-ID> <server-name> <server-status> <open-time>")
		fmt.Println("GxTool system-table")
		fmt.Println("GxTool server-table <server-ID>")
		return
	}

	GxMisc.LoadConfig("config.json")

	if os.Args[1] == "system-table" {
		systemTable()
		return
	} else if os.Args[1] == "server-table" {
		serverTable()
		return
	} else if os.Args[1] == "new-server" {
		newServer()
		return
	} else if os.Args[1] == "test" {
		test()
		return
	} else {
		fmt.Println("paramter error")
		fmt.Println("GxTool system-table")
		fmt.Println("GxTool server-table <server-ID>")
	}
}
