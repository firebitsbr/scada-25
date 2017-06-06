//风功预测
package wf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"opapi4/opevent/opeventer"
	"os"
	"scada/points"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"wj/estring"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

//const (
//	CFT  = 1 //测风塔
//	DQ   = 2 //短期
//	CDQ  = 3 //超短期
//	YXZT = 4 //运行状态
//)

//风功预测
type WindForecast struct {
	pointPath  string
	pointSnMap map[string]*points.Point

	dataPath     string //数据文件路径
	bakPath      string //备份路径
	saveDays     int64  //保存天数
	scanInterval int    //扫描间隔
}

func New(conf config.Configer) (das points.Daser, err error) {
	wf := new(WindForecast)
	wf.pointPath = conf.DefaultString("service_name", points.GetAppName())
	wf.dataPath = conf.DefaultString("file_path", "data")
	wf.bakPath = conf.DefaultString("file_bak", "bak")
	wf.saveDays = conf.DefaultInt64("save_days", 30)
	wf.scanInterval = conf.DefaultInt("source_interval", 30)
	if wf.scanInterval < 5 {
		wf.scanInterval = 5
	}
	wf.LoadPointList()
	das = wf
	if !strings.HasSuffix(wf.dataPath, "/") {
		wf.dataPath += "/"
	}
	if !strings.HasSuffix(wf.bakPath, "/") {
		wf.bakPath += "/"
	}
	os.Mkdir(wf.bakPath, os.ModeDir)
	return
}

func (this *WindForecast) Check() {
	if points.IsUpdatePointList() {
		logs.Notice("正在重新加载点表")
		err := this.LoadPointList()
		if err != nil {
			logs.Error("动态加载点表错误", err)
		} else {
			logs.Notice("正在重新加载点表完成.")
		}
		points.ResetPointListStatus()
	}
}

func FormatAddress(point *points.Point, address string) (err error) {
	address = strings.ToUpper(address)
	//	//格式如下 CFT.1.1 (类型.ID.列号) address = ID * 1000 + 类型 * 100 + 列号
	//	ss := strings.Split(address, ".")
	//	if len(ss) != 3 {
	//		return errors.New("Format address error.")
	//	}
	//	id, err := strconv.Atoi(ss[1])
	//	if err != nil {
	//		return err
	//	}
	//	col, err := strconv.Atoi(ss[2])
	//	if err != nil {
	//		return err
	//	}
	point.Extend = address
	//logs.Debug(address, point.Extend)
	//	switch ss[0] {
	//	case "CFT":
	//		point.Address = int32(id)*1000 + CFT*100 + int32(col)
	//	case "DQ":
	//		point.Address = int32(id)*1000 + DQ*100 + int32(col)
	//	case "CDQ":
	//		point.Address = int32(id)*1000 + CDQ*100 + int32(col)
	//	case "YXZT":
	//		point.Address = int32(id)*1000 + YXZT*100 + int32(col)
	//	default:
	//		return errors.New("Format address error when parser type.")
	//	}
	return
}

//加载点表
func (this *WindForecast) LoadPointList() (err error) {
	points.PFuncFormatAddress = FormatAddress
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}
	this.pointSnMap = map[string]*points.Point{}
	for _, point := range pointArray {
		this.pointSnMap[point.Extend] = point

		//logs.Debug(point.Extend, point.Name)
	}

	points.Append(pointArray...)
	return
}

func (this *WindForecast) Work(conn interface{}, control opeventer.Controler) (err error) {
	logs.Debug(this.dataPath)
	for {
		logs.Debug(this.dataPath)
		fs, err := ioutil.ReadDir(this.dataPath)
		if err == nil {
			for _, f := range fs {
				logs.Debug(f.Name())
				if f.IsDir() {
					continue
				}
				if !strings.HasSuffix(f.Name(), ".WPD") {
					if !strings.HasSuffix(f.Name(), ".txt") {
						continue
					}
				}
				logs.Debug(f.Name(), f.ModTime().Unix(), time.Now().Unix())
				//为了保证文件的完整性 30秒内创建的文件不操作
				if f.ModTime().Unix() > (time.Now().Unix() - 30) {
					logs.Debug(f.Name(), f.ModTime().Unix(), time.Now().Unix())
					continue
				}
				dir := time.Now().Format("20060102")
				path := this.bakPath + dir
				os.Mkdir(path, os.ModeDir)
				logs.Info("Parser file", this.dataPath+f.Name())
				err = this.ParserFile(this.dataPath + f.Name())
				if err != nil {
					logs.Info(err)
					path += "/" + f.Name() + ".error"
				} else {
					path += "/" + f.Name()
				}
				Move(this.dataPath+f.Name(), path)
			}
		}

		//删除过期文件
		DeleteTimeoutFile(this.bakPath, this.saveDays)
		for i := 0; i < this.scanInterval; i++ {
			this.Check()
			time.Sleep(time.Second)
		}

		//刷新一次数据
		for _, point := range this.pointSnMap {
			if point.Time == 0 {
				continue
			}
			if point.Status&points.OP_TIMEOUT != 0 {
				continue
			}
			points.SendPoint(point)
		}
	}
	return
}

//删除过期文件
func DeleteTimeoutFile(path string, days int64) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		logs.Info("Delete timeout file", err)
		return
	}
	timeout := time.Now().Truncate(time.Duration(days) * time.Hour * 24).Format("20060102")
	for _, f := range fs {
		if f.IsDir() {
			if f.Name() < timeout {
				logs.Info("Delete directory:", path+f.Name())
				os.RemoveAll(path + f.Name())
			}
		}
	}
}

