package GxMisc

/**
作者： Kyle Ding
模块：ID控制器
说明：
创建时间：2015-10-30
**/

import (
	"sync"
)

//Counter ID控制器
type Counter struct {
	C     uint32
	Mutex *sync.Mutex
}

//NewCounter 生成一个新的ID控制器
func NewCounter() *Counter {
	counter := new(Counter)
	counter.C = 0
	counter.Mutex = new(sync.Mutex)
	return counter
}

//Genarate 返回一个ID
func (counter *Counter) Genarate() uint32 {
	counter.Mutex.Lock()
	defer counter.Mutex.Unlock()

	counter.C++
	return counter.C
}
