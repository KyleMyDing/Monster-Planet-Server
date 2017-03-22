package GxDict

/**
作者: guangbo
模块:卡牌分解字典模块
说明:
创建时间：2015-11-19
**/

// import (
// 	"git.oschina.net/jkkkls/goxiang/GxMisc"
// )

//CardDecompressInfo 卡牌分解掉落表
var cardDecompressInfo *map[int]DropTable

//LoadCardDecompress 初始化卡牌分解掉落表
func LoadCardDecompress(d *map[string]*DictSource) error {
	c := make(map[int]DropTable)

	for ID, info := range (*d)["DecompositionDrop"].D {
		var err error
		c[ID], err = LoadDropTable(info["SoulStone"].(string))
		if err != nil {
			return err
		}
	}

	cardDecompressInfo = &c
	return nil
}

//GetDrop4CardDecompress
func GetDrop4CardDecompress(dropID int) []*DropItem {
	return GetDrop4DropTable((*cardDecompressInfo)[dropID])
}
