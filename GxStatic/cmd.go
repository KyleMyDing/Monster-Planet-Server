package GxStatic

/**
作者： Kyle Ding
模块：消息命令字模块
说明：
创建时间：2015-10-30
**/

//增加新命令需要同时修改CmdString
const (
	//login
	CmdRegister     = 1000 //注册帐号 ok
	CmdLogin        = 1001 //登录帐号 ok
	CmdGetGatesInfo = 1002 //获取所有网关信息 ok

	//gate cmd begin
	CmdGateBegin   = 2000 //=========gate begin=========
	CmdHeartbeat   = 2000 //心跳 ok
	CmdGetRoleList = 2001 //获取角色列表 ok
	CmdCreateRole  = 2002 //创建角色进入游戏 ok
	CmdSelectRole  = 2003 //选择角色进入游戏 ok

	CmdAdminInfoChange    = 2004 //新物品，内部使用
	CmdAdminNewMail       = 2005 //新邮件，内部使用
	CmdAdminAddFriend     = 2006 //添加好友通知，内部使用
	CmdAdminDealFriendYes = 2007 //同意添加好友通知，内部使用
	CmdAdminDealFriendNo  = 2008 //拒绝添加好友通知，内部使用
	CmdAdminDeleteFriend  = 2009 //删除好友通知，内部使用
	CmdAdminFriendSend    = 2010 //赠送物品通知，内部使用
	CmdAdminRecharge      = 2011 //充值到账通知，内部使用
	CmdGateEnd            = 2999 //=========gate end=========

	CmdRoleBegin              = 2010 //=========role begin=========
	CmdPowerAdd               = 2010 //体力恢复通知 ok
	CmdLoginConflict          = 2011 //登录冲突通知 ok
	CmdInfoChange             = 2012 //角色信息变更通知 ok
	CmdSendChat               = 2013 //聊天请求 ok
	CmdNewChat                = 2014 //新聊天消息 ok
	CmdStrengthenEquipment    = 2100 //强化或提升品质装备
	CmdReplaceEquipment       = 2101 //更换装备 ok
	CmdUnloadEquipment        = 2102 //卸下装备 ok
	CmdUseItem                = 2103 //使用道具 ok
	CmdSellItem               = 2104 //出售道具 ok
	CmdNewFightCardBag        = 2105 //新建卡组 ok
	CmdDelFightCardBag        = 2106 //删除卡组 ok
	CmdFuseCard               = 2108 //融合卡牌 ok
	CmdDecomposeCard          = 2109 //分解卡牌 ok
	CmdAddFightCard           = 2110 //添加战斗卡牌 ok
	CmdDelFightCard           = 2111 //删除战斗卡牌 ok
	CmdOrderBag               = 2112 //整理背包 ok
	CmdExpandBag              = 2113 //扩展背包 ok
	CmdExtractCard            = 2114 //抽取武将卡牌
	CmdReceiveCard            = 2115 //领取抽卡赠送的卡牌
	CmdExtractSkillCard       = 2116 //抽取技能卡牌
	CmdBuyCardInfo            = 2117 //获取抽卡和奖励信息
	CmdBuyPower               = 2118 //购买体力
	CmdGetFreePower           = 2119 //领取免费体力
	CmdGetLevelInfo           = 2120 //读取章节关卡信息
	CmdGetPointInfo           = 2121 //读取关卡战斗点信息
	CmdChapterShuffle         = 2122 //征战洗牌
	CmdChapterBattleBegin     = 2123 //征战战斗开始
	CmdChapterBattleEnd       = 2124 //征战战斗结束
	CmdGetMailList            = 2125 //获取邮件列表
	CmdGetMailItem            = 2126 //领取(删除)邮件
	CmdNewMail                = 2127 //新邮件通知
	CmdSelectVocation         = 2128 //选择职业
	CmdGetFriend              = 2129 //获取好友列表
	CmdGetReferrer            = 2130 //获取推荐好友列表
	CmdAddFriend              = 2131 //添加好友
	CmdAddFriendNotify        = 2132 //添加好友通知
	CmdDealFriend             = 2133 //同意或者拒绝添加好友
	CmdDealFriendNotify       = 2134 //同意或者拒绝添加好友通知
	CmdDeleteFriend           = 2135 //删除好友
	CmdDeleteFriendNotify     = 2136 //删除好友通知
	CmdFriendSend             = 2137 //赠送物品
	CmdFriendSendNotify       = 2138 //赠送物品通知
	CmdFriendRecv             = 2139 //领取赠送物品
	CmdFriendInfoChangeNotify = 2140 //好友状态变更通知
	CmdGetTaskList            = 2141 //获取任务列表
	CmdGetRoleFightInfo       = 2142 //获取玩家战斗信息

	CmdReceiveTaskReward = 2146 //领取任务奖励

	CmdReadMail       = 2143 //阅读邮件
	CmdRecharge       = 2144 //充值
	CmdRechargeNotify = 2145 //充值到账通知

	CmdRoleEnd = 3999 //=========role end=========

	//gate cmd end

	//server
	CmdServerConnectGate = 4000 //逻辑服务器连接网关
	CmdClientLogout      = 4001 //客户端登出
	CmdAdminConnect      = 4002 //管理后台连接网关
	CmdAdminMessage      = 4003 //管理后台操作
)

