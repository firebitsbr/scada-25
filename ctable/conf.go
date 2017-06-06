package ctable

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func init() {
	config.Register("table", &Conf{})
}

//读取配置文件
type Conf struct {
	name string
	lock *sync.RWMutex
	m    map[string]string
}

func (this *Conf) Set(key, val string) (err error) { //support section::key type in given key when using ini type.
	if len(key) == 0 {
		return errors.New("key is empty")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m[key] = val
	return
}
func (this *Conf) String(key string) (s string) { //support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
	this.lock.RLock()
	defer this.lock.RUnlock()
	s = this.m[key]
	return
}

//TODO 暂没实现
func (this *Conf) Strings(key string) (ss []string) { //get string slice
	return
}

func (this *Conf) Int(key string) (int, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	s := this.m[key]
	return strconv.Atoi(s)
}

func (this *Conf) Int64(key string) (int64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	s := this.m[key]
	return strconv.ParseInt(s, 10, 64)
}
func (this *Conf) Bool(key string) (bool, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	s := this.m[key]
	return ParseBool(s)
}
func (this *Conf) Float(key string) (float64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	s := this.m[key]
	return strconv.ParseFloat(s, 64)
}
func (this *Conf) DefaultString(key string, defaultVal string) string { // support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
	this.lock.RLock()
	defer this.lock.RUnlock()
	s, ok := this.m[key]
	if ok {
		return s
	}
	return defaultVal
}

//TODO 暂没实现
func (this *Conf) DefaultStrings(key string, defaultVal []string) []string { //get string slice
	return nil
}

func (this *Conf) DefaultInt(key string, defaultVal int) int {
	i, err := this.Int(key)
	if err != nil {
		return defaultVal
	}
	return i
}
func (this *Conf) DefaultInt64(key string, defaultVal int64) int64 {
	i, err := this.Int64(key)
	if err != nil {
		return defaultVal
	}
	return i
}
func (this *Conf) DefaultBool(key string, defaultVal bool) bool {
	b, err := this.Bool(key)
	if err != nil {
		return defaultVal
	}
	return b
}
func (this *Conf) DefaultFloat(key string, defaultVal float64) float64 {
	f, err := this.Float(key)
	if err != nil {
		return defaultVal
	}
	return f
}

//TODO 暂没实现
func (this *Conf) DIY(key string) (interface{}, error) {
	return nil, nil
}

//TODO 暂没实现
func (this *Conf) GetSection(section string) (map[string]string, error) {
	return nil, nil
}

//TODO 暂没实现
func (this *Conf) SaveConfigFile(filename string) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	return nil
}

//从关系库conf表中读取配置
func (this *Conf) Parse(servName string) (c config.Configer, err error) {
	if this.lock == nil {
		this.lock = new(sync.RWMutex)
	}
	if this.m == nil {
		this.m = map[string]string{}
	}
	this.name = servName
	err = CheckDB()
	if err != nil {
		return
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	sqlString := "select key, value from das_conf where ex_scope == 'addServer' and driver = ?"
	//sqlString := "select key, value from conf where driver = ?"
	rows, err := db.Query(sqlString, this.name)
	if err != nil {
		return
	}
	var key string
	var value string
	for rows.Next() {
		err = rows.Scan(&key, &value)
		if err != nil {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		logs.Info("table:", key, value)
		this.m[key] = value
	}
	c = this
	return
}

//TODO 暂没实现
func (this *Conf) ParseData(data []byte) (config.Configer, error) {
	return this, errors.New("Not to do")
}

func ParseBool(val interface{}) (value bool, err error) {
	if val != nil {
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			switch v {
			case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "Y", "y", "ON", "on", "On":
				return true, nil
			case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "N", "n", "OFF", "off", "Off":
				return false, nil
			}
		case int8, int32, int64:
			strV := fmt.Sprintf("%s", v)
			if strV == "1" {
				return true, nil
			} else if strV == "0" {
				return false, nil
			}
		case float64:
			if v == 1 {
				return true, nil
			} else if v == 0 {
				return false, nil
			}
		}
		return false, fmt.Errorf("parsing %q: invalid syntax", val)
	}
	return false, fmt.Errorf("parsing <nil>: invalid syntax")
}
