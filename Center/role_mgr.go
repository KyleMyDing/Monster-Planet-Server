package main

/**
作者： Kyle Ding
模块：角色基础信息管理模块
说明：
创建时间：2015-11-02
**/

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
	"sync"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//RoleRunInfo 角色运行时信息，每个角色登录时候，都会运行一个goroutine来处理他的所有操作
type RoleRunInfo struct {
	Status    int                       //0-未认证 1-未选择角色 2-游戏状态
	R         *GxStatic.Role            //role信息
	Info      *GxStatic.LoginInfo       //网关连接信息
	Conn      *GxNet.GxTCPConn          //tcp连接信息
	MsgQue    chan *GxMessage.GxMessage //要处理的消息队列
	Quit      chan int                  //登出消息队列
	Init      chan int                  //登录角色时初始化角色信息消息队列
	T         *GxMisc.GxTimer           //定时器
	NotifyQue chan *GxMessage.GxMessage //内部消息或者通知消息
	Remote    string
}

//RolesRunInfo 所有角色运行列表，主键为GateID:ConnId
type RolesRunInfo map[string]*RoleRunInfo

//RoleMessageCallback 角色需要处理消息的回调
type RoleMessageCallback func(*RoleRunInfo, *redis.Client, *GxMessage.GxMessage)

var runInfoMutex sync.Mutex
var rolesRunInfo RolesRunInfo

//rolesGateInfo  角色id对应GateID:ConnId表
var rolesGateInfo map[int]string

//回调
var logicCb map[uint16]RoleMessageCallback

func registerCallback(cmd uint16, cb RoleMessageCallback) {
	_, ok := logicCb[cmd]
	if ok {
		GxMisc.Error("callback has been registered, cmd: %d", cmd)
		return
	}
	logicCb[cmd] = cb
}

func init() {
	rolesGateInfo = make(map[int]string)
	rolesRunInfo = make(RolesRunInfo)
	logicCb = make(map[uint16]RoleMessageCallback)

	registerCallback(GxStatic.CmdGetRoleList, GetRoleListCallback)
	registerCallback(GxStatic.CmdSelectRole, SelectRoleCallback)
	registerCallback(GxStatic.CmdCreateRole, CreateRoleCallback)
	registerCallback(GxStatic.CmdClientLogout, ClientLogout)
	registerCallback(GxStatic.CmdSelectVocation, SelectVocationCallback)

	registerCallback(GxStatic.CmdSendChat, SendChatCallback)
	registerCallback(GxStatic.CmdUseItem, UseItemCallback)
	registerCallback(GxStatic.CmdSellItem, SellItemCallback)

	registerCallback(GxStatic.CmdReplaceEquipment, ReplaceEquipmentCallback)
	registerCallback(GxStatic.CmdUnloadEquipment, UnloadEquipmentCallback)
	registerCallback(GxStatic.CmdStrengthenEquipment, StrengthenEquipmentCallback)

	registerCallback(GxStatic.CmdFuseCard, FuseCardCallback)
	registerCallback(GxStatic.CmdDecomposeCard, DecomposeCardCallback)
	registerCallback(GxStatic.CmdAddFightCard, AddFightCardCallback)
	registerCallback(GxStatic.CmdDelFightCard, DelFightCardCallback)
	registerCallback(GxStatic.CmdNewFightCardBag, NewFightCardBagCallback)
	registerCallback(GxStatic.CmdDelFightCardBag, DelFightCardBagCallback)

	registerCallback(GxStatic.CmdOrderBag, OrderBagCallback)
	registerCallback(GxStatic.CmdExpandBag, ExpandBagCallback)

	/*Written by Jie 抽卡*/
	registerCallback(GxStatic.CmdExtractCard, ExtractCardCallback)
	registerCallback(GxStatic.CmdReceiveCard, ReceiveCardCallback)
	registerCallback(GxStatic.CmdExtractSkillCard, ExtractSkillCardCallback)
	registerCallback(GxStatic.CmdBuyCardInfo, GetBuyCardInfoCallback)

	registerCallback(GxStatic.CmdGetFreePower, GetFreePowerCallback)
	registerCallback(GxStatic.CmdBuyPower, BuyPowerCallback)

	//征战
	registerCallback(GxStatic.CmdGetLevelInfo, GetLevelInfoCallback)
	registerCallback(GxStatic.CmdGetPointInfo, GetPointInfoCallback)
	registerCallback(GxStatic.CmdChapterShuffle, ChapterShuffleCallback)
	registerCallback(GxStatic.CmdChapterBattleBegin, ChapterBattleBeginCallback)
	registerCallback(GxStatic.CmdChapterBattleEnd, ChapterBattleEndCallback)

	//邮件
	registerCallback(GxStatic.CmdGetMailList, GetMailListCallback)
	registerCallback(GxStatic.CmdGetMailItem, GetMailItemCallback)
	registerCallback(GxStatic.CmdReadMail, ReadMailCallback)

	//好友
	registerCallback(GxStatic.CmdGetFriend, GetFriendCallback)
	registerCallback(GxStatic.CmdGetReferrer, GetReferrerCallback)
	registerCallback(GxStatic.CmdAddFriend, AddFriendCallback)
	registerCallback(GxStatic.CmdDealFriend, DealFriendCallback)
	registerCallback(GxStatic.CmdDeleteFriend, DelelteFriendCallback)
	registerCallback(GxStatic.CmdFriendSend, FriendSendCallback)
	registerCallback(GxStatic.CmdFriendRecv, FriendRecvCallback)

	/*Written by Jie 任务*/
	registerCallback(GxStatic.CmdGetTaskList, GetTaskListCallback)
	registerCallback(GxStatic.CmdReceiveTaskReward, ReceiveTaskRewardCallback)
}

