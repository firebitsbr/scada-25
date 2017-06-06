package main

import (
	"fmt"
	"scada/ctable"
)

func main() {
	sql := " INSERT INTO point (uid,event,sn,cp,hw,rt,pn,ed,fk,fb,pt,an,tv,bv,ph,pl,tm,os) VALUES('12345','1','BBB','','1.3.0','AX','111','','1.000000','0.000000','COILS','111','0.000000','0.000000','0.000000','0.000000','2017-05-02 00:06:47','adding'); "
	err := ctable.Insert(sql)
	if err != nil {
		fmt.Println(err)
	}
}