//CmdString 命令字描述列表
var CmdString map[uint16]string = map[uint16]string{
	CmdRegister:     "注册帐号",
	CmdLogin:        "登录帐号",
	CmdGetGatesInfo: "获取所有网关信息",

	CmdHeartbeat:   "心跳",
	CmdGetRoleList: "获取角色列表",
	CmdCreateRole:  "创建角色进入游戏",
	CmdSelectRole:  "选择角色进入游戏",

	CmdPowerAdd:      "体力恢复通知",
	CmdLoginConflict: "登录冲突通知",
	CmdInfoChange:    "角色信息变更通知",
	CmdSendChat:      "聊天请求",
	CmdNewChat:       "新聊天消息",

	CmdStrengthenEquipment: "强化或提升品质装备",
	CmdReplaceEquipment:    "更换装备",
	CmdUnloadEquipment:     "卸下装备",

	CmdUseItem:                "使用道具",
	CmdSellItem:               "出售道具",
	CmdNewFightCardBag:        "新建卡组",
	CmdDelFightCardBag:        "删除卡组",
	CmdFuseCard:               "融合卡牌",
	CmdDecomposeCard:          "分解卡牌",
	CmdAddFightCard:           "添加战斗卡牌",
	CmdDelFightCard:           "删除战斗卡牌",
	CmdOrderBag:               "整理背包",
	CmdExpandBag:              "扩展背包",
	CmdExtractCard:            "抽取武将卡牌",
	CmdReceiveCard:            "领取抽卡赠送的卡牌",
	CmdExtractSkillCard:       "抽取技能卡牌",
	CmdBuyCardInfo:            "获取抽卡和奖励信息",
	CmdBuyPower:               "购买体力",
	CmdGetFreePower:           "领取免费体力",
	CmdGetLevelInfo:           "读取章节关卡信息",
	CmdGetPointInfo:           "读取关卡战斗点信息",
	CmdChapterShuffle:         "征战洗牌",
	CmdChapterBattleBegin:     "征战战斗开始",
	CmdChapterBattleEnd:       "征战战斗结束",
	CmdGetMailList:            "获取邮件列表",
	CmdGetMailItem:            "领取(删除)邮件",
	CmdNewMail:                "新邮件通知",
	CmdSelectVocation:         "选择职业",
	CmdGetReferrer:            "获取推荐好友列表",
	CmdGetFriend:              "获取好友列表",
	CmdAddFriend:              "添加好友",
	CmdAddFriendNotify:        "添加好友通知",
	CmdDealFriend:             "同意或者拒绝添加好友",
	CmdDealFriendNotify:       "同意或者拒绝添加好友",
	CmdDeleteFriend:           "删除好友",
	CmdDeleteFriendNotify:     "删除好友通知",
	CmdFriendSend:             "赠送物品",
	CmdFriendSendNotify:       "赠送物品通知",
	CmdFriendRecv:             "领取赠送物品",
	CmdFriendInfoChangeNotify: "好友状态变更通知",
	CmdGetTaskList:            "获取任务列表",
	CmdGetRoleFightInfo:       "获取玩家战斗信息",

	CmdReceiveTaskReward: "领取任务奖励",

	CmdReadMail:       "阅读邮件",
	CmdRecharge:       "充值",
	CmdRechargeNotify: "充值到账通知",

	//server
	CmdServerConnectGate: "逻辑服务器连接网关",
	CmdClientLogout:      "客户端登出",
	CmdAdminConnect:      "管理后台连接网关",
	CmdAdminMessage:      "管理后台操作",
}

