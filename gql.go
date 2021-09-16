package gql

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

var sqlDB *sql.DB

func Connect(driver string,dataSource string) *sql.DB{
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	sqlDB=db
	return db
}


func (m *Model) Select(cols ...string) *Model{
	for _,col := range cols {
		if inStringArray(m.query.selected,col) == false{
			m.query.selected = append(m.query.selected,col)
		}
	}
	return m
}


func (m *Model) UseScanner(scanner func() interface{}) *Model {
	m.Scanner = scanner
	return m
}

func (m *Model) Exists()  (bool,error){
	m.query.exists = true
	m.UseScanner(func() interface{} {
		return &BoolScanner{}
	})
	data,err :=m.Get()
	if err != nil{
		return false,err
	}
	result:=data[0].(*BoolScanner)
	return result.Result,err
}

func (m *Model) Where(column string ,op string,value string) *Model{
	if m.query.whereCombination != true{

		if stringHasDot(column) == false{
			column = m.Table+"."+column
		}

		m.query.query = append(m.query.query, whereQuery{
			column: column,
			op:     op,
			value:  value,
			query:"where",
		})
	}

	return m
}


func (m *Model) OrWhere(column string,op string,value string) *Model{
	m.query.query = append(m.query.query, whereQuery{
		column: column,
		op:     op,
		value:  value,
		query:"or",
	})
	return m
}

func (m *Model) WhereCombination(query func(m *Model)) *Model {
	lastQueryIndex:=len(m.query.query)
	query(m)
	currentQueryIndex:=len(m.query.query) -1
	checkWhereCombinationMap(m)
	m.query.combinationWhere[lastQueryIndex]=currentQueryIndex
	return m
}

func (m *Model) WhereExists(query func() *Model) *Model {
	model:=query()
	queryString,params:=buildQuery(model)
	m.query.query = append(m.query.query, whereQuery{
		query:"exists",
		existsParams: params,
		existsQuery: queryString,
	})
	return m
}

func (m *Model) Union(query func() *Model) *Model {
	model:=query()
	queryString,params:=buildQuery(model)
	m.query.union = append(m.query.union, unionQuery{
		unionQuery:  queryString,
		unionParams: params,
	})
	return m
}


func (m *Model) WhereIn(column string,value []string) *Model {
	m.query.query = append(m.query.query, whereQuery{
		column: column,
		in:value,
		query:"in",
	})
	return m
}

func (m *Model) GroupBy(groupBy ...string)  *Model{
	for _,col := range groupBy {
		if inStringArray(m.query.groupBy,col) == false{
			m.query.groupBy = append(m.query.groupBy,col)
		}
	}
	return m
}

func (m *Model) OrderBy(column string,orderType string)  *Model{
	m.query.order = column+" "+orderType
	return m
}

func (m *Model) With(relationName string) *Model{
	m.query.joins = append(m.query.joins, relationName)
	return m
}


func (m *Model) Find(primaryKeyValue int64) (DataItem,error){
	m.Limit(1)
	m.query.query = []whereQuery{}
	m.query.query = append(m.query.query,whereQuery{
		column: getPrimaryKey(m),
		op:     "=",
		value:  strconv.FormatInt(primaryKeyValue, 10),
		query:"where",

	})

	items,err:=m.Get()
	if err != nil{
		return nil,err
	}
	return items[0],err
}

func (m *Model) Limit(limit int)  *Model{
	m.query.limit = strconv.Itoa(limit)
	return m
}

func (m *Model) First() (DataItem,error){
	m.Limit(1)
	m.OrderBy(getPrimaryKey(m),"asc")
	items,err:=m.Get()
	if err != nil{
		return nil,err
	}
	return items[0],err
}

func (m *Model) Latest() (DataItem,error){
	m.Limit(1)
	m.OrderBy(getPrimaryKey(m),"desc")
	items,err:=m.Get()
	if err != nil{
		return nil,err
	}
	return items[0],err
}

func (m *Model) Get() ([]DataItem,error) {
	return sqlSelectQuery(m)
}


func (m *Model) HasRelation(relationName string,relatedTable string,foreignKey string,localKey string) *Model{
	checkRelationsMap(m)
	m.Relations[relationName]=Relation{relationType: "hasRelation",relationTable: relatedTable,foreignKey: foreignKey,localKey: localKey}
	return m
}

func (m *Model) BelongsToMany(relationName string,relatedTable string,foreignKey string,localKey string,relatedForeignKey string,relatedLocalKey string,middleTable string) *Model{
	checkRelationsMap(m)
	m.Relations[relationName]=Relation{relationType: "belongsToMany",relationTable: relatedTable,foreignKey: foreignKey,localKey: localKey,relationModelForeignKey: relatedForeignKey,relationModelLocalKey: relatedLocalKey,middleTable: middleTable}
	return m
}


func (m *Model) Insert(insertObject interface{}) (int64,error){
	insertStmt,params:=buildInsertStmt(m,insertObject)
	result,err:=prepareAndExec(m,insertStmt,params,true)
	return result.LastId,err
}

func (m *Model) InsertAndReturn(insertObject interface{}) (DataItem,error) {
	id,err:= m.Insert(insertObject)
	if err != nil{
		return -1,err
	}
	return m.Find(id)
}


func (m *Model) Update(updatedObject interface{}) (int64,error) {
		updateStmt,params:=buildUpdateStmt(m,updatedObject)
		result,err:=prepareAndExec(m,updateStmt,params,false)
		return result.Affected,err
}

func (m *Model) UpdateAndReturn(updatedObject interface{}) ([]DataItem,error) {
	updateStmt,params:=buildUpdateStmt(m,updatedObject)
	_,err:=prepareAndExec(m,updateStmt,params,false)
	if err != nil{
		return nil,err
	}
	return m.Get()
}

func (m *Model) Delete() (int64,error) {
	if len(m.query.query) == 0{
		return 0,errors.New("you want to delete with out any conditions , so will delete all data, if you want this please use Truncate func")
	}
	var deleteStmt string
	var params []interface{}
	deleteStmt+="DELETE FROM "+m.Table
	buildWhereQuery(m,&deleteStmt,&params)

	result,err:=prepareAndExec(m,&deleteStmt,&params,false)

	return result.Affected,err
}

func (m *Model) Truncate() error {
	_,err:=sqlDB.Query("truncate table "+m.Table)
	return err
}


func (m *Model) Transaction(Tx *sql.Tx) *Model {
	m.query.transaction = true
	m.query.sqlTransaction = Tx
	return m
}

func Transaction(BeginContext *context.Context,TxOptions *sql.TxOptions,transaction func(tx *sql.Tx) error) error {
	tx, err := sqlDB.BeginTx(*BeginContext, TxOptions)
	if err != nil {
		return err
	}
	err=transaction(tx)
	if err != nil{
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func (m *Model) Context(Context *context.Context) *Model{
	m.query.queryContext = Context
	return m
}

func (m *Model) LockForUpdate() *Model {
	m.query.lock = " for update"
	return m
}

func (m *Model) ToSql() string {
	queryString,_:=buildQuery(m)
	return *queryString
}

func GetSqlConnection()  *sql.DB{
	return sqlDB
}




