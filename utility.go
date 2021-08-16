package gql

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

func checkRelationsMap(m *Model)  {
	if m.Relations == nil{
		relations:=make(map[string]Relation)
		m.Relations = relations
	}
}

func inStringArray(selected []string,col string) bool {
	for i:=0;i<len(selected);i++ {
		if selected[i] == col{
			return true
		}
	}
	return false
}


func getStrutFields(u interface{},columns []string,fields *map[int]int)  []interface{}{
	val := reflect.ValueOf(u).Elem()
	fieldsMap:=*fields
	v := make([]interface{}, len(fieldsMap))
	for  i := 0; i < len(columns); i++ {
		valueField := val.Field(fieldsMap[i])
		v[i] = valueField.Addr().Interface()
	}
	return v
}

func mapColumnsFields(u interface{},columns []string) *map[int]int{
	val := reflect.ValueOf(u).Elem()
	matched := 0
	fields:=make(map[int]int)
	if val.NumField() < len(columns){
		log.Fatal("the scanner not has all columns you selected it , your scanner must be has fields for this columns:",columns)
	}

	for i := 0; i < val.NumField(); i++ {
		if InColumns(val.Type().Field(i).Name,val.Type().Field(i).Tag.Get("db"),columns){
			fields[matched] = i
			matched++
		}
	}

	if len(columns) > matched{
		log.Fatal("the scanner not has all columns you selected it , your scanner must be has fields for this columns:",columns)
	}

	return &fields
}


func InColumns(field string,tag string,columns []string)  bool{
	for i := 0; i < len(columns); i++ {
		if tag == columns[i] || strings.ToLower(field) == columns[i]{
			return true
		}
	}
	return false
}



func structFieldsValues(insertObject interface{})  (*[]string,*[]interface{}){
	val := reflect.ValueOf(insertObject).Elem()
	var columns []string
	var values []interface{}
	for i := 0; i < val.NumField(); i++ {

		var column string
		if val.Type().Field(i).Tag.Get("db") != ""{
			column = val.Type().Field(i).Tag.Get("db")
		}else{
			column = strings.ToLower(val.Type().Field(i).Name)
		}
		field := val.Field(i).Interface()
		fieldVal := reflect.ValueOf(field)

		columns=append(columns,column)
		values = append(values, fmt.Sprintf("%v",fieldVal))
	}
	return &columns,&values
}



func checkWhereCombinationMap(m *Model)  {
	if m.query.combinationWhere == nil{
		combinationWhere:=make(map[int]int)
		m.query.combinationWhere = combinationWhere
	}
}

func stringHasDot(text string)  bool {
	result, _ := regexp.MatchString("\\.", text)
	return result
}

func removeFromStrings(s *[]string, r *string) *[]string {
	slice:=*s
	for i, v := range slice {
		if v == *r {
			slice =  append(slice[:i], slice[i+1:]...)
			return &slice
		}
	}
	return &slice
}
