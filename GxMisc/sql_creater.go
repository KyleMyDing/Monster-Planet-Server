package GxMisc

/**
作者： Kyle Ding
模块：sql生成模块
说明：
创建时间：2015-10-30
**/

import (
	"reflect"
	"strconv"
)

func getMysqlTableName(info interface{}, tableID string) string {
	tableName := "tb_" + GetTableName(info)
	if tableID != "" {
		tableName += "_" + tableID
	}
	return tableName
}

//GenerateCreateSQL 结构的create sql生成函数
//结构标签说明：
//ignore : 为ture时忽略这个字段
//pk     : 为ture时该字段为主键
//type   : int(time)
//         string(text)时是TEXT类型
//len    : 字符串最大长度
func GenerateCreateSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	key := ""

	ret := "CREATE TABLE " + tableName + " ("

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	first := true
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)

		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if fieldTag.Get("pk") == "true" {
			if key == "" {
				key = fieldType.Name
			} else {
				key += ", " + fieldType.Name
			}
		}

		if !first {
			ret += ","
		}

		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				ret += " " + fieldType.Name + " DATETIME NOT NULL"
			} else {
				ret += " " + fieldType.Name + " INT NOT NULL"
			}
		case reflect.Float32, reflect.Float64:
			ret += " " + fieldType.Name + " FLOAT NOT NULL"
		case reflect.String:
			if fieldTag.Get("type") == "text" {
				ret += " " + fieldType.Name + " TEXT NOT NULL"
			} else {
				lenStr := fieldTag.Get("len")
				if lenStr != "" {
					ret += " " + fieldType.Name + " VARCHAR(" + lenStr + ") BINARY NOT NULL"
				} else {
					ret += " " + fieldType.Name + " VARCHAR(32) BINARY NOT NULL"
				}
			}

		case reflect.Bool:
			ret += " " + fieldType.Name + " INT NOT NULL"
		case reflect.Slice:
			ret += " " + fieldType.Name + " BLOB NOT NULL"
		}
		first = false
	}

	if key != "" {
		ret += ", PRIMARY KEY (" + key + "))"
	} else {
		ret += ")"
	}

	return ret
}

//GenerateIndexSQL 结构的index sql生成函数
//用多个索引的时候，根据ID区别就行了
func GenerateIndexSQL(info interface{}, tableID string, indexID string) string {
	tableName := getMysqlTableName(info, tableID)
	indexName := "index_" + GetTableName(info) + "_" + tableID

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	ret := ""
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if fieldTag.Get("index") != indexID {
			continue
		}

		if ret == "" {
			ret = fieldType.Name
		} else {
			ret += ", " + fieldType.Name
		}
		indexName += "_" + fieldType.Name
	}
	return "CREATE INDEX " + indexName + " ON " + tableName + " (" + ret + ")"
}

//GenerateInsertSQL 结构的insert sql生成函数
func GenerateInsertSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	ret := "INSERT INTO " + tableName + " VALUES("

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	first := true
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if !first {
			ret += ","
		}

		str := ""
		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if fieldTag.Get("type") == "time" {
				str = "FROM_UNIXTIME(" + strconv.FormatInt(fieldValue.Int(), 10) + ")"
			} else {
				str = strconv.FormatInt(fieldValue.Int(), 10)
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				str = "FROM_UNIXTIME(" + strconv.FormatUint(fieldValue.Uint(), 10) + ")"
			} else {
				str = strconv.FormatUint(fieldValue.Uint(), 10)
			}
		case reflect.Float32, reflect.Float64:
			str = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.String:
			str = "'" + fieldValue.String() + "'"
		case reflect.Bool:
			if fieldValue.Bool() {
				str = "1"
			} else {
				str = "0"
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Uint8 {
				str = string(fieldValue.Interface().([]byte))
			}
		}

		first = false
		ret += str
	}

	ret += ")"
	return ret
}

