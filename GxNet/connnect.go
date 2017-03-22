package GxNet

/**
作者： Kyle Ding
模块：tcp连接接口模块
说明：
创建时间：2015-10-30
**/

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/monnand/dhkx"
	"io"
	"net"
	"sync"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMessage"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//GxTCPConn tcp连接
type GxTCPConn struct {
	ID        uint32 //连接ID
	Conn      net.Conn
	ConnMutex sync.Mutex
	Connected bool //连接状态

	TimeoutCount int          //超时次数
	T            *time.Ticker //超时检测定时器
	Toc          chan int

	Data     []uint16 //支持的cmd列表
	ServerID uint32   //连接是服务器时，服务器ID
	M        string   //模块名 Cli，Adm，Srv
	Remote   string

	Key []byte //加密时使用的密钥

}

//NewTCPConn 生成一个新的GxTCPConn
func NewTCPConn() *GxTCPConn {
	tcpConn := new(GxTCPConn)
	tcpConn.Connected = false
	tcpConn.TimeoutCount = 0
	tcpConn.Toc = make(chan int, 1)
	tcpConn.T = time.NewTicker(5 * time.Second)
	tcpConn.M = "Cli" //默认

	tcpConn.ServerID = 0
	return tcpConn
}

//runHeartbeat 处理心跳函数，用协程启动
func (conn *GxTCPConn) runHeartbeat() {
	for {
		select {
		case state := <-conn.Toc:
			if state == 0XFFFF {
				return
			}
			conn.TimeoutCount = state
		case <-conn.T.C:
			if conn.TimeoutCount > 3 {
				//超时超过三次关闭连接
				conn.Conn.Close()

				GxMisc.Debug("client[%d] %s timeout", conn.ID, conn.Remote)
				return
			} else if conn.TimeoutCount >= 0 {
				conn.TimeoutCount = conn.TimeoutCount + 1
			} else {
				break
			}
		}
	}
}

func (conn *GxTCPConn) read(buf []byte) error {
	total := len(buf)
	readLen := 0
	for {
		if total == 0 {
			return nil
		}
		n, err := conn.Conn.Read(buf[readLen:])
		if err != nil {
			return err
		}
		total -= n
		readLen += n
	}
}

func (conn *GxTCPConn) write(buf []byte) error {
	total := len(buf)
	writeLen := 0
	for {
		if total == 0 {
			return nil
		}
		n, err := conn.Conn.Write(buf[writeLen:])
		if err != nil {
			return err
		}
		total -= n
		writeLen += n
	}
}

