package GxStatic

/**
作者： Kyle Ding
模块：返回值
说明：
创建时间：2015-10-30
**/

//增加新返回值需要同时修改RetString
const (
	RetSucc              = 0 // 成功
	RetFail              = 1 // 失败
	RetMessageNotSupport = 2 // 不支持这个消息
	RetMsgFormatError    = 3 // 消息格式错误
	RetPwdError          = 4 // 密码错误
	RetUserNotExists     = 5 // 用户不存在
	RetUserExists        = 6 // 用户已经存在
	RetServerNotExists   = 7 // 服务器不存在
	RetServerMaintain    = 8 // 服务器正在维护
	RetAdminDistrust     = 9 // 管理平台客户端连接不在信任列表中
	//
	RetTokenError       = 11 // 鉴权失败
	RetRoleNotExists    = 12 // 游戏角色不存在
	RetRoleExists       = 13 // 游戏角色已经存在
	RetRoleNameConflict = 14 // 角色名重复
	RetNotLogin         = 15 // 尚未登录

	RetLoginConflict  = 16 //登录冲突
	RetLevNotEnough   = 17 //等级不足
	RetMoneyNotEnough = 18 //金币不足
	RetGoldNotEnough  = 19 //元宝不足

	RetPowerFull      = 20 //体力已经达到上限
	RetPowerNotEnough = 21 //体力不足

	RetEquiNotExists  = 22 //装备不存在
	RetItemNotReplace = 23 //道具不能装备

	RetCardNotEnough = 24 //卡片不足
	RetCardNotExists = 25 //卡片不存在
	RetCardMaxLevel  = 26 //卡片已经最高等级

	RetItemNotEnough = 27 //道具数量不足
	RetItemCanotSell = 28 //物品不能出售
	RetItemNotExists = 29 //道具不存在
	RetItemCanotOpen = 30 //道具不能打开

	RetBagFull            = 31 // 背包已满
	RetBagZoneIsNotEnough = 32 // 背包空间不足
	RetFightBagMax        = 33 //卡组数量已经达到最大限制
	RetFightBagNotExists  = 34 //卡组不存在
	RetBagCannotExpand    = 36 //背包已不能再扩充
	RetRoleNotOnline      = 37 //游戏角色不在线
	RetNotFreePowerTime   = 38 //不是领取免费体力时间
	RetFreePowerBeGet     = 39 //免费体力已经领取
	RetMaxBuyPowerCnt     = 40 //已经达到当天最大购买体力的次数

	RetEquipCannotStrengen = 41 //该装备不能强化
	RetEquipCannotColorup  = 42 //该装备不能进阶
	RetEquipColorupErr     = 43 //只能使用同类型装备做强化材料
	RetChapterFighting     = 44 //正在征战副本中
	RetChapterNotFighting  = 45 //你没有征战副本中

	RetExtractCardType      = 46 // 抽卡的类型
	RetCostGoldNumNotEnough = 47 //抽奖使用的元宝还没到达领取奖励的要求
	RetReceiveCardNotExists = 48 //拿过该类型奖励了
	RetExtractSkillCardType = 49 //抽技能卡的类型

	RetMailNotExists     = 50 //该邮件不存在
	RetVocationNotExists = 51 //该职业不存在
	RetVocationSelected  = 52 //职业已经选择过了
	RetHasBeenFriend     = 53 //已经是好友
	RetNotUnacceptFriend = 54 //不是待确认好友
	RetNotFriend         = 55 //不是正式好友

	RetHasSendFriend         = 56 //今天已经赠送过该好友
	RetHasRecvFriend         = 57 //今天已经领取过该好友赠送的物品
	RetRecvFriendMax         = 58 //已经达到当天最大领取次数
	RetCompleteTask          = 59 //完成任务
	RetTaskNotComplete       = 60 //任务未完成
	RetTaskRewardwasReceived = 61 //任务奖励已经领取了
)

//RetString 返回值描述信息
var RetString = map[uint16]string{
	RetSucc:              "成功",
	RetFail:              "失败",
	RetMessageNotSupport: "不支持这个消息",
	RetMsgFormatError:    "消息格式错误",
	RetPwdError:          "密码错误",
	RetUserNotExists:     "用户不存在",
	RetUserExists:        "用户已经存在",
	RetServerNotExists:   "服务器不存在",
	RetServerMaintain:    "服务器正在维护",
	RetAdminDistrust:     "管理平台客户端连接不在信任列表中",
	//
	RetTokenError:       "鉴权失败",
	RetRoleNotExists:    "游戏角色不存在",
	RetRoleExists:       "游戏角色已经存在",
	RetRoleNameConflict: "角色名重复",
	RetNotLogin:         "尚未登录",

	RetLoginConflict:  "登录冲突",
	RetLevNotEnough:   "等级不足",
	RetMoneyNotEnough: "金币不足",
	RetGoldNotEnough:  "元宝不足",

	RetPowerFull:      "体力已经达到上限",
	RetPowerNotEnough: "体力不足",

	RetEquiNotExists:  "装备不存在",
	RetItemNotReplace: "道具不能装备",

	RetCardNotEnough: "卡片不足",
	RetCardNotExists: "卡片不存在",
	RetCardMaxLevel:  "卡片已经最高等级",

	RetItemNotEnough: "道具数量不足",
	RetItemCanotSell: "物品不能出售",
	RetItemNotExists: "道具不存在",
	RetItemCanotOpen: "道具不能打开",

	RetBagFull:             "背包已满",
	RetBagZoneIsNotEnough:  "背包空间不足",
	RetFightBagMax:         "卡组数量已经达到最大限制",
	RetFightBagNotExists:   "卡组不存在",
	RetBagCannotExpand:     "背包已不能再扩充",
	RetRoleNotOnline:       "游戏角色不在线",
	RetNotFreePowerTime:    "不是领取免费体力时间",
	RetFreePowerBeGet:      "免费体力已经领取",
	RetMaxBuyPowerCnt:      "已经达到当天最大购买体力的次数",
	RetEquipCannotStrengen: "该装备不能强化",
	RetEquipCannotColorup:  "该装备不能进阶",
	RetEquipColorupErr:     "只能使用同类型装备做强化材料",
	RetChapterFighting:     "正在征战副本中",
	RetChapterNotFighting:  "你没有征战副本中",

	RetExtractCardType:      "抽卡的类型",
	RetCostGoldNumNotEnough: "抽奖使用的元宝还没到达领取奖励的要求",
	RetReceiveCardNotExists: "拿过该类型奖励了",
	RetExtractSkillCardType: "抽技能卡的类型",

	RetMailNotExists:         "该邮件不存在",
	RetVocationNotExists:     "该职业不存在",
	RetVocationSelected:      "职业已经选择过了",
	RetHasBeenFriend:         "已经是好友",
	RetNotUnacceptFriend:     "不是待确认好友",
	RetNotFriend:             "不是正式好友",
	RetHasSendFriend:         "今天已经赠送过该好友",
	RetHasRecvFriend:         "今天已经领取过该好友赠送的物品",
	RetRecvFriendMax:         "已经达到当天最大领取次数",
	RetCompleteTask:          "完成任务",
	RetTaskNotComplete:       "任务未完成",
	RetTaskRewardwasReceived: "任务奖励已经领取了",
}
