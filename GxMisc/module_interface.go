package GxMisc

/**
作者： Kyle Ding
模块：
说明：
创建时间：2015-10-30
**/

import (
	"database/sql"
	"gopkg.in/redis.v3"
)

//GxModule 系统模块接口
type GxModule interface {
	Set4Redis(client *redis.Client) error
	Get4Redis(client *redis.Client) error
	Del4Redis(client *redis.Client) error
	SetField4Redis(client *redis.Client, field string) error
	GetField4Redis(client *redis.Client, field string) error
	Set4Mysql(db *sql.DB) error
	Get4Mysql(db *sql.DB) error
	Del4Mysql(db *sql.DB) error
}
