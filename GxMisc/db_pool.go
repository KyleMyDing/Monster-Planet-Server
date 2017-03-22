package GxMisc

/**
作者： Kyle Ding
模块：mysql连接池
说明：
创建时间：2015-10-30
**/

import (
	"database/sql"
	"fmt"
	// "github.com/go-sql-driver/mysql"
)

//MysqlPool mysql连接池
var MysqlPool *sql.DB

//InitMysql mysql连接池初始化函数，在程序启动时候调用
func InitMysql(user string, pwd string, host string, port int, dbs string, charset string) error {
	var err error
	connInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", user, pwd, host, port, dbs, charset)
	MysqlPool, err = sql.Open("mysql", connInfo)
	if err != nil {
		return err
	}
	MysqlPool.SetMaxOpenConns(128)
	MysqlPool.SetMaxIdleConns(64)
	MysqlPool.Ping()

	return nil
}
