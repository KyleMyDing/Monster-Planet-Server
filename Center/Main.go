package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

var r *rand.Rand

func main() {
	if len(os.Args) != 2 {
		fmt.Println("paramter error")
		fmt.Println("gxcenter <config-file-name>")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	GxMisc.InitLogger("center")
	err := LoadCenterConfig()
	if err != nil {
		GxMisc.Error("load config fail, err: %s", err)
		return
	}
	GxMisc.SetLevel(config.LogLevel)

	GxMisc.GeneratePIDFile("center", config.ID)

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	if config.MemoryPool > 0 {
		GxMisc.Info("open memory pool ok, init-size: %d", config.MemoryPool)
		GxMisc.OpenGxMemoryPool(config.MemoryPool)
	}
	if config.MessagePool > 0 {
		GxMisc.Info("open message pool ok, init-size: %d", config.MessagePool)
		GxMessage.OpenGxMessagePool(config.MessagePool)
	}

	if (config.MemoryPool + config.MessagePool) > 0 {
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

	err = GxMisc.InitMysql(config.DbUser, config.DbPwd, config.DbHost, config.DbPort, config.DbDb, config.DbCharset)
	if err != nil {
		GxMisc.Debug("connect mysql fail, err: %s", err)
		return
	}
	GxMisc.Debug("connect mysql ok, host: %s:%d", config.DbHost, config.DbPort)

	err = GxDict.LoadAllDict(config.DictDir)
	if err != nil {
		GxMisc.Debug("load dict fail, err: %s", err)
		return
	}
	GxMisc.Debug("load dict ok, dir: %s", config.DictDir)

	//同步角色信息
	SyncRole()

	//玩家操作记录入库
	go GxStatic.RunOperateLog()
	go GxStatic.RunRechargeLog()

	//pprof
	if config.PprofPort != 0 {
		go func() {
			http.ListenAndServe(":"+strconv.Itoa(config.PprofPort), nil)
		}()
	}

	err = GxNet.ConnectAllGate(uint32(config.ID))
	if err != nil {
		GxMisc.Debug("ConnectAllGate fail, %s", err)
		return
	}
}
