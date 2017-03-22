package GxMessage

/**
作者： Kyle Ding
模块：消息池
说明：
创建时间：2015-12-10
**/

import (
	"container/list"
	"sync"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

var gxMessagePoolMutex sync.Mutex

//gxMessagePool 消息池
var gxMessagePool *list.List

//gxMessagePoolStatus 消息池是否开启
var gxMessagePoolStatus = 0

//总消息数
var gxMessagePoolCount = 0

func OpenGxMessagePool(size int) {
	gxMessagePoolStatus = size
	gxMessagePool = list.New()

	for i := 0; i < size; i++ {
		gxMessagePoolCount++
		gxMessagePool.PushBack(NewGxMessage())
	}
}

//GetGxMessage 返回一个新消息
func GetGxMessage() *GxMessage {
	if gxMessagePoolStatus == 0 {
		return NewGxMessage()
	}
	gxMessagePoolMutex.Lock()
	defer gxMessagePoolMutex.Unlock()

	var msg *GxMessage = nil
	if gxMessagePool.Len() == 0 {
		gxMessagePoolCount++
		msg = NewGxMessage()
	} else {
		msg = gxMessagePool.Front().Value.(*GxMessage)
		gxMessagePool.Remove(gxMessagePool.Front())

		msg.SetLen(0)
		msg.SetID(0)
		msg.SetCmd(0)
		msg.SetRet(0)
		msg.SetSeq(0)
		msg.ClearMask()
	}
	return msg
}

//ReturnGxMessage 归还一个新消息
//消息流程结束后，系统会回收消息
//如果回调里需要保存该消息，就复制一个新信息出来自己使用
func ReturnGxMessage(msg *GxMessage) {
	if gxMessagePoolStatus == 0 {
		return
	}
	gxMessagePoolMutex.Lock()
	defer gxMessagePoolMutex.Unlock()
	gxMessagePool.PushBack(msg)
}

func PrintfMessagePool() {
	if gxMessagePoolStatus == 0 {
		return
	}
	gxMessagePoolMutex.Lock()
	defer gxMessagePoolMutex.Unlock()
	GxMisc.Info("==============Message Pool Info===============")
	GxMisc.Info("free count: %d, total count: %d", gxMessagePool.Len(), gxMessagePoolCount)
	GxMisc.Info("==============================================")
}

func Copy4MessagePool(msg *GxMessage) *GxMessage {
	newMsg := GetGxMessage()
	newMsg.Data = GxMisc.GxMalloc(int(msg.GetLen()))
	copy(newMsg.Header[:], msg.Header)
	copy(newMsg.Data[:], msg.Data)
	return newMsg
}

func FreeMessage(msg *GxMessage) {
	if msg.GetLen() > 0 && GxMisc.GxFree(msg.Data) {
		msg.Data = nil
	}
	ReturnGxMessage(msg)
}
