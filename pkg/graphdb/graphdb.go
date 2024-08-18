package graphdb

import (
	"database/sql"
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Database 存储数据库连接和读写锁
type Database struct {
	conn *sql.DB
	mux  sync.RWMutex
}

// NewDatabase 创建 Database 对象，并且初始化表
func NewDatabase(conn *sql.DB, initDatabase bool) (*Database, error) {
	if conn == nil {
		return nil, errors.New("database connection cannot be nil")
	}
	database := &Database{
		conn: conn,
	}

	if initDatabase { // 需要创建表
		for tableName, sqlStr := range databaseSql {
			_, err := conn.Exec(sqlStr)
			log.Infof("Create sqlite database table: %s", tableName)
			if err != nil {
				return nil, err
			}
		}
		// rollback 回调函数
		rollback := func() {
			conn.Exec("ROLLBACK ")
		}

		// 创建初始化的 init 数据
		if _, err := conn.Exec("BEGIN"); err != nil {
			return nil, err
		}
		for _, sqlStr := range initSql {
			if _, err := conn.Exec(sqlStr); err != nil {
				rollback()
				return nil, err
			}
		}
		if _, err := conn.Exec("COMMIT"); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
