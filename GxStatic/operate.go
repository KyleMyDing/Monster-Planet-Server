package GxStatic

/**
作者： Kyle Ding
模块：操作日志模块
说明：
创建时间：2015-11-12
**/

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//Operate 角色操作信息
type Operate struct {
	PlayerName string `len:"32" index:"1"`
	RoleID     int    //角色ID
	ServerId   int
	GateId     int
	ConnId     uint32
	Ts         int64 `type:"time"` //操作时间
	Cmd        int
	Seq        int
	Ret        int
	Mask       int
	Remote     string
	Req        string `type:"text"` //
	Rsp        string `type:"text"` //
}

var logChan chan *Operate
var marshaler jsonpb.Marshaler

func init() {
	logChan = make(chan *Operate, 1000)
}

func PutRoleOperateLog(info *LoginInfo, remote string, req proto.Message, mask uint16, cmd uint16, seq uint16, ret uint16, rsp proto.Message) {
	if info.PlayerName == "" {
		return
	}
	op := &Operate{
		PlayerName: info.PlayerName,
		GateId:     info.GateID,
		ConnId:     info.ConnID,
		RoleID:     info.RoleID,
		ServerId:   info.ServerID,
		Ts:         time.Now().Unix(),
		Cmd:        int(cmd),
		Seq:        int(seq),
		Ret:        int(ret),
		Mask:       int(mask),
		Remote:     remote,
	}
	if req != nil {
		op.Req, _ = marshaler.MarshalToString(req)
	}
	if rsp != nil {
		op.Rsp, _ = marshaler.MarshalToString(rsp)
	}
	logChan <- op
}

func RunOperateLog() {
	for {
		select {
		case op := <-logChan:
			log := GxMisc.GenerateInsertSQL(op, strconv.Itoa(op.ServerId))

			_, err := GxMisc.MysqlPool.Exec(log)
			if err != nil {
				GxMisc.Error("execute log fail ,error: %s, sql: %s", err, log)
				return
			}
		}
	}

}
