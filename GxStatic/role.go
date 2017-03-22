package GxStatic

/**
作者： Kyle Ding
模块：游戏角色信息模块
说明：
创建时间：2015-10-30
**/

import (
	"errors"
	"fmt"
	"gopkg.in/redis.v3"
	"strconv"
	"sync"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxDict"
	"git.oschina.net/jkkkls/goxiang/GxMisc"
)

//RoleListTableName 帐号指定服务器角色列表，redis的表名
var RoleListTableName = "l_role_list:"

//RoleNameListTableName 指定服务器角色名列表，redis的表名，用来判断角色是否重复
var RoleNameListTableName = "s_role_name:"

//RoleIDTableName 角色ID表，redis的表名
var RoleIDTableName = "k_role_id"

//RoleCreateList 创建角色列表，redis的表名，用来更新redis数据到mysql
var RoleCreateList = "s_role_create:"

//RoleIdList 服务器所有id列表，redis的表名
var RoleIdList = "s_role_list:"

//RoleNameIdTable 服务器角色名对应id表
var RoleNameIdTable = "h_role_name_id:"

//RoleUngetItemList 角色未领取物品列表，redis的表名
var RoleUngetItemList = "l_role_unget_list:"

var roleIDMutex sync.Mutex

//Role 角色信息
type Role struct {
	ID         int    `pk:"true"` //角色ID
	ServerID   int    `ignore:"true"`
	PlayerName string `len:"32" index:"1"` //所属帐号名
	Name       string `len:"32"`           //角色名
	Sex        int    //性别 1-男 0-女
	VocationID int    //职业
	Vip        int    //vip等级
	Expr       int64  //累计经验
	Money      int64  //金币
	Gold       int64  //元宝
	Power      int64  //体力
	PowerTs    int64  `type:"time"` //体力恢复时间
	Honor      int64  //荣誉
	ArmyPay    int64  //军团贡献

	Wings    string `len:"128"` //翅膀装备实例ID
	Weapon   string `len:"128"` //武器装备实例ID
	Coat     string `len:"128"` //上衣装备实例ID
	Cloakd   string `len:"128"` //披风装备实例ID
	Trouser  string `len:"128"` //下衣装备实例ID
	Armguard string `len:"128"` //护手装备实例ID
	Shoes    string `len:"128"` //鞋子装备实例ID

	FightCardDistri int   //已经创建卡套的ID列表，位表示
	BuyPowerCnt     int   //当天购买体力次数
	BuyPowerTs      int64 //最近一次购买体力的时间
	GetFreePowerTs  int64 //最近一次领取免费体力的时间
}

//NewRoleID 生成一个新的角色ID
func NewRoleID(client *redis.Client) int {
	roleIDMutex.Lock()
	defer roleIDMutex.Unlock()

	if !client.Exists(RoleIDTableName).Val() {
		client.Set(RoleIDTableName, "10000000", 0)

	}
	return int(client.Incr(RoleIDTableName).Val())
}

func (role *Role) String() string {
	return "[" + strconv.Itoa(role.ID) + " " + role.Name + "]"
}

//Set4Redis ...
func (role *Role) Set4Redis(client *redis.Client) error {
	GxMisc.SaveToRedis(client, role)
	return nil
}

//Get4Redis ...
func (role *Role) Get4Redis(client *redis.Client) error {
	return GxMisc.LoadFromRedis(client, role)
}

//Del4Redis ...
func (role *Role) Del4Redis(client *redis.Client) error {
	return GxMisc.DelFromRedis(client, role)
}

//SetField4Redis ...
func (role *Role) SetField4Redis(client *redis.Client, field string) error {
	return GxMisc.SetFieldFromRedis(client, role, field)
}

//GetField4Redis ...
func (role *Role) GetField4Redis(client *redis.Client, field string) error {
	return GxMisc.GetFieldFromRedis(client, role, field)
}