func checkRoleRunInfo(conn *GxNet.GxTCPConn, gateID int, connId int) *RoleRunInfo {
	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()

	//如果角色运行时信息已经存在，则返回空
	runInfo, ok := rolesRunInfo[fmt.Sprintf("%d:%d", gateID, connId)]
	if ok {
		return nil
	}

	//没有则新建一个，并初始化
	runInfo = new(RoleRunInfo)
	runInfo.MsgQue = make(chan *GxMessage.GxMessage, 100) //数量？
	runInfo.Quit = make(chan int, 1)
	runInfo.Init = make(chan int, 1)
	runInfo.T = GxMisc.NewGxTimer()
	runInfo.Conn = conn
	runInfo.Info = &GxStatic.LoginInfo{
		GateID: gateID,
		ConnID: uint32(connId),
	}
	runInfo.R = nil
	runInfo.Status = 0
	runInfo.NotifyQue = make(chan *GxMessage.GxMessage, 10)

	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)
	runInfo.Remote = GxStatic.GetRemote4Gate(rdClient, gateID, connId)

	rolesRunInfo[fmt.Sprintf("%d:%d", gateID, connId)] = runInfo
	go roleRun(runInfo)
	return runInfo
}

//roleRun 角色运行主函数，每个角色登录上来，都会生成一个RoleRunInfo，并且分配一个新的goroutine运行roleRun函数
func roleRun(runInfo *RoleRunInfo) {
	GxMisc.Info("=====> role[%s] goroutine start run, connInfo: %d:%d", runInfo.Remote, runInfo.Info.GateID, runInfo.Info.ConnID)
	for {
		select {
		case msg := <-runInfo.MsgQue:
			//角色消息处理
			if msg.GetCmd() >= GxStatic.CmdRoleBegin && runInfo.Status != 2 {
				sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetNotLogin, nil)
				GxMessage.FreeMessage(msg)
				GxMisc.Debug("role has not been login, connInfo: %d:%d", runInfo.Info.GateID, runInfo.Info.ConnID)
				break
			}

			cb, ok := logicCb[msg.GetCmd()]
			if ok {
				client := GxMisc.PopRedisClient()
				cb(runInfo, client, msg)
				GxMisc.PushRedisClient(client)
			} else {
				sendMessage(runInfo, nil, 0, msg.GetID(), msg.GetCmd(), msg.GetSeq(), GxStatic.RetMessageNotSupport, nil)
				GxMisc.Debug("message has not been registered, msg: %s", msg.String())
			}
			//这个消息要回收
			GxMessage.FreeMessage(msg)
		case msg := <-runInfo.NotifyQue:
			//内部消息或者通知
			if msg == nil {
				continue
			}
			checkInternalMessage(runInfo, msg)
			GxMessage.FreeMessage(msg)
		case <-runInfo.Init:
			//角色登录初始化或者数据验证
			roleInit(runInfo)
		case <-runInfo.Quit:
			//角色登出
			GxMisc.Debug("<===== stop role goroutine, roleID: %d", runInfo.R.ID)
			client := GxMisc.PopRedisClient()
			GxStatic.DelGateLoginInfo(client, runInfo.Info.GateID, runInfo.Info.ConnID)
			runInfo.Info.Del4Redis(client)
			GxMisc.PushRedisClient(client)

			runInfoMutex.Lock()
			delete(rolesRunInfo, fmt.Sprintf("%d:%d", runInfo.Info.GateID, runInfo.Info.ConnID))
			delete(rolesGateInfo, runInfo.R.ID)
			runInfoMutex.Unlock()
			close(runInfo.MsgQue)
			close(runInfo.Quit)
			close(runInfo.Init)
			runInfo.T.T.Stop()
			return
		case <-runInfo.T.T.C:
			//定时器
			runInfo.T.Run()
		}
	}
}

