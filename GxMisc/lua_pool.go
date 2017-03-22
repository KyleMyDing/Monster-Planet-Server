package GxMisc

/**
作者: guangbo
模块：lua虚拟机池
说明：当池容量小于最大容量时，以最大容量扩充一倍
创建时间：2015-11-5
**/

import (
	"container/list"
	"github.com/yuin/gopher-lua"
	"sync"
)

type initLuaCallback func(L *lua.LState)

//lStatePool lua虚拟机连接池
//M 锁
//Pool 池
//C 当前数量
//Cb 初始化虚拟机回调函数
type lStatePool struct {
	M    sync.Mutex
	Pool *list.List
	C    int
	Cb   initLuaCallback
}

//LuaPool lua虚拟机连接池
var LuaPool *lStatePool

//initLuaPool 初始化lua虚拟机连接池函数,在程序启动时调用
func initLuaPool(count int, cb initLuaCallback) {
	LuaPool = &lStatePool{
		Pool: list.New(),
		Cb:   cb,
	}
	LuaPool.C = count
	for i := 0; i < LuaPool.C; i++ {
		LuaPool.Pool.PushBack(LuaPool.New())
	}
}

func (pl *lStatePool) expand(count int) {
	pl.M.Lock()
	defer pl.M.Unlock()
	for i := 0; i < count; i++ {
		LuaPool.Pool.PushBack(LuaPool.New())
	}
	LuaPool.C += count
}

func (pl *lStatePool) Get() *lua.LState {
	pl.M.Lock()
	defer pl.M.Unlock()

	if pl.Pool.Len() <= (LuaPool.C / 5) {
		go pl.expand(LuaPool.C)
		if pl.Pool.Len() == 0 {
			return LuaPool.New()
		}
	}
	l := LuaPool.Pool.Front().Value.(*lua.LState)
	LuaPool.Pool.Remove(LuaPool.Pool.Front())

	return l
}

func (pl *lStatePool) New() *lua.LState {
	l := lua.NewState()
	pl.Cb(l)
	return l
}

func (pl *lStatePool) Put(L *lua.LState) {
	pl.M.Lock()
	defer pl.M.Unlock()
	LuaPool.Pool.PushBack(L)
}

func (pl *lStatePool) Shutdown() {
	for L := LuaPool.Pool.Front(); L != nil; L = L.Next() {
		L.Value.(*lua.LState).Close()
	}
}
