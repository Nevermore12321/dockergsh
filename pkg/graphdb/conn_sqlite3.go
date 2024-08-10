package graphdb

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3" // 注册 sqlite引擎
)

// NewSqliteConn 新建 sqlite 连接
func NewSqliteConn(root string) (*Database, error) {
	initDatabase := false

	// 判断存储文件存不存在
	stat, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) { // 存储文件不存在，需要创建
			initDatabase = true
		} else {
			return nil, err
		}
	}

	// 如果文件大小为0,说明没有初始化 database
	if stat != nil && stat.Size() == 0 {
		initDatabase = true
	}

	conn, err := sql.Open("sqlite3", root)
	if err != nil {
		return nil, err
	}
	return NewDatabase(conn, initDatabase)
}
