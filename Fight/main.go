package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

import (
	. "git.oschina.net/jkkkls/goxiang/GxMisc"
	. "git.oschina.net/jkkkls/goxiang/GxNet"
)

var r *rand.Rand

func main() {
	if len(os.Args) != 2 {
		fmt.Println("paramter error")
		fmt.Println("gxfight <config-file-name>")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := LoadFightConfig()
	if err != nil {
		Debug("load config fail, err: %s", err)
		return
	}
	InitLogger("fight")

	GeneratePIDFile("gxfight", 0)

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	err = ConnectRedis(config.RedisHost, config.RedisPort, config.RedisDb)
	if err != nil {
		Debug("connect redis fail, err: %s", err)
		return
	}

	err = InitMysql(config.DbUser, config.DbPwd, config.DbHost, config.DbPort, config.DbDb, config.DbCharset)
	if err != nil {
		Debug("connect mysql fail, err: %s", err)
		return
	}
	Debug("connect mysql ok, host: %s:%d", config.DbHost, config.DbPort)

	err = ConnectAllGate(0)
	if err != nil {
		Debug("ConnectAllGate fail, %s", err)
		return
	}
}
