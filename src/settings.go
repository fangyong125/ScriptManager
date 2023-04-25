package src

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func AddSettingsTab(tabs *container.AppTabs) {
	body := container.New(layout.NewFormLayout(),
		widget.NewLabel("工作目录"), CreateBoundItem("Workspace"),
		widget.NewLabel("排除目录"), CreateBoundItem("ExcludeDir"),
		widget.NewLabel("日志文件路径"), CreateBoundItem("LogFilePath"),
		widget.NewLabel("并行度"), CreateBoundItem("ThreadCount"))
	tabs.Append(container.NewTabItemWithIcon("设置", theme.SettingsIcon(), body))
}
