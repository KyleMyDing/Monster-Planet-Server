package main

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

var server *GxNet.GxTCPServer

//NewConn 新客户端连接
func NewConn(conn *GxNet.GxTCPConn) {
	GxMisc.Debug("new connnect, remote: %s", conn.Remote)
}

//DisConn 客户端断开连接
func DisConn(conn *GxNet.GxTCPConn) {
	GxMisc.Debug("dis connnect, remote: %s", conn.Remote)
}

//NewMessage 不需要处理的消息,目前Login没有不处理的函数
func NewMessage(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	GxMisc.Debug("new message, remote: %s", conn.Remote)
	conn.Send(msg)
	return errors.New("close")
}

func startServer() {
	server = GxNet.NewGxTCPServer(0x00FFFFFF, NewConn, DisConn, NewMessage, true)
	server.RegisterClientCmd(GxStatic.CmdLogin, login)
	server.RegisterClientCmd(GxStatic.CmdRegister, register)
	server.RegisterClientCmd(GxStatic.CmdGetGatesInfo, getGatesInfo)
	server.Start(":" + strconv.Itoa(config.Port))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("paramter error")
		fmt.Println("gxlogin <config-file-name>")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := LoadGateConfig()
	if err != nil {
		GxMisc.Debug("load config fail, err: %s", err)
		return
	}

	GxMisc.InitLogger("login")
	GxMisc.SetLevel(config.LogLevel)

	GxMisc.GeneratePIDFile("login", 0)

	if config.MemoryPool > 0 {
		GxMisc.Info("open memory pool ok, init-size: %d", config.MemoryPool)
		GxMisc.OpenGxMemoryPool(config.MemoryPool)
	}
	if config.MessagePool > 0 {
		GxMisc.Info("open message pool ok, init-size: %d", config.MessagePool)
		GxMessage.OpenGxMessagePool(config.MessagePool)
	}

	if config.MemoryPool > 0 || config.MessagePool > 0 {
		go func() {
			t := time.NewTicker(3600 * time.Second)

			for {
				select {
				case <-t.C:
					GxMisc.PrintfMemoryPool()
					GxMessage.PrintfMessagePool()
				}
			}
		}()
	}

	err = GxMisc.ConnectRedis(config.RedisHost, config.RedisPort, config.RedisDb)
	if err != nil {
		GxMisc.Debug("connect redis fail, err: %s", err)
		return
	}
	GxMisc.Debug("connect redis ok, host: %s:%d", config.RedisHost, config.RedisPort)

	err = GxMisc.InitMysql(config.DbUser, config.DbPwd, config.DbHost, config.DbPort, config.DbDb, config.DbCharset)
	if err != nil {
		GxMisc.Debug("connect mysql fail, err: %s", err)
		return
	}
	GxMisc.Debug("connect mysql ok, host: %s:%d", config.DbHost, config.DbPort)

	err = LoadPlayer()
	if err != nil {
		GxMisc.Debug("load player fail, err: %s", err)
		return
	}
	GxMisc.Debug("load player ok")

	//pprof
	if config.PprofPort != 0 {
		go func() {
			http.ListenAndServe(":"+strconv.Itoa(config.PprofPort), nil)
		}()
	}

	//
	startServer()
	GxMisc.Debug("connect redis fail")
}