const (
	//ServerIDCenterBegin 游戏服务器起始ID
	ServerIDCenterBegin = 1
	//ServerIDPublicBegin 公共服务器起始ID
	ServerIDPublicBegin = 10001
)

//ServerStatus 服务器状态
type ServerStatus int

const (
	//ServerStatusHot 火热
	ServerStatusHot ServerStatus = iota
	//ServerStatusNew 新服
	ServerStatusNew
	//ServerStatusMaintain 维护
	ServerStatusMaintain
)

//定时器时间ID列表
const (
	//TimerPower 体力恢复
	TimerPower = 0
)

// 元宝	40001
// 金币	40002
// 体力	40003
// 荣誉	40004
// 个人贡献	40005
// 军团贡献	40006
// 回廊币	40007
//常用属性ID
const (
	//IDGold 元宝
	IDGold = 40001
	//IDMoney 金币
	IDMoney = 40002
	//IDPower 体力
	IDPower = 40003
	//IDHonor 荣誉
	IDHonor = 40004
	//IDPay 个人贡献
	IDPay = 40005
	//IDArmyPay 军团贡献
	IDArmyPay = 40006
	// IDCoin 回廊币
	IDCoin = 40007
	// IDExpr 经验
	IDExpr = 40008
)

//时间
const (
	Day    = 86400
	Hour   = 3600
	Minute = 60
)

//颜色
const (
	//ColorWhite 白色
	ColorWhite = 0
	//ColorGreen 绿色
	ColorGreen = 1
	//ColorBlue 蓝色
	ColorBlue = 2
	//ColorPurple 紫色
	ColorPurple = 3
	//ColorOrange 橙色
	ColorOrange = 4
	//ColorRed 红色
	ColorRed = 5
)

//性别
const (
	//SexMale 男
	SexMale = 1
	//SexFemale 女
	SexFemale = 2
)

//MaxP 概率基数
var MaxP = 10000

//ServerID 当前游戏服务器ID
var ServerID int

//IsEquipment 指定ID是否属于装备
func IsEquipment(ID int) bool {
	return (ID / 10000) == 1
}

//EquipmentPos 返回指定装备的位置
func EquipmentPos(ID int) int {
	return (ID / 1000) % 10
}

//IsCard 指定ID是否属于卡牌
func IsCard(ID int) bool {
	return IsHeroCard(ID) || IsSkillCard(ID)
}

//IsHeroCard 指定ID是否属于英雄卡牌
func IsHeroCard(ID int) bool {
	return (ID / 10000) == 2
}

//IsSkillCard 指定ID是否属于技能卡牌
func IsSkillCard(ID int) bool {
	return (ID / 10000) == 3
}

//IsItem 指定ID是否属于道具
func IsItem(ID int) bool {
	return (ID / 10000) == 4
}
