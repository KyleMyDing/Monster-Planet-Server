package main

/**
作者： Kyle Ding
模块：游戏服务器配置模块
说明：
创建时间：2015-11-2
**/

import (
	"fmt"
	"os"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//PublicConfig 游戏服务器配置
type PublicConfig struct {
	ID          int
	PprofPort   int
	DictDir     string
	LogLevel    int
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

var config *PublicConfig

//LoadCenterConfig 配置加载函数,程序启动时候调用
func LoadPublicConfig() error {
	err := GxMisc.LoadConfig(os.Args[1])
	if err != nil {
		return err
	}

	config = new(PublicConfig)

	config.ID, _ = GxMisc.Config.Get("id").Int()
	config.PprofPort, _ = GxMisc.Config.Get("pprofPort").Int()
	config.DictDir, _ = GxMisc.Config.Get("dictDir").String()
	config.LogLevel, _ = GxMisc.Config.Get("logLevel").Int()

	config.MemoryPool, _ = GxMisc.Config.Get("memoryPool").Int()
	config.MessagePool, _ = GxMisc.Config.Get("memoryPool").Int()

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
	fmt.Println("ID          : ", config.ID)
	fmt.Println("PprofPort   : ", config.PprofPort)
	fmt.Println("DictDir     : ", config.DictDir)
	fmt.Println("LogLevel    : ", config.LogLevel)
	fmt.Println("MemoryPool  : ", config.MemoryPool)
	fmt.Println("MessagePool : ", config.MessagePool)
	fmt.Println("DbHost      : ", config.DbHost)
	fmt.Println("DbPort      : ", config.DbPort)
	fmt.Println("DbUser      : ", config.DbUser)
	fmt.Println("DbPwd       : ", config.DbPwd)
	fmt.Println("DbDb        : ", config.DbDb)
	fmt.Println("DbCharset   : ", config.DbCharset)
	fmt.Println("RedisHost   : ", config.RedisHost)
	fmt.Println("RedisPort   : ", config.RedisPort)
	fmt.Println("RedisDb     : ", config.RedisDb)
	fmt.Println("=================================================")

	GxStatic.ServerID = config.ID

	return nil
}
