package gql

import (
	"context"
	"database/sql"
)

type Model struct {
	Table string
	Connection string
	Fillable []string
	Scanner func() interface{}
	Relations map[string]Relation
	PrimaryKey string
	query query
}

type unionQuery struct {
	unionQuery *string
	unionParams *[]interface{}
}


type query struct {
	query []whereQuery
	combinationWhere map[int]int
	whereCombination bool
	params []interface{}
	selected []string
	append string
	order string
	groupBy []string
	limit string
	joins []string
	transaction bool
	sqlTransaction *sql.Tx
	queryContext *context.Context
	lock string
	exists bool
	union []unionQuery
}

type whereQuery struct {
	column string
	op string
	value string
	in []string
	query string
	existsQuery *string
	existsParams *[]interface{}

}


type Relation struct {
	relationType string
	relationTable string
	foreignKey string
	localKey string
	middleTable string
	relationModelForeignKey string
	relationModelLocalKey string
}

type DataItem interface{}

type ExecResult struct {
	Affected int64
	LastId int64
}

type BoolScanner struct {
	Result bool `db:"result"`
}