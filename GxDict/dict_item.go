package GxDict

/**
作者: guangbo
模块:道具打开字典模块
说明:
创建时间：2015-11-16
**/

// import (
// 	"git.oschina.net/jkkkls/goxiang/GxMisc"
// )

//ItemOpenInfo 道具使用后得到的物品表
var itemOpenInfo *map[int]DropTable

//LoadItemDict4Drop 道具使用后得到的物品表的初始化函数,在load字典的时候设置他为回调函数
func LoadItemDict4Drop(d *map[string]*DictSource) error {
	temp := make(map[int]DropTable)

	for ID, info := range (*d)["Items"].D {
		var err error
		temp[ID], err = LoadDropTable(info["gets"].(string))
		if err != nil {
			return err
		}
	}

	itemOpenInfo = &temp

	return nil
}

func GetItemDict4Drop(id int) DropTable {
	if _, ok := (*itemOpenInfo)[id]; ok {
		return (*itemOpenInfo)[id]
	} else {
		return nil
	}
}
