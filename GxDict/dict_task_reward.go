/**
作者: Jie
模块:task reward模块
说明:
创建时间：2015-12-30
**/

package GxDict

var TaskRewardArray *map[int]DropTable

func LoadTaskReward4Task(d *map[string]*DictSource) error {
	c := make(map[int]DropTable)

	for ID, info := range (*d)["Tasks"].D {
		var err error
		c[ID], err = LoadDropTable(info["Reward"].(string))
		if err != nil {
			return err
		}
	}

	TaskRewardArray = &c
	return nil
}

func GetTaskReward4TaskRewardTable(taskID int) []*DropItem {
	return GetDrop4DropTable((*TaskRewardArray)[taskID])
}