func Move(src, dest string) (err error) {
	fs, err := os.Open(src)
	if err != nil {
		return
	}
	fd, err := os.Create(dest)
	if err != nil {
		fs.Close()
		return
	}
	_, err = io.Copy(fd, fs)
	fs.Close()
	fd.Close()
	os.Remove(src)
	return
}

func getMode(buf *bufio.Reader) (mode string) {
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		if strings.Contains(line, "UltraShortTermForcast") {
			mode = "CDQ"
			break
		}
		if strings.Contains(line, "UltraShortTermForecast") {
			mode = "CDQ"
			break
		}
		if strings.Contains(line, "ShortTermForcast") {
			mode = "DQ"
			break
		}
		if strings.Contains(line, "ShortTermForecast") {
			mode = "DQ"
			break
		}
		if strings.Contains(line, "MastData") {
			mode = "CFT"
			break
		}
		if strings.Contains(line, "FANDATA") {
			mode = "YXZT"
			break
		}
	}
	return mode
}

func getTime(buf *bufio.Reader) int64 {
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "'!>") || strings.HasSuffix(line, "' !>") {
			//判断是否有时间
			idx := strings.LastIndex(line, "='")
			if idx >= 0 {
				line = line[idx+2:]
				line = strings.TrimRight(line, "' !>")
				line = strings.Replace(line, "_", " ", -1)
				line += ":00"
				t, err := estring.Time(line)
				if err != nil {
					break
				}
				return t.Unix()
			}
		}
	}
	return 0
}

func (this *WindForecast) ParserFile(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	buf := bufio.NewReader(f)

	t := getTime(buf)
	if t == 0 {
		err = errors.New("Cant't find time")
		return
	}
	logs.Debug("time", estring.TimeFormat(t))
	mode := getMode(buf)
	if mode == "" {
		err = errors.New("Error mode")
		return
	}
	logs.Debug("Mode", mode)

	//查找数据开始部分
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@") {
			break
		}
	}

	//
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		ss := strings.Fields(line)
		if len(ss[0]) == 0 {
			continue
		}
		for i := range ss {
			address := ""
			if ss[0][0] < utf8.RuneSelf {
				address = fmt.Sprintf("%s.%s.%d", mode, ss[0], i)
			} else {
				address = fmt.Sprintf("%s.%s.%d", mode, ss[0]+ss[1], i)
			}
			val, err := strconv.ParseFloat(ss[i], 64)
			if err != nil {
				continue
			}
			//fmt.Println(address, val)
			if point, ok := this.pointSnMap[address]; ok {
				point.Update(val, 0, t)
			}
		}
	}
	return
}

//func (this *WindForecast) ParserFileOld(path string) (err error) {
//	data, err := ioutil.ReadFile(path)
//	if err != nil {
//		return
//	}
//	s := string(data)
//	n := strings.Count(s, "ShortTermForcast")
//	if n == 2 {
//		//短期数据
//		err = this.ParserData(data, DQ, "ShortTermForcast")
//		return
//	}
//	n = strings.Count(s, "MastData")
//	if n == 2 {
//		//测风塔
//		err = this.ParserData(data, CFT, "MastData")
//		return
//	}
//	n = strings.Count(s, "UltraShortTermForcast")
//	if n == 2 {
//		//超短期
//		err = this.ParserData(data, CDQ, "UltraShortTermForcast")
//		return
//	}
//	n = strings.Count(s, "FANDATA")
//	if n == 2 {
//		//运行状态
//		err = this.ParserData(data, YXZT, "FANDATA")
//		return
//	}
//	return errors.New("Error file")
//}

//func (this *WindForecast) ParserData(data []byte, mode int, key string) (err error) {
//	buf := bytes.NewBuffer(data)

//	//读取时间
//	t := int64(0)
//	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
//		line = strings.TrimSpace(line)
//		//logs.Debug(line)
//		pos := strings.Index(line, "time='")
//		if pos > 0 {
//			line = strings.TrimRight(line, "' !>")
//			line = strings.Replace(line, "_", " ", -1)
//			line += ":00"
//			pos += 6
//			if pos >= len(line) {
//				continue
//			}

//			tt, err := estring.Time(line[pos:])
//			//logs.Debug(line[pos:])

//			if err != nil {
//				return err
//			}
//			t = tt.Unix()
//			break
//		}
//	}

//	//找到开始数据
//	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
//		line = strings.TrimSpace(line)
//		//logs.Debug(line)
//		if strings.Contains(line, key) {
//			break
//		}
//	}

//	//解析数据
//	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
//		line = strings.TrimSpace(line)
//		if !strings.HasPrefix(line, "#") {
//			continue
//		}
//		//logs.Debug(line)
//		line = strings.TrimPrefix(line, "#")
//		ss := strings.Fields(line)
//		if len(ss) <= 1 {
//			continue
//		}

//		id, err := strconv.Atoi(ss[0])
//		if err != nil {
//			continue
//		}
//		id = id*1000 + mode*100
//		n := len(ss)
//		for i := 1; i < n; i++ {
//			val, err := strconv.ParseFloat(ss[i], 64)
//			if err != nil {
//				continue
//			}
//			if point, ok := this.pointSnMap[int32(id+i)]; ok {
//				point.Update(val, 0, t)
//			}
//		}
//	}
//	return
//}
