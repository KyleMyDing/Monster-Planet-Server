package main

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"os"
	"strconv"
	"time"
)
import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//LoginServerAddr 登录服务器地址
var LoginServerAddr = "127.0.0.1:9000"

//GateServerAddr 网关服务器地址
var GateServerAddr = "127.0.0.1:13001"

var username = "guang6"
var pwd = "guang123"
var token = "7f4dab1fc8925f0790197dcbbfe02a86"

var serverID = 1
var roleID = 0

var counter *GxMisc.Counter

var quit chan int

func stressEnroll() {
	count, _ := strconv.Atoi(os.Args[7])
	q := make(chan int, count)
	for i := 0; i < count; i++ {
		go func(i int) {
			conn := GxNet.NewTCPConn()
			err := conn.Connect(LoginServerAddr)
			if err != nil {
				GxMisc.Debug("new connnect, remote: %s", err)
				return
			}

			var req GxProto.LoginServerReq
			var rsp GxProto.LoginServerRsp

			req.Raw = &GxProto.PlayerRaw{
				Username: proto.String(username + fmt.Sprintf("%d", i)),
				Pwd:      proto.String(pwd),
			}

			pt := GxProto.PlatformType_GC_PT_91_ASSISTANT
			req.Pt = &pt
			GxNet.SendPbMessage(conn, 0, 0, GxStatic.CmdLogin, uint16(counter.Genarate()), 0, &req)
			//
			msg, err := conn.Recv()
			if err != nil {
				GxMisc.Debug("recv error, %s", err)
				return
			}
			err = msg.UnpackagePbmsg(&rsp)
			if err != nil {
				GxMisc.Debug("unpackage error, %s", err)
				return
			}
			if msg.GetRet() != 0 {
				GxMisc.Error("test_register fail, errorID: %d", msg.GetID())
				q <- 0
				return
			}
			q <- 1
		}(i)
	}
	ret := 0
	for i := 0; i < count; i++ {
		ret += <-q
	}
	fmt.Println("count: ", count, ", ret: ", ret)
}

func stressLogin() {
	begin := time.Now().UnixNano()
	stressEnroll()
	end := time.Now().UnixNano()
	fmt.Println("=================", (end-begin)/1000000, (end-begin)%1000000)
	// begin := time.Now().UnixNano()
	// count, _ := strconv.Atoi(os.Args[4])
	// q := make(chan int, count)
	// ok := 0
	// fail := 0
	// for i := 0; i < count; i++ {
	// 	conn := GxNet.NewTCPConn()
	// 	err := conn.Connect(LoginServerAddr)
	// 	if err != nil {
	// 		GxMisc.Debug("new connnect, remote: %s", err)
	// 		return
	// 	}

	// 	go func(conn *GxNet.GxTCPConn) {
	// 		defer conn.Conn.Close()
	// 		defer func() {
	// 			q <- 1
	// 		}()

	// 		msg, err := conn.Recv()
	// 		if err != nil {
	// 			fail++
	// 			GxMisc.Error("recv error, %s", err)
	// 			return
	// 		}
	// 		// if msg.get
	// 		if msg.GetRet() == GxStatic.RetRoleExists {
	// 			ok++
	// 		}

	// 	}(conn)
	// 	testLogin1(conn)
	// }
	// end1 := time.Now().UnixNano()
	// fmt.Println("=================", (end1-begin)/1000000, (end1-begin)%1000000)
	// for {
	// 	select {
	// 	case <-q:
	// 		count--
	// 		if count == 0 {
	// 			end := time.Now().UnixNano()
	// 			fmt.Println("=================", ok, fail, (end-begin)/1000000, (end-begin)%1000000)
	// 			return
	// 		}
	// 		break
	// 	}
	// }

}

func testHeartbeat(conn *GxNet.GxTCPConn) error {
	GxNet.SendPbMessage(conn, 0, 0, GxStatic.CmdHeartbeat, uint16(counter.Genarate()), 0, nil)
	return nil
}

func testLogin1(conn *GxNet.GxTCPConn) error {
	var req GxProto.LoginServerReq
	// var rsp LoginServerRsp

	req.Raw = &GxProto.PlayerRaw{
		Username: proto.String(username),
		Pwd:      proto.String(pwd),
	}
	GxNet.SendPbMessage(conn, 0, 0, GxStatic.CmdLogin, uint16(counter.Genarate()), 0, &req)
	return nil
}

