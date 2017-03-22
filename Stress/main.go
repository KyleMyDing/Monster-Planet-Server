/**
作者： Kyle Ding
模块：Stress Test Programing
说明：
创建时间：2015-11-26
**/
package main

import (
	"os"
	"strconv"
	"sync"
	"time"
)

import (
	. "git.oschina.net/jkkkls/goxiang/GxMisc"
	. "git.oschina.net/jkkkls/goxiang/GxNet"
	. "git.oschina.net/jkkkls/goxiang/GxStatic"
)

var GateServerAddr = "192.168.1.120:13001"
var connNum = 100
var quit chan int
var counter *Counter
var ret = 0

var mutex sync.Mutex

func heartbeat(conn *GxTcpConn) {
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			test_heartbeat(conn)
		case <-quit:
			return
		}
	}
}

func test_heartbeat(conn *GxTcpConn) error {
	SendPbMessage(conn, 0, 0, CmdHeartbeat, uint16(counter.Genarate()), 0, nil)
	return nil
}

func RecvMsg(conn *GxTcpConn) {
	for {
		_, err := conn.Recv()
		if err != nil {
			Debug("recv error, %s", err)
			quit <- 1
			quit <- 1
			return
		}
		mutex.Lock()
		ret++
		mutex.Unlock()
		//Debug("<==== recv buff msg, info: %s", msg.String())
	}
}

func RecvRet() {
	for {
		time.Sleep(10 * time.Second)
		mutex.Lock()
		Debug("ret: %d", ret)
		mutex.Unlock()
	}
}

func StressTest() {
	for i := 0; i < connNum; i++ {
		connGate := NewTcpConn()
		err := connGate.Connect(GateServerAddr)
		if err != nil {
			Error("new connnect, remote: %s", err)
			return
		}
		go heartbeat(connGate)
		go RecvMsg(connGate)
	}
	go RecvRet()
	for {
		time.Sleep(100 * time.Second)
	}
}

func main() {
	InitLogger("stress")

	if len(os.Args) != 4 {
		Info("Paramter Error")
		Info("gxstress <host> <port> <connNum>")
		return
	}
	counter = NewCounter()
	GateServerAddr = os.Args[1] + ":" + os.Args[2]
	connNum, _ = strconv.Atoi(os.Args[3])
	StressTest()
}
