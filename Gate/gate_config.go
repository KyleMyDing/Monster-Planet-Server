package main

import (
	"fmt"
	"os"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//GateConfig 网关配置
type GateConfig struct {
	ID        int
	PprofPort int
	LogLevel  int
	Host1     string
	Port1     int
	Host2     string
	Port2     int

	MemoryPool  int
	MessagePool int

	DbHost    string
	DbPort    int
	DbUser    string
	DbPwd     string
	DbDb      string
	DbCharset string

	RedisHost string
	RedisPort int
	RedisDb   int64
}

var config *GateConfig

//LoadGateConfig 加载网关配置函数,在程序启动时候调用
func LoadGateConfig() error {
	err := GxMisc.LoadConfig(os.Args[1])
	if err != nil {
		return err
	}

	config = new(GateConfig)

	config.ID, _ = GxMisc.Config.Get("server").Get("id").Int()
	config.PprofPort, _ = GxMisc.Config.Get("server").Get("pprofPort").Int()
	config.LogLevel, _ = GxMisc.Config.Get("server").Get("logLevel").Int()

	config.Host1, _ = GxMisc.Config.Get("server").Get("host1").String()
	config.Port1, _ = GxMisc.Config.Get("server").Get("port1").Int()
	config.Host2, _ = GxMisc.Config.Get("server").Get("host2").String()
	config.Port2, _ = GxMisc.Config.Get("server").Get("port2").Int()

	config.MemoryPool, _ = GxMisc.Config.Get("server").Get("memoryPool").Int()
	config.MessagePool, _ = GxMisc.Config.Get("server").Get("messagePool").Int()

	config.DbHost, _ = GxMisc.Config.Get("db").Get("host").String()
	config.DbPort, _ = GxMisc.Config.Get("db").Get("port").Int()
	config.DbUser, _ = GxMisc.Config.Get("db").Get("user").String()
	config.DbPwd, _ = GxMisc.Config.Get("db").Get("pwd").String()
	config.DbDb, _ = GxMisc.Config.Get("db").Get("db").String()
	config.DbCharset, _ = GxMisc.Config.Get("db").Get("charset").String()

	config.RedisHost, _ = GxMisc.Config.Get("redis").Get("host").String()
	config.RedisPort, _ = GxMisc.Config.Get("redis").Get("port").Int()
	config.RedisDb, _ = GxMisc.Config.Get("redis").Get("db").Int64()

	fmt.Println("=================config info=====================")
	fmt.Println("ID         : ", config.ID)
	fmt.Println("PprofPort  : ", config.PprofPort)
	fmt.Println("LogLevel   : ", config.LogLevel)
	fmt.Println("Host1      : ", config.Host1)
	fmt.Println("Port1      : ", config.Port1)
	fmt.Println("Host2      : ", config.Host2)
	fmt.Println("Port2      : ", config.Port2)
	fmt.Println("MemoryPool : ", config.MemoryPool)
	fmt.Println("MessagePool: ", config.MessagePool)
	fmt.Println("DbHost     : ", config.DbHost)
	fmt.Println("DbPort     : ", config.DbPort)
	fmt.Println("DbUser     : ", config.DbUser)
	fmt.Println("DbPwd      : ", config.DbPwd)
	fmt.Println("DbDb       : ", config.DbDb)
	fmt.Println("DbCharset  : ", config.DbCharset)
	fmt.Println("RedisHost  : ", config.RedisHost)
	fmt.Println("RedisPort  : ", config.RedisPort)
	fmt.Println("RedisDb    : ", config.RedisDb)
	fmt.Println("=================================================")

	return nil
}