func testLogin(conn *GxNet.GxTCPConn) error {
	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp

	req.Raw = &GxProto.PlayerRaw{
		Username: proto.String(username),
		Pwd:      proto.String(pwd),
	}
	GxNet.SendPbMessage(conn, 0, 0, GxStatic.CmdLogin, uint16(counter.Genarate()), 0, &req)

	msg, err := conn.Recv()
	if err != nil {
		GxMisc.Error("recv error, %s", err)
		return err
	}

	err = msg.UnpackagePbmsg(&rsp)
	if err != nil {
		GxMisc.Error("unpackage error, %s", err)
		return err
	}
	GxMisc.Debug("recv buff msg, info: %s\n\t%s", msg.String(), rsp.String())
	if msg.GetRet() != 0 {
		GxMisc.Error("test_login fail, errorID: %d", msg.GetID())
		return errors.New("test_login fail")
	}
	token = rsp.GetInfo().GetToken()
	GateServerAddr = rsp.GetInfo().GetHost() + ":" + strconv.Itoa(int(rsp.GetInfo().GetPort()))

	return nil
}

func testRegister(conn *GxNet.GxTCPConn) error {
	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp

	req.Raw = &GxProto.PlayerRaw{
		Username: proto.String(username),
		Pwd:      proto.String(pwd),
	}

	pt := GxProto.PlatformType_GC_PT_91_ASSISTANT
	req.Pt = &pt
	GxNet.SendPbMessage(conn, 0, 0, GxStatic.CmdRegister, uint16(counter.Genarate()), 0, &req)
	//
	msg, err := conn.Recv()
	if err != nil {
		GxMisc.Debug("recv error, %s", err)
		return err
	}
	err = msg.UnpackagePbmsg(&rsp)
	if err != nil {
		GxMisc.Debug("unpackage error, %s", err)
		return err
	}
	if msg.GetRet() != 0 {
		GxMisc.Error("test_register fail, errorID: %d", msg.GetID())
		return errors.New("test_register fail")
	}
	GxMisc.Debug("recv buff msg, info: %s\n\t%s", msg.String(), rsp.String())
	token = rsp.GetInfo().GetToken()
	GateServerAddr = rsp.GetInfo().GetHost() + ":" + strconv.Itoa(int(rsp.GetInfo().GetPort()))
	return nil
}

func testGetRoleList(conn *GxNet.GxTCPConn) {
	var req GxProto.GetRoleListReq
	var rsp GxProto.GetRoleListRsp
	req.Info = &GxProto.RequestInfo{
		Token: proto.String(token),
	}

	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetRoleList, uint16(counter.Genarate()), 0, &req)
	//
	msg, err := conn.Recv()
	if err != nil {
		GxMisc.Debug("recv error, %s", err)
		return
	}
	err = msg.UnpackagePbmsg(&rsp)
	if err != nil {
		GxMisc.Debug("unpackage error, %s", err)
		return
	}
	GxMisc.Debug("recv buff msg, info: %s\n\t%s", msg.String(), rsp.String())
	if len(rsp.Roles) > 0 {
		roleID = int(rsp.Roles[0].GetId())
	}
}

func testSelectRole(conn *GxNet.GxTCPConn) error {
	var req GxProto.SelectRoleReq
	var rsp GxProto.SelectRoleRsp
	req.Info = &GxProto.RequestInfo{
		Token: proto.String(token),
	}
	req.RoleId = proto.Int(roleID)

	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdSelectRole, uint16(counter.Genarate()), 0, &req)
	//
	msg, err := conn.Recv()
	if err != nil {
		GxMisc.Debug("recv error, %s", err)
		return err
	}
	err = msg.UnpackagePbmsg(&rsp)
	if err != nil {
		GxMisc.Debug("unpackage error, %s", err)
		return err
	}
	if msg.GetRet() != 0 {
		GxMisc.Error("test_select_role fail, errorID: %d", msg.GetID())
		return errors.New("test_select_role fail")
	}
	GxMisc.Debug("recv buff msg, info: %s\n\t%s", msg.String(), rsp.String())
	return nil
}

func testCreateRole(conn *GxNet.GxTCPConn) {
	var req GxProto.CreateRoleReq
	var rsp GxProto.CreateRoleRsp

	req.Info = &GxProto.RequestInfo{
		Token: proto.String(token),
	}
	req.Name = proto.String(username)
	req.Sex = proto.Int(1)

	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdCreateRole, uint16(counter.Genarate()), 0, &req)
	//
	msg, err := conn.Recv()
	if err != nil {
		GxMisc.Debug("recv error, %s", err)
		return
	}
	err = msg.UnpackagePbmsg(&rsp)
	if err != nil {
		GxMisc.Debug("unpackage error, %s", err)
		return
	}
	GxMisc.Debug("recv buff msg, info: %s\n\t%s", msg.String(), rsp.String())
}

