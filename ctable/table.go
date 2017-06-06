package ctable

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var connectString string
var db *sql.DB

func GetAppName() string {
	binaryPath, _ := filepath.Abs(os.Args[0])
	app := strings.TrimPrefix(binaryPath, filepath.Dir(binaryPath))
	app = strings.TrimRight(app, filepath.Ext(app))
	app = strings.TrimLeft(app, "/\\")
	return app
}

func init() {
	connectString = DBPath
}

//定义配置表结构
type ConfTable struct {
	Name     string //页面显示的数据
	Key      string //显示数据对应的KEY
	Value    string //需要配置的值
	Desc     string //描述信息
	Scope    string //测点类型
	TextType string //填充模式
	TextData string //字段候选值
}

func AddRow(name, key, value, desc, scope, tType, tData string) *ConfTable {
	if key == "" {
		return nil
	}
	if name == "" {
		return nil
	}
	row := new(ConfTable)
	row.Name = name
	row.Key = key
	row.Value = value
	row.Desc = desc
	row.Scope = scope
	row.TextType = tType
	row.TextData = tData
	return row
}

func GetDB() *sql.DB {
	if CheckDB() != nil {
		return nil
	}
	return db
}

func ConfCount(servName string) (count int, err error) {
	if err = CheckDB(); err != nil {
		return
	}
	sqlString := "SELECT count(*) FROM das_conf WHERE driver='" + servName + "';"
	row := db.QueryRow(sqlString)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

//同步到数据库
func Sync(table []*ConfTable, driverName string, tableName string) (err error) {
	if err = CheckDB(); err != nil {
		return
	}

	//1. 清空服务下的全部数据
	//2. 插入新的数据
	tx, err := db.Begin()
	if err != nil {
		return
	}
	sqlString := "DELETE FROM " + tableName
	sqlString += " WHERE driver = ?"
	_, err = tx.Exec(sqlString, driverName)
	if err != nil {
		tx.Rollback()
		err = errors.New("清空旧数据:" + err.Error())
		return
	}
	for _, row := range table {
		sqlString = fmt.Sprintf("INSERT INTO %s (driver,name,key,value,ed,ex_scope,text_type,text_data) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s');",
			tableName, driverName, row.Name, row.Key, row.Value, row.Desc, row.Scope, row.TextType, row.TextData)
		_, err = tx.Exec(sqlString)
		if err != nil {
			log.Println(sqlString)
		}
	}
	tx.Commit()
	return
}

//检查数据库是否可用
func CheckDB() (err error) {
	if db == nil {
		db, err = sql.Open("sqlite3", connectString)
		if err != nil {
			return
		}
	}
	err = db.Ping()
	if err != nil {
		return
	}
	return
}

//清空表数据
//servName 服务名
func ClearTable(servName string) (err error) {
	if err = CheckDB(); err != nil {
		return
	}
	//清空表数据
	sqlString := "DELETE  FROM das_conf WHERE driver= ?"
	_, err = db.Exec(sqlString, servName)
	return
}

//插入数据
func Insert(sqlString string) (err error) {
	if err = CheckDB(); err != nil {
		return
	}

	//建立表结构
	_, err = db.Exec(sqlString)
	if err != nil {
		return
	}
	return
}
