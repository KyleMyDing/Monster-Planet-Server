package main

/**
作者： Kyle Ding
模块：征战消息管理模块
说明：
创建时间：2015-11-12
**/

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//GetLevelInfoCallback 读取章节关卡信息
func GetLevelInfoCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.GetLevelInfoReq
	var rsp GxProto.GetLevelInfoRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	GxMisc.Trace("get GetLevelInfoReq request, role: %s, msg: %s", role.String(), req.String())
	chapterId := int(req.GetChapterId())
	//if (*GxDict.Dicts)["Chapter"].D[chapterId] == nil {
	if !GxDict.CheckDictId("Chapter", chapterId) {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	chapter := GxStatic.NewChapter(role.ID)
	chapter.Get4Redis(client)

	levelStr, _ := GxDict.GetDictString("Chapter", chapterId, "HaveLevels")
	levels := strings.Split(levelStr, ",")
	for i := 0; i < len(levels); i++ {
		levelId, _ := strconv.Atoi(levels[i])
		pointStr, _ := GxDict.GetDictString("ChapterLevel", levelId, "HaveBattles")
		points := strings.Split(pointStr, ",")

		pass := 1
		for j := 0; j < len(points); j++ {
			pointId, _ := strconv.Atoi(points[j])
			if !chapter.PointExists(client, chapterId, levelId, pointId) {
				pass = 0
				break
			}
		}

		rsp.Levels = append(rsp.Levels, &GxProto.PbLevelInfo{
			Id:   proto.Int(levelId),
			Pass: proto.Int(pass),
		})
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//GetPointInfoCallback 读取关卡战斗点信息
func GetPointInfoCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.GetPointInfoReq
	var rsp GxProto.GetPointInfoRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get GetPointInfoReq request, role: %s, msg: %s", role.String(), req.String())

	chapterId := int(req.GetChapterId())
	levelId := int(req.GetLevelId())
	if !GxDict.CheckDictId("Chapter", chapterId) || !GxDict.CheckDictId("ChapterLevel", levelId) {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	chapter := GxStatic.NewChapter(role.ID)
	chapter.Get4Redis(client)

	pointStr, _ := GxDict.GetDictString("ChapterLevel", levelId, "HaveBattles")
	points := strings.Split(pointStr, ",")
	for i := 0; i < len(points); i++ {
		pointId, _ := strconv.Atoi(points[i])
		point := chapter.GetPoint(client, chapterId, levelId, pointId)
		if point == nil {
			continue
		}
		rsp.Points = append(rsp.Points, &GxProto.PbPointInfo{
			Id:       proto.Int(levelId),
			PassCnt:  proto.Int(point.PassCnt),
			FightCnt: proto.Int(point.FightCnt),
		})
	}

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//ChapterShuffleCallback 战斗洗牌
func ChapterShuffleCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {

}

//ChapterBattleBeginCallback 征战战斗开始
func ChapterBattleBeginCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ChapterBattleBeginReq
	var rsp GxProto.ChapterBattleBeginRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ChapterBattleBeginReq request, role: %s, msg: %s", role.String(), req.String())

	chapterId := int(req.GetChapterId())
	levelId := int(req.GetLevelId())
	pointId := int(req.GetPointId())
	if !GxDict.CheckDictId("Chapter", chapterId) || !GxDict.CheckDictId("ChapterLevel", levelId) || //
		!GxDict.CheckDictId("ChapterLevelBattle", pointId) {
		GxMisc.Warn("msg format error, role: %s, msg: %s", role.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return
	}

	chapterLev, _ := GxDict.GetDictInt("Chapter", chapterId, "RequirePlayerLevel")
	levelLev, _ := GxDict.GetDictInt("ChapterLevelBattle", pointId, "RequirePlayerLevel")
	power, _ := GxDict.GetDictInt("ChapterLevelBattle", pointId, "CostVigor")
	lev := role.GetLev()
	if lev < chapterLev || lev < levelLev {
		//等级不足
		GxMisc.Warn("level is not enough, roleID: %d, lev: %d, chapterLev: %d, levelLev: %d", role.ID, lev, chapterLev, levelLev)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetLevNotEnough, nil)
		return
	}

	if role.Power < int64(power) {
		//体力不足
		GxMisc.Warn("level is not enough, roleID: %d, lev: %d, chapterLev: %d, levelLev: %d", role.ID, lev, chapterLev, levelLev)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetPowerNotEnough, nil)
		return
	}

	chapter := GxStatic.NewChapter(role.ID)
	chapter.GetFightChapter(client)

	if chapter.PointId != 0 {
		//正在征战副本中
		GxMisc.Warn("role is fighting Chapter, role: %s", role.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetChapterFighting, nil)
		return
	}
	chapter.ChapterId = chapterId
	chapter.LevelId = levelId
	chapter.PointId = pointId
	chapter.SetFightChapter(client)
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

//ChapterBattleEndCallback 征战战斗结束
func ChapterBattleEndCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	role := runInfo.R
	var req GxProto.ChapterBattleEndReq
	var rsp GxProto.ChapterBattleEndRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ChapterBattleEndReq request, role: %s, msg: %s", role.String(), req.String())

	chapter := GxStatic.NewChapter(role.ID)
	chapter.GetFightChapter(client)
	if chapter.PointId == 0 {
		//你没有征战副本中
		GxMisc.Warn("role is not fighting Chapter, role: %s", role.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetChapterNotFighting, nil)
		return
	}

	point := chapter.GetPoint(client, chapter.ChapterId, chapter.LevelId, chapter.PointId)
	if point == nil {
		point = &GxStatic.ChapterPointInfo{0, 0}
	}

	if req.GetWin() == 1 {
		rsp.Info = new(GxProto.RespondInfo)

		//胜利 扣除体力
		power, _ := GxDict.GetDictInt("ChapterLevelBattle", chapter.PointId, "CostVigor")
		DelItem(client, runInfo, -1, GxStatic.IDPower, power, rsp.GetInfo())

		//更新副本信息，更新掉落
		point.PassCnt++
		point.FightCnt++

		//检查当前章节是否完成
		nextPointId, _ := GxDict.GetDictInt("ChapterLevelBattle", chapter.PointId, "SufBattle")
		if nextPointId != -1 {
			nextLevelId, _ := GxDict.GetDictInt("ChapterLevelBattle", chapter.PointId, "BelongLevel")
			nextChapterId, _ := GxDict.GetDictInt("ChapterLevel", chapter.PointId, "Chapter")

			if !chapter.PointExists(client, nextChapterId, nextLevelId, nextPointId) {
				chapter.NowChapterId = nextChapterId
				chapter.NowLevelId = nextLevelId
				chapter.NowPointId = nextPointId

				chapter.SetChapter(client)
			}
		}

		var items []*GxDict.DropItem
		if point.PassCnt == 1 {
			items = GxDict.Random4ChapterDrop(chapter.PointId, true)
		} else {
			items = GxDict.Random4ChapterDrop(chapter.PointId, false)
		}
		for i := 0; i < len(items); i++ {
			GetItem(client, runInfo, items[i].ID, items[i].Cnt, rsp.GetInfo())
		}

		rsp.ChapterId = proto.Int(chapter.ChapterId)
		rsp.LevelId = proto.Int(chapter.LevelId)
		rsp.PointId = proto.Int(chapter.PointId)
	} else {
		point.FightCnt++
	}

	//更新征战信息

	chapter.SetPoint(client, chapter.ChapterId, chapter.LevelId, chapter.PointId, point)
	chapter.ChapterId = 0
	chapter.LevelId = 0
	chapter.PointId = 0
	chapter.SetFightChapter(client)

	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}