////////////////////////////////////////////////////////////////

func recvMsg(conn *GxNet.GxTCPConn) {
	for {
		msg, err := conn.Recv()
		if err != nil {
			GxMisc.Debug("recv error, %s", err)
			quit <- 1
			quit <- 1
			return
		}

		GxMisc.Debug("<<==== recv buff msg, info: %s", msg.String())
	}
}

func testUseItem(conn *GxNet.GxTCPConn) {
	var req GxProto.UseItemReq
	req.Indx = proto.Int(3)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdUseItem, uint16(counter.Genarate()), 0, &req)
}

func testSellItem(conn *GxNet.GxTCPConn) {
	var req GxProto.SellItemReq
	req.Indx = proto.Int(16)
	req.Cnt = proto.Int(1000)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdSellItem, uint16(counter.Genarate()), 0, &req)
}

func testReplaceEquipment(conn *GxNet.GxTCPConn) {
	var req GxProto.ReplaceEquipmentReq
	req.Indx = proto.Int(0)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdReplaceEquipment, uint16(counter.Genarate()), 0, &req)
}

func testUnloadEquipment(conn *GxNet.GxTCPConn) {
	var req GxProto.UnloadEquipmentReq
	req.Eid = proto.Int(16001)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdUnloadEquipment, uint16(counter.Genarate()), 0, &req)
}

/*Jie*/
//testExtractCardType    testing ExtractCard function
func testExtractCardType(conn *GxNet.GxTCPConn) {
	var req GxProto.ExtractCardReq
	var test int32
	test = 304
	req.ExtractCardType = &test
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdExtractCard, uint16(counter.Genarate()), 0, &req)
}

//testReceiveCardType   testing ReceiveCard function
func testReceiveCardType(conn *GxNet.GxTCPConn) {
	var req GxProto.ReceiveCardReq
	var test int32
	test = 202
	req.ReceiveCardType = &test
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdReceiveCard, uint16(counter.Genarate()), 0, &req)
}

func testExtractSkillCardType(conn *GxNet.GxTCPConn) {
	var req GxProto.ExtractSkillCardReq
	var extractSkillCardType int32
	extractSkillCardType = 505
	req.ExtractSkillCardType = &extractSkillCardType
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdExtractSkillCard, uint16(counter.Genarate()), 0, &req)
}

func testReceiveTaskReward(conn *GxNet.GxTCPConn) {
	var req GxProto.ReceiveTaskRewardReq
	var receiveTaskRewardType int32
	receiveTaskRewardType = 94001
	req.ReceiveTaskRewardType = &receiveTaskRewardType
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdReceiveTaskReward, uint16(counter.Genarate()), 0, &req)
}

func testOrderBag(conn *GxNet.GxTCPConn) {
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdOrderBag, uint16(counter.Genarate()), 0, nil)
}

func testGetFriend(conn *GxNet.GxTCPConn) {
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetFriend, uint16(counter.Genarate()), 0, nil)
}

func testGetReferrer(conn *GxNet.GxTCPConn) {
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetReferrer, uint16(counter.Genarate()), 0, nil)
}

func testAddFriend(conn *GxNet.GxTCPConn) {
	var req GxProto.AddFriendReq
	req.RoleId = append(req.RoleId, 10000003)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdAddFriend, uint16(counter.Genarate()), 0, &req)
}

func testDealFriend(conn *GxNet.GxTCPConn) {
	var req GxProto.DealFriendReq
	req.RoleId = append(req.RoleId, 10000004)
	req.Agree = proto.Int(1)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdDealFriend, uint16(counter.Genarate()), 0, &req)
}

func testGetMailList(conn *GxNet.GxTCPConn) {
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetMailList, uint16(counter.Genarate()), 0, nil)
}

func testGetMailItem(conn *GxNet.GxTCPConn) {
	var req GxProto.GetMailItemReq

	req.Id = append(req.Id, 2)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetMailItem, uint16(counter.Genarate()), 0, &req)
}

func testSendChat(conn *GxNet.GxTCPConn) {
	var req GxProto.SendChatReq
	req.RoleId = proto.Int(10000007)
	req.Type = proto.Int(3) //1-世界频道 2-军团频道 3-私聊频道
	req.Text = proto.String("aaaaa")
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdSendChat, uint16(counter.Genarate()), 0, &req)
}

