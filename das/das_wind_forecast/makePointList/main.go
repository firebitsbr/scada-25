package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

func parser2array(line string) []int {
	ss := strings.Split(line, ",")
	array := make([]int, 0, len(ss))
	for i := range ss {
		id, err := strconv.Atoi(ss[i])
		if err != nil {
			continue
		}
		array = append(array, id)
	}
	return array
}

func isInArray(array []int, v int) bool {
	for i := range array {
		if array[i] == v {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println(filepath.Base(os.Args[0]), "file", "TAG", "列数n,列数m")
		fmt.Println("Example", filepath.Base(os.Args[0]), "CDQ_2017.txt", "WP4", "2,3")
		return
	}
	file := os.Args[1]
	colsName := os.Args[3] //"5,6,7,8,9,10"
	array := parser2array(colsName)
	tag := os.Args[2]
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	mode := ""
	defer f.Close()
	buf := bufio.NewReader(f)
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		if strings.Contains(line, "MastData") {
			mode = "CFT"
			break
		}
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
		if strings.Contains(line, "FANDATA") {
			mode = "YXZT"
			break
		}
	}
	if mode == "" {
		fmt.Println(file, "不是正确的文件")
		return
	}
	var cols []string
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@") {
			cols = strings.Fields(line)
			break
		}
	}
	if len(cols) == 0 {
		fmt.Println("没有找到需要的列")
		return
	}
	idx := 0
	//fmt.Println("PN,RT,AD,ED")
	for line, err := buf.ReadString('\n'); err == nil || len(line) > 0; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		ss := strings.Fields(line)
		idx++
		col := 0
		for i := range ss {
			if isInArray(array, i) {
				col++
				if ss[0][0] < utf8.RuneSelf {
					fmt.Printf("%s_%s_%d_%d,AX,%s.%s.%d,%s%s\n", tag, mode, idx, col, mode, ss[0], i, ss[0], cols[i])
				} else {
					fmt.Printf("%s_%s_%d_%d,AX,%s.%s.%d,%s%s\n", tag, mode, idx, col, mode, ss[0]+ss[1], i, ss[0]+ss[1], cols[i])
				}
			}
		}
	}
}
