package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego/logs"
)

var (
	buildDate string
	gitDate   string
	gitCommit string
)

func version() {
	if buildDate != "" {
		logs.Info("Build date:", buildDate)
	}
	if gitDate != "" {
		logs.Info("Git date:", gitDate)
	}
	if gitCommit != "" {
		logs.Info("Git version:", gitCommit)
	}
}

//设置日志级别
func setLogLevel(level string) {
	level = strings.ToLower(level)
	switch level {
	case "emergency":
		logs.SetLevel(logs.LevelEmergency)
	case "alert":
		logs.SetLevel(logs.LevelAlert)
	case "error":
		logs.SetLevel(logs.LevelError)
	case "warn", "warning":
		logs.SetLevel(logs.LevelWarn)
	case "info", "informational":
		logs.SetLevel(logs.LevelInfo)
	case "debug":
		logs.SetLevel(logs.LevelDebug)
	}
}

func reset() {
	binaryPath, _ := filepath.Abs(os.Args[0])
	os.Chdir(filepath.Dir(binaryPath))
	os.Mkdir("logs", os.ModeDir)

	app := strings.TrimPrefix(binaryPath, filepath.Dir(binaryPath))
	app = strings.TrimRight(app, filepath.Ext(app))
	app = strings.TrimLeft(app, "/\\")

	logs.SetLogFuncCall(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterConsole)
	logs.SetLogger("file", `{"filename":"logs/`+app+`.log"}`)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	version()
}