//Send 发送消息函数
func (conn *GxTCPConn) Send(msg *GxMessage.GxMessage) error {
	conn.ConnMutex.Lock()
	defer conn.ConnMutex.Unlock()

	//写消息头
	err := conn.write(msg.Header)
	if err != nil {
		fmt.Println(err)
	}
	if err = msg.CheckFormat(); err != nil {
		return err
	}

	//如果消息体没有数据，直接返回
	if msg.GetLen() == 0 {
		return nil
	}

	//写消息体
	err = conn.write(msg.Data)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

//Recv 接受消息函数
func (conn *GxTCPConn) Recv() (*GxMessage.GxMessage, error) {
	//写消息头
	//如果读取消息失败，消息要归还给消息池
	msg := GxMessage.GetGxMessage()
	err := conn.read(msg.Header)
	if err != nil {
		GxMessage.FreeMessage(msg)
		conn.Connected = false
		return nil, err
	}
	if err = msg.CheckFormat(); err != nil {
		GxMisc.Warn("recv error format message, remote: %s", conn.Remote)
		GxMessage.FreeMessage(msg)
		return nil, err
	}

	//消息头没有数据，则返回
	if msg.GetLen() == 0 {
		return msg, nil
	}

	//写消息体
	msg.InitData()
	err = conn.read(msg.Data)
	if err != nil {
		GxMessage.FreeMessage(msg)
		conn.Connected = false
		return nil, err
	}
	return msg, nil
}

//Connect 连接指定host
func (conn *GxTCPConn) Connect(host string) error {
	c, err := net.Dial("tcp", host)

	if err != nil {
		return err
	}
	conn.Conn = c
	conn.Connected = true
	conn.Remote = c.RemoteAddr().String()
	return nil
}

//ServerKey 服务器生成一个简单的密钥，需要和客户端交互
func (conn *GxTCPConn) ServerKey() error {
	//给客户端发送19字节的随机字符串
	randomStr := GxMisc.GetRandomString(16)
	n, err := conn.Conn.Write([]byte(randomStr))
	if err != nil {
		GxMisc.Error("send random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send random string ok, len: %d", n)

	clientRandomStr := make([]byte, 16)
	n, err = conn.Conn.Read(clientRandomStr)
	if err != nil {
		GxMisc.Error("recv client random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("recv client random string ok, len: %d, clientRandomStr: %s", n, clientRandomStr)

	h := md5.New()
	io.WriteString(h, "ruiyue")
	io.WriteString(h, randomStr)
	io.WriteString(h, string(clientRandomStr))
	conn.Key = []byte(fmt.Sprintf("%x", h.Sum(nil)))[:8]
	GxMisc.Trace("new crype key, len: %d, key: %s", len(conn.Key), conn.Key)

	// 读取客户端发送的加密随机字符串
	enStr := make([]byte, 24)
	n, err = conn.Conn.Read(enStr)
	if err != nil {
		GxMisc.Error("recv cleint encrypt random string, error: %s", err)
		return err
	}
	GxMisc.Trace("recv cleint encrypt random string ok, len: %d, keylen: %d, enStr: %x", n, len(enStr), enStr)

	//解密客户端发送的加密随机字符串
	deStr, _ := GxMisc.DesDecrypt(enStr, conn.Key)
	if randomStr != string(deStr) {
		GxMisc.Error("random string error, randomStr: %s, destr: %s", randomStr, deStr)
		return errors.New("crypt key error")
	}

	GxMisc.Trace("get key ok, key: %s", conn.Key)
	return nil
}

//ClientKey 客户端生成一个简单的密钥，需要和服务端交互
func (conn *GxTCPConn) ClientKey() error {
	serverRandomStr := make([]byte, 16)
	n, err := conn.Conn.Read(serverRandomStr)
	if err != nil {
		GxMisc.Error("recv server random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("recv server random string ok, len: %d, clientRandomStr: %s", n, serverRandomStr)

	//给客户端发送19字节的随机字符串
	randomStr := GxMisc.GetRandomString(16)
	n, err = conn.Conn.Write([]byte(randomStr))
	if err != nil {
		GxMisc.Error("send random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send random string ok, len: %d", n)

	h := md5.New()
	io.WriteString(h, "ruiyue")
	io.WriteString(h, string(serverRandomStr))
	io.WriteString(h, randomStr)
	conn.Key = []byte(fmt.Sprintf("%x", h.Sum(nil)))[:8]
	GxMisc.Trace("new crype key, len: %d, key: %s", len(conn.Key), conn.Key)

	//加密随机字符串
	enStr, err1 := GxMisc.DesEncrypt([]byte(randomStr), conn.Key)
	if err1 != err {
		GxMisc.Error("encrype random fail, error: %s", err)
		return err1
	}
	GxMisc.Trace("encrype random ok, enStr: %x", enStr)

	// 发送加密随机字符串
	n, err = conn.Conn.Write(enStr)
	if err != nil {
		GxMisc.Error("send encrypt random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send encrypt random string ok, key: %s", conn.Key)
	return nil
}

//ServerDhKey 服务器生成一个DHkey交换算法的密钥，需要和客户端交互
func (conn *GxTCPConn) ServerDhKey() error {
	//给客户端发送19字节的随机字符串
	randomStr := GxMisc.GetRandomString(16)
	n, err := conn.Conn.Write([]byte(randomStr))
	if err != nil {
		GxMisc.Error("send random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send random string ok, len: %d", n)

	g, _ := dhkx.GetGroup(0)
	priv, _ := g.GeneratePrivateKey(nil)
	pub := priv.Bytes()
	GxMisc.Trace("new private key ok, pubkeylen: %d", len(pub))

	// 接受客户端的DH公钥
	b := make([]byte, len(pub))
	n, err = conn.Conn.Read(b)
	if err != nil {
		GxMisc.Error("reav client public key, error: %s", err)
		return err
	}
	GxMisc.Trace("recv client public key ok, len: %d, keylen: %d", n, len(b))

	// 发送服务端的DH公钥到服务端
	n, err = conn.Conn.Write(pub)
	if err != nil {
		GxMisc.Error("send server public key fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send server public key ok, len: %d, keylen: %d", n, len(pub))

	//获取加密公钥
	clientPub := dhkx.NewPublicKey(b)
	k, _ := g.ComputeKey(clientPub, priv)
	conn.Key = k.Bytes()
	GxMisc.Trace("new crype key, len: %d, key: %x", len(conn.Key), conn.Key)

	// 读取客户端发送的加密随机字符串
	enStr := make([]byte, 24)
	n, err = conn.Conn.Read(enStr)
	if err != nil {
		GxMisc.Error("recv cleint encrypt random string, error: %s", err)
		return err
	}
	GxMisc.Trace("recv cleint encrypt random string ok, len: %d, keylen: %d, enStr: %x", n, len(enStr), enStr)

	//解密客户端发送的加密随机字符串
	deStr, _ := GxMisc.DesDecrypt(enStr, conn.Key[:8])
	if randomStr != string(deStr) {
		GxMisc.Error("random string error, randomStr: %s, destr: %s", randomStr, deStr)
		return errors.New("crypt key error")
	}

	GxMisc.Trace("Dh excharhe key ok, key: %x", conn.Key[:8])
	return nil
}

//ClientDhKey 客户端生成一个DHkey交换算法的密钥，需要和服务端交互
func (conn *GxTCPConn) ClientDhKey() error {
	//接受服务端发送的随机字符串
	randomStr := make([]byte, 16)
	n, err := conn.Conn.Read(randomStr)
	if err != nil {
		GxMisc.Error("recv random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("recv random string ok, randomStr: %s", randomStr)

	g, _ := dhkx.GetGroup(0)
	priv, _ := g.GeneratePrivateKey(nil)
	pub := priv.Bytes()
	GxMisc.Trace("new private key ok, pubkeylen: %d", len(pub))

	// 发送客户端端的DH公钥到服务端
	n, err = conn.Conn.Write(pub)
	if err != nil {
		GxMisc.Error("send client public key fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send client public key ok, len: %d, keylen: %d", n, len(pub))

	// 接受服务端的DH公钥
	b := make([]byte, len(pub))
	n, err = conn.Conn.Read(b)
	if err != nil {
		GxMisc.Error("reav server public key, error: %s", err)
		return err
	}
	GxMisc.Trace("recv server public key ok, len: %d, keylen: %d", n, len(b))

	//获取加密公钥
	servertPub := dhkx.NewPublicKey(b)
	k, _ := g.ComputeKey(servertPub, priv)
	conn.Key = k.Bytes()
	GxMisc.Trace("new crype key, len: %d, key: %x", len(conn.Key), conn.Key)

	//加密随机字符串
	enStr, err1 := GxMisc.DesEncrypt([]byte(randomStr), conn.Key[:8])
	if err1 != err {
		GxMisc.Error("encrype random fail, error: %s", err)
		return err1
	}
	GxMisc.Trace("encrype random ok, enStr: %x", enStr)

	// 发送加密随机字符串
	n, err = conn.Conn.Write(enStr)
	if err != nil {
		GxMisc.Error("send encrypt random string fail, error: %s", err)
		return err
	}
	GxMisc.Trace("send encrypt random string ok, key: %x", conn.Key[:8])
	return nil
}
