package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:mysql123@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000) // 最大连接数
	err := db.Ping()
	if err != nil {
		panic(err)
	}
}

// 对外返回数据库连接对象
func DBConn() *sql.DB {
	return db
}
