package GxDict

/**
作者: Jie
模块: 抽技能卡物品列表模块
说明:
创建时间：2015-12-9
**/
import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//SkillCardArrayType 技能卡片列表类型
type SkillCardArrayType []int

//BuySkillCardArrayType 抽卡物品列表类型
type BuySkillCardArrayType map[int]SkillCardArrayType

//BuySkillCardArray 抽卡物品列表
var buySkillCardArray *BuySkillCardArrayType

//LoadBuySkillCardArray4BuyCard 加载抽技能卡物品列表回调
func LoadBuySkillCardArray4BuyCard(d *map[string]*DictSource) error {
	b := make(BuySkillCardArrayType)
	for k, v := range (*d)["BuySkillCard"].D {

		color := v["Color"].(int)
		for k1, v1 := range (*d)["HerosCard"].D {
			color1 := v1["Color"].(int)
			cradOrSkill := v1["CardOrSkill"].(int)
			if color == color1 && cradOrSkill == 2 {
				b[k] = append(b[k], k1)
			}
		}
	}

	buySkillCardArray = &b

	return nil
}

//Random4BuySkillCardArray 获取一张指定抽技能卡物品列表中的卡牌
func Random4BuySkillCardArray(id int) int {
	return (*buySkillCardArray)[id][GxMisc.GetRandomInterval(0, len((*buySkillCardArray)[id])-1)]
}
