package main

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
	// "time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxNet"
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

func fillLoginRsp(rdClient *redis.Client, player *GxStatic.Player, rsp *GxProto.LoginServerRsp) {
	rsp.Info = &GxProto.LoginRspInfo{
		Token: proto.String(player.SaveToken(rdClient)),
	}

	var gates []*GxStatic.GateInfo
	GxStatic.GetAllGate(rdClient, &gates)
	var gate *GxStatic.GateInfo = nil
	for i := 0; i < len(gates); i++ {
		// if (time.Now().Unix() - gates[i].Ts) > 20 {
		// 	continue
		// }
		if gate == nil {
			gate = gates[i]
		} else {
			if gate.Count > gates[i].Count {
				gate = gates[i]
			}
		}
	}

	if gate != nil {
		rsp.GetInfo().Host = proto.String(gate.Host1)
		rsp.GetInfo().Port = proto.Uint32(uint32(gate.Port1))
	}

	serverID, ts := GxStatic.GetPlayerLastServer(rdClient, player.Username)
	var servers []*GxStatic.GameServer
	GxStatic.GetAllGameServer(rdClient, &servers)
	for i := 0; i < len(servers); i++ {
		GxMisc.Debug("server-ID: %d, name: %s", servers[i].ID, servers[i].Name)
		var lastts int64 = 0
		if serverID == servers[i].ID {
			lastts = ts
		}
		rsp.GetInfo().Srvs = append(rsp.GetInfo().Srvs, &GxProto.GameSrvInfo{
			Index:  proto.Uint32(uint32(servers[i].ID)),
			Name:   proto.String(servers[i].Name),
			Status: proto.Uint32(servers[i].Status),
			Lastts: proto.Uint32(uint32(lastts)),
		})
	}

}

func login(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxMisc.Debug("UnpackagePbmsg error")
		return errors.New("close")
	}
	if req.GetRaw() == nil {
		GxMisc.Debug("login message miss filed: raw")
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return errors.New("close")
	}

	GxMisc.Debug("login player, username: %s, pwd: %s", req.GetRaw().GetUsername(), req.GetRaw().GetPwd())

	player := FindPlayer(req.GetRaw().GetUsername())
	if player == nil {
		GxMisc.Debug("user is not exists, username: %s", req.GetRaw().GetUsername())
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetUserNotExists, nil)
		return errors.New("close")
	}

	if !GxStatic.VerifyPassword(player, req.GetRaw().GetPwd()) {
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetPwdError, nil)
		return errors.New("close")
	}

	GxMisc.Debug("old user: %s login from %s", req.GetRaw().GetUsername(), conn.Remote)
	fillLoginRsp(rdClient, player, &rsp)

	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	return errors.New("close")
}

func register(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		GxMisc.Debug("UnpackagePbmsg error")
		return errors.New("close")
	}

	if req.GetRaw() == nil {
		GxMisc.Debug("register message miss filed: raw")
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetMsgFormatError, nil)
		return errors.New("close")
	}

	GxMisc.Debug("new player, username: %s, pwd: %s", req.GetRaw().GetUsername(), req.GetRaw().GetPwd())

	player := FindPlayer(req.GetRaw().GetUsername())
	if player != nil {
		GxMisc.Debug("user has been exists, username: %s", req.GetRaw().GetUsername())
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetUserExists, nil)
		return errors.New("close")
	}

	player = GxStatic.NewPlayer(rdClient, req.GetRaw().GetUsername(), req.GetRaw().GetPwd(), uint32(req.GetPt()))
	err = AddPlayer(player)
	if err != nil {
		GxMisc.Debug("user has been exists, username: %s", req.GetRaw().GetUsername())
		GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetUserExists, nil)
		return errors.New("close")
	}

	fillLoginRsp(rdClient, player, &rsp)

	GxMisc.Debug("new user: %s login from %s", req.GetRaw().GetUsername(), conn.Remote)
	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)

	return errors.New("close")
}

func getGatesInfo(conn *GxNet.GxTCPConn, msg *GxMessage.GxMessage) error {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	var rsp GxProto.GetGatesInfoRsp
	var gates []*GxStatic.GateInfo
	GxStatic.GetAllGate(rdClient, &gates)
	for i := 0; i < len(gates); i++ {
		rsp.Host = append(rsp.Host, gates[i].Host1)
		rsp.Port = append(rsp.Port, uint32(gates[i].Port1))
	}
	GxNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), GxStatic.RetSucc, &rsp)
	return errors.New("close")
}
