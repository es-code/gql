package gql

import (
	"database/sql"
	"log"
	"strings"
)

func buildQuery(m *Model) (*string,*[]interface{}){
	var sqlQuery string
	var params []interface{}

	sqlQuery += "select "+getSelected(m)+" from "+m.Table
	if len(m.query.joins) > 0 {
		buildJoins(m,&sqlQuery)
	}

	if len(m.query.query) > 0{
		buildWhereQuery(m,&sqlQuery,&params)
	}

	if len(m.query.groupBy) > 0{
		sqlQuery+=" group by "+strings.Join(m.query.groupBy,",")
	}

	if m.query.order !=""{
		sqlQuery+=" order by "+m.query.order
	}

	if m.query.limit !=""{
		sqlQuery+=" limit ?"
		params = append(params,m.query.limit)
	}

	if m.query.lock !=""{
		sqlQuery+=m.query.lock
	}
	sqlQuery = buildExists(m,sqlQuery)

	buildUnion(m,&sqlQuery,&params)

	return &sqlQuery,&params
}

func buildExists(m *Model,sqlQuery string) string {

	var queryWithExists string
	if m.query.exists == true{
		queryWithExists = "select exists("+sqlQuery+") as result"
		return queryWithExists
	}
	return sqlQuery
}

func buildUnion(m *Model,sqlQuery *string,params *[]interface{})  {
	if len(m.query.union) > 0{
		for i:=0;i<len(m.query.union);i++ {
			*sqlQuery+= " UNION ("+*m.query.union[i].unionQuery+")"
			unionParams:= *m.query.union[i].unionParams
			for x:=0; x < len(unionParams); x++{
				*params = append(*params,unionParams[x] )
			}
		}
	}
}


func sqlSelectQuery(model *Model)  ([]DataItem,error){

	query,params:=buildQuery(model)
	data:=[]DataItem{}
	var rows *sql.Rows
	var err error

	if model.query.transaction == true{
		rows,err=queryInTransaction(model,query,params)
	}else{
		rows,err=queryWithOutTransaction(model,query,params)
	}
	if err != nil {
		log.Fatal(err)
		return data,err
	}
	//// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return data,err
	}
	//get fields from struct
	scanner := model.Scanner()
	fields:=mapColumnsFields(scanner,columns)

	// Fetch rows
	for rows.Next() {
		Scan := model.Scanner()
		err = rows.Scan(getStrutFields(Scan,columns,fields)...)
		if err != nil {
			return data,err
		}
		data = append(data,Scan)
	}
	if err = rows.Err(); err != nil {
		return data,err
	}

	return data,err
}

func prepareAndExec(m *Model, stmt *string,params *[]interface{},returnInsertId bool) (ExecResult,error) {
	 Result := ExecResult{
		 Affected: 0,
		 LastId:   0,
	 }
	var sqlResult sql.Result
	var err error

	if m.query.transaction == true{
		sqlResult,err=execInTransaction(m,stmt,params)
	}else{
		sqlResult,err=execWithOutTransaction(m,stmt,params)
	}

	if err != nil{
		return Result,err
	}

	rowsAffected,err:=sqlResult.RowsAffected()
	if err!=nil{
		return Result,err
	}

	Result.Affected = rowsAffected

	if returnInsertId == true{
		LastId,err :=sqlResult.LastInsertId()
		if err != nil{
			return Result,err
		}
		Result.LastId = LastId
	}

	return Result,err

}

func execInTransaction(m *Model,stmt *string,params *[]interface{}) (sql.Result,error) {
	var result sql.Result
	var err error
	if m.query.queryContext != nil{
		result, err = m.query.sqlTransaction.ExecContext(*m.query.queryContext,*stmt, *params...)
	}else{
		result, err =  m.query.sqlTransaction.Exec(*stmt, *params...)
	}

	return result,err
}

func execWithOutTransaction(m *Model,stmt *string,params *[]interface{}) (sql.Result,error)  {
	var result sql.Result
	stmtOut, err := sqlDB.Prepare(*stmt)
	if err != nil {
		return result,err
	}
	defer stmtOut.Close()

	if m.query.queryContext != nil{
		result, err = stmtOut.ExecContext(*m.query.queryContext, *params...)
	}else{
		result, err = stmtOut.Exec(*params...)
	}

	if err != nil {
		return result,err
	}
	return result,err
}

func queryInTransaction(m *Model,query *string,params *[]interface{}) (*sql.Rows,error)  {
	var result *sql.Rows
	var err error
	if m.query.queryContext != nil{
		result, err = m.query.sqlTransaction.QueryContext(*m.query.queryContext,*query, *params...)
	}else{
		result, err = m.query.sqlTransaction.Query(*query, *params...)
	}
	return result,err
}

