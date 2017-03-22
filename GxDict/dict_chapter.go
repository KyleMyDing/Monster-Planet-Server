package GxDict

/**
作者: guangbo
模块:战斗点掉落表
说明:
创建时间：2015-12-11
**/

// import (
// 	"git.oschina.net/jkkkls/goxiang/GxMisc"
// )

type ChapterDropType struct {
	First  DropTable //第一次掉落
	Common DropTable //正常掉落
}

//chapterDrops 战斗点掉落表
var chapterDrops *map[int]*ChapterDropType

//LoadChapterDropDict 加载战斗点掉落表
func LoadChapterDropDict(d *map[string]*DictSource) error {
	c := make(map[int]*ChapterDropType)

	var err error
	for k, v := range (*d)["ChapterLevelBattle"].D {
		cd := new(ChapterDropType)
		cd.First, err = LoadDropTable(v["FirstPassDrop"].(string))
		if err != nil {
			return err
		}
		cd.Common, err = LoadDropTable(v["CommonDrop"].(string))
		if err != nil {
			return err
		}
		c[k] = cd
	}
	chapterDrops = &c
	return nil
}

//Random4ChapterDrop 获取战斗点掉落
func Random4ChapterDrop(id int, first bool) []*DropItem {
	if (*chapterDrops)[id] == nil {
		return nil
	}
	if first {
		return GetDrop4DropTable((*chapterDrops)[id].First)
	}
	return GetDrop4DropTable((*chapterDrops)[id].Common)
}