//roleInit 定义各种定时器任务,和一些初始化或者检查数据的任务
func roleInit(runInfo *RoleRunInfo) {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	//检查体力
	CheckPower(runInfo, rdClient)
	//检查是否有物品未领取
	CheckNewItem(runInfo, rdClient)
	//检查是否有新邮件
	CheckNewMail(runInfo, rdClient)
	//检查好友通知
	CheckFriend(runInfo, rdClient)
	//检查充值到账通知
	CheckRecharge(runInfo, rdClient)
}

//checkInternalMessage 内部消息或者通知
func checkInternalMessage(runInfo *RoleRunInfo, msg *GxMessage.GxMessage) {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)
	switch msg.GetCmd() {
	case GxStatic.CmdAdminInfoChange:
		CheckNewItem(runInfo, rdClient)
	case GxStatic.CmdAdminNewMail:
		CheckNewMail(runInfo, rdClient)
	case GxStatic.CmdAdminAddFriend:
		CheckAddFriend(runInfo, rdClient, int(msg.GetCmd()))
	case GxStatic.CmdAdminDealFriendYes:
		CheckDealFriendYes(runInfo, rdClient, int(msg.GetCmd()))
	case GxStatic.CmdAdminDealFriendNo:
		CheckDealFriendNo(runInfo, rdClient, int(msg.GetCmd()))
	case GxStatic.CmdAdminDeleteFriend:
		CheckDeleteFriend(runInfo, rdClient, int(msg.GetCmd()))
	case GxStatic.CmdAdminFriendSend:
		CheckFriendSend(runInfo, rdClient, int(msg.GetCmd()))
	case GxStatic.CmdAdminRecharge:
		CheckRecharge(runInfo, rdClient)
	default:
		//通知消息，直接转发回客户端
		runInfo.Conn.Send(msg)
	}
}

func roleMsg(gateID int, connId int, msg *GxMessage.GxMessage) {
	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()
	runInfo, ok := rolesRunInfo[fmt.Sprintf("%d:%d", gateID, connId)]
	if ok {
		runInfo.MsgQue <- msg
	} else {
		GxMisc.Error("role goroutine has not been running, connInfo: %d:%d", gateID, connId)
	}
}

func NewSendMessage(id uint32, cmd uint16, seq uint16, ret uint16, pb proto.Message) *GxMessage.GxMessage {
	msg := GxMessage.GetGxMessage()

	msg.SetID(id)
	msg.SetCmd(cmd)
	msg.SetRet(ret)
	msg.SetSeq(seq)
	msg.ClearMask()

	if pb == nil {
		msg.SetLen(0)
	} else {
		err := msg.PackagePbmsg(pb)
		if err != nil {
			GxMessage.FreeMessage(msg)
			GxMisc.Debug("PackagePbmsg error")
			return nil
		}
	}
	return msg
}

//sendMessage 统一使用此函数往客户端发送消息
func sendMessage(runInfo *RoleRunInfo, req proto.Message, mask uint16, ID uint32, cmd uint16, seq uint16, ret uint16, rsp proto.Message) {
	GxNet.SendPbMessage(runInfo.Conn, mask, ID, cmd, seq, ret, rsp)

	//保存操作
	GxStatic.PutRoleOperateLog(runInfo.Info, runInfo.Remote, req, mask, cmd, seq, ret, rsp)

	//检查任务

	// if rsp != nil {
	// 	checkTask(runInfo, mask, ID, cmd, seq, ret, rsp)
	// }

}