func queryWithOutTransaction(m *Model,query *string,params *[]interface{}) (*sql.Rows,error) {
	stmtOut, err := sqlDB.Prepare(*query)
	var rows *sql.Rows
	if err != nil {
		return rows,err
	}
	defer stmtOut.Close()

	if m.query.queryContext != nil{
		rows, err = stmtOut.QueryContext(*m.query.queryContext,*params...)
	}else{
		rows, err = stmtOut.Query(*params...)
	}

	if err != nil {
		return rows,err
	}
	return rows,err
}

func buildUpdateStmt(m *Model,updatedObject interface{}) (*string,*[]interface{}) {
	var updateStmt string
	columns,values := structFieldsValues(updatedObject)

	updateStmt+="UPDATE "+m.Table+" set "
	for i:=0;i<len(*columns);i++ {
		checkPrimaryKeyValue(m,i,columns,values)
		updateStmt+=(*columns)[i]+" = ?,"
	}

	updateStmt = strings.TrimRight(updateStmt, ",")
	if len(m.query.query) > 0{
		buildWhereQuery(m,&updateStmt,values)
	}
	return &updateStmt,values
}

func buildInsertStmt(m *Model,insertObject interface{}) (*string,*[]interface{}) {
	columns,values := structFieldsValues(insertObject)
	removeUnFillable(m,columns,values)
	var stmtQuery string
	var perpareValues []string

	for i:=0;i<len(*columns);i++ {
		//remove primary key from insert object when it has empty value
		checkPrimaryKeyValue(m,i,columns,values)
		perpareValues =append(perpareValues,"?")

	}
	stmtQuery+="INSERT INTO "+m.Table+" ("+strings.Join(*columns,",")+") VALUES ("+strings.Join(perpareValues,",")+")"
	return &stmtQuery,values
}

func buildJoins(m *Model,sqlQuery *string)  {
	for i:=0; i< len(m.query.joins); i++ {
		relation:=m.Relations[m.query.joins[i]]
		if relation.relationType == "hasMany"{
			*sqlQuery +=" inner join "+relation.relationTable+" on "+m.Table+"."+relation.localKey+" = "+relation.relationTable+"."+relation.foreignKey
		}else if relation.relationType == "belongsToMany"{
			*sqlQuery +=" inner join "+relation.middleTable+" on "+m.Table+"."+relation.localKey+" = "+relation.middleTable+"."+relation.foreignKey+" inner join "+relation.relationTable+" on "+relation.relationTable+"."+relation.relationModelLocalKey+" = "+relation.middleTable+"."+relation.relationModelForeignKey
		}
	}

}

func buildWhereQuery(m *Model,sqlQuery *string, params *[]interface{}){
	stmt:=" where "
	var lastComp int
	var comp string

	for i:=0;i<len(m.query.query);i++{
		_,ok:=m.query.combinationWhere[i]

		if ok{
			lastComp = i
			comp = "("
		}else{
			comp = ""
		}

		if m.query.query[i].query =="where"{
			*sqlQuery+=stmt+comp+m.query.query[i].column+" "+m.query.query[i].op+" ?"
			*params = append(*params, m.query.query[i].value)
		}else if m.query.query[i].query =="or" {
			*sqlQuery+=" or "+comp+m.query.query[i].column+" "+m.query.query[i].op+" ?"
			*params = append(*params, m.query.query[i].value)
		}else if m.query.query[i].query =="in"{
				var in []string
				for j:=0;j<len(m.query.query[j].in); j++{
					in = append(in, "?")
					*params = append(*params, m.query.query[j].in)
				}
			*sqlQuery+=stmt+comp+m.query.query[i].column+" in ("+strings.Join(in,",")+")"

		}else if m.query.query[i].query =="exists"{
			*sqlQuery+=stmt+comp+" exists ("+*m.query.query[i].existsQuery+")"
			existsParams:= *m.query.query[i].existsParams
			for x:=0; x < len(existsParams); x++{
				*params = append(*params,existsParams[x] )
			}
		}

		stmt=" and "
		_,check:=m.query.combinationWhere[lastComp]
		if check{
			if  i == m.query.combinationWhere[lastComp]{
				*sqlQuery+=")"
			}
		}
	}

}

func getSelected(m *Model)  string{
	if len(m.query.selected) == 0{
		m.query.selected = append(m.query.selected,"*" )
	}
	return strings.Join(m.query.selected,",")
}

func getPrimaryKey(m *Model) string  {
	if m.PrimaryKey == ""{
		return "id"
	}
	return m.PrimaryKey
}

func checkPrimaryKeyValue(m *Model,index int, columns *[]string,values *[]interface{})  {
	//remove primary key from columns and values when it has empty value
	if (*columns)[index] == getPrimaryKey(m) && ((*values)[index] == "0" || (*values)[index] == "")  {
		*columns =  append((*columns)[:index], (*columns)[index+1:]...)
		*values =  append((*values)[:index], (*values)[index+1:]...)
	}
}

