package mysql

import (
	"database/sql"
	"filestore-server/config"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=%s",
		config.Config.MysqlConfig.DBUser,
		config.Config.MysqlConfig.DBPassword,
		config.Config.MysqlConfig.DBHost,
		config.Config.MysqlConfig.DBPort,
		config.Config.MysqlConfig.DBName,
		config.Config.MysqlConfig.DBCharset)
	fmt.Println(dsn)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1000) // 最大连接数
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("mysql连接失败")
	}
	log.Println("mysql连接成功")
}

// 对外返回数据库连接对象
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
