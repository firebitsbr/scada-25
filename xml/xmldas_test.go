package xmldas

import (
	"fmt"
	"testing"
)

func TestV1ToScada(t *testing.T) {
	source, dest, conf, err := GetXmlConf("das.xml")
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Source:", source)
	fmt.Println("Dest:", dest)
	fmt.Println("Conf:", conf)

}
