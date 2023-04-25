package src

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

// config
var Config TypeConfig
var BtnConfig TypeBtnConfig

// ui
var StatusBar *widget.Entry
var BindConfig binding.Struct
var TopWin fyne.Window

//runtime
var RootPath string
var configJson string
var fieldMap map[string]fyne.CanvasObject
var alertMutex sync.Mutex
var pDialog dialog.Dialog = nil
var Cancel bool

func Init() {
	BindConfig = binding.BindStruct(&Config)
	ReadConfig()
	logConfig()
	fieldMap = make(map[string]fyne.CanvasObject)
}

func ReadConfig() {
	readBtnConfig()
	readStateConfig()
}

func readStateConfig() {
	cnfPath := filepath.Join(RootPath, "conf", "conf.json")
	content, err := os.ReadFile(cnfPath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &Config)
	if err != nil {
		log.Fatal(err)
	}
	configJson = string(content)
}

func logConfig() {
	logPath := filepath.Join(RootPath, "logs", "ScriptManager.log")
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0766)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func readBtnConfig() {
	btnConfig := filepath.Join(RootPath, "conf", "btns.json")
	content, err := os.ReadFile(btnConfig)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &BtnConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func InitRootPath() {
	release := IsRelease()
	var cur string
	if release {
		cur = getCurrentPathByExecutable()
	} else {
		cur = getCurrentAbPathByCaller()
	}

	fontDirPath := filepath.Join(cur, "font")
	if !IsDir(fontDirPath) {
		parent, err := filepath.Abs(filepath.Join(cur, "..", "font"))
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		fontDirPath = parent
	}
	abs, err := filepath.Abs(filepath.Join(fontDirPath, ".."))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	RootPath = abs
	fmt.Println("RootPath: " + RootPath)
}

func MainIcon() *fyne.StaticResource {
	btnConfig := filepath.Join(RootPath, "main.ico")
	content, err := os.ReadFile(btnConfig)
	if err != nil {
		icon := theme.ConfirmIcon()
		resource := fyne.NewStaticResource("MainIcon", icon.Content())
		return resource
	}
	resource := fyne.NewStaticResource("MainIcon", content)
	return resource
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsRelease() bool {
	arg1 := strings.ToLower(os.Args[0])
	name := filepath.Base(arg1)
	return strings.Index(name, "go_build") < 0 && strings.Index(arg1, "go-build") < 0
}

func getCurrentPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func SetConfig(field string, value any) {
	v, err := BindConfig.GetItem(field)

	if err != nil {
		return
	}
	fo := fieldMap[field]
	switch v.(type) {
	case binding.Bool:
		fo.(*widget.Check).SetChecked(value.(bool))
	case binding.Float:
		fo.(*widget.Slider).SetValue(value.(float64))
	case binding.Int:
		fo.(*widget.Entry).SetText(string(value.(int)))
	case binding.String:
		fo.(*widget.Entry).SetText(value.(string))
	default:
		log.Println("nothing todo")
	}
	SaveConfig()
}

func CreateBoundItem(field string) fyne.CanvasObject {
	v, err := BindConfig.GetItem(field)

	if err != nil {
		return nil
	}
	switch val := v.(type) {
	case binding.Bool:
		data := widget.NewCheckWithData("", val)
		fieldMap[field] = data
		return data
	case binding.Float:
		data := widget.NewSliderWithData(0, 1, val)
		fieldMap[field] = data
		s := data
		s.Step = 0.01
		return s
	case binding.Int:
		data := widget.NewEntryWithData(binding.IntToString(val))
		fieldMap[field] = data
		data.OnCursorChanged = SaveConfig
		return data
	case binding.String:
		data := widget.NewEntryWithData(val)
		fieldMap[field] = data
		data.OnCursorChanged = SaveConfig
		return data
	default:
		return widget.NewLabel("")
	}
}

func SaveConfig() {
	jsonByte, err := json.Marshal(Config)
	if err != nil {
		ErrorDialog(err)
		return
	}
	jsonStr := string(jsonByte)
	if strings.Compare(jsonStr, configJson) == 0 {
		return
	}
	configJson = jsonStr
	cnfPath := filepath.Join(RootPath, "conf", "conf.json")
	file, err := os.OpenFile(cnfPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		ErrorDialog(err)
		return
	}
	defer file.Close()
	file.Write(jsonByte)
}

func ErrorDialog(err error) {
	lock := alertMutex.TryLock()
	if lock {
		dialog.NewConfirm("执行错误", err.Error(), func(b bool) {
			alertMutex.Unlock()
		}, TopWin).Show()
	}
}

func ShowProgress() {
	Cancel = false
	if pDialog != nil {
		pDialog.Show()
		return
	}
	progress := widget.NewProgressBarInfinite()
	pDialog = dialog.NewCustom("progress", "               取消               ", progress, TopWin)
	InitDismissBtn()
	pDialog.Show()
}
func HideProgress() {
	Cancel = true
	pDialog.Hide()
}

func InitDismissBtn() {
	v0 := reflect.ValueOf(&pDialog)
	v1 := v0.Elem().Elem().Elem()
	v := v1.FieldByName("dismiss")
	vb := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface()
	btn := vb.(*widget.Button)
	btn.OnTapped = func() {
		Cancel = true
		pDialog.Hide()
		log.Println("Cancel")
	}
}