//Set4Mysql ...
func (role *Role) Set4Mysql() error {
	r := new(Role)
	r.ID = role.ID
	var str string
	if r.Get4Mysql() == nil {
		str = GxMisc.GenerateUpdateSQL(role, strconv.Itoa(ServerID))
	} else {
		str = GxMisc.GenerateInsertSQL(role, strconv.Itoa(ServerID))
	}

	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		fmt.Println("set error", err)
		return err
	}
	return nil
}

//Get4Mysql ...
func (role *Role) Get4Mysql() error {
	role.ServerID = ServerID
	str := GxMisc.GenerateSelectSQL(role, strconv.Itoa(ServerID))

	rows, err := GxMisc.MysqlPool.Query(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&role.PlayerName, &role.Name, &role.Sex, &role.VocationID, &role.Vip, &role.Expr, //
			&role.Money, &role.Gold, &role.Power, &role.PowerTs, &role.Honor, &role.ArmyPay, //
			&role.Wings, &role.Weapon, &role.Coat, &role.Cloakd, &role.Trouser, &role.Armguard, &role.Shoes, //
			&role.FightCardDistri, &role.BuyPowerCnt, &role.BuyPowerTs, &role.GetFreePowerTs)
	}
	return errors.New("null")
}

//Del4Mysql ...
func (role *Role) Del4Mysql() error {
	str := GxMisc.GenerateDeleteSQL(role, strconv.Itoa(ServerID))
	_, err := GxMisc.MysqlPool.Exec(str)
	if err != nil {
		return err
	}
	return err
}

//SaveRoleID4RoleList 保存玩家角色ID列表.玩家登录服务器时候,需要拉取所有角色信息
func SaveRoleID4RoleList(client *redis.Client, playerName string, gameServerID int, roleID int) {
	client.LPush(RoleListTableName+playerName+":"+strconv.Itoa(gameServerID), strconv.Itoa(roleID))
}

//GetRoleIDs4RoleList 获取玩家角色ID列表
func GetRoleIDs4RoleList(client *redis.Client, playerName string, gameServerID int) []string {
	return client.LRange(RoleListTableName+playerName+":"+strconv.Itoa(gameServerID), 0, -1).Val()
}

//SaveRoleName 创建角色时候检查角色名是否冲突
func SaveRoleName(client *redis.Client, gameServerID int, name string) bool {
	return client.SAdd(RoleNameListTableName+strconv.Itoa(gameServerID), name).Val() == 1
}

//SaveCreateRoleId4Server 保存新创建角色的id
func SaveCreateRoleId4Server(client *redis.Client, roleID int) {
	client.SAdd(RoleCreateList+strconv.Itoa(ServerID), strconv.Itoa(roleID))
	client.SAdd(RoleIdList+strconv.Itoa(ServerID), strconv.Itoa(roleID))
}

