package graphdb

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
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
	EntityId string // 实体 id
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
			return nil, fmt.Errorf("cannot find child for %s", name)
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

// Exists 判断名为 name 的 entity 存不存在
func (db *Database) Exists(name string) bool {
	db.mux.RLock()
	defer db.mux.RUnlock()

	e, err := db.get(name)
	if err != nil {
		return false
	}
	return e != nil
}

type WalkMeta struct {
	Parent   *Entity // 父 Entity
	Entity   *Entity // 当前 Entity
	FullPath string  // 完成的路径，通过组合父实体的路径和当前边的名称来构建，便于追踪实体在树形结构中的位置。
	Edge     *Edge   // 当前 Entity 所有关联的 Entity，从父实体到当前实体的连接信息
}

// 查找所有 e 下所有 子 entity，查找深度是 depth
func (db *Database) children(e *Entity, name string, depth int, entities []WalkMeta) ([]WalkMeta, error) {
	// 如果当前 entity 不存在，直接返回
	if e == nil {
		return entities, nil
	}

	// 根据当前实体 entity id，查找所有 子 entity
	rows, err := db.conn.Query("SELECT entity_id, name FROM edge where parent_id = ?", e.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 遍历每一个 子 entity，构建出子 Entity 实例
	for rows.Next() {
		var entityId, entityName string
		if err := rows.Scan(&entityId, &entityName); err != nil { // 获取下一个子entity
			return nil, err
		}
		// 构建 子Entity 和 Edge 实例
		child := &Entity{
			id: entityId,
		}
		edge := &Edge{
			EntityId: entityId,
			Name:     entityName,
			ParentId: e.id,
		}
		// 构建 WalkMeta
		meta := &WalkMeta{
			Parent:   e,
			Entity:   child,
			FullPath: filepath.Join(name, edge.Name),
			Edge:     edge,
		}
		entities = append(entities, *meta)

		// 如果传入的 查找深度 depth 不为0,需要递归查找 孙子 entity 等等
		if depth != 0 {
			nDepth := depth
			if depth != -1 { // 递归深度 -1
				nDepth -= 1
			}
			entities, err = db.children(child, meta.FullPath, nDepth, entities)
			if err != nil {
				return nil, err
			}
		}
	}

	return entities, nil
}

// List 按名称列出所有实体 键将是实体的完整路径
// 根据深度找到所有 名为 name 的 实体
func (db *Database) List(name string, depth int) Entities {
	db.mux.RLock()
	defer db.mux.RUnlock()

	// 获取最上层的名为  name 的实体
	out := Entities{}
	e, err := db.get(name)
	if err != nil {
		return out
	}

	// 查找所有子节点
	children, err := db.children(e, name, depth, nil)
	if err != nil {
		return out
	}

	for _, child := range children {
		out[child.FullPath] = child.Entity
	}

	return out
}

// Id 获取 Entity 实例中的 id
func (e *Entity) Id() string {
	return e.id
}

// Paths 返回按深度排序的路径
func (es *Entities) Paths() []string {
	out := make([]string, len(*es))
	var i int
	for _, e := range *es {
		out[i] = e.id
		i++
	}

	sortByDepth(out)

	return out
}
