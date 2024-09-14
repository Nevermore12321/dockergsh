package graphdb

import (
	"database/sql"
	"errors"
	"fmt"
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

// Entity 通过 id 关联唯一存储实体
// entity 表通常用来存储关于 Docker 镜像、容器或层 (layer) 的元数据。每个 entity 代表 Docker 存储中的一个组件，比如一个容器层或镜像层
type Entity struct {
	id string
}

// Edge 不同实体间的关联关系
// edge 表主要用于定义这些实体之间的关系。在 Docker 的分层文件系统中，不同的镜像层（实体）会彼此依赖，edge 表用来维护这些依赖关系。
type Edge struct {
	Entity   string // 实体 id
	Name     string // 实体 名称
	ParentId string // 实体的父级
}

type Entities map[string]*Entity
type Edges []*Edge

func (db *Database) Close() error {
	return db.conn.Close()
}

// Get 根据路径层级 name，一直寻找到最后一层的 Entity
func (db *Database) Get(name string) *Entity {
	db.mux.RLock()
	defer db.mux.RUnlock()

	entity, err := db.get(name)
	if err != nil {
		return nil
	}
	return entity
}

// get 根据路径层级 name，一直寻找到最后一层的 Entity
func (db *Database) get(name string) (*Entity, error) {
	entity := db.RootEntity()

	if name == "/" { // 如果就是要找根实体，直接返回
		return entity, nil
	}

	// 根据 "/" 分隔成多个部分
	parts := split(name)

	// 遍历每一层，寻找存在 Edge 表中的实体
	for i := 1; i < len(parts); i++ { // 第0个元素是空，因为路径是 /
		p := parts[i]
		if p == "" {
			continue
		}

		// 从根开始，一直往下层寻找
		next := db.child(entity, p)
		if next == nil { // 没找到
			return nil, fmt.Errorf("Cannot find child for %s", name)
		}

		entity = next
	}

	return entity, nil
}

// RootEntity 返回实体的根 "/"，所有的实体都是从 0 实体开始
func (db *Database) RootEntity() *Entity {
	return &Entity{
		id: "0",
	}
}

// 根据 parent 和 name 可以从 Edge 表中唯一找到一条 Entity 数据
func (db *Database) child(parent *Entity, name string) *Entity {
	var id string
	if err := db.conn.QueryRow("SELECT entity_id FROM edge WHERE parent_id = ? and name = ?", parent.id, name).Scan(&id); err != nil {
		return nil
	}
	return &Entity{
		id: id,
	}
}

// Set 根据 id 添加实体，并根据提供的 path 添加实体关系
func (db *Database) Set(fullPath, id string) (*Entity, error) {
	// 同一时间只有一个 goroutine 可以写
	db.mux.Lock()
	defer db.mux.Unlock()

	// rollback 函数
	rollback := func() {
		db.conn.Exec("ROLLBACK ")
	}

	// 开启排他锁
	if _, err := db.conn.Exec("BEGIN EXCLUSIVE"); err != nil {
		return nil, err
	}

	var entityId string

	// 先查询是否已经存在此 Entity
	if err := db.conn.QueryRow("SELECT id FROM entity WHERE id = ?", id).Scan(&entityId); err != nil {
		if err == sql.ErrNoRows { // 不存在，插入新 entity
			if _, err := db.conn.Exec("INSERT INTO entity (id) VALUES (?)", id); err != nil {
				rollback() // 插入失败，则回滚当前事务
				return nil, err
			}
		} else { // 已存在，直接报错返回
			rollback()
			return nil, err
		}
	}

	e := &Entity{
		id: entityId,
	}

	// 插入此 Entity 的关联关系，也就是其父ParentId
	// fullPath 可以有多层
	parentPath, name := splitPath(fullPath)
	if err := db.setEdge(parentPath, name, e); err != nil {
		rollback()
		return nil, err
	}
	// 提交事务，释放锁
	if _, err := db.conn.Exec("COMMIT"); err != nil {
		return nil, err
	}

	return e, nil
}

// 插入 Edge 表数据，在路径层 parentPath 的叶子节点后，添加一个名为 name 的 entity e，
func (db *Database) setEdge(parentPath, name string, e *Entity) error {
	// 寻找路径为 parentPath 的叶子 entity
	parent, err := db.get(parentPath)
	if err != nil {
		return nil
	}

	if parent.id == e.id { // 如果当前叶子 Entity 和要插入的 Entity 相同，报错
		return fmt.Errorf("Cannot set self as child")
	}

	if _, err := db.conn.Exec("INSERT INTO edge (parent_id, name, entity_id) VALUES (?, ?, ?)", parent.id, name, e.id); err != nil {
		return err
	}
	return nil
}