//GenerateUpdateSQL 结构的update sql生成函数
func GenerateUpdateSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	where := ""
	ret := "UPDATE " + tableName + " set "

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	first := true
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		pk := fieldTag.Get("pk") == "true"

		if !pk && !first {
			ret += ", "
		}

		str := ""
		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatInt(fieldValue.Int(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatInt(fieldValue.Int(), 10)
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatUint(fieldValue.Uint(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatUint(fieldValue.Uint(), 10)
			}
		case reflect.Float32, reflect.Float64:
			str = fieldType.Name + "=" + strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.String:
			str = fieldType.Name + "='" + fieldValue.String() + "'"
		case reflect.Bool:
			if fieldValue.Bool() {
				str = fieldType.Name + "=1"
			} else {
				str = fieldType.Name + "=0"
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Uint8 {
				str = fieldType.Name + string(fieldValue.Interface().([]byte))
			}
		}

		if pk {
			if where == "" {
				where = str
			} else {
				where += " AND " + str
			}
		} else {
			first = false
			ret += str
		}
	}

	if where != "" {
		ret += " WHERE " + where
	}
	return ret
}

//GenerateSelectSQL 结构的select sql生成函数
func GenerateSelectSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	where := ""
	ret := "SELECT "

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	first := true
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		pk := fieldTag.Get("pk") == "true"

		if pk {
			str := ""
			switch fieldType.Type.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				if fieldTag.Get("type") == "time" {
					str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatInt(fieldValue.Int(), 10) + ")"
				} else {
					str = fieldType.Name + "=" + strconv.FormatInt(fieldValue.Int(), 10)
				}
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fieldTag.Get("type") == "time" {
					str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatUint(fieldValue.Uint(), 10) + ")"
				} else {
					str = fieldType.Name + "=" + strconv.FormatUint(fieldValue.Uint(), 10)
				}
			case reflect.Float32, reflect.Float64:
				str = fieldType.Name + "=" + strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
			case reflect.String:
				str = fieldType.Name + "='" + fieldValue.String() + "'"
			case reflect.Bool:
				if fieldValue.Bool() {
					str = fieldType.Name + "=1"
				} else {
					str = fieldType.Name + "=0"
				}
			case reflect.Slice:
				if fieldType.Type.Elem().Kind() == reflect.Uint8 {
					str = fieldType.Name + string(fieldValue.Interface().([]byte))
				}
			}

			if where == "" {
				where = str
			} else {
				where += " AND " + str
			}
		} else {
			if !first {
				ret += ", "
			}

			str := ""
			switch fieldType.Type.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fieldTag.Get("type") == "time" {
					str = "UNIX_TIMESTAMP(" + fieldType.Name + ")"
				} else {
					str = fieldType.Name
				}
			case reflect.Float32, reflect.Float64:
				str = fieldType.Name
			case reflect.String:
				str = fieldType.Name
			case reflect.Bool:
				str = fieldType.Name
			case reflect.Slice:
				str = fieldType.Name
			}
			first = false
			ret += str
		}
	}

	if where != "" {
		ret += " FROM " + tableName + " WHERE " + where
	}
	return ret
}

//GenerateSelectAllSQL 结构的select * sql生成函数
func GenerateSelectAllSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	ret := "SELECT "

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	first := true
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if !first {
			ret += ", "
		}

		str := ""
		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				str = "UNIX_TIMESTAMP(" + fieldType.Name + ")"
			} else {
				str = fieldType.Name
			}
		case reflect.Float32, reflect.Float64:
			str = fieldType.Name
		case reflect.String:
			str = fieldType.Name
		case reflect.Bool:
			str = fieldType.Name
		case reflect.Slice:
			str = fieldType.Name
		}
		first = false
		ret += str
	}
	ret += " FROM " + tableName
	return ret
}

//GenerateSelectOneFieldSQL 结构的select field sql生成函数
func GenerateSelectOneFieldSQL(info interface{}, field string, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	where := ""
	ret := "SELECT " + field

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if fieldTag.Get("pk") != "true" {
			continue
		}

		str := ""
		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatInt(fieldValue.Int(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatInt(fieldValue.Int(), 10)
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatUint(fieldValue.Uint(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatUint(fieldValue.Uint(), 10)
			}
		case reflect.Float32, reflect.Float64:
			str = fieldType.Name + "=" + strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.String:
			str = fieldType.Name + "='" + fieldValue.String() + "'"
		case reflect.Bool:
			if fieldValue.Bool() {
				str = fieldType.Name + "=1"
			} else {
				str = fieldType.Name + "=0"
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Uint8 {
				str = fieldType.Name + string(fieldValue.Interface().([]byte))
			}
		}

		if where == "" {
			where = str
		} else {
			where += " AND " + str
		}

	}

	if where != "" {
		ret += " FROM " + tableName + " WHERE " + where
	}
	return ret
}

//GenerateDeleteSQL 结构的delete sql生成函数
func GenerateDeleteSQL(info interface{}, tableID string) string {
	tableName := getMysqlTableName(info, tableID)
	where := ""
	ret := "DELETE FROM " + tableName

	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)
		fieldTag := fieldType.Tag

		if fieldTag.Get("ignore") == "true" {
			continue
		}

		if fieldTag.Get("pk") != "true" {
			continue
		}

		str := ""
		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatInt(fieldValue.Int(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatInt(fieldValue.Int(), 10)
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldTag.Get("type") == "time" {
				str = fieldType.Name + "=" + "FROM_UNIXTIME(" + strconv.FormatUint(fieldValue.Uint(), 10) + ")"
			} else {
				str = fieldType.Name + "=" + strconv.FormatUint(fieldValue.Uint(), 10)
			}
		case reflect.Float32, reflect.Float64:
			str = fieldType.Name + "=" + strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.String:
			str = fieldType.Name + "='" + fieldValue.String() + "'"
		case reflect.Bool:
			if fieldValue.Bool() {
				str = fieldType.Name + "=1"
			} else {
				str = fieldType.Name + "=0"
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Uint8 {
				str = fieldType.Name + string(fieldValue.Interface().([]byte))
			}
		}

		if where == "" {
			where = str
		} else {
			where += " AND " + str
		}
	}

	if where != "" {
		ret += " WHERE " + where
	}
	return ret
}
