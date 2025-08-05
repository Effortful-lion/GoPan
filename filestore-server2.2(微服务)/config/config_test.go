package config

import (
	"fmt"
	"testing"
)

func TestInitConfig(t *testing.T) {
	InitConfig()

	// 检查配置是否正确初始化
	if Config.MysqlConfig.DBUser == "" || Config.MysqlConfig.DBHost == "" || Config.MysqlConfig.DBName == "" {
		t.Errorf("MySQL configuration is not properly initialized")
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=%s",
		Config.MysqlConfig.DBUser,
		Config.MysqlConfig.DBPassword,
		Config.MysqlConfig.DBHost,
		Config.MysqlConfig.DBPort,
		Config.MysqlConfig.DBName,
		Config.MysqlConfig.DBCharset)
	t.Log(dsn)

	// 屏蔽密码部分
	maskedDSN := fmt.Sprintf("%s:******@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=%s",
		Config.MysqlConfig.DBUser,
		Config.MysqlConfig.DBHost,
		Config.MysqlConfig.DBPort,
		Config.MysqlConfig.DBName,
		Config.MysqlConfig.DBCharset)
	t.Log(maskedDSN)
}