func checkTask(runInfo *RoleRunInfo, mask uint16, ID uint32, cmd uint16, seq uint16, ret uint16, rsp proto.Message) {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	info := GxStatic.GetInfo(rsp)

	if info != nil {
		var notify GxProto.CompleteTaskRsp

		for _, v := range info.Obtain.Items {
			//判断是否为武将卡牌
			if GxStatic.IsHeroCard(int(v.GetItem().GetId())) {

				tSchedule := GxStatic.NewTaskSchedule(runInfo.R.ID)
				iColor, _ := GxDict.GetDictInt("HerosCard", int(v.GetItem().GetId()), "Color")

				if iColor == GxStatic.ColorBlue {
					//判断是否为蓝色武将卡牌
					GxMisc.Debug("Is a BuleHeroCard,ID: %d", v.GetItem().GetId())
					checkHeroCardTask(runInfo, rdClient, tSchedule, &notify, v, "BlueHerosCnt", checkTaskSchedule(runInfo, rdClient, tSchedule, 94001))
				} else if iColor == GxStatic.ColorPurple {
					//判断是否为紫色武将卡牌
					GxMisc.Debug("Is a PurpleHeroCard,ID: %d", v.GetItem().GetId())
					checkHeroCardTask(runInfo, rdClient, tSchedule, &notify, v, "PurpleHerosCnt", checkTaskSchedule(runInfo, rdClient, tSchedule, 94002))

				} else if iColor == GxStatic.ColorOrange {
					//判断是否为橙色武将卡牌
					GxMisc.Debug("Is a OrangeHerosCard,ID: %d", v.GetItem().GetId())
					checkHeroCardTask(runInfo, rdClient, tSchedule, &notify, v, "OrangeHerosCnt", checkTaskSchedule(runInfo, rdClient, tSchedule, 94003))
				}
			}
		}
		if len(notify.CompleteTaskType) > 0 || len(notify.PostTask) > 0 {
			GxNet.SendPbMessage(runInfo.Conn, mask, ID, cmd, seq, GxStatic.RetCompleteTask, &notify)
		}
	}
}

//checkTaskSchedule 使用递归判断任务的进度
func checkTaskSchedule(runInfo *RoleRunInfo, rdClient *redis.Client, tSchedule *GxStatic.TaskSchedule, taskID int) int {
	flag, err := tSchedule.GetIsCompleteTaskAndReward4Redis(rdClient, taskID)
	if err != nil {
		GxMisc.Debug("GetIsCompleteTaskAndReward4Redis error: %s", err)
		return -1
	}
	if flag == 0 || taskID == -1 {
		return taskID
	} else {
		taskID, _ = GxDict.GetDictInt("Tasks", taskID, "SuffixID")
		return checkTaskSchedule(runInfo, rdClient, tSchedule, taskID)
	}
}

func checkHeroCardTask(runInfo *RoleRunInfo, rdClient *redis.Client, tSchedule *GxStatic.TaskSchedule, rsp *GxProto.CompleteTaskRsp, v *GxProto.PbCellInfo, field string, taskType int) {
	if taskType != -1 {
		tasks := GxStatic.NewTasks(runInfo.R.ID)
		//在redis中保存武将卡牌数
		tasks.SetHerosCnt4Redis(rdClient, int(v.GetItem().GetCnt()), field)
		herosCnt, err := tasks.GetHerosCnt4Redis(rdClient, field)
		if err != nil {
			GxMisc.Debug("GetBuleHerosCnt4Redis error: %s", err)
			return
		}
		completeTask, _ := GxDict.GetDictInt("Tasks", taskType, "TaskValue")
		tSchedule := GxStatic.NewTaskSchedule(runInfo.R.ID)
		//判断是否完成武任务
		if herosCnt >= completeTask {
			//  填充完成任务响应中的已完成任务消息
			(*rsp).CompleteTaskType = append(rsp.CompleteTaskType, int32(taskType))
			completedTaskflag := 1
			tSchedule.SetIsCompleteTaskAndReward4Redis(rdClient, taskType, completedTaskflag)
			//填充完成任务响应中的后置任务
			posttask, _ := GxDict.GetDictInt("Tasks", taskType, "SuffixID")
			preTaskStr, _ := GxDict.GetDictString("Tasks", posttask, "PreIDs")
			arr := strings.Split(preTaskStr, ",")
			for _, v := range arr {
				var flag int
				iID, _ := strconv.Atoi(v)
				flag, err = tSchedule.GetIsCompleteTaskAndReward4Redis(rdClient, iID)
				if err != nil {
					GxMisc.Debug("GetIsCompleteTask4Redis error:%s", err)
					return
				}
				if flag == 1 {
					continue
				} else {
					return
				}
			}
			(*rsp).PostTask = append(rsp.PostTask, int32(posttask))
		}
	}
}

//RoleSendMsg 用于给其他角色发送消息
func RoleSendMsg(runInfo *RoleRunInfo, cmd uint16, seq uint16, ret uint16, pb proto.Message) {
	conn := runInfo.Conn
	msg := NewSendMessage(runInfo.Info.ConnID, cmd, seq, ret, pb)
	if msg == nil {
		return
	}

	if pb == nil {
		GxMisc.Debug("====>> remote[%s:%s], info: %s", conn.M, conn.Remote, msg.String())
	} else {
		GxMisc.Debug("====>> remote[%s:%s], info: %s, rsp: \r\n\t%s", conn.M, conn.Remote, msg.String(), pb.String())
	}
	runInfo.NotifyQue <- msg

	GxStatic.PutRoleOperateLog(runInfo.Info, runInfo.Remote, nil, 0, cmd, seq, ret, pb)
}

//OnlineRoleSendMsg 给所有在线的玩家发送消息
func OnlineRoleSendMsg(role *GxStatic.Role, cmd uint16, seq uint16, ret uint16, pb proto.Message) {
	msg := NewSendMessage(0, cmd, seq, ret, pb)
	if msg == nil {
		return
	}
	defer GxMessage.FreeMessage(msg)

	str := pb.String()
	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()
	for _, runInfo := range rolesRunInfo {
		conn := runInfo.Conn
		newMsg := GxMessage.Copy4MessagePool(msg)
		newMsg.SetID(runInfo.Info.ConnID)
		if pb == nil {
			GxMisc.Debug("====>> remote[%s:%s], info: %s", conn.M, runInfo.Remote, newMsg.String())
		} else {
			GxMisc.Debug("====>> remote[%s:%s], info: %s, rsp: \r\n\t%s", conn.M, runInfo.Remote, newMsg.String(), str)
		}
		runInfo.NotifyQue <- newMsg
	}

	//群发消息日志是否入库？？
	// for _, runInfo := range rolesRunInfo {
	// 	putRoleOperateLog(runInfo, nil, 0, cmd, seq, ret, pb)
	// }
}

//RefRole 引用某个角色,回调里不要有锁
func RefRole(roleID int, f func(rri *RoleRunInfo)) {
	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()
	gateInfo, ok := rolesGateInfo[roleID]
	if !ok {
		f(nil)
		return
	}
	var rri *RoleRunInfo
	rri, ok = rolesRunInfo[gateInfo]
	if ok {
		f(rri)
	} else {
		f(nil)
	}
}

//RoleIsOnline 检查角色是否在线
func RoleIsOnline(roleID int) bool {
	runInfoMutex.Lock()
	defer runInfoMutex.Unlock()
	_, ok := rolesGateInfo[roleID]
	return ok
}

