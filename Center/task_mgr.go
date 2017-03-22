/**
作者: Jie
模块：任务消息模块
说明：
创建时间：2015-12-22
**/
package main

import (
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func GetTaskListCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	var rsp GxProto.GetTasksListRsp
	role := runInfo.R
	task := GxStatic.NewTasks(role.ID)
	err := task.Get4Redis(client)
	if err != nil {
		GxMisc.Warn("GetTaskListCallback Get4Redis error, Get4Redis: %s", err)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	tSchedule := GxStatic.NewTaskSchedule(runInfo.R.ID)
	err = tSchedule.Get4Redis(client)
	if err != nil {
		GxMisc.Warn("GetTaskListCallback Get4Redis error, Get4Redis: %s", err)
		sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	//任务具体属性
	rsp.Tasks = new(GxProto.PbTaskInfo)
	rsp.Tasks.PassChapterTs = proto.Int64(task.PassChapterTs)
	rsp.Tasks.PassChapterCnt = proto.Int(task.PassChapterCnt)
	rsp.Tasks.PurpleHerosCnt = proto.Int(task.PassChapterCnt)
	rsp.Tasks.BlueHerosCnt = proto.Int(task.BlueHerosCnt)
	rsp.Tasks.PurpleHerosCnt = proto.Int(task.PurpleHerosCnt)
	rsp.Tasks.OrangeHerosCnt = proto.Int(task.OrangeHerosCnt)
	rsp.Tasks.ReceiveFreePowerCnt = proto.Int(task.ReceiveFreePowerCnt)
	rsp.Tasks.BuyPowerCnt = proto.Int(task.BuyPowerCnt)
	rsp.Tasks.SignInDaily = proto.Int(task.SignInDaily)
	rsp.Tasks.SignInMonthCnt = proto.Int(task.SignInMonthCnt)
	rsp.Tasks.MoneyPointCnt = proto.Int(task.MoneyPointCnt)
	rsp.Tasks.CompetitiveDailyCnt = proto.Int(task.CompetitiveDailyCnt)
	rsp.Tasks.CompetitiveTotalCnt = proto.Int(task.CompetitiveTotalCnt)
	rsp.Tasks.PlayerLevel = proto.Int(task.PlayerLevel)
	rsp.Tasks.IsPassTaskChapter = proto.Int(task.IsPassTaskChapter)
	//任务进度
	for k, v := range tSchedule.IsCompleteTaskAndReward {

		rsp.TaskSchedule = append(rsp.TaskSchedule, &GxProto.TaskState{
			TaskType:                proto.Int(k),
			IsCompleteTaskAndReward: proto.Int(v),
		})
	}
	sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}

func ReceiveTaskRewardCallback(runInfo *RoleRunInfo, client *redis.Client, msg *GxMessage.GxMessage) {
	var req GxProto.ReceiveTaskRewardReq
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}
	GxMisc.Trace("get ReceiveTaskReward request, role: %s, msg: %s", runInfo.R.String(), req.String())

	if !GxDict.CheckDictId("Tasks", int(req.GetReceiveTaskRewardType())) {
		//判断领取任务奖励请求的类型是否在合理的范围内
		GxMisc.Warn("ReceiveTaskRewardType is not exists, role: %s, msg: %s", runInfo.R.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetExtractCardType, nil)
		return
	}
	taskReward := GxStatic.NewTaskSchedule(runInfo.R.ID)
	var flag int
	flag, err = taskReward.GetIsCompleteTaskAndReward4Redis(client, int(req.GetReceiveTaskRewardType()))
	if err != nil {
		GxMisc.Debug("GetIsCompleteTaskAndReward4Redis error: %s", err)
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetFail, nil)
		return
	}

	var rsp GxProto.ReceiveTaskRewardRsp
	rsp.Info = new(GxProto.RespondInfo)

	if flag == 1 {
		//任务完成了并且还没有领取奖励
		items := GxDict.GetTaskReward4TaskRewardTable(int(req.GetReceiveTaskRewardType()))
		size := len(items)
		for i := 0; i < size; i++ {
			GetItem(client, runInfo, items[i].ID, items[i].Cnt, rsp.GetInfo())
		}
		flag = 2
		taskReward.SetIsCompleteTaskAndReward4Redis(client, int(req.GetReceiveTaskRewardType()), flag)
	} else if flag == 2 {
		//任务奖励已经领取了
		GxMisc.Warn("task Reward was received, role: %s, msg: %s", runInfo.R.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetTaskRewardwasReceived, nil)
		return
	} else {
		//任务还没完成
		GxMisc.Warn("task is not complete, role: %s, msg: %s", runInfo.R.String(), req.String())
		sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetTaskNotComplete, nil)
		return
	}
	sendMessage(runInfo, &req, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
}
