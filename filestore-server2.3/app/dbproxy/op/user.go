package op

import (
	mydb "filestore-server/app/dbproxy/mysql"
	"filestore-server/common"
	"fmt"
)

func UserSignup(username string, passwd string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`, `user_pwd`) values (?, ?)")
	if err != nil {
		fmt.Println("Failed to insert, err: ", err.Error())
		return false
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, passwd)
	if err != nil {
		fmt.Println("Failed to insert, err: ", err.Error())
		return false
	}
	if rowsAffected, err := res.RowsAffected(); err == nil {
		if rowsAffected <= 0 {
			// sql执行成功，但是重复没有插入数据
			fmt.Printf("User with username: %s has been signed up", username)
		}
		return true
	}
	// sql执行失败
	return false
}

// 判断密码是否一致
func UserSignin(username string, passwd string) bool {
	// limit作用：一定那找到记录，就会立即停止搜索，避免对剩余数据进行不必要的扫描，从而提高查询性能。
	stmt, err := mydb.DBConn().Prepare(
		"select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println("Failed to select, err: ", err.Error())
		return false
	}
	defer stmt.Close()
	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println("Failed to select, err: ", err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found:", username)
		return false
	}

	prows := mydb.ParseRows(rows)
	if len(prows) > 0 && string(prows[0]["user_pwd"].([]byte)) == passwd {
		return true
	}
	return false
}

// 刷新用户token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token(`user_name`, `user_token`) values (?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err: ", err.Error())
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println("Failed to insert, err: ", err.Error())
		return false
	}
	return true
}

// 查询用户信息
func GetUserInfo(username string) (common.User, error) {
	user := common.User{}
	fmt.Println("开始数据库查询user")
	stmt, err := mydb.DBConn().Prepare(
		"select user_name, signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println("Failed to select, err: ", err.Error())
		return user, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(username).Scan(&user.UserName, &user.SignupAt)
	if err != nil {
		fmt.Println("Failed to select, err: ", err.Error())
		return user, err
	}
	fmt.Println("数据库查询user成功：", user)
	return user, nil
}