//GetRoleIds4Server 返回服务器所有角色id
func GetRoleIds4Server(client *redis.Client) []int {
	var ids []int
	ret := client.SMembers(RoleIdList + strconv.Itoa(ServerID))
	s, err := ret.Result()
	if err != nil {
		return nil
	}
	for i := 0; i < len(s); i++ {
		id, err1 := strconv.Atoi(s[i])
		if err1 != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

//SaveRoleUngetItemList 保存角色未领取物品列表，登录时检查，或者其他模块通知
func SaveRoleUngetItemList(client *redis.Client, roleID int, item *Item) {
	buf, err := GxMisc.MsgToBuf(item)
	if err != nil {
		return
	}
	client.LPush(RoleUngetItemList+strconv.Itoa(roleID), string(buf))
}

//GetRoleUngetItemList 获取角色未领取物品列表
func GetRoleUngetItemList(client *redis.Client, roleID int) []*Item {
	tablename := RoleUngetItemList + strconv.Itoa(roleID)
	var items []*Item
	arr := client.LRange(tablename, 0, -1).Val()
	client.Del(tablename)
	for i := 0; i < len(arr); i++ {
		j, err := GxMisc.BufToMsg([]byte(arr[i]))
		if err == nil {
			item := new(Item)
			GxMisc.JSONToStruct(j, item)
			items = append(items, item)
		}
	}
	return items
}

//GetLev 获取角色等级
func (role *Role) GetLev() int {
	i := 1
	for {
		//if (*GxDict.Dicts)["Playergrowup"].D[i] == nil || (*GxDict.Dicts)["Playergrowup"].D[i]["TotalValue"].(int) > int(role.Expr) {
		iVal, _ := GxDict.GetDictInt("Playergrowup", i, "TotalValue")

		if !GxDict.CheckDictId("Playergrowup", i) || iVal > int(role.Expr) {
			break
		}
	}
	return i
}

//GetBagCellsCount 获取角色最大格子数量，返回角色当前最大格子数，等级最大格子数+已经购买的格子
func (role *Role) GetBagCellsCount(client *redis.Client) int {
	//return (*GxDict.Dicts)["CommonConfig"].D[1]["InitialGrid"].(int) + GetRoleBuyCount4Bag(client, role.ID)
	iVal, _ := GxDict.GetDictInt("CommonConfig", 1, "InitialGrid")
	return iVal + GetRoleBuyCount4Bag(client, role.ID)
}

//NewEquipment 角色增加一个新装备
func (role *Role) NewEquipment(client *redis.Client, eID, lev int) int {
	count := role.GetBagCellsCount(client)
	distri := GetRoleDistri4Bag(client, role.ID)

	index := -1
	for i := 0; i < count; i++ {
		if ((1 << uint32(i%64)) & distri[i/64]) == 0 {
			index = i
			break
		}
	}
	if index == -1 {
		return -1
	}

	equi := &Item{
		ID:  eID,
		Cnt: lev,
	}

	distri[index/64] |= (1 << uint32(index%64))
	SetRoleItem4Bag(client, role.ID, index, equi)
	SetRoleDistri4Bag(client, role.ID, distri)
	return index
}

//UpdateItem 角色更新一个新装备
func (role *Role) UpdateItem(client *redis.Client, index int, item *Item) {
	SetRoleItem4Bag(client, role.ID, index, item)
}

//DelEquipment 角色删除一个新装备
func (role *Role) DelEquipment(client *redis.Client, index int) uint16 {
	distri := GetRoleDistri4Bag(client, role.ID)
	if (distri[index/64] & (1 << uint32(index%64))) == 0 {
		GxMisc.Error("error equi ID, index: %d", index)
		return RetEquiNotExists
	}

	distri[index/64] = distri[index/64] & (^(1 << uint32(index%64)))
	DelRoleItem4Bag(client, role.ID, index)
	SetRoleDistri4Bag(client, role.ID, distri)
	return RetSucc
}

//NewItem 角色增加一个新道具
func (role *Role) NewItem(client *redis.Client, ID, cnt int) int {
	return role.NewEquipment(client, ID, cnt)
}

//DelItem 角色删除一个新道具
func (role *Role) DelItem(client *redis.Client, ID, cnt int) uint16 {
	bag := &Bag{
		RoleID: role.ID,
		Cells:  make(map[int]string),
	}
	bag.Get4Redis(client)

	last := cnt
	total := 0
	items := make(map[int]*Item)

	//先计算背包该道具数量
	for k, v := range bag.Cells {
		if v == "" {
			continue
		}

		j, err := GxMisc.BufToMsg([]byte(v))
		if err != nil {
			return RetFail
		}
		item := new(Item)
		GxMisc.JSONToStruct(j, item)

		if item.ID != ID {
			continue
		}

		items[k] = item
		total += item.Cnt

		if total >= cnt {
			break
		}
	}

	if total < cnt {
		return RetItemNotEnough
	}

	//依次减掉背包对应单元格数量，并更新
	for k, v := range items {
		if last == 0 {
			break
		}
		if last > v.Cnt {
			last -= v.Cnt
			v.Cnt = 0
		} else {
			last = 0
			v.Cnt -= last
		}
		if v.Cnt == 0 {
			bag.Distri[k/64] &= ^(1 << uint32(k%64))
			bag.Cells[k] = ""
		} else {
			info, _ := GxMisc.MsgToBuf(v)

			bag.Cells[k] = string(info)
		}
	}
	bag.Set4Redis(client)
	return RetSucc
}

//删除背包指定单元格数量物品
func (role *Role) DelBagCell(client *redis.Client, index, ID, cnt int) uint16 {
	item := GetRoleItem4Bag(client, role.ID, index)
	if item == nil || item.ID != ID {
		return RetEquiNotExists
	}
	if IsEquipment(ID) {
		role.DelEquipment(client, index)
	} else {
		if item.Cnt > cnt {
			item.Cnt -= cnt
			SetRoleItem4Bag(client, role.ID, index, item)
		} else {
			DelRoleItem4Bag(client, role.ID, index)
		}
	}
	// distri[index/64] = distri[index/64] & (^(1 << uint32(index%64)))
	// DelRoleItem4Bag(client, role.ID, index)
	// SetRoleDistri4Bag(client, role.ID, distri)
	return RetSucc
}

//BagEmptyCount 背包空余格子的数量
func (role *Role) BagEmptyCount(client *redis.Client) int {
	count := role.GetBagCellsCount(client)
	distri := GetRoleDistri4Bag(client, role.ID)

	ret := 0
	for i := 0; i < count; i++ {
		if ((1 << uint32(i%64)) & distri[i/64]) == 0 {
			ret++
		}
	}
	return ret
}

//NewCard 角色新增加一张卡片
func (role *Role) NewCard(client *redis.Client, ID, cnt int) {
	card := GetRoleCard(client, role.ID, ID)
	if card == nil {
		card = &(Item{
			ID:  ID,
			Cnt: int(cnt),
		})
	} else {
		card.Cnt += cnt
	}

	SetRoleCard(client, role.ID, card)
}

//NewCard 角色新删除一张卡片
func (role *Role) DelCard(client *redis.Client, ID, cnt int) uint16 {
	card := GetRoleCard(client, role.ID, ID)
	if card == nil {
		return RetSucc
	}
	if card.Cnt < cnt {
		return RetCardNotEnough
	}
	card.Cnt -= cnt
	SetRoleCard(client, role.ID, card)
	return RetSucc
}

//NewFightCardBag 角色新建一个卡组
func (role *Role) NewFightCardBag(client *redis.Client, bagName string) (int, int64) {
	index := -1
	for i := 0; i < 32; i++ {
		if role.FightCardDistri&(1<<uint32(i)) == 0 {
			index = i
			break
		}
	}
	if index == -1 {
		return index, 0
	}
	bag := NewFightCard(role.ID, index, bagName)
	bag.CreateTs = time.Now().Unix()
	bag.Set4Redis(client)

	role.FightCardDistri |= 1 << uint32(index)
	role.SetField4Redis(client, "FightCardDistri")
	return index, bag.CreateTs
}

//DelFightCardBag 角色删除一个卡组
func (role *Role) DelFightCardBag(client *redis.Client, bagID int) bool {
	if role.FightCardDistri&(1<<uint32(bagID)) == 0 {
		return false
	}
	bag := NewFightCard(role.ID, bagID, "")
	bag.Del4Redis(client)

	role.FightCardDistri &= ^(1 << uint32(bagID))
	role.SetField4Redis(client, "FightCardDistri")
	return true
}

//RoleExists 角色是否存在
func RoleExists(client *redis.Client, roleID int) bool {
	if !client.Exists("h_role:" + fmt.Sprintf("%d", roleID)).Val() {
		return false
	}
	role := new(Role)
	role.ID = roleID
	role.GetField4Redis(client, "ServerID")
	return role.ServerID == ServerID
}

func SetRoleNameId(client *redis.Client, name string, id int) {
	client.HSet(RoleNameIdTable+fmt.Sprintf("%d", ServerID), name, fmt.Sprintf("%d", id))
}

func GetRoleNameId(client *redis.Client, name string) int {
	I64, err := client.HGet(RoleNameIdTable+fmt.Sprintf("%d", ServerID), name).Int64()
	if err != nil {
		return 0
	}
	return int(I64)
}
