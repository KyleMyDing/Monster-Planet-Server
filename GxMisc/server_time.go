package GxMisc

/**
作者： Kyle Ding
模块：系统运行时间管理
说明：
创建时间：2015-12-10
**/

import (
	"time"
)

var serverRunTs int64

func init() {
	serverRunTs = time.Now().Unix()
}

func GetRunSecond() int64 {
	return time.Now().Unix() - serverRunTs
}
