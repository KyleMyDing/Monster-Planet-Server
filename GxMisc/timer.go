package GxMisc

/**
作者： Kyle Ding
模块：定时器接口
说明：
创建时间：2015-10-30
**/

import (
	"sort"
	"sync"
	"time"
)

//GxTimerCallback 定时器回调
type GxTimerCallback func(roleID int, eventID int) int64

//GxTimerTask 定时器任务
//EndTs 触发时间
//RoleID 所属role的ID
//EventID 事件ID
//Cb 事件回调
type GxTimerTask struct {
	EndTs   int64
	RoleID  int
	EventID int
	Cb      GxTimerCallback
}

//GxTimerTaskSlice 定时器任务列表
type GxTimerTaskSlice []*GxTimerTask

func (s GxTimerTaskSlice) Len() int {
	return len(s)
}

func (s GxTimerTaskSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s GxTimerTaskSlice) Less(i, j int) bool {
	return s[j].EndTs > s[i].EndTs
}

//GxTimer 定时器
//T 时间调度器
//L 定时器任务列表
//M 锁
type GxTimer struct {
	T *time.Timer
	L GxTimerTaskSlice
	M sync.Mutex
}

//NewGxTimer 生成一个新的定时器
func NewGxTimer() *GxTimer {
	t := &GxTimer{
		T: time.NewTimer(1 * time.Second),
	}
	t.T.Stop()
	t.AddTimer(0, 0, time.Now().Unix()+86400, func(roleID int, eventID int) int64 {
		return time.Now().Unix() + 86400
	})
	return t
}

//AddTimer 将一个新任务添加到定时器中
func (t *GxTimer) AddTimer(roleID int, eventID int, endTs int64, cb GxTimerCallback) {
	t.M.Lock()
	defer t.M.Unlock()

	now := time.Now().Unix()
	if now >= endTs {
		Error("add timer error, role-ID", roleID)
		return
	}

	//同一ID的定时器只能存在一个
	for k, v := range t.L {
		if v.RoleID == roleID && v.EventID == eventID {
			Error("add timer repeated, del it, role: %d, eventID: %d", roleID, eventID)
			kk := k + 1
			t.L = append(t.L[:k], t.L[kk:]...)

			//删除第一个，重置定时器
			if k == 0 {
				n := t.L[0].EndTs - time.Now().Unix()
				if n == 0 {
					n = 1
				}
				t.T.Reset(time.Duration(n) * time.Second)
			}
			break
		}
	}

	t.L = append(t.L, &GxTimerTask{
		EndTs:   endTs,
		RoleID:  roleID,
		EventID: eventID,
		Cb:      cb,
	})
	sort.Sort(t.L)

	n := t.L[0].EndTs - now
	if n == 0 {
		n = 1
	}
	t.T.Reset(time.Duration(n) * time.Second)

}

//DelTimer 从定时器中删除一个事件
func (t *GxTimer) DelTimer(roleID int, eventID int) {
	t.M.Lock()
	defer t.M.Unlock()

	for k, v := range t.L {
		if v.RoleID == roleID && v.EventID == eventID {
			kk := k + 1
			t.L = append(t.L[:k], t.L[kk:]...)

			//删除第一个，重置定时器
			if k == 0 {
				n := t.L[0].EndTs - time.Now().Unix()
				if n == 0 {
					n = 1
				}
				t.T.Reset(time.Duration(n) * time.Second)
			}
			break
		}
	}
}

//Run 定时器启动函数
func (t *GxTimer) Run() {
	t.M.Lock()
	defer t.M.Unlock()

	now := time.Now().Unix()
	for {
		if t.L.Len() == 0 {
			break
		}
		if now >= t.L[0].EndTs {
			newEndTs := t.L[0].Cb(t.L[0].RoleID, t.L[0].EventID)
			if newEndTs == 0 {
				t.L = t.L[1:]
			} else {
				t.L[0].EndTs = newEndTs
				sort.Sort(t.L)
			}
		} else {
			t.T.Reset(time.Duration(t.L[0].EndTs-now) * time.Second)
			break
		}
	}
}