func testRecharge(conn *GxNet.GxTCPConn) {
	var req GxProto.RechargeReq

	req.RechargeType = proto.Int(10001)
	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdRecharge, uint16(counter.Genarate()), 0, &req)
}

func main() {
	GxMisc.InitLogger("client")

	if len(os.Args) < 7 {
		GxMisc.Info("gxclient <op> <host> <port> <serverID> <username> <pwd>")
		GxMisc.Info("0-register")
		GxMisc.Info("1-login")
		GxMisc.Info("2-getRoleList")
		GxMisc.Info("3-SelectRole")
		GxMisc.Info("4-CreateRole")
		GxMisc.Info("5-stressLogin")
		return
	}

	LoginServerAddr = os.Args[2] + ":" + os.Args[3]
	serverID, _ = strconv.Atoi(os.Args[4])
	username = os.Args[5]
	pwd = os.Args[6]
	counter = GxMisc.NewCounter()
	quit = make(chan int, 100)

	conn := GxNet.NewTCPConn()
	err := conn.Connect(LoginServerAddr)
	if err != nil {
		GxMisc.Debug("new connnect, remote: %s", err)
		return
	}
	defer conn.Conn.Close()

	// if conn.ClientKey() != nil {
	// 	conn.Conn.Close()
	// 	return
	// }

	connGate := GxNet.NewTCPConn()
	// f := func(conn *GxNet.GxTCPConn) {

	// }

	switch os.Args[1] {
	case "0":
		testRegister(conn)
		return
	case "1":
		testLogin(conn)
		return
	case "2":
		err = testLogin(conn)
		if err != nil {
			return
		}
		GxMisc.Debug("connect %s", GateServerAddr)
		err = connGate.Connect(GateServerAddr)
		if err != nil {
			GxMisc.Error("new connnect, remote: %s", err)
			return
		}
		defer connGate.Conn.Close()
		testGetRoleList(connGate)
		return
	case "3":
		err = testLogin(conn)
		if err != nil {
			return
		}
		GxMisc.Debug("connect %s", GateServerAddr)
		err = connGate.Connect(GateServerAddr)
		if err != nil {
			GxMisc.Error("new connnect, remote: %s", err)
			return
		}
		defer connGate.Conn.Close()
		go heartbeat(connGate)

		testGetRoleList(connGate)
		if roleID == 0 {
			GxMisc.Debug("role has not been created")
			return
		}
		testSelectRole(connGate)

		GxMisc.Debug("==========begin process test msg============\n")
		go recvMsg(connGate)
		// test_use_item(connGate)
		//testSellItem(connGate)
		/*Jie*/

		//testExtractCardType(connGate)

		// testExtractCardType(connGate)

		//testReceiveCardType(connGate)
		//testExtractSkillCardType(connGate)
		// func(conn *GxNet.GxTCPConn) {
		// 	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdBuyCardInfo, uint16(counter.Genarate()), 0, nil)
		// }(connGate)

		// func(conn *GxNet.GxTCPConn) {
		// 	GxNet.SendPbMessage(conn, 0, uint32(serverID), GxStatic.CmdGetTaskList, uint16(counter.Genarate()), 0, nil)
		// }(connGate)
		testReceiveTaskReward(connGate)

		// testOrderBag(connGate)
		//testSendChat(connGate)
		//testRecharge(connGate)

		// testGetFriend(connGate)
		// testGetReferrer(connGate)
		// testAddFriend(connGate)
		// testDealFriend(connGate)

		// testGetMailList(connGate)
		// testGetMailItem(connGate)

		// t := time.NewTicker(5 * time.Second)
		// test_unload_equipment(connGate)
		for {
			select {
			// case <-t.C:
			// 	testSendChat(connGate)
			case <-quit:
				GxMisc.Warn("test exit")
				return
			}
		}

		return
	case "4":
		err = testLogin(conn)
		if err != nil {
			return
		}
		GxMisc.Debug("connect %s", GateServerAddr)
		err = connGate.Connect(GateServerAddr)
		if err != nil {
			GxMisc.Error("new connnect, remote: %s", err)
			return
		}
		defer connGate.Conn.Close()
		testGetRoleList(connGate)
		if roleID != 0 {
			GxMisc.Debug("role has been created")
			return
		}
		testCreateRole(connGate)
		return
	}
	return
}

func heartbeat(conn *GxNet.GxTCPConn) {
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			testHeartbeat(conn)
		case <-quit:
			return
		}
	}
}
