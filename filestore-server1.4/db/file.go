package db

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"fmt"
)

// 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (file_sha1,file_name,file_size,file_addr,status) values(?,?,?,?,1)",
	)
	if err != nil {
		fmt.Println("Failed to prepare statement, err: " + err.Error())
		return false
	}
	defer stmt.Close()
	res, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println("Failed to execute statement, err: " + err.Error())
		return false
	}
	if affected, err := res.RowsAffected(); err == nil {
		if affected <= 0 {
			fmt.Printf("File with hash: %s has been uploaded before\n", filehash)
		}
		return true
	}
	// 操作失败
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// 查询文件信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file where file_sha1=? and status=1 limit 1",
	)
	if err != nil {
		fmt.Println("Failed to prepare statement, err: " + err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			fmt.Println("Failed to execute statement, err: " + err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}
