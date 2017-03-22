package GxDict

/**
作者： Kyle Ding
模块：字典模块
说明：
创建时间：2015-11-9
**/

import (
	"fmt"
	"github.com/tealeg/xlsx"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//DictSource 字典信息表
//S 字典名
//D 字典数据
//C 字典回调,可能会初始化一些掉落信息
type DictSource struct {
	S string
	D GxMisc.Dict
	C func(d *map[string]*DictSource) error
}

//Dicts 游戏所有字典信息表,根据map的key去读取字典
var dicts *map[string]*DictSource

func initDicts() map[string]*DictSource {
	dicts := make(map[string]*DictSource)
	dicts["Player"] = &DictSource{
		S: "Player",
	}
	dicts["CommonConfig"] = &DictSource{
		S: "CommonConfig",
	}
	dicts["Playergrowup"] = &DictSource{
		S: "Playergrowup",
	}
	dicts["HerosCard"] = &DictSource{
		S: "HerosCard",
	}
	dicts["Package"] = &DictSource{
		S: "Package",
	}
	dicts["Items"] = &DictSource{
		S: "Items",
		C: LoadItemDict4Drop,
	}
	dicts["Cardsfusion"] = &DictSource{
		S: "Cardsfusion",
	}
	dicts["Cardsdecomposition"] = &DictSource{
		S: "Cardsdecomposition",
	}
	dicts["DecompositionDrop"] = &DictSource{
		S: "DecompositionDrop",
		C: LoadCardDecompress,
	}
	/*Jie*/
	//抽卡配置
	dicts["BuyCard"] = &DictSource{
		S: "BuyCard",
		C: LoadBuyCardArray4BuyCard,
	}
	dicts["BuyCardCost"] = &DictSource{
		S: "BuyCardCost",
	}
	dicts["BuyCardSpecial"] = &DictSource{
		S: "BuyCardSpecial",
	}
	dicts["BuySkillCard"] = &DictSource{
		S: "BuySkillCard",
		C: LoadBuySkillCardArray4BuyCard,
	}
	dicts["BuySkillCardCost"] = &DictSource{
		S: "BuySkillCardCost",
	}
	//副本
	dicts["Chapter"] = &DictSource{
		S: "Chapter",
	}
	dicts["ChapterLevel"] = &DictSource{
		S: "ChapterLevel",
	}
	dicts["ChapterLevelBattle"] = &DictSource{
		S: "ChapterLevelBattle",
		C: LoadChapterDropDict,
	}
	/*written by Jie 2015-12-22 */
	dicts["Tasks"] = &DictSource{
		S: "Tasks",
		C: LoadTaskReward4Task,
	}
	return dicts
}

//LoadDict 读取单个字典函数
func (d *DictSource) LoadDict(dictDir string) error {
	str := dictDir + d.S + ".xlsx"
	xlFile, err := xlsx.OpenFile(str)
	if err != nil {
		GxMisc.Error("can find file: %s", str)
		return err
	}

	for _, sheet := range xlFile.Sheets {
		if sheet.Name == d.S {
			//GxMisc.Debug("load dict[%s] ok", d.S)
			err = d.D.Load(sheet)
			if err != nil {
				return err
			}
			GxMisc.Debug("load dict[%s] ok", d.S)
			break
		}
	}
	return nil
}

//LoadAllDict 读取字典函数，在程序启动时候调用该函数
func LoadAllDict(dictDir string) error {
	d := initDicts()

	for _, v := range d {
		v.D = make(GxMisc.Dict)
		if err := v.LoadDict(dictDir); err != nil {
			return err
		}
	}

	for _, v := range d {
		if v.C != nil {
			err := v.C(&d)
			if err != nil {
				return err
			}
		}
	}

	dicts = &d

	return nil
}

//CheckDictName 判断字典是否存在
func CheckDictName(dictName string) bool {
	return (*dicts)[dictName] != nil
}

//CheckDictId  判断字典的ID是否存在
func CheckDictId(dictName string, id int) bool {
	return (*dicts)[dictName].D[id] != nil
}

//CheckDictfield  判断字典的字段是否存在
func CheckDictfield(dictName string, id int, fieldName string) error {
	if (*dicts)[dictName] == nil {
		return fmt.Errorf("dict[%s] is not exists", dictName)
	} else if (*dicts)[dictName].D[id] == nil {
		return fmt.Errorf("dict[%s] id[%d] is not exists", dictName, id)
	} else if (*dicts)[dictName].D[id][fieldName] == nil {
		return fmt.Errorf("dict[%s] id[%d] fieldName[%s] is not exists", dictName, id, fieldName)
	} else {
		return nil
	}

}

//GetDictInt  获取字典中int类型的字段值
func GetDictInt(dictName string, id int, fieldName string) (int, error) {
	err := CheckDictfield(dictName, id, fieldName)
	if err != nil {
		return -1, err
	}
	return (*dicts)[dictName].D[id][fieldName].(int), nil
}

//GetDictString 获取字典中string类型的字段值
func GetDictString(dictName string, id int, fieldName string) (string, error) {
	err := CheckDictfield(dictName, id, fieldName)
	if err != nil {
		return "", err
	}
	return (*dicts)[dictName].D[id][fieldName].(string), nil
}

func GetDict(dictName string) (*GxMisc.Dict, error) {
	if !CheckDictName(dictName) {
		return nil, fmt.Errorf("dict[%s] is not exists", dictName)
	}
	return &(*dicts)[dictName].D, nil
}
