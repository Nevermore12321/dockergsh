package graphdb

const (
	createEntityTable = `
    CREATE TABLE IF NOT EXISTS entity (
        id text NOT NULL PRIMARY KEY
    );`

	createEdgeTable = `
    CREATE TABLE IF NOT EXISTS edge (
        "entity_id" text NOT NULL,
        "parent_id" text NULL,
        "name" text NOT NULL,
        CONSTRAINT "parent_fk" FOREIGN KEY ("parent_id") REFERENCES "entity" ("id"),
        CONSTRAINT "entity_fk" FOREIGN KEY ("entity_id") REFERENCES "entity" ("id")
        );
    `

	createEdgeIndices = `
    CREATE UNIQUE INDEX IF NOT EXISTS "name_parent_ix" ON "edge" (parent_id, name);
    `
)

var databaseSql map[string]string = map[string]string{
	"Entity":      createEntityTable,
	"Edge":        createEdgeTable,
	"EdgeIndices": createEdgeIndices,
}

const (
	insertEntitySql = "INSERT INTO entity (id) VALUES (?);"
	insertEdgeSql   = "INSERT INTO edge (entity_id, name) VALUES(?,?);"
)

var initSql map[string]string = map[string]string{
	"Entity": insertEntitySql,
	"Edge":   insertEdgeSql,
}
