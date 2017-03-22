package main

/**
作者： Kyle Ding
模块：网关信息更新模块
说明：
创建时间：2015-10-30
**/

import (
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func gateRun() {
	go func() {
		rdClient := GxMisc.PopRedisClient()
		defer GxMisc.PushRedisClient(rdClient)

		t := time.NewTicker(10 * time.Second)
		Self := new(GxStatic.GateInfo)
		Self.ID = config.ID
		Self.Host1 = config.Host1
		Self.Port1 = config.Port1
		Self.Host2 = config.Host2
		Self.Port2 = config.Port2
		Self.Count = 0

		for {
			select {
			case <-t.C:
				//定时更新自己的信息到缓存中
				Self.Ts = time.Now().Unix()
				if clientRouter.ConnectCount() > AdminCount() {
					Self.Count = clientRouter.ConnectCount() - AdminCount()
				} else {
					Self.Count = 0
				}
				GxStatic.SaveGate(rdClient, Self)
			}
		}
	}()
}
