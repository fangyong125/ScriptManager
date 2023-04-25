package src

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

func AddBatchScriptTab(tabs *container.AppTabs) {
	commandEntry := CreateBoundItem("BatchCommand").(*widget.Entry)
	engineEntry := CreateBoundItem("BatchCommandEngine").(*widget.Entry)
	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("引擎"), engineEntry,
		widget.NewLabel("工作目录"), CreateBoundItem("Workspace"),
		widget.NewLabel("命令"), commandEntry)

	rows, cols := getBatchGridSize()
	grid := container.NewGridWithColumns(cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			grid.Add(newBatchContent(i, j, commandEntry, engineEntry))
		}
	}

	execBtn := widget.NewButton("执行", batchExecute)
	body := container.NewVBox(form, execBtn, grid)
	tabs.Append(container.NewTabItem("批量脚本", body))
}

func newBatchContent(row int, col int, commandEntry *widget.Entry, engineEntry *widget.Entry) fyne.CanvasObject {
	item := BtnConfig.Batch[row]
	if col == 0 {
		return widget.NewLabel(item.GroupName)
	}
	i := col - 1
	if len(item.Commands) <= i {
		return widget.NewLabel("")
	}
	command := item.Commands[i]
	btn := widget.NewButton(command.Label, func() {
		if len(command.PWD) != 0 {
			SetConfig("Workspace", command.PWD)
		}
		commandEntry.SetText(command.Command)
		engineEntry.SetText(command.Engine)
	})
	return btn
}

func getBatchGridSize() (int, int) {
	maxCols := 0
	for _, item := range BtnConfig.Batch {
		l := len(item.Commands)
		if l > maxCols {
			maxCols = l
		}
	}
	return len(BtnConfig.Batch), maxCols + 1
}

var taskCount int
var beforeTime time.Time

var taskIndex int
var taskIndexMutex sync.Mutex

func getTaskIndex() int {
	taskIndexMutex.Lock()
	index := taskIndex
	taskIndex++
	taskIndexMutex.Unlock()
	return index
}

var submitCount int
var submitCountMutex sync.Mutex

func increaseSubmitCount() {
	submitCountMutex.Lock()
	submitCount++
	submitCountMutex.Unlock()
}

var finishCount int
var finishCountMutex sync.Mutex

func increaseFinishCount() {
	finishCountMutex.Lock()
	finishCount++
	finishCountMutex.Unlock()
}

func batchExecute() {
	go doBatchExecute()
}
func doBatchExecute() {
	ShowProgress()
	beforeTime = time.Now()
	file, err := os.OpenFile(Config.LogFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		ErrorDialog(err)
		return
	}
	defer file.Close()

	var mutex sync.Mutex
	fileLog := TypeFileLog{file, &mutex}

	tasks := getAllTasks()
	taskCount = len(tasks)

	taskIndex = 0
	submitCount = 0
	finishCount = 0

	wg := sync.WaitGroup{}
	wg.Add(Config.ThreadCount)
	log.Println("taskCount:", taskCount)
	StatusBar.SetText("Batch exec... " + Config.BatchCommand)
	for i := 0; i < Config.ThreadCount; i++ {
		go threadRun("thread-"+strconv.Itoa(i), tasks, &fileLog, &wg)
	}

	log.Println("main wait")
	wg.Wait()
	log.Println("main wait finish")

	HideProgress()
	displayProgress()
}

func threadRun(threadName string, tasks []string, fileLog *TypeFileLog, waitGroup *sync.WaitGroup) {
	log.Println(threadName, "started")
	for {
		if Cancel {
			break
		}
		index := getTaskIndex()
		if index >= taskCount {
			break
		}

		increaseSubmitCount()
		displayProgress()

		pwd := tasks[index]
		log.Println(threadName, "begin ", pwd)
		BatchExec(pwd, Config.BatchCommandEngine, Config.BatchCommand, fileLog)
		log.Println(threadName, "finish", pwd)
		increaseFinishCount()
		displayProgress()
	}
	log.Println(threadName, "exit")
	waitGroup.Done()
}

func displayProgress() {
	after := time.Now()
	duration := after.Sub(beforeTime)

	msg := fmt.Sprintf("submit: %d, finish: %d, total: %d, time: %s", submitCount, finishCount, taskCount, duration.String())
	StatusBar.SetText(msg)
}

func getAllTasks() []string {
	var rs []string
	dirs, err := os.ReadDir(Config.Workspace)
	if err != nil {
		ErrorDialog(err)
		return rs
	}

	excludes := make(map[string]bool)
	splitReg, _ := regexp.Compile(`[;,]`)
	split := splitReg.Split(Config.ExcludeDir, -1)
	for _, s := range split {
		excludes[s] = true
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if excludes[dir.Name()] {
			continue
		}
		pwd := filepath.Join(Config.Workspace, dir.Name())
		rs = append(rs, pwd)
	}
	return rs
}
