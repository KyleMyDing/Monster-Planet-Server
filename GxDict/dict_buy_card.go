package GxDict

/**
作者: guangbo
模块: 抽卡物品列表模块
说明:
创建时间：2015-12-16
**/

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//CardArrayType 卡片列表类型
type CardArrayType []int

//BuyCardArrayType 抽卡物品列表类型
type BuyCardArrayType map[int]CardArrayType

//BuyCardArray 抽卡物品列表
var buyCardArray *BuyCardArrayType

//LoadBuyCardArray4BuyCard 加载抽卡物品列表回调
func LoadBuyCardArray4BuyCard(d *map[string]*DictSource) error {
	b := make(BuyCardArrayType)
	for k, v := range (*d)["BuyCard"].D {
		country := v["Country"].(int)
		color := v["Color"].(int)
		sex := v["Sex"].(int)
		for k1, v1 := range (*d)["HerosCard"].D {
			country1 := v1["CountrySign"].(int)
			color1 := v1["Color"].(int)
			sex1 := v1["SexSign"].(int) //
			isJoinBuy := v1["IsJoinBuy"].(int)
			if country == country1 && color == color1 && sex == sex1 && isJoinBuy == 1 {
				b[k] = append(b[k], k1)
			}
		}
	}
	buyCardArray = &b
	return nil
}

//Random4BuyCardArray 获取一张指定抽卡物品列表中的卡牌
func Random4BuyCardArray(id, level int) int {
	var ExtractCardArray []int
	size := len((*buyCardArray)[id])
	for i := 0; i < size; i++ {
		HeroID := (*buyCardArray)[id][i]
		//if (*Dicts)["HerosCard"].D[HeroID]["CardLevelLimit"].(int) <= level {
		lev, _ := GetDictInt("HerosCard", HeroID, "CardLevelLimit")
		if lev <= level {
			ExtractCardArray = append(ExtractCardArray, HeroID)
		}
	}
	return ExtractCardArray[GxMisc.GetRandomInterval(0, len(ExtractCardArray)-1)]
}
