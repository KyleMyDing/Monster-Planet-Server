package GxMisc

/**
作者： Kyle Ding
模块：配置读写接口模块
说明：
创建时间：2015-10-30
**/

import (
	"github.com/bitly/go-simplejson"
	"os"
)

//Config 配置文件
var Config *simplejson.Json

//LoadConfig 配置文件初始化函数,在程序启动时候调用
func LoadConfig(filename string) error {
	buf := make([]byte, 1024)
	f, err := os.Open(filename)
	if err != nil {
		// fmt.Printf("Error: %s\n", err)
		return err
	}
	defer f.Close()
	f.Read(buf)
	//
	Config, _ = simplejson.NewJson(buf)
	return nil
}
