package src

import (
	"os"
	"sync"
)

type TypeConfig struct {
	Workspace           string
	ExcludeDir          string
	LogFilePath         string
	ThreadCount         int
	SingleCommand       string
	SingleCommandEngine string
	BatchCommand        string
	BatchCommandEngine  string
}

type TypeCmdItem struct {
	Label   string `json:"label"`
	Engine  string `json:"engine"`
	Command string `json:"command"`
	CmdType string `json:"cmdType"`
	PWD     string `json:"pwd"`
	Help    string `json:"help"`
}
type TypeCmdGroupItem struct {
	GroupName string        `json:"groupName""`
	Commands  []TypeCmdItem `json:"commands"`
}

type TypeBtnConfig struct {
	Single []TypeCmdGroupItem `json:"single"`
	Batch  []TypeCmdGroupItem `json:"batch"`
}

type TypeFileLog struct {
	file  *os.File
	mutex *sync.Mutex
}
