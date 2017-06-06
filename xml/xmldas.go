//读取xml配置文件
package xmldas

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/golang/glog"
)

type Das struct {
	XMLName    xml.Name      `xml:"DAS"`
	DataSource XmlDataSource `xml:"DataSource"`
	DataDest   XmlDataDest   `xml:"DataDest"`
	Config     XmlConfig     `xml:"Config"`
}

type XmlConfig struct {
	LN string `xml:"LN,attr"`
}

type XmlDataSource struct {
	LN string `xml:"LN,attr"`
}

type XmlDataDest struct {
	LN string `xml:"LN,attr"`
}

func GetXmlConf(path string) (source, dest, config string, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorln("Read \"" + path + "\" " + err.Error())
		return
	}
	das := Das{}
	err = xml.Unmarshal(data, &das)
	if err != nil {
		glog.Errorln("Parser \"" + path + "\" " + err.Error())
		return
	}
	source = das.DataSource.LN
	dest = das.DataDest.LN
	config = das.Config.LN
	return
}
