package GxDict

/**
作者: guangbo
模块:字典模块
说明:
创建时间：2015-11-16
**/

import (
	"fmt"
	"strconv"
	"strings"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//DropItem 物品掉落表
//ID 物品ID
//Cnt 物品数量
//P 概率，范围1到10000，单组物品里的所有物品的P的和不应该大于10000
type DropItem struct {
	ID  int
	Cnt int
	P   int //1-10000
}

//DropInfo 单组物品,其中最多只会掉落一个
type DropInfo []*DropItem

//DropTable 多组物品,每组物品掉一个
type DropTable []DropInfo

//LoadDropTable 从配置表中读取物品掉落表
func LoadDropTable(str string) (DropTable, error) {
	var dict DropTable
	arr := strings.Split(str, ";")
	for i := 0; i < len(arr); i++ {
		if arr[i] == "" {
			continue
		}
		arr1 := strings.Split(arr[i], ",")
		if (len(arr1) % 3) != 0 {
			return nil, fmt.Errorf("drop format error, str: %s", arr[i])
		}

		var info DropInfo
		n := len(arr1) / 3
		for j := 0; j < n; j++ {
			ID, _ := strconv.Atoi(arr1[j*3])
			cnt, _ := strconv.Atoi(arr1[j*3+1])
			p, _ := strconv.Atoi(arr1[j*3+2])

			info = append(info, &DropItem{
				ID:  ID,
				Cnt: cnt,
				P:   p,
			})
		}
		dict = append(dict, info)
	}

	return dict, nil
}

//GetDrop4DropTable 物品掉落表中获取掉落物品
func GetDrop4DropTable(t DropTable) []*DropItem {
	var items []*DropItem
	for i := 0; i < len(t); i++ {
		if len(t[i]) == 0 {
			continue
		}
		n := GxMisc.GetRandomInterval(1, 10000)
		for j := 0; j < len(t[i]); j++ {
			if n <= t[i][j].P {
				items = append(items, t[i][j])
				break
			}
			n -= t[i][j].P
		}
	}
	return items
}
