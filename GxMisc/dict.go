package GxMisc

/**
作者： Kyle Ding
模块：字典接口
说明：
创建时间：2015-10-30
**/

import (
	"fmt"
	"github.com/tealeg/xlsx"
)

//DictInfo xlsx文档的每一行的数据
type DictInfo map[string]interface{}

//Dict xlsx文档的每一页对应一个dict
type Dict map[int]DictInfo

//Load xlsx文件读取函数，每个sheet页要调用一次 CommonConfig
func (d Dict) Load(sheet *xlsx.Sheet) error {
	//Debug("====1====sheet: %s", sheet.Name)
	for i, row := range sheet.Rows {
		if i < 3 {
			continue
		}
		//Debug("====3====sheet: %s, i: %d, cell: %d", sheet.Name, i, len(row.Cells))
		if len(row.Cells) < 2 {
			break
		}
		ID, err := row.Cells[1].Int()
		if err != nil {
			break
		}
		d[ID] = make(DictInfo)
		for j, cell := range row.Cells {
			if j < 1 {
				continue
			}
			name, _ := sheet.Rows[1].Cells[j].String()
			typename, _ := sheet.Rows[2].Cells[j].String()
			if name == "" {
				break
			}
			if typename == "int" {
				d[ID][name], err = cell.Int()
				if err != nil {
					return fmt.Errorf("xlsx file: %s, col: %d, name: %s, text is not integer", sheet.Name, j, name)
				}
			} else if typename == "string" {
				d[ID][name], _ = cell.String()
			} else if typename == "plan" {
				//ignro
			} else {
				return fmt.Errorf("xlsx file: %s, col: %d, name: %s, name error", sheet.Name, j, name)
			}
		}
	}
	return nil
}
