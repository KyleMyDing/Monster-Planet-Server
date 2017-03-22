package main

/**
作者： Kyle Ding
模块：角色体力管理模块
说明：
创建时间：2015-11-2
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

var powerTvl int64 = 1800

//GetFreePowerCallback 领取免费体力
func GetFreePowerCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	GxMisc.Trace("get GetFreePowerReq request, role: %s", role.String())

	timeStr, _ := GxDict.GetDictString("CommonConfig", 1, "FreeVigorTime")
	timeArr := strings.Split(timeStr, ",")
	power, _ := GxDict.GetDictInt("CommonConfig", 1, "FreeVigorNum")
	now := time.Now().Unix()
	count := len(timeArr) / 2
	for i := 0; i < count; i++ {
		beginHour, _ := strconv.Atoi(timeArr[i*2])
		endHour, _ := strconv.Atoi(timeArr[i*2+1])
		begin := GxMisc.NextTime(0, 0, 0) + int64(time.Hour)*int64(beginHour-24)
		end := GxMisc.NextTime(0, 0, 0) + int64(time.Hour)*int64(endHour-24)

		GxMisc.Debug("check power time, role: %s, begin: %s, end: %s, now: %s", role.String(), GxMisc.TimeToStr(begin), GxMisc.TimeToStr(end), GxMisc.TimeToStr(now))
		if now < begin || now > end {
			//现在不是领取免费体力的时间
			GxMisc.Warn("now is not free power time, role: %s", role.String())
			sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotFreePowerTime, nil)
			return
		}
		if role.GetFreePowerTs >= begin && role.GetFreePowerTs <= end {
			//体力已经领取过了
			GxMisc.Warn("free power has been get, role: %s", role.String())
			sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFreePowerBeGet, nil)
			return
		}
	}

	var rsp GxProto.GetFreePowerRsp
	rsp.Info = new(GxProto.RespondInfo)

	role.GetFreePowerTs = now
	role.SetField4Redis(client, "GetFreePowerTs")

	GetItem(client, runInfo, GxStatic.IDPower, power, rsp.Info)
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//BuyPowerCallback 购买体力
func BuyPowerCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	GxMisc.Trace("get BuyPowerReq request, role: %s", role.String())

	costStr, _ := GxDict.GetDictString("CommonConfig", 1, "BuyVigorCost")
	costArr := strings.Split(costStr, ",")
	power, _ := GxDict.GetDictInt("CommonConfig", 1, "IncreaseVigor")
	max := len(costArr)

	count := role.BuyPowerCnt
	if role.BuyPowerTs < GxMisc.NextTime(0, 0, 0)-int64(time.Hour)*24 {
		//上次购买不是今天
		count = 0
	} else {
		if count >= max {
			//已经超过今天的最大领取次数
			GxMisc.Warn("today buy power count is max, role: %s， count: %d", role.String(), count)
			sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMaxBuyPowerCnt, nil)
			return
		}
	}
	gold, _ := strconv.Atoi(costArr[count])
	count++

	if role.Gold < int64(gold) {
		//元宝不足
		GxMisc.Warn("gold is not enough, role: %s, gold: %d", role.String(), role.Gold)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetGoldNotEnough, nil)
		return
	}

	var rsp GxProto.BuyPowerRsp
	rsp.Info = new(GxProto.RespondInfo)

	DelItem(client, runInfo, -1, GxStatic.IDGold, gold, rsp.Info)
	role.BuyPowerCnt = count
	role.SetField4Redis(client, "BuyPowerCnt")
	role.BuyPowerTs = time.Now().Unix()
	role.SetField4Redis(client, "BuyPowerTs")
	rsp.BuyPowerCnt = proto.Int(role.BuyPowerCnt)
	rsp.BuyPowerTs = proto.Int64(role.BuyPowerTs)

	GetItem(client, runInfo, GxStatic.IDPower, power, rsp.Info)
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//addPower 增加体力 返回实际增加的体力，如果大于最大值，删除定时器事件
//调用后自己更新redis
func addPower(runInfo *RoleRunInfo, cnt int64, limit bool) int64 {
	tmp, _ := GxDict.GetDictInt("Playergrowup", runInfo.R.GetLev(), "PowerUpper")
	max := int64(tmp)
	if runInfo.R.Power >= max && limit {
		return 0
	}
	n := cnt
	runInfo.R.Power += cnt
	if runInfo.R.Power > max && limit {
		n -= runInfo.R.Power - max //实际增加的体力
		runInfo.R.Power = max
		runInfo.T.DelTimer(runInfo.R.ID, GxStatic.TimerPower)
	}

	return n
}

//addTimer4Power 增加一个恢复体力的定时起，nextTime为下次恢复时间
func addTimer4Power(runInfo *RoleRunInfo, nextTime int64) {
	runInfo.T.AddTimer(runInfo.R.ID, GxStatic.TimerPower, nextTime, func(roleID int, eventID int) int64 {
		rdClient := GxMisc.PopRedisClient()
		defer GxMisc.PushRedisClient(rdClient)
		var ret int64 = 0
		RefRole(roleID, func(rri *RoleRunInfo) {
			if rri == nil {
				GxMisc.Error("role goroutine has not been running, roleID: %d", roleID)
			} else {
				if timerAddPower(rdClient, runInfo) {
					ret = 0
				} else {
					ret = time.Now().Unix() + powerTvl
				}
			}
		})
		return ret
	})
}

//subPower 扣除体力，如果小于最大值，触发定时器
//调用后自己更新redis
func subPower(runInfo *RoleRunInfo, cnt int64, rspInfo *GxProto.RespondInfo) uint16 {
	if cnt > runInfo.R.Power {
		return GxStatic.RetPowerNotEnough
	}
	runInfo.R.Power -= cnt
	tmp, _ := GxDict.GetDictInt("Playergrowup", runInfo.R.GetLev(), "PowerUpper")
	max := int64(tmp)
	if max > runInfo.R.Power && (runInfo.R.Power+cnt) > max {
		addTimer4Power(runInfo, time.Now().Unix()+powerTvl)
	}
	return GxStatic.RetSucc
}

//timerAddPower 每30分钟恢复2点体力,如果达到最大限制，返回true
func timerAddPower(client *redis.Client, runInfo *RoleRunInfo) bool {
	tmp, _ := GxDict.GetDictInt("Playergrowup", runInfo.R.GetLev(), "PowerUpper")
	max := int64(tmp)
	if runInfo.R.Power >= max {
		return true
	}

	now := time.Now().Unix()
	var tvl int64 = 0
	if now > runInfo.R.PowerTs {
		tvl = (now - runInfo.R.PowerTs) / powerTvl
	}
	GxMisc.Debug("check power, role: %s, now power: %d, tvl: %d, now: %d, PowerTs: %d", runInfo.R.String(), runInfo.R.Power, tvl, now, runInfo.R.PowerTs)
	if tvl != 0 {
		power := addPower(runInfo, tvl*2, true)
		if power > 0 {
			notify := new(GxProto.PowerAddNotify)
			notify.Info = new(GxProto.RespondInfo)
			fillObtain(-1, GxStatic.IDPower, int(power), notify.GetInfo())
			GxNet.SendPbMessage(runInfo.Conn, 0, runInfo.Info.ConnID, GxStatic.CmdPowerAdd, 0, GxStatic.RetSucc, notify)
		}
		runInfo.R.PowerTs += tvl * powerTvl
		runInfo.R.SetField4Redis(client, "Power")
		runInfo.R.SetField4Redis(client, "PowerTs")
	}
	return runInfo.R.Power >= max
}

func CheckPower(runInfo *RoleRunInfo, client *redis.Client) uint16 {
	if timerAddPower(client, runInfo) {
		return GxStatic.RetSucc
	}

	//计算下一次恢复体力是什么时候
	nextTime := runInfo.R.PowerTs + powerTvl
	addTimer4Power(runInfo, nextTime)
	return GxStatic.RetSucc
}
