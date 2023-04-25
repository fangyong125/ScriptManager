package main

import (
	"ScriptManager/src"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"time"
)

func main() {
	before := time.Now()

	src.InitRootPath()
	src.SetFont()
	src.Init()

	myApp := app.New()
	src.TopWin = myApp.NewWindow("ScriptManager")
	src.TopWin.SetMaster()
	src.TopWin.SetIcon(src.MainIcon())

	tabs := container.NewAppTabs()

	src.AddSettingsTab(tabs)
	src.AddScriptTab(tabs)
	src.AddBatchScriptTab(tabs)

	tabs.SetTabLocation(container.TabLocationTop)
	src.StatusBar = widget.NewEntry()
	src.StatusBar.Disable()
	mainBody := container.NewVBox(tabs, src.StatusBar)

	src.TopWin.SetContent(mainBody)
	src.TopWin.Resize(fyne.NewSize(600, mainBody.MinSize().Height))
	src.TopWin.SetFixedSize(true)

	after := time.Now()
	duration := after.Sub(before)
	src.StatusBar.SetText("Started in " + duration.String())

	src.TopWin.ShowAndRun()
}
