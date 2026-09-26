# coj — CandyOJ 命令行客户端

用 Go 编写的 CandyOJ 在线评测教学平台 CLI，命令结构与 [gh](https://cli.github.com/) 一致，
遵循 REST 资源模型：

```
coj <resource> <action> [flags]
```

后端 190 个接口全部可访问，覆盖 17 个控制器、53 条路由、47 个页面。

---

## 1. 安装

```bash
cd coj-cli
go build -o coj .
sudo mv coj /usr/local/bin/
```

需要 Go 1.21+，**无第三方依赖**（命令解析、表格输出、jq 子集均为标准库实现）。

---

## 2. 快速开始

```bash
# 登录（会话保存在 ~/.config/coj/session）
coj auth login --username alice --password secret
coj auth login --username alice              # 交互式输入密码

# 短信登录（两步）
coj auth sms --phone 13800138000 --send
coj auth sms --phone 13800138000 --code 123456

# 查看身份
coj auth status

# 列比赛
coj match list
coj match list --page 2
coj match list --paginate          # 自动翻页拉全量
```

---

## 3. 命令结构

### 3.1 资源命令（190 个，自动生成）

资源即后端控制器，动作即接口。每个动作同时提供完整名与短别名：

| 命令 | 接口 |
|---|---|
| `coj match list` | `POST /Match/getMatchList` |
| `coj match get-match-list` | 同上（完整动作名） |
| `coj match rank --field F_MatchID=42` | `POST /Match/getMatchRank` |
| `coj subject list` | `POST /Subject/getSubjectListNew` |
| `coj testlog do-test --field ...` | `POST /TestLog/doTest` |
| `coj user one-key-reset-pwd 1001` | `POST /User/oneKeyResetPwd` |

资源清单（括号内为别名）：

| 资源 | 接口数 | 说明 |
|---|---|---|
| `match` | 33 | 比赛：创建、答题、榜单、导出 |
| `user` | 32 | 用户：登录、注册、改密、建号 |
| `homework` | 18 | 作业 |
| `class` (`fclass`) | 15 | 班级 |
| `bag` (`paper-bag`) | 13 | 试卷袋 |
| `exam` | 13 | 考试/题库 |
| `course` | 12 | 课程 |
| `subject` (`problem`) | 12 | 题目（含测试数据、Hydro 同步） |
| `chapter` | 10 | 章节 |
| `chapter-subject` (`cs`) | 8 | 章节题目 |
| `testlog` (`log`) | 7 | 评测记录 |
| `tag` / `origin` / `difficulty` / `grade` | 11 | 标签体系 |
| `home` | 4 | 首页统计（四种角色面板） |
| `hydro` | 2 | Hydro OJ 同步 |

随时查看某个动作的用法与字段：

```bash
coj match rank --help
# 获取排名
#
# 接口: POST /Match/getMatchRank
# 参数位置: URL 查询串（注意：即使是 POST 也走查询串）
#
# 已知字段
#   F_MatchID              前端取值: this.$route.query.kid
#   F_ClassID              前端取值: this.class_id
```

### 3.2 逃逸口 `coj api`

直接调用任意接口，覆盖 CLI 未封装的能力：

```bash
coj api POST /Subject/getSubjectListNew --field page=1
coj api GET  '/User/login?user_name=alice&user_pwd=secret'
coj api POST /TestLog/doTest --body --field F_CodeContent=@solution.cpp
coj api --list                          # 列出全部 190 个接口
coj api --list --filter tongbu          # 按关键字过滤
```

---

## 4. 参数传递

### 4.1 三种编码方式自动选择

这是本后端最容易踩的坑，CLI 已按接口定义自动处理，**无需手动指定**：

| 前端方法 | 参数位置 | 接口数 | CLI 行为 |
|---|---|---|---|
| `$get` | URL 查询串 | 18 | 自动 |
| `$post` | **URL 查询串** | 256 处调用 | 自动 |
| `$post2` | JSON 请求体 | 17 | 自动 |
| `$upfile` | multipart/form-data | 20 | 自动 |

> 注意：绝大多数 POST 接口把参数放在 URL 查询串而非请求体。
> 用 `--verbose` 可以确认实际发出的请求。

### 4.2 标志

| 标志 | 说明 |
|---|---|
| `--field k=v` / `-F k=v` | 请求字段，可重复 |
| `--field k=@path` | 从文件读取值（提交代码时很好用） |
| `--raw-field k=v` | 强制放 URL 查询串 |
| `--file k=/path` | 上传文件字段 |
| `--page N` | 页码 |
| `--paginate` | 自动翻页拉全量 |
| `--force` | 超出限额(461)时确认继续 |
| `--host` | 站点地址，默认 `https://candyoj.com` |
| `--session` | 直接指定 PHPSESSID |

位置参数会填充为主键字段：

```bash
coj match del-match 42          # 等价于 --field F_ID=42
coj user delete-user 1001,1002  # 多个 ID 用逗号（后端批量接口如此）
```

---

## 5. 输出控制

默认人类可读表格，gh 风格的三个标志任选其一：

```bash
coj match list                              # 表格
coj match list --json                       # 原始 JSON
coj match list --jq '.list[].F_Name'        # 字段提取
coj match list --template '{{.count}} 场'   # Go 模板
coj match list --no-header                  # 去表头（便于管道处理）
```

`--jq` 支持子集语法：

```
.count                 顶层字段
.list                  整个数组
.list[].F_Name         展开后取字段
.list[0].F_ID          按下标取
.data.list[].F_Title   多级路径
```

示例：批量导出某场比赛的参赛学生账号

```bash
coj match rank --field F_MatchID=42 --jq '.list[].F_Account' > accounts.txt
```

---

## 6. 错误处理

后端响应码已映射为可操作的提示：

| code | CLI 行为 |
|---|---|
| 200 | 正常输出 |
| 440 | `未登录或会话已过期（440）` → 请重新 `coj auth login` |
| 461 | `超出限额（461）…加 --force 重发` |
| 其他 | `接口 /Xxx 返回 code=NNN: <msg>` |

```bash
$ coj match exportrecord --field F_MatchID=42
coj: 超出限额（461）：超出导出限额；如确需继续，加 --force 重发

$ coj match exportrecord --field F_MatchID=42 --force
EXPORT_ID
EXP-77
```

---

## 7. 典型场景

### 7.1 提交代码并查看结果

```bash
# 提交（代码从文件读取）
coj testlog do-test \
  --field F_SubjectID=7 \
  --field F_Language=cpp \
  --field F_CodeContent=@solution.cpp

# 查看评测记录
coj testlog get-test-log-list --paginate --jq '.list[].F_Res'

# 取回某次提交的代码
coj testlog get-code-content --field F_ID=90001
```

`F_Res` 判题状态：0等待 1评测中 2编译错误 3答案错误 4超时 5内存超限
6运行时错误 7输出超限 8部分正确 9通过 10编译中 11提交失败

### 7.2 批量建号

```bash
# 先预览账号
coj user account-preview --field F_Names="zhangsan,lisi" --field F_Pre=stu

# 正式创建（Excel 走上传接口）
coj user student-create --file file=./students.xlsx --field F_Type=4
```

### 7.3 题目管理

```bash
coj subject list --paginate --json
coj subject change-status --field F_ID=7,8,9 --field F_Status=1
coj hydro tongbufromhydrooj --field problemID=1000   # 从 Hydro 同步
```

### 7.4 比赛运营

```bash
coj match list --paginate
coj match get-match-rank --field F_MatchID=42 --field F_ClassID=7
coj match publish-score --field F_MatchID=42
coj match exportrecord --field F_MatchID=42 --force
```

---

## 8. 配置与凭据

| 路径 | 用途 |
|---|---|
| `~/.config/coj/config.json` | 站点、账号、角色、超时 |
| `~/.config/coj/session` | PHPSESSID（权限 0600） |

可用 `COJ_CONFIG_DIR` 覆盖目录，`COJ_USERNAME` / `COJ_PASSWORD` 供非交互环境使用。

```bash
coj --host http://120.24.228.173 status    # 切换站点
coj status                                 # 查看当前配置
coj auth logout                            # 清除会话
```

---

## 9. 会话机制说明

后端 `Common::checkLogin()` 用 session 中的 timestamp 做 **120 分钟滑动窗口**。
按设计，`/User/heartbeat` 会刷新窗口，而 `searchInvite`、`getCanMoveClassList`
等轮询接口**只校验不续期**。长时间挂起的脚本建议周期性执行：

```bash
coj auth ping
```

---

## 10. 代码结构

```
coj-cli/
├── main.go                       入口
├── internal/
│   ├── cli/cli.go                零依赖命令框架（命令树、标志解析、帮助）
│   ├── api/
│   │   ├── client.go             HTTP 客户端：会话、编码、错误映射
│   │   └── endpoints_gen.go      190 个端点定义（由前端产物自动生成）
│   ├── cmd/
│   │   ├── root.go               根命令与全局标志
│   │   ├── auth.go               登录/登出/心跳
│   │   ├── api.go                coj api 逃逸口
│   │   └── resource.go           190 个资源命令（自动生成）
│   ├── iostreams/iostreams.go    终端能力：CanPrompt / TTY 判定（对标 gh pkg/iostreams）
│   ├── prompter/                 交互式提示（对标 gh internal/prompter）
│   │   ├── prompter.go           Prompter 接口与终端实现
│   │   ├── search.go             MultiSelectWithSearch：Search 哨兵 + 按需搜索
│   │   ├── select.go             可过滤选择器（raw mode，方向键/空格）
│   │   ├── input.go               单行输入与密码（无回显）
│   │   └── prompter_mock.go      测试用 Mock
│   ├── browser/browser.go        Browser 接口 Browse(string)（对标 gh internal/browser）
│   ├── config/config.go          配置与凭据存储
│   ├── term/                     跨平台密码读取（按 build tag 分文件）
│   │   ├── term.go               公共接口 ReadPassword / IsTerminal
│   │   ├── term_linux.go         TCGETS / TCSETS
│   │   ├── term_darwin.go        TIOCGETA / TIOCSETA（macOS & BSD）
│   │   ├── term_windows.go       GetConsoleMode / SetConsoleMode
│   │   └── term_fallback.go      其余平台：stty -echo
│   └── output/output.go          表格 / JSON / jq 子集 / Go 模板
└── README.md
```

`endpoints_gen.go` 由 `extract/gen_go.py` 从前端 bundle 提取的
`deep_api_all.json` 生成，包含每个接口的路径、方法、参数位置（query/body/upload）
与已知字段——**这是 CLI 能正确处理参数编码的关键**。

---

## 10.1 交互模型（对齐 gh）

交互不是"缺参数就弹个输入框"。coj 照搬 gh 的 `internal/prompter` 模型：

### Prompter 是接口，不是具体类型

```go
type Prompter interface {
    Select(prompt, defaultValue string, options []string) (int, error)
    MultiSelect(prompt string, defaults, options []string) ([]int, error)
    MultiSelectWithSearch(prompt, searchPrompt string, defaults, persistentOptions []string,
        searchFunc func(string) MultiSelectSearchResult) ([]string, error)
    Input(prompt, defaultValue string) (string, error)
    Password(prompt string) (string, error)
    Confirm(prompt string, defaultValue bool) (bool, error)
    ConfirmDeletion(requiredValue string) error
}
```

好处和 gh 一样：命令依赖行为而非实现，测试可以注入
`prompter.Mock` 断言"问了什么、选了什么"，无需终端。

### MultiSelectWithSearch 是主力

不一次性拉全量再让你挑，而是：

- 列表首项是 **Search 哨兵**（有更多结果时显示 `Search (N more)`）
- 选中 Search → 输入关键词 → 再次调用 `searchFunc` 重新拉
- 已选项在后续搜索中**持续保留**
- 因为列表是动态的，返回的是 **Keys（值）** 而不是下标

```
? Select Class ID (type to filter)
> [ ] Search
    [ ] 高三(1)班 (7)
    [ ] 高三(2)班 (8)
Use arrows to move, Space to toggle, Enter to confirm
```

### 只在能交互时才提示

判定沿用 gh 的 `CanPrompt()`：

```
CanPrompt = stdin 是 TTY && stdout 是 TTY && !neverPrompt
```

`neverPrompt` 由 `--no-prompt`、`COJ_PROMPT_DISABLED`、
`GH_PROMPT_DISABLED` 任一触发。**不满足时直接报错并给出示例命令，绝不静默挂起**。

### 不烦人

- 显式传入的字段永远优先，不再追问
- 只要调用方提供了任一 ID，其余上下文 ID 一律当作可选，保持安静
  （`doTest` 的 `F_CourseID`/`F_MatchID` 常为空，逐个追问会让提交变成审讯）
- `page`、`F_Status` 这类分页与筛选项从不询问

### 写操作的收尾确认

参照 gh 的 `confirmSubmission`("What's next?")：交互式填完参数后，
写操作会给出 **Submit / Continue in browser / Cancel**。
仅当本次真的做过提示时才出现——脚本化运行必须保持确定性。

```
? What's next?  (Submit)
  Submit
  Continue in browser
  Cancel
```

破坏性操作（`delete`/`oneKeyResetPwd` 等，且 id 为逗号批量）走
`ConfirmDeletion`：必须手动输入目标值才能执行。

```bash
$ coj user one-key-reset-pwd 1001,1002,1003
About to run /User/oneKeyResetPwd on 3 ids.
? Type "1001,1002,1003" to confirm:
```

### --web 是交互的另一种出口，不是放弃

gh `pr view --web` 仍会先解析出 PR（只取 `url` 字段），再打开浏览器。
coj 同理：`--web` 时不做业务调用，把已定位到的页面打开；
只有网页确实需要 id 而调用方没给时才询问。

---

## 11. 跨平台说明（密码读取）

`coj auth login` 在终端中读取密码时需要关闭回显。终端控制常量是**平台相关**的：

| 平台 | 常量 / API |
|---|---|
| Linux | `TCGETS` / `TCSETS` |
| macOS / BSD | `TIOCGETA` / `TIOCSETA` |
| Windows | `GetConsoleMode` / `SetConsoleMode` |

直接在代码里写 `syscall.TCGETS` 会导致 **macOS 编译失败**（该常量仅在 Linux 存在）。
本项目因此把密码读取拆到 `internal/term/`，用 build tag 分平台实现，
**不引入任何第三方依赖**：

```go
// term_darwin.go
//go:build darwin || freebsd || netbsd || openbsd || dragonfly
const (
    getTermios = 0x40487413 // TIOCGETA
    setTermios = 0x80487414 // TIOCSETA
)
```

非交互环境（管道、CI、脚本）自动退化为从 stdin 读一行，无需特殊处理：

```bash
echo 'secret' | coj auth login --username alice
COJ_PASSWORD=secret coj auth login --username alice
```

### 若你更想用 golang.org/x/term

把 `internal/term` 换成官方库只需两步：

1. `go get golang.org/x/term`
2. 将 `term_linux.go` / `term_darwin.go` / `term_windows.go` / `term_fallback.go`
   全部删除，只留一个 `term.go`：

```go
package term

import (
    "bufio"
    "os"
    "strings"

    "golang.org/x/term"
)

func ReadPassword(prompt string) (string, error) {
    fd := int(os.Stdin.Fd())
    if !term.IsTerminal(fd) {
        // 非交互：读一行
        if prompt != "" {
            os.Stderr.WriteString(prompt)
        }
        line, err := bufio.NewReader(os.Stdin).ReadString('\n')
        if err != nil && line == "" {
            return "", err
        }
        return strings.TrimRight(line, "\r\n"), nil
    }
    os.Stderr.WriteString(prompt)
    raw, err := term.ReadPassword(fd)
    os.Stderr.WriteString("\n")
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(raw)), nil
}

func IsTerminal(fd int) bool { return term.IsTerminal(fd) }
```

代价是引入一个外部依赖（本项目其余部分零依赖，故默认未采用）。

---

## 12. 已知限制

- **后端未实现的分页**：部分接口忽略 `limit`，只在前端截断；`--paginate` 依赖后端返回 `count`。
- **样式字段**：47 个页面组件的 `<style scoped>` 在独立 css-loader 模块中，未合并进还原的 `.vue`。
- **站点可达性**：`candyoj.com` 与 `120.24.228.173` 部署在阿里云 WAF 之后，
  部分来源 IP 会被边缘策略拒绝（`403 policy_default_denied`），此时需通过 WAF 放行。