//CheckAddFriend 检查是否有新好友添加通知
func CheckAddFriend(runInfo *RoleRunInfo, client *redis.Client, cmd int) {
	ids := GxStatic.GetRoleUndealFriendLMessageist(client, runInfo.R.ID, cmd)
	if len(ids) == 0 {
		return
	}

	friend := GxStatic.NewFriend(runInfo.R.ID)
	friend.Get4Redis(client)

	var notify GxProto.AddFriendNotify
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		if friend.Friends[id] != 0 {
			continue
		}
		friend.Friends[id] = 1 << uint32(GxStatic.FriendMaskUnaccept)
		friend.SetFriend4Redis(client, id)

		newFriend := &GxStatic.Role{
			ID: id,
		}
		newFriend.Get4Redis(client)

		online := 0
		if RoleIsOnline(id) {
			online = 1
		}

		//返回id,name,vocationId,expr,FightValue,online等字段
		notify.Roles = append(notify.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(id),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
			FightValue: proto.Int(0),
			Online:     proto.Int(online),
		})
	}

	if len(notify.Roles) > 0 {
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdAddFriendNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckDealFriendYes 检查是否有新好友同意通知
func CheckDealFriendYes(runInfo *RoleRunInfo, client *redis.Client, cmd int) {
	ids := GxStatic.GetRoleUndealFriendLMessageist(client, runInfo.R.ID, cmd)
	if len(ids) == 0 {
		return
	}

	friend := GxStatic.NewFriend(runInfo.R.ID)
	friend.Get4Redis(client)

	var notify GxProto.DealFriendNotify
	notify.Agree = proto.Int(1)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		friend.Friends[id] = 1 << uint32(GxStatic.FriendMaskAccept)
		friend.SetFriend4Redis(client, id)

		newFriend := &GxStatic.Role{
			ID: id,
		}
		newFriend.Get4Redis(client)

		online := 0
		if RoleIsOnline(id) {
			online = 1
		}

		//返回id,name,vocationId,expr,FightValue,online等字段
		notify.Roles = append(notify.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(id),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
			FightValue: proto.Int(0),
			Online:     proto.Int(online),
		})
	}

	if len(notify.Roles) > 0 {
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdDealFriendNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckDealFriendNo 检查是否有新好友拒绝通知
func CheckDealFriendNo(runInfo *RoleRunInfo, client *redis.Client, cmd int) {
	ids := GxStatic.GetRoleUndealFriendLMessageist(client, runInfo.R.ID, cmd)
	if len(ids) == 0 {
		return
	}

	var notify GxProto.DealFriendNotify
	notify.Agree = proto.Int(0)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		newFriend := &GxStatic.Role{
			ID: id,
		}
		newFriend.Get4Redis(client)

		online := 0
		if RoleIsOnline(id) {
			online = 1
		}

		//返回id,name,vocationId,expr,FightValue,online等字段
		notify.Roles = append(notify.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(id),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
			FightValue: proto.Int(0),
			Online:     proto.Int(online),
		})
	}

	if len(notify.Roles) > 0 {
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdDealFriendNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckDeleteFriend 检查是否有删除好友通知
func CheckDeleteFriend(runInfo *RoleRunInfo, client *redis.Client, cmd int) {
	ids := GxStatic.GetRoleUndealFriendLMessageist(client, runInfo.R.ID, cmd)
	if len(ids) == 0 {
		return
	}

	friend := GxStatic.NewFriend(runInfo.R.ID)
	friend.Get4Redis(client)

	var notify GxProto.DeleteFriendNotify
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		friend.DeleteFriend4Redis(client, id)

		newFriend := &GxStatic.Role{
			ID: id,
		}
		newFriend.Get4Redis(client)

		//返回id,name,vocationId,expr,FightValue,online等字段
		notify.Roles = append(notify.Roles, &GxProto.RoleCommonInfo{
			Id:         proto.Int(id),
			Name:       proto.String(newFriend.Name),
			VocationId: proto.Int(newFriend.VocationID),
			Expr:       proto.Int64(newFriend.Expr),
		})
	}

	if len(notify.Roles) > 0 {
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdDeleteFriendNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckDeleteFriend 检查是否有好友赠送物品通知
func CheckFriendSend(runInfo *RoleRunInfo, client *redis.Client, cmd int) {
	ids := GxStatic.GetRoleUndealFriendLMessageist(client, runInfo.R.ID, cmd)
	if len(ids) == 0 {
		return
	}

	friend := GxStatic.NewFriend(runInfo.R.ID)
	friend.Get4Redis(client)

	var notify GxProto.FriendSendNotify
	for i := 0; i < len(ids); i++ {
		id := ids[i]

		//收到赠送物品通知，更新自己的可领取物品状态
		friend.Friends[id] |= 1 << uint32(GxStatic.FriendMaskRecv)
		friend.SetFriend4Redis(client, id)

		notify.RoleId = append(notify.RoleId, int32(id))
	}

	if len(notify.RoleId) > 0 {
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdFriendSendNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckFriend 检查好友通知
func CheckFriend(runInfo *RoleRunInfo, client *redis.Client) {
	CheckAddFriend(runInfo, client, GxStatic.CmdAdminAddFriend)
	CheckDealFriendYes(runInfo, client, GxStatic.CmdAdminDealFriendYes)
	CheckDealFriendNo(runInfo, client, GxStatic.CmdAdminDealFriendNo)
	CheckDeleteFriend(runInfo, client, GxStatic.CmdAdminDeleteFriend)
	CheckFriendSend(runInfo, client, GxStatic.CmdAdminFriendSend)
}

func CheckRecharge(runInfo *RoleRunInfo, client *redis.Client) {
	recharges := GxStatic.GetRoleRechargeList(client, runInfo.R.ID)
	if len(recharges) == 0 {
		return
	}

	for i := 0; i < len(recharges); i++ {
		recharge := recharges[i]
		GxStatic.PutRechargeLog(recharge)

		var notify GxProto.RechargeNotify
		notify.Info = new(GxProto.RespondInfo)
		notify.RechargeType = proto.Int(recharge.RechargeType)
		GetItem(client, runInfo, GxStatic.IDGold, recharge.Gold, notify.GetInfo())
		sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdRechargeNotify, 0, GxStatic.RetSucc, &notify)
	}
}

//CheckNewMail 检查是否有新邮件
func CheckNewMail(runInfo *RoleRunInfo, client *redis.Client) {
	mails := GxStatic.GetRoleUngetMailList(client, runInfo.R.ID)
	if len(mails) == 0 {
		return
	}

	for i := 0; i < len(mails); i++ {
		GxStatic.SetRoleMail(client, mails[i].RoleID, mails[i])
	}
	var notify GxProto.NewMailNotify
	mailsToPbMails(&mails, &notify.Mails)
	sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdNewMail, 0, GxStatic.RetSucc, &notify)
}

//CheckNewItem 检查新物品
func CheckNewItem(runInfo *RoleRunInfo, client *redis.Client) {
	items := GxStatic.GetRoleUngetItemList(client, runInfo.R.ID)
	if len(items) == 0 {
		return
	}

	var notify GxProto.InfoChangeNotify
	notify.Info = new(GxProto.RespondInfo)
	for i := 0; i < len(items); i++ {
		GetItem(client, runInfo, items[i].ID, items[i].Cnt, notify.GetInfo())
	}
	sendMessage(runInfo, nil, 0, runInfo.Info.ConnID, GxStatic.CmdInfoChange, 0, GxStatic.RetSucc, &notify)
}

//DelItem 角色消耗物品
func DelItem(client *redis.Client, runInfo *RoleRunInfo, index, ID, cnt int, rspInfo *GxProto.RespondInfo) {
	switch ID {
	case GxStatic.IDGold: // 元宝	40001
		if runInfo.R.Gold >= int64(cnt) {
			runInfo.R.Gold -= int64(cnt)
		} else {
			runInfo.R.Gold = 0
		}
		runInfo.R.SetField4Redis(client, "Gold")
		break
	case GxStatic.IDMoney: // 金币	40002
		if runInfo.R.Money >= int64(cnt) {
			runInfo.R.Money -= int64(cnt)
		} else {
			runInfo.R.Money = 0
		}
		runInfo.R.SetField4Redis(client, "Money")
		break
	case GxStatic.IDPower: // 体力	40003
		subPower(runInfo, int64(cnt), nil)
		runInfo.R.SetField4Redis(client, "Power")
		break
	case GxStatic.IDHonor: // 荣誉	40004
		if runInfo.R.Honor >= int64(cnt) {
			runInfo.R.Honor -= int64(cnt)
		} else {
			runInfo.R.Honor = 0
		}
		runInfo.R.SetField4Redis(client, "Honor")
		break
	case GxStatic.IDPay: // 个人贡献	40005
		break
	case GxStatic.IDArmyPay: // 军团贡献	40006
		if runInfo.R.ArmyPay >= int64(cnt) {
			runInfo.R.ArmyPay -= int64(cnt)
		} else {
			runInfo.R.ArmyPay = 0
		}
		runInfo.R.SetField4Redis(client, "ArmyPay")
		break
	case GxStatic.IDCoin: // 回廊币	40007
		break
	case GxStatic.IDExpr: // 经验	40008
		break
	default:
		if GxStatic.IsEquipment(ID) || GxStatic.IsItem(ID) {
			runInfo.R.DelBagCell(client, index, ID, cnt)
		} else if GxStatic.IsCard(ID) {
			runInfo.R.DelCard(client, ID, cnt)
		} else {
			GxMisc.Error("new item error, role: %S, ID: %d", runInfo.R.String(), ID)
			return
		}
		break
	}
	GxMisc.Debug("[####] role: %s delete item[%d %d %d]", runInfo.R.String(), index, ID, cnt)
	fillConsume(index, ID, cnt, rspInfo)
}

//GetItem 角色领取物品
func GetItem(client *redis.Client, runInfo *RoleRunInfo, ID, cnt int, rspInfo *GxProto.RespondInfo) {
	index := -1
	switch ID {
	case GxStatic.IDGold: // 元宝	40001
		runInfo.R.Gold += int64(cnt)
		runInfo.R.SetField4Redis(client, "Gold")
		break
	case GxStatic.IDMoney: // 金币	40002
		runInfo.R.Money += int64(cnt)
		runInfo.R.SetField4Redis(client, "Money")
		break
	case GxStatic.IDPower: // 体力	40003
		addPower(runInfo, int64(cnt), false)
		runInfo.R.SetField4Redis(client, "Power")
		break
	case GxStatic.IDHonor: // 荣誉	40004
		runInfo.R.Honor += int64(cnt)
		runInfo.R.SetField4Redis(client, "Honor")
		break
	case GxStatic.IDPay: // 个人贡献	40005
		break
	case GxStatic.IDArmyPay: // 军团贡献	40006
		runInfo.R.ArmyPay += int64(cnt)
		runInfo.R.SetField4Redis(client, "ArmyPay")
		break
	case GxStatic.IDCoin: // 回廊币	40007
		break
	case GxStatic.IDExpr: // 经验	40008
		runInfo.R.Expr += int64(cnt)
		runInfo.R.SetField4Redis(client, "Expr")
		break
	default:
		if GxStatic.IsEquipment(ID) || GxStatic.IsItem(ID) {
			index = runInfo.R.NewItem(client, ID, cnt)
		} else if GxStatic.IsCard(ID) {
			runInfo.R.NewCard(client, ID, cnt)
		} else {
			GxMisc.Error("new item error, role: %S, ID: %d", runInfo.R.String(), ID)
			return
		}
		break
	}
	GxMisc.Debug("[****] role: %s new item[%d %d %d]", runInfo.R.String(), index, ID, cnt)
	fillObtain(index, ID, cnt, rspInfo)
}

//fillObtain 领取物品填充响应消息
func fillObtain(index, ID, cnt int, rspInfo *GxProto.RespondInfo) {
	if rspInfo != nil {
		if rspInfo.GetObtain() == nil {
			rspInfo.Obtain = new(GxProto.RoleChangeInfo)
		}
		n := -1
		for i := 0; i < len(rspInfo.GetObtain().GetItems()); i++ {
			if index == int(rspInfo.GetObtain().GetItems()[i].GetIndx()) && ID == int(rspInfo.GetObtain().GetItems()[i].GetItem().GetId()) {
				n = i
				break
			}
		}
		if n == -1 {
			rspInfo.GetObtain().Items = append(rspInfo.GetObtain().GetItems(), &GxProto.PbCellInfo{
				Indx: proto.Int(index),
				Item: &GxProto.PbItemInfo{
					Id:  proto.Int(ID),
					Cnt: proto.Int(cnt),
				},
			})
		} else {
			rspInfo.GetObtain().GetItems()[n].GetItem().Cnt = proto.Int(cnt + int(rspInfo.GetObtain().GetItems()[n].GetItem().GetCnt()))
		}

	}
}

//fillConsume 消耗物品填充响应消息
func fillConsume(index, ID, cnt int, rspInfo *GxProto.RespondInfo) {
	if rspInfo != nil {
		if rspInfo.GetConsume() == nil {
			rspInfo.Consume = new(GxProto.RoleChangeInfo)
		}
		n := -1
		for i := 0; i < len(rspInfo.GetConsume().GetItems()); i++ {
			if index == int(rspInfo.GetConsume().GetItems()[i].GetIndx()) && ID == int(rspInfo.GetConsume().GetItems()[i].GetItem().GetId()) {
				n = i
				break
			}
		}
		if n == -1 {
			rspInfo.GetConsume().Items = append(rspInfo.GetConsume().GetItems(), &GxProto.PbCellInfo{
				Indx: proto.Int(index),
				Item: &GxProto.PbItemInfo{
					Id:  proto.Int(ID),
					Cnt: proto.Int(int(cnt)),
				},
			})
		} else {
			rspInfo.GetConsume().GetItems()[n].GetItem().Cnt = proto.Int(cnt + int(rspInfo.GetConsume().GetItems()[n].GetItem().GetCnt()))
		}

	}
}
