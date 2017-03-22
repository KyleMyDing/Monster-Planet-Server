package main

/**
作者： Kyle Ding
模块：
说明：
创建时间：2015-10-30
**/

import (
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
)

func startServer() {
	go clientRouter.Start(":" + strconv.Itoa(config.Port1))

	serverRouter.Start(":" + strconv.Itoa(config.Port2))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("paramter error")
		fmt.Println("gxcenter <config-file-name>")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := LoadGateConfig()
	if err != nil {
		GxMisc.Debug("load config fail, err: %s", err)
		return
	}

	GxMisc.InitLogger("gate")
	GxMisc.SetLevel(config.LogLevel)

	GxMisc.GeneratePIDFile("gate", config.ID)

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

	//pprof
	if config.PprofPort != 0 {
		go func() {
			http.ListenAndServe(":"+strconv.Itoa(config.PprofPort), nil)
		}()
	}

	gateRun()
	//
	startServer()
}
