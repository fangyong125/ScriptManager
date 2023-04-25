package src

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"strings"
	"time"
)

func AddScriptTab(tabs *container.AppTabs) {
	commandEntry := CreateBoundItem("SingleCommand").(*widget.Entry)
	engineEntry := CreateBoundItem("SingleCommandEngine").(*widget.Entry)
	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("引擎"), engineEntry,
		widget.NewLabel("工作目录"), CreateBoundItem("Workspace"),
		widget.NewLabel("命令"), commandEntry)

	rows, cols := getGridSize()
	grid := container.NewGridWithColumns(cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			grid.Add(newContent(i, j, commandEntry, engineEntry))
		}
	}

	execBtn := widget.NewButton("执行", func() {
		go executeScript()
	})
	body := container.NewVBox(form, execBtn, grid)
	tabs.Append(container.NewTabItem("脚本", body))
}

func executeScript() {
	before := time.Now()

	StatusBar.SetText("exec... " + Config.SingleCommand)
	ShowProgress()
	Exec(Config.Workspace, Config.SingleCommandEngine, Config.SingleCommand, Config.LogFilePath)
	HideProgress()

	after := time.Now()
	duration := after.Sub(before)
	StatusBar.SetText("Done in " + duration.String())
}

func newContent(row int, col int, commandEntry *widget.Entry, engineEntry *widget.Entry) fyne.CanvasObject {
	item := BtnConfig.Single[row]
	if col == 0 {
		return widget.NewLabel(item.GroupName)
	}
	i := col - 1
	if len(item.Commands) <= i {
		return widget.NewLabel("")
	}
	command := item.Commands[i]
	btn := widget.NewButton(command.Label, func() {

		if strings.EqualFold(command.CmdType, "launcher") {
			pwd := Config.Workspace
			if len(command.PWD) != 0 {
				pwd = command.PWD
			}
			go Launcher(pwd, command.Engine, command.Command)
		} else {
			if len(command.PWD) != 0 {
				SetConfig("Workspace", command.PWD)
			}
			commandEntry.SetText(command.Command)
			engineEntry.SetText(command.Engine)
		}
	})
	return btn
}

func getGridSize() (int, int) {
	maxCols := 0
	for _, item := range BtnConfig.Single {
		l := len(item.Commands)
		if l > maxCols {
			maxCols = l
		}
	}
	return len(BtnConfig.Single), maxCols + 1
}
