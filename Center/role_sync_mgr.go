package main

/**
作者： Kyle Ding
模块：角色同步模块
说明：定时将缓存数据同步到mysql中
创建时间：2015-11-2
**/

import (
	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxMisc"
	"git.oschina.net/jkkkls/goxiang/GxStatic"
)

//SyncRole 更新新创建的角色到mysql，同时每天定时同步role
func SyncRole() {
	go SyncCreateRole()
	go DailySyncRole()
}

//SyncCreateRole 更新新创建的角色到mysql
func SyncCreateRole() {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)
	key := GxStatic.RoleCreateList + strconv.Itoa(config.ID)

	for {
		//每五分钟检查是否有新数据
		t := time.NewTicker(5 * time.Minute)
		select {
		case <-t.C:
			for {
				if rdClient.SCard(key).Val() == 0 { // 没有数据更新
					break
				}

				roleID := rdClient.SRandMember(key).Val()
				rdClient.SRem(key, roleID)
				GxMisc.Debug("sync create role: %s", roleID)

				//common
				ID, _ := strconv.Atoi(roleID)
				saveRole(rdClient, ID)
			}
		}
	}
}

//DailySyncRole 每天定时同步role
func DailySyncRole() {
	rdClient := GxMisc.PopRedisClient()
	defer GxMisc.PushRedisClient(rdClient)

	for {
		now := time.Now()
		tvl := GxMisc.NextTime(4, 0, 0) - now.Unix()

		t := time.NewTicker(time.Duration(tvl) * time.Second)
		select {
		case <-t.C:
			var infos []*GxStatic.RoleLoginInfo
			GxStatic.GetAllRoleLogin(rdClient, uint32(config.ID), &infos)
			n1 := len(infos)
			n2 := 0
			for i := 0; i < len(infos); i++ {
				saveRole(rdClient, infos[i].RoleID)

				//七天没有登录过，则清除部分缓存 /*Jie Del表示缓存是否被清理 , Ts:保存玩家登陆的时间*/
				if infos[i].Del == 1 && (now.Unix()-infos[i].Ts) >= int64(7*24*time.Hour.Seconds()) {
					n2++
					delRoleCache(rdClient, infos[i].RoleID)
					//
					infos[i].Del = 1
					GxStatic.SaveRoleLogin(rdClient, uint32(config.ID), infos[i])
				}
			}
			GxMisc.Debug("Daily Sync Role, time: %s, sync-role-count: %d, del-role-cache-count: %d, next-ts", //
				GxMisc.TimeToStr(now.Unix()), n1, n2, GxMisc.TimeToStr(GxMisc.NextTime(4, 0, 0)))
		}
	}
}

func saveBag(client *redis.Client, roleID int) {
	b := GxStatic.NewBag(roleID)
	b.Get4Redis(client)
	b.Set4Mysql()
}

func saveCard(client *redis.Client, roleID int) {
	c := GxStatic.NewCard(roleID)
	c.Get4Redis(client)
	c.Set4Mysql()
}

func saveFightCard(client *redis.Client, roleID int) {
	for i := 0; i < 32; i++ {
		c := GxStatic.NewFightCard(roleID, i, "")
		err := c.Get4Redis(client)
		if err != nil {
			c.Del4Mysql()
		} else {
			c.Set4Mysql()
		}

	}
}

func saveBuyCard(client *redis.Client, roleID int) {
	bc := GxStatic.NewBuyCard(roleID)
	bc.Get4Redis(client)
	bc.Set4Mysql()
}

func saveChapter(client *redis.Client, roleID int) {
	c := GxStatic.NewChapter(roleID)
	c.Get4Redis(client)
	c.Set4Mysql()
}

func saveMail(client *redis.Client, roleID int) {
	mails := GxStatic.GetRoleAllMail(client, roleID)
	for i := 0; i < len(mails); i++ {
		if mails[i].Del == 0 {
			mails[i].Set4Mysql()
		} else {
			//彻底删除文件
			mails[i].Del4Mysql()
			GxStatic.DelRoleMail(client, roleID, mails[i].ID)
		}

	}
}

func saveFriend(client *redis.Client, roleID int) {
	c := GxStatic.NewFriend(roleID)
	c.Get4Redis(client)
	c.Set4Mysql()
}

func saveRole(client *redis.Client, roleID int) {
	r := new(GxStatic.Role)
	r.ID = roleID
	r.Get4Redis(client)
	r.Set4Mysql()

	//bag
	saveBag(client, r.ID)

	//card
	saveCard(client, r.ID)

	//fight card
	saveFightCard(client, r.ID)

	//buy card
	saveBuyCard(client, roleID)

	//chapter
	saveChapter(client, r.ID)

	//mail
	saveMail(client, r.ID)

	//friend
	saveFriend(client, r.ID)
}

func delRoleCache(client *redis.Client, roleID int) {
	//常用信息，装备,卡组等不用清除

	//bag
	bag := &GxStatic.Bag{
		RoleID: roleID,
	}
	bag.Del4Redis(client)

	card := &GxStatic.Card{
		RoleID: roleID,
	}
	card.Del4Redis(client)

	//buycard
	bc := &GxStatic.BuyCard{
		RoleID: roleID,
	}
	bc.Del4Redis(client)

	//chapter
	chapter := &GxStatic.Chapter{
		RoleID: roleID,
	}
	chapter.Del4Redis(client)

	//mail
	mails := GxStatic.GetRoleAllMail(client, roleID)
	for i := 0; i < len(mails); i++ {
		//删除缓存
		mails[i].Del4Redis(client)
	}

	//friend
	friend := &GxStatic.Friend{
		RoleID: roleID,
	}
	friend.Del4Redis(client)
}
