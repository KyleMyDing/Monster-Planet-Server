package GxMessage

/**
作者: guangbo
模块: 消息模块
说明:
创建时间：2015-11-19
**/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

const (
	//MessageMaskDisconn 是否断开连接
	MessageMaskDisconn = 1
	//MessageMaskNotify 是否通知
	MessageMaskNotify = 2
	//MessageMaskInternal 是否内部消息
	MessageMaskInternal = 4
)

const (
	//MessageIDBit 消息来源或者目的ID
	MessageIDBit = 0
	//MessageLenBit 消息体长度，可为0
	MessageLenBit = 4
	//MessageCmdBit 消息命令字
	MessageCmdBit = 6
	//MessageSeqBit 消息序号
	MessageSeqBit = 8
	//MessageRetBit 消息返回值，消息返回时使用
	MessageRetBit = 10
	//MessageMaskBit 一些标志
	MessageMaskBit = 12

	//MessageHeaderLen 消息长度
	MessageHeaderLen = 14
)

//暂时写死加密密钥
var key = "c14faab0"

//GxMessage 消息类，包括消息头和消息体
type GxMessage struct {
	Header []byte
	Data   []byte
}

//NewGxMessage 生成一个新消息
func NewGxMessage() *GxMessage {
	msg := new(GxMessage)
	msg.Header = make([]byte, MessageHeaderLen)
	return msg
}

//CheckFormat 检查消息长度，如果不是8的倍数(因为des加密后长度是8的倍数)，则返回错误
func (msg *GxMessage) CheckFormat() error {
	if (msg.GetLen() % 8) != 0 {
		return errors.New("msg info error")
	}
	return nil
}

//Package 打包原生字符串，目前是des加密
func (msg *GxMessage) Package(buff []byte) error {
	if len(buff) == 0 {
		return nil
	}

	enbuff, _ := GxMisc.DesEncrypt(buff, []byte(key))
	l := len(enbuff)

	msg.FreeDate()
	msg.SetLen(uint16(l))
	msg.InitData()
	copy(msg.Data[:], enbuff)
	return nil
}

//Unpackage 解包原生字符串，目前是des解密
func (msg *GxMessage) Unpackage() ([]byte, error) {
	if msg.GetLen() == 0 {
		return []byte(""), nil
	}
	return GxMisc.DesDecrypt(msg.Data, []byte(key))
}

//PackagePbmsg 打包protobuf消息
func (msg *GxMessage) PackagePbmsg(pb proto.Message) error {
	buff, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	return msg.Package(buff)
}

//UnpackagePbmsg 解包protobuf消息
func (msg *GxMessage) UnpackagePbmsg(pb proto.Message) error {
	data, err := msg.Unpackage()
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, pb)
}

//InitData 根据消息长度初始化消息体内存
func (msg *GxMessage) InitData() {
	if msg.GetLen() == 0 {
		return
	}
	msg.Data = GxMisc.GxMalloc(int(msg.GetLen()))
}

func (msg *GxMessage) FreeDate() {
	if msg.GetLen() == 0 {
		return
	}

	GxMisc.GxFree(msg.Data)
	msg.Data = nil
	msg.SetLen(0)
}

func (msg *GxMessage) get16(t uint32) uint16 {
	buf := bytes.NewBuffer(make([]byte, 0, MessageHeaderLen))
	buf.Write(msg.Header[t : t+2])

	i16 := make([]byte, 2)
	buf.Read(i16)
	return binary.BigEndian.Uint16(i16)
}

func (msg *GxMessage) set16(t uint32, ID uint16) {
	binary.BigEndian.PutUint16(msg.Header[t:t+2], ID)
}

func (msg *GxMessage) get32(t uint32) uint32 {
	buf := bytes.NewBuffer(make([]byte, 0, MessageHeaderLen))
	buf.Write(msg.Header[t : t+4])

	i32 := make([]byte, 4)
	buf.Read(i32)
	return binary.BigEndian.Uint32(i32)
}

func (msg *GxMessage) set32(t uint32, ID uint32) {
	binary.BigEndian.PutUint32(msg.Header[t:t+4], ID)
}

//GetID 返回消息来源或者目的ID
func (msg *GxMessage) GetID() uint32 {
	return msg.get32(MessageIDBit)
}

//SetID 设置消息来源或者目的ID
func (msg *GxMessage) SetID(ID uint32) {
	msg.set32(MessageIDBit, ID)
}

//GetCmd 返回消息命令字
func (msg *GxMessage) GetCmd() uint16 {
	return msg.get16(MessageCmdBit)
}

//SetCmd 设置消息命令字
func (msg *GxMessage) SetCmd(cmd uint16) {
	msg.set16(MessageCmdBit, cmd)
}

//GetSeq 返回消息序列号
func (msg *GxMessage) GetSeq() uint16 {
	return msg.get16(MessageSeqBit)
}

//SetSeq 设置消息序列号
func (msg *GxMessage) SetSeq(seq uint16) {
	msg.set16(MessageSeqBit, seq)
}

//GetLen 返回消息体长度
func (msg *GxMessage) GetLen() uint16 {
	return msg.get16(MessageLenBit)
}

//SetLen 设置消息体长度
func (msg *GxMessage) SetLen(len uint16) {
	msg.set16(MessageLenBit, len)
}

//GetRet 返回消息返回值
func (msg *GxMessage) GetRet() uint16 {
	return msg.get16(MessageRetBit)
}

//SetRet 设置消息返回值
func (msg *GxMessage) SetRet(ret uint16) {
	msg.set16(MessageRetBit, ret)
}

//GetMask 返回消息掩码
func (msg *GxMessage) GetMask(mask uint16) bool {
	i := msg.get16(MessageMaskBit)
	return (i & mask) != 0
}

//SetMask 设置消息掩码
func (msg *GxMessage) SetMask(mask uint16) {
	msg.set16(MessageMaskBit, msg.get16(MessageMaskBit)|mask)
}

//XorMask 异或消息掩码
func (msg *GxMessage) XorMask(mask uint16) {
	msg.set16(MessageMaskBit, msg.get16(MessageMaskBit)^mask)
}

func (msg *GxMessage) ClearMask() {
	msg.set16(MessageMaskBit, 0)
}

func (msg *GxMessage) String() string {
	return fmt.Sprintf("ID: %d, len: %d, cmd: [%d:%s], seq: %d, ret: [%d:%s], mask: %d", msg.GetID(), msg.GetLen(), //
		msg.GetCmd(), GxStatic.CmdString[msg.GetCmd()], msg.GetSeq(), msg.GetRet(), GxStatic.RetString[msg.GetRet()], msg.get16(MessageMaskBit))
}

//Copy 消息复制函数,返回一个新消息
func (msg *GxMessage) Copy() *GxMessage {
	newMsg := NewGxMessage()
	newMsg.Data = make([]byte, msg.GetLen())
	copy(newMsg.Header[:], msg.Header)
	copy(newMsg.Data[:], msg.Data)
	return newMsg
}
