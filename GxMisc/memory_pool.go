package GxMisc

/**
作者： Kyle Ding
模块：内存池
说明：消息消息体内存使用内存池
创建时间：2015-12-10
**/

import (
	"container/list"
	"sync"
)

type GxMemoryPoolType map[int]*list.List
type gxMemoryRecordType map[*byte]int

var gxMemoryPoolMutex sync.Mutex

//gxMmemoryPool 内存池
var gxMemoryPool GxMemoryPoolType

var gxPoolCount map[int]int

//gxMemoryRecord 内存记录
var gxMemoryRecord gxMemoryRecordType

//gxMemoryPoolStatus 内存池是否开启
var gxMemoryPoolStatus = 0

//gxMemoryMaxSize 内存申请单位大小
var gxMemoryAllocSize = 16

//gxMemoryMaxSize 内存管理最大的内存大小
var gxMemoryMaxSize = 1024

func OpenGxMemoryPool(size int) {
	gxMemoryPoolStatus = size
	gxMemoryPool = make(GxMemoryPoolType)
	gxMemoryRecord = make(gxMemoryRecordType)
	gxPoolCount = make(map[int]int)

	//16-128初始化size个

	for i := 1; i <= 8; i++ {
		actualLen := i * gxMemoryAllocSize
		for j := 0; j < size; j++ {
			gxPoolCount[actualLen] = gxPoolCount[actualLen] + 1
			mem := make([]byte, actualLen)
			gxMemoryRecord[&mem[0]] = actualLen
			if gxMemoryPool[actualLen] == nil {
				gxMemoryPool[actualLen] = list.New()
			}
			gxMemoryPool[actualLen].PushBack(mem)
		}
	}
}

func GxMalloc(len int) []byte {
	if len <= 0 {
		return nil
	}
	if len > gxMemoryMaxSize {
		mem := make([]byte, len)
		return mem
	}
	if gxMemoryPoolStatus == 0 {
		return make([]byte, len)
	}
	gxMemoryPoolMutex.Lock()
	defer gxMemoryPoolMutex.Unlock()

	actualLen := len
	n := actualLen % gxMemoryAllocSize
	if n != 0 {
		actualLen += gxMemoryAllocSize - n
	}

	if gxMemoryPool[actualLen] == nil {
		gxMemoryPool[actualLen] = list.New()
	}

	//如果内存池是空的，重新申请新内存
	if gxMemoryPool[actualLen].Len() == 0 {
		//Debug("new memcpy, len: %d, actualLen: %d", len, actualLen)
		gxPoolCount[actualLen] = gxPoolCount[actualLen] + 1

		mem := make([]byte, actualLen)
		gxMemoryRecord[&mem[0]] = actualLen
		return mem[:len]
	}
	//Debug("get memcpy, len: %d, actualLen: %d", len, actualLen)
	mem := gxMemoryPool[actualLen].Front().Value.([]byte)
	gxMemoryPool[actualLen].Remove(gxMemoryPool[actualLen].Front())
	return mem[:len]
}

func GxFree(mem []byte) bool {
	if gxMemoryPoolStatus == 0 {
		return false
	}
	if len(mem) == 0 {
		Error("[!!!!!], free memory!!!")
	}
	if len(mem) > gxMemoryMaxSize {
		Debug("skip free big memcpy, len: %d", len(mem))
		return false
	}
	gxMemoryPoolMutex.Lock()
	defer gxMemoryPoolMutex.Unlock()

	length := gxMemoryRecord[&mem[0]]
	//Debug("free memcpy, len: %d, actual-len: %d", len(mem), length)
	if length == 0 {
		Error("error free memory, memory: %v", &mem[0])
		return false
	}
	gxMemoryPool[length].PushBack(mem[:length])

	return true
}

func PrintfMemoryPool() {
	if gxMemoryPoolStatus == 0 {
		return
	}
	gxMemoryPoolMutex.Lock()
	defer gxMemoryPoolMutex.Unlock()
	Info("==========Memory Pool Info[%d]===============", len(gxMemoryRecord))
	for k, v := range gxPoolCount {
		if v == 0 {
			continue
		}
		Info("memory[%4d], free count: %d, total count: %d", k, gxMemoryPool[k].Len(), v)
	}
	Info("==============================================")
}
