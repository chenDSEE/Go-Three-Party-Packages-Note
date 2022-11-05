> [cobra(v1.5.0)](https://github.com/spf13/cobra/tree/v1.5.0)，CLI/command
> [pflag(v1.0.5)](https://github.com/spf13/pflag/tree/v1.0.5)，命令行参数
> [viper(v1.13.0)](https://github.com/spf13/viper/tree/v1.13.0)，配置文件

CLI 命令、命令参数、配置文件，这三者并不是独立存在的，他们都共同涉及到了一个核心的问题：应用程序启动过程及其初始化。而在这一个过程中，最为常规的三件事则是：读取并解析配置文件（viper），为应用程序添加命令行参数（pflag），为应用程序以 CLI 的使用风格添加功能（cobra）。

配置文件很简单，就是在服务器上那些 `.conf`, `.xml`, `.json`, `.yaml` 等等结尾的实际文件。但是 `cobra` 跟 `pflag` 呢？举些栗子：

```bash
命令行（CLI）：
# app-name start
# app-name stop
# app-name restart

命令行参数：
# app-name --host 127.0.0.1 --port 8080
# app-name --config /PATH/TO/FILE

命令与参数混合：
# app-name start --host 127.0.0.1 --port 8080

命令、参数、配置文件同时指定
# app-name start --host 127.0.0.1 --port 8080 --config /PATH/TO/FILE
```

从上面的例子，很显然，命令、参数、配置文件之间其实是相互纠缠在一起的。

> corba 对于三者的区分：**Commands** represent actions, **Args** are things and **Flags** are modifiers for those actions.
>
> command 意思是：你想要做出什么操作？
>
> args 则是想要操作的对象是谁？
>
> flag 则是对这个操作动作做出怎样的微调？





下面，我们先分开看看它们各自是怎么使用的



# pflag(v1.0.5)

![pflag-overview](pflag-overview.svg)

如上图所示，`pflag` package 的核心是：
1. 使用者如何向 `pflag` package 注册自己的 flag 名字、用法、类型
2. `pflag` package 针对 CLI 输入进行解析处理（string --> int/string/IP/IPNet/..../[]int 等等）
3. 使用者按照通过访问变量数据的方法，获得相应的 CLI flag
4. 为了使用者的方便，实际上 `pflag` package 提供了很多类型的转换（如：time，IP，IPNet 等等）

值得注意的是：每次 CLI 输入后，程序中先读到的必然是 string，而我们想要的是通过不同类型的变量去访问相应的输入值；所以 `pflag` 的核心是：解决 string 向不同类型变量的转换问题，并提供一个简单易用的框架。

## demo

```go
package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"time"
)

func main() {
	fmt.Printf("======> app start\n")

	pflagParse_Demo()
	//shorthandFlag_Demo()
	//CLIFlagSyntax_Demo()
	//BoolCLIFlagSyntax_Demo()
	//multiFlagValue_Demo()
	//count_Demo()
	//time_Demo()
	//network_Demo()
	//slice_Demo()
	//bytes_Demo()

	fmt.Printf("======> app end\n")
}

func pflagParse_Demo() {
	// 默认值在通过 pflag 设置的这一瞬间，就会顺便写入响应变量中
	var port *int = pflag.Int("port", 80, "server port to listen on")

	var ip string
	pflag.StringVar(&ip, "ip", "127.0.0.1", "server ip to listen on")

	// test case:
	// 1. go build -o pflag_demo  pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo  pflag_demo.go && ./pflag_demo --port 8080
	// 3. go build -o pflag_demo  pflag_demo.go && ./pflag_demo --ip 192.168.0.1
	fmt.Printf("========== pflagParse_Demo() ==========\n")
	fmt.Printf("before pflag.Parse()==> ip:[%s], port:[%d]\n", ip, *port)

	// pflag.Parse() 必须显式调用，然后 pflag package 才会去处理 CLI 输入
	// 不显式调用这个 pflag.Parse(), 甚至连 -h 都不会响应的
	// 这其实也就意味着：pflag 相关的变量设置，要在 pflag.Parse() 调用之前完成
	pflag.Parse()
	fmt.Printf("after pflag.Parse()==>ip:[%s], port:[%d]\n", ip, *port)
}

func shorthandFlag_Demo() {
	// 用带 P 后缀的函数，需要额外指定简写形式
	var port *int = pflag.IntP("port", "p", 80, "server port to listen on")

	var ip string
	pflag.StringVarP(&ip, "ip", "i", "127.0.0.1", "server port to listen on")

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo --ip 2.2.2.2 --port 8081
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -i 1.1.1.1 -p 8080

	pflag.Parse()
	fmt.Printf("========== shorthandFlag_Demo() ==========\n")
	fmt.Printf("ip:[%s], port:[%d]\n", ip, *port)
}

func CLIFlagSyntax_Demo() {
	var port *int = pflag.IntP("port", "p", 80, "server port to listen on")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -p 81
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -p81
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo -p=81
	// 5. go build -o pflag_demo pflag_demo.go && ./pflag_demo --port 81
	// 6. go build -o pflag_demo pflag_demo.go && ./pflag_demo --port81 不允许
	// 7. go build -o pflag_demo pflag_demo.go && ./pflag_demo --port=81

	fmt.Printf("========== CLIFlagSyntax_Demo() ==========\n")
	fmt.Printf("port:[%d]\n", *port)
}

func BoolCLIFlagSyntax_Demo() {
	var a, b, c bool
	pflag.BoolVarP(&a, "bool-a", "a", false, "bool type flag for a")
	pflag.BoolVarP(&b, "bool-b", "b", false, "bool type flag for b")
	pflag.BoolVarP(&c, "bool-c", "c", false, "bool type flag for c")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -abc
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -a -b -c
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo -a=true -b -c

	fmt.Printf("========== BoolCLIFlagSyntax_Demo() ==========\n")
	fmt.Printf("a[%v], b[%v], c[%v]\n", a, b, c)
}

func count_Demo() {
	// 其实就像是 tcpdump -vvv, 根据 v 出现的次数，作为等级、程度衡量
	cntA := pflag.CountP("cnt-a", "a", "number for cntA")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -a
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -aaa
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo --cnt-a
	// 5. go build -o pflag_demo pflag_demo.go && ./pflag_demo --cnt-a --cnt-a

	fmt.Printf("========== count_Demo() ==========\n")
	fmt.Printf("count for a[%d]\n", *cntA)
}

func time_Demo() {
	// string 转换为标准库 time.Duration

	t := pflag.DurationP("time", "t", 1*time.Second, "time in second")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -t 2(失败，因为没有指定时间单位)
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -t 2s
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo --time 3ms
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo --time 3000ms
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo -t 0.5s

	fmt.Printf("========== count_Demo() ==========\n")
	fmt.Printf("time[%v]\n", *t)
}

func network_Demo() {
	// string 转换为 net package 的参数
	ip := pflag.IPP("host", "h", net.ParseIP("1.1.1.1"), "server IP to listen on")
	_, network, _ := net.ParseCIDR("1.1.1.0/24")
	ipNet := pflag.IPNetP("net", "n", *network, "server net to handle")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h 2.2.2.2
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h 3:3::3
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo -n 4.4.4.0/24
	// 5. go build -o pflag_demo pflag_demo.go && ./pflag_demo -n 5:5::5/64
	fmt.Printf("========== network_Demo() ==========\n")
	fmt.Printf("ip[%s], ipNet[%s]\n", ip.String(), ipNet.String())
}

func slice_Demo() {
	var numSlice []int
	pflag.IntSliceVarP(&numSlice, "port", "p", []int{1, 2}, "server port listen on")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -p 1,2,3,4,5
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo --port 6,7,8,9
	// 5. go build -o pflag_demo pflag_demo.go && ./pflag_demo -p 1 -p 3
	fmt.Printf("========== slice_Demo() ==========\n")
	for _, port := range numSlice {
		fmt.Printf("port[%d]\n", port)
	}
}

func bytes_Demo() {
	var b []byte
	pflag.BytesHexVarP(&b, "byte", "b", []byte{}, "data in HEX byte")
	pflag.Parse()

	// test case:
	// 1. go build -o pflag_demo pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo pflag_demo.go && ./pflag_demo
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -b 0x1234567890(失败。不需要 0x)
	// 3. go build -o pflag_demo pflag_demo.go && ./pflag_demo -b 1234567890
	// 4. go build -o pflag_demo pflag_demo.go && ./pflag_demo --byte 1234567890
	// 5. go build -o pflag_demo pflag_demo.go && ./pflag_demo -b 1230 -b 5670(前面的会被覆盖掉)
	fmt.Printf("========== slice_Demo() ==========\n")
	fmt.Printf("Hex Byte[0x%x]\n", b)
}

```

## 主要 object
![pflag-object](pflag-object.svg)

### `FlagSet` struct
```go
// A FlagSet represents a set of defined flags.
// 多个 Flag 的管理集合。因为 pflag 的具有很多不同的 flag 用法，实际上不同的用法，都会导致
// FlagSet 这里不得不增加一个相应的管理结构体，
// 如：FlagSet.formal，FlagSet.actual，FlagSet.orderedFormal，FlagSet.sortedFormal，FlagSet.shorthands 等等
type FlagSet struct {
	........
	name              string
	parsed            bool // FlagSet.Parse() 调用的标识。基本没用，对应 FlagSet.Parsed()
	
	// Q&A(DONE): actual 跟 formal 有什么区别？
	// actual 是实际命令中输入了那些命令
	// formal 是代码编写者写了哪些命令
	actual            map[NormalizedName]*Flag
	formal            map[NormalizedName]*Flag // 一般化的 flag-name 作为 key，Flag struct 作为 value
	shorthands        map[byte]*Flag // shorthand 简写 mapping to Flag struct
	
	// 这些 ordered 是为了在不同场景下，采用不同的顺序进行遍历
	orderedActual     []*Flag
	orderedFormal     []*Flag // 按照 Flag 的注册顺序排列
	sortedActual      []*Flag
	sortedFormal      []*Flag

	args              []string // arguments after flags（pflag 剩下没有处理的 flag）
	..........
	normalizeNameFunc func(f *FlagSet, name string) NormalizedName

	addedGoFlagSets []*goflag.FlagSet // 跟 Go 标准库中的 flag 兼容
}
```
- 通常情况下，只使用 `pflag` 内置的 `CommandLine` 其实就足够了。而且自己创建一个新的 `FlagSet` 是一件挺麻烦的事情
- 所有由使用者注册的 `Flag` 都保存在 `FlagSet`；`FlagSet` 也是所有 CLI 输入 flag 的解析合集
- 因为 flag 字符上是不允许重复的，所以 `FlagSet` 采用了 hash map 的形式对注册进来的 `Flag` 进行管理


### `Flag` struct
```go
// A Flag represents the state of a flag.
// 当使用者调用了一堆 pflag 的函数对一个自己想要创建的 flag 进行描述之后
// 最终所有的描述信息将会被整合成一个 pflag.Flag 在 pflag 内部进行流动
// 因为 Flag struct 以及所有的 field 都是 public 的，所以你甚至可以自定义一个 Flag 让后加入 FlagSet 里面
type Flag struct {
	Name                string              // name as it appears on command line
	Shorthand           string              // one-letter abbreviated flag
	Usage               string              // help message
	Value               Value               // value as set
	DefValue            string              // default value (as text); for usage message
	Changed             bool                // If the user set the value (or if left to default)
	NoOptDefVal         string              // default value (as text); if the flag is on the command line without any options
	Deprecated          string              // If this flag is deprecated, this string is the new or now thing to use
	Hidden              bool                // used by cobra.Command to allow flags to be hidden from help/usage text
	ShorthandDeprecated string              // If the shorthand of this flag is deprecated, this string is the new or now thing to use
	Annotations         map[string][]string // used by cobra.Command bash autocomple code
}
```
- `Flag` struct 拥有了一个 flag 所需要的所有信息，包括：全称(`--port`)、简写(`-p`)、help 信息、默认值、输入值等等信息


### `Value` interface
```go
// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
// 将 int、int32、string 等等所有 flag 对应的值，都抽象为 pflag.Value interface，然后通过 FlagSet.VarP() 进行统一处理
// Q&A(DONE): 解析 pflag.Value interface 的设计思路
// 所有 flag value 在输入的那一瞬间起，必然是 string，
// 那么 string --> 代码的过程则是 string 向不同类型的转换。
// 而这之间的桥梁则是 Value interface
type Value interface {
	String() string // 从具体的 Value 转换为 string

	// 需要注意的是：Value.Set() 是要把相应的值，设置到相应变量的内存上面的
	Set(string) error // FlagSet.Set() 针对 Flag 进行调用，为这个 Flag 进行 string 向 Value.Type() 的转换
	Type() string
}
```
- `pflag.Value` interface 是不同种类变量的抽象。
	- 通过提供 string 向具体变量种类转换的 method（`Value.Set()`）
	- 查询 `Value` interface 实际的变量类型（`Value.Type()`）
	- 实际变量值向 string 的转换（`Value.String()`）
- 这也就是说，`Value` interface 封装、并抽象了：默认值、变量类型、转换方式、实际值
- `Flag` 通过 Has-a `Value` 的方式，来处理这个 flag 对应的值


## 如何向 `FlagSet` 注册一个 flag ？
![pflag-register](pflag-register.svg)

1. 首先一块内存地址(`&ip`)以及这块内存地址的变量类型(`FlagSet.StringVar()`)向 `FlagSet` 进行注册;
1.1 或者是由 `FlagSet` 直接分配一块内存地址(`var port *int`)，然后指定地址的种类(`FlagSet.Int()`)
2. 在这个注册的过程中，由 `FlagSet.VarPF()` 将所有注册信息，打包成一个 `Flag` struct 并保存进 `FlagSet` 中
3. 在 `FlagSet.StringVar()` 向 `FlagSet.VarP()` 传递的时候，会将 string 这种变量类型，通过 `newStringValue()` 统一为 `pflag.Value` interface;（默认值、值类型、string 解析处理都会被封装进这个 `pflag.Value` interface 中）
4. 将使用者为这个 flag 设置的全称、简写、使用信息、`Value` 封装为一个 `Flag` struct
5. 将 `Flag` struct 放入 `FlagSet` struct 相应的 hash-map 中，待后续解析过程中进行查找


具体代码流程如下：
```go
func pflagParse_Demo() {
	/* ===== step 1 ===== */
	// 默认值在通过 pflag 设置的这一瞬间，就会顺便写入相应变量中
	var port *int = pflag.Int("port", 80, "server port to listen on")

	var ip string
	pflag.StringVar(&ip, "ip", "127.0.0.1", "server ip to listen on")

	// test case:
	// 1. go build -o pflag_demo  pflag_demo.go && ./pflag_demo -h
	// 2. go build -o pflag_demo  pflag_demo.go && ./pflag_demo --port 8080
	// 3. go build -o pflag_demo  pflag_demo.go && ./pflag_demo --ip 192.168.0.1
	fmt.Printf("========== pflagParse_Demo() ==========\n")
	fmt.Printf("before pflag.Parse()==> ip:[%s], port:[%d]\n", ip, *port)

	// pflag.Parse() 必须显式调用，然后 pflag package 才会去处理 CLI 输入
	// 不显式调用这个 pflag.Parse(), 甚至连 -h 都不会响应的
	// 这其实也就意味着：pflag 相关的变量设置，要在 pflag.Parse() 调用之前完成
	pflag.Parse()
	fmt.Printf("after pflag.Parse()==>ip:[%s], port:[%d]\n", ip, *port)
}

/* ===== step 2 ===== */
// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
	CommandLine.VarP(newStringValue(value, p), name, "", usage) /* ===== step 3 ===== */
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {
	// newStringValue() 返回 stringValue
	// VarP() 接收 Value
	f.VarP(newStringValue(value, p), name, "", usage) /* ===== step 3 ===== */
}

/* ===== step 3 ===== */
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}


/* ===== step 4 ===== */
// VarP is like Var, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) VarP(value Value, name, shorthand, usage string) {
	f.VarPF(value, name, shorthand, usage)
}

// 把使用者输入的 flag-name、shorthand、usage 信息通通变成一个 pflag.Flag struct
// VarPF is like VarP, but returns the flag created
func (f *FlagSet) VarPF(value Value, name, shorthand, usage string) *Flag {
	// Remember the default value as a string; it won't change.
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.AddFlag(flag) // 将 Flag 放入 FlagSet 中
	return flag
}


/* ===== step 5 ===== */
// 将 Flag 放入 FlagSet 中的 hash map
// AddFlag will add the flag to the FlagSet
func (f *FlagSet) AddFlag(flag *Flag) {
	/* long-name 的处理 */
	normalizedFlagName := f.normalizeFlagName(flag.Name) // 对 flag-name 进行一般化（统一化）

	// NOTE: 可以从一个 nil 的 map 中 read，但是不可以往一个 nil 的 map write
	_, alreadyThere := f.formal[normalizedFlagName]
	if alreadyThere {
		// 不接受 flag 的 redefine
		msg := fmt.Sprintf("%s flag redefined: %s", f.name, flag.Name)
		fmt.Fprintln(f.out(), msg)
		panic(msg) // Happens only if flags are declared with identical names
	}
	if f.formal == nil {
		f.formal = make(map[NormalizedName]*Flag) // lazy 的初始化
	}

	flag.Name = string(normalizedFlagName)
	f.formal[normalizedFlagName] = flag
	f.orderedFormal = append(f.orderedFormal, flag)

	/* short-name 的处理 */
	if flag.Shorthand == "" {
		return
	}

	.......

	if f.shorthands == nil {
		f.shorthands = make(map[byte]*Flag)
	}
	c := flag.Shorthand[0]
	used, alreadyThere := f.shorthands[c]
	if alreadyThere {
		msg := fmt.Sprintf("unable to redefine %q shorthand in %q flagset: it's already used for %q flag", c, f.name, used.Name)
		fmt.Fprintf(f.out(), msg)
		panic(msg)
	}
	f.shorthands[c] = flag
}


```


## `pflag.Parse()` 做了什么？
![pflag-parse](pflag-parse.svg)

1. 调用 `pflag.Parse()` 开始对 CLI 输入参数（`os.Args[1:]`）进行字符处理与转换；
2. `FlagSet.parseArgs()` 区分输入的参数中（`--port 8080`），flag 部分（`-p`，`--port`）与参数部分（`80`, `8080`）
3.1 `--` 开头，则是通过全称来进行 flag 指定的
3.2 `-` 开头，则是通过简称来进行 flag 指定的
4. 通过 `FlagSet.Set()` 调用 `Flag.Value.Set()` 进而调用到实际 struct 的 `Set()` method
5. 在不同的 `xxxValue.Set()` 中，从字符串向具体类型进行转换，并保存到相应的变量中
6. 通过访问变量的方式，获取 CLI 输入的参数值

具体代码流程如下：
```go
/* ===== step 1 ===== */
// Parse parses the command-line flags from os.Args[1:].  Must be called
// after all flags are defined and before flags are accessed by the program.
// pflag 利用 os.Args[1:] 拿到相应的命令行输入
func Parse() {
	// Ignore errors; CommandLine is set for ExitOnError.
	// os.Args[0] 就是可执行文件的名称，忽略
	CommandLine.Parse(os.Args[1:])
}

// Parse parses flag definitions from the argument list, which should not
// include the command name.  Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help was set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	..........
	f.parsed = true

	if len(arguments) < 0 {
		// 没有 flag 输入，不需要处理
		return nil
	}

	// 存放 pflag 没有处理的 CLI 输入
	f.args = make([]string, 0, len(arguments)) // 确实，放进 FlagSet.parseArgs() 区初始化更好，但是没关系了

	set := func(flag *Flag, value string) error {
		// 把 f 这个 FlagSet 直接闭包进来，后续的 set() 就直接省去指定 Flag.Name 跟 FlagSet.Set()
		// 说白了就是偷懒
		return f.Set(flag.Name, value)
	}

	err := f.parseArgs(arguments, set) /* ===== step 2 ===== */
	if err != nil {
		// 注意，这里没有根据 err 的种类来决定接下来的行为
		// 而是通过在处理过程中设置的 flag，决定接下来的行为
		// 正式因为 default FlagSet 设置了 ExitOnError，所以输入错误、-h 才会直接退出运行
		switch f.errorHandling {
		case ContinueOnError:
			// 使用者自行决定要不要终止
			return err
		case ExitOnError:
			// 强制终止
			fmt.Println(err)
			os.Exit(2)
		case PanicOnError:
			// 强制终止
			panic(err)
		}
	}

	// 没有任何错误发生，可以继续运行
	return nil
}

/* ===== step 2 ===== */
// 1. 从命令行输入的 []string 中，找到指定的是哪一个 Flag，
// 2. 通过 parseFunc 将输入的 string 设置为 Flag.Value
func (f *FlagSet) parseArgs(args []string, fn parseFunc) (err error) {
	for len(args) > 0 {
		/* 找到当前命令行中指定的 Flag */
		s := args[0]
		args = args[1:]
		if len(s) == 0 || s[0] != '-' || len(s) == 1 {
			if !f.interspersed {
				// 异常输入，直接输出
				f.args = append(f.args, s)
				f.args = append(f.args, args...)
				return nil
			}

			// 这个 flag 有多个输入值, 当前的 arg 是 value 之一
			// 但是 pflag 不支持这样的输入方式：-p val-1 val-2 val-3
			f.args = append(f.args, s)
			continue
		}

		if s[1] == '-' { // 长 flag，len(s) = 1 的情况已经被上面拦住了
			if len(s) == 2 { // "--" terminates the flags
				// https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html:
				// The argument -- terminates all options;
				// any following arguments are treated as non-option arguments,
				// even if they begin with a hyphen.
				//
				// man bash
				// A  --  signals the end of options and disables further option processing.
				// Any arguments after the -- are treated as filenames and arguments.
				// An argument of - is equivalent to --.
				//
				// 标识所有 flag 的结束（K8S 的参数中就有这样的形式）
				// 而在位于 -- 后面的字符串，是给接下来的命令用的
				// 1. 所有 flag 都输入完了：docker run -it -- centos:8.1.1911 /bin/bash
				//    docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
				//    -- 是 [OPTIONS] 跟 IMAGE 的分隔符
				// 2. 后面的参数是给接下来的命令的：kubectl exec abc -c ddf -it -- sh
				f.argsLenAtDash = len(f.args)
				f.args = append(f.args, args...)
				break
			}

			// 因为不断去掉已经处理过的 args slice，所以返回的时候，要不停的接住
			args, err = f.parseLongArg(s, args, fn) /* ===== step 3.1 ===== */
		} else {
			args, err = f.parseShortArg(s, args, fn) /* ===== step 3.2 ===== */
		}
		if err != nil {
			return
		}
	}
	return
}

/* ===== step 3.1 ===== */
func (f *FlagSet) parseLongArg(s string, args []string, fn parseFunc) (a []string, err error) {
	a = args
	name := s[2:] // 去掉 '--'

	......

	split := strings.SplitN(name, "=", 2) // 可能有 "=", 也可能没有，但是都是拿 split[0]
	name = split[0]
	flag, exists := f.formal[f.normalizeFlagName(name)] // 找到相应的 Flag，后续才能将解析出来了的值，放到相应的内存上

	......

	var value string
	if len(split) == 2 {
		// '--flag=arg'
		value = split[1]
	} else if flag.NoOptDefVal != "" {
		// '--flag' (arg was optional)
		value = flag.NoOptDefVal
	} else if len(a) > 0 {
		// '--flag arg'
		value = a[0]
		a = a[1:]
	} else {
		// '--flag' (arg was required)
		err = f.failf("flag needs an argument: %s", s)
		return
	}

	err = fn(flag, value) // FlagSet.Set()
	if err != nil {
		f.failf(err.Error())
	}
	return
}

/* ===== step 3.2 ===== */
// 输入的 s 包含破折号
func (f *FlagSet) parseShortArg(s string, args []string, fn parseFunc) (a []string, err error) {
	a = args
	shorthands := s[1:]

	// "shorthands" can be a series of shorthand letters of flags (e.g. "-vvv").
	for len(shorthands) > 0 {
		shorthands, a, err = f.parseSingleShortArg(shorthands, args, fn)
		if err != nil {
			return
		}
	}

	return
}

// shorthands 可能的 case 有：
// 1. -a
// 2. -abc
func (f *FlagSet) parseSingleShortArg(shorthands string, args []string, fn parseFunc) (outShorts string, outArgs []string, err error) {
	outArgs = args

	outShorts = shorthands[1:] // 去掉 '-'
	c := shorthands[0]

	flag, exists := f.shorthands[c] // 找到相应的 Flag，后续才能将解析出来了的值，放到相应的内存上

	.......

	var value string
	if len(shorthands) > 2 && shorthands[1] == '=' {
		// '-f=arg'
		value = shorthands[2:]
		outShorts = ""
	} else if flag.NoOptDefVal != "" {
		// '-f' (arg was optional)
		value = flag.NoOptDefVal
	} else if len(shorthands) > 1 {
		// '-farg'
		value = shorthands[1:]
		outShorts = ""
	} else if len(args) > 0 {
		// '-f arg'
		value = args[0]
		outArgs = args[1:]
	} else {
		// '-f' (arg was required)
		err = f.failf("flag needs an argument: %q in -%s", c, shorthands)
		return
	}

	........

	err = fn(flag, value) // FlagSet.Set()

	........

	return
}

/* ===== step 4 ===== */
// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	normalName := f.normalizeFlagName(name)
	flag, ok := f.formal[normalName]
	......

	// 多态到不同的 Value，如：stringValue, intValue
	// 通过 interface 调用 intValue, stringValue 相应的转换函数
	// 并复制到相应的变量内存上
	err := flag.Value.Set(value) /* ===== step 5 ===== */

	......

	return nil
}

/* ===== step 5 ===== */
// string 向 stringValue 转换
func (s *stringValue) Set(val string) error {
	*s = stringValue(val) // 前置转换就完事了
	return nil
}

// string 向 intValue 转换
func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64) // 字符串向 int 转换
	*i = intValue(v)
	return err
}

// string 向 intSliceValue 转换
func (s *intSliceValue) Set(val string) error {
	ss := strings.Split(val, ",")
	out := make([]int, len(ss)) // 注册时，slice 是可以不初始化的
	for i, d := range ss {
		var err error
		out[i], err = strconv.Atoi(d)
		if err != nil {
			return err
		}

	}
	if !s.changed {
		*s.value = out
	} else {
		*s.value = append(*s.value, out...)
	}
	s.changed = true
	return nil
}

// string 向 ipValue 转换（net.IP）
func (i *ipValue) Set(s string) error {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return fmt.Errorf("failed to parse IP: %q", s)
	}
	*i = ipValue(ip)
	return nil
}

/* ===== step 6 ===== */
// 直接通过访问变量来获取值
fmt.Printf("after pflag.Parse()==>ip:[%s], port:[%d]\n", ip, *port)

```



### `pflag.Value` interface 的作用是什么？
我们先看看 `pflag` 中的文件分布，你会发现不同的数据类型都有一个相应的文件。而在这些文件里，做的事情都非常相同：将类型信息 + string 向本类型的转换方法(`Set()`) 封装为 `pflag.Value` interface.
`pflag.Value` interface 的存在，不仅仅是将每一种 flag 统一为了 `Flag` struct，而且为后续更多种类的数据结构，提供了很方便的横向扩展方式。
当你想让自己定义的数据结构作为一个 `Flag` 加入 `FlagSet` 中时，只需要模仿下面的文件就好了
```bash
[root@LC pflag]# ls | grep -v test
bool.go
bool_slice.go
bytes.go
count.go
.......
string_to_int64.go
string_to_int.go
string_to_string.go
uint16.go
uint32.go
uint64.go
uint8.go
uint.go
uint_slice.go
[root@LC pflag]# 
```










# viper(v1.13.0)

Viper 提供的主要功能是：解析配置文件。而要实现配置解析的功能，那就是要做到 string 向不同类型的数据结构进行转换。
Viper 则是采用开源的配置文件解析方案（毕竟配置文件的解析没有性能要求，只要不是慢的离谱）。
都直接用开源库了，Viper 还干了些什么？
- 用户通过代码指定相应的配置文件名称，以及若干个搜索路径
- 除了通过 map struct 的方式进行数据读取，Viper 还支持 k1.k2.k3 的方式直接读取某一个配置项
- 与环境变量、flag、默认配置整合，最终实现具有优先级的配置系统，具体优先级如下：
  - explicit call to `Set`（在代码中写死）
  - flag（临时更改某一两个配置项）
  - env（敏感配置内容，通过环境变量输入）
  - config（普通的配置文件）
  - key/value store（远端的 key-value 存储系统，比如 etcd）
  - default（代码中设置的默认值）


注意：
- 个人不推荐使用配置修改功能，除非你的配置是采用 remote K/V store 的方式，或者是 host 机部署的方案
- 个人不推荐使用配置监控并热加载

## demo
```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// configuration file information
	cfgName = "config_demo"
	cfgType = "yaml"
	cfgFullName = cfgName + "." + cfgType
	cfgPath = "." // in the working directory
)

// test case:
// 1. go run viper_demo.go
func main() {
	// set default value
	viper.SetDefault("default-value.key-2", 20)
	viper.SetDefault("default-value.key-3", 30)

	// bind env
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("PREFIX")
	_ = viper.BindEnv("env.k1") // for ${PREFIX_K2}
	_ = viper.BindEnv("env.k2") // for ${PREFIX_K2}
	_ = viper.BindEnv("env.k3") // for ${PREFIX_K3}

	fmt.Println("========> demo start")

	loadConfig()
	//loadErrorConfig()
	fmt.Printf("========> load config [%s] from [%s] done\n", cfgFullName, cfgPath)

	fmt.Println("============= defaultValue_Demo() ===============")
	defaultValue_Demo()
	fmt.Println("===============================================")

	fmt.Println("============= readConfig_Demo() ===============")
	readConfig_Demo()
	fmt.Println("===============================================")

	fmt.Println("=========== accessNestedKey_Demo() ============")
	accessNestedKey_Demo()
	fmt.Println("===============================================")

	fmt.Println("============ loadSecondCfg_Demo() =============")
	loadSecondCfg_Demo()
	fmt.Println("===============================================")

	fmt.Println("================= ENV_Demo() ==================")
	ENV_Demo()
	fmt.Println("===============================================")

	fmt.Println("=============== deepNest_Demo() ===============")
	deepNest_Demo()
	fmt.Println("===============================================")

	fmt.Println("=============== mapping_demo() ================")
	mapping_demo()
	fmt.Println("===============================================")

	fmt.Println("========> demo end")
}

func loadConfig() {
	viper.SetConfigName(cfgName) // name of config file (without extension)
	viper.SetConfigType(cfgType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(cfgPath) // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s] in path[%s]", cfgFullName, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s] in path[%s], but something error when parsing", cfgFullName, cfgPath)
		}
		panic(info)
	}
}

func loadErrorConfig() {
	errCfgName := "error" + cfgName
	viper.SetConfigName(errCfgName) // name of config file (without extension)
	viper.SetConfigType(cfgType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(cfgPath) // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s.%s] in path[%s]", errCfgName, cfgType, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s.%s] in path[%s], but something error when parsing", errCfgName, cfgType, cfgPath)
		}
		panic(info)
	}
}

func defaultValue_Demo() {
	// default value had been set before configuration file loaded
	var key string
	key = "default-value.key-1"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "default-value.key-2"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "default-value.key-3"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
}

func readConfig_Demo() {
	// case-insensitive Setting & Getting
	fmt.Printf("key-1:[%s]\n", viper.GetString("key-1"))
	fmt.Printf("not-exist-key:[%s]\n", viper.GetString("not-exist-keys"))
	fmt.Printf("CaseInsensitive:[%s]\n", viper.GetString("CaseInsensitive"))
	fmt.Printf("caseinsensitive:[%s]\n", viper.GetString("caseinsensitive"))

	fmt.Printf("key-1 IsSet [%v]\n", viper.IsSet("key-1"))
	fmt.Printf("not-exist-keys IsSet [%v]\n", viper.IsSet("not-exist-keys"))
	fmt.Printf("CaseInsensitive IsSet [%v]\n", viper.IsSet("CaseInsensitive"))
	fmt.Printf("caseinsensitive IsSet [%v]\n", viper.IsSet("caseinsensitive"))
}

func accessNestedKey_Demo() {
	// Accessing nested keys
	var key string
	key = "nest.bool-key";            fmt.Printf("%s:[%v]\n", key, viper.GetBool(key))
	key = "nest.float64-key";         fmt.Printf("%s:[%v]\n", key, viper.GetFloat64(key))
	key = "nest.int-key";             fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "nest.int-slice-key";       fmt.Printf("%s:[%v]\n", key, viper.GetIntSlice(key))
	key = "nest.int-slice-key.1";     fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	// not panic but get zero-value get by nest.int-slice-key.5
	key = "nest.int-slice-key.5";     fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "nest.string-key";          fmt.Printf("%s:[%v]\n", key, viper.GetString(key))
	key = "nest.string-slice-key";    fmt.Printf("%s:[%v]\n", key, viper.GetStringSlice(key))
	key = "nest.duration-key-3s";     fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-0.5s";   fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-30ms";   fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-3000ms"; fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
}

func loadSecondCfg_Demo() {
	newCfgName := cfgName + "_2"
	cfgParser := viper.New()
	cfgParser.SetConfigName(newCfgName) // name of config file (without extension)
	cfgParser.SetConfigType(cfgType) // REQUIRED if the config file does not have the extension in the name
	cfgParser.AddConfigPath(cfgPath) // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := cfgParser.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s] in path[%s]", newCfgName + "." + cfgType, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s] in path[%s], but something error when parsing", newCfgName + "." + cfgType, cfgPath)
		}
		panic(info)
	}

	fmt.Printf("========> load config [%s] from [%s] done\n", newCfgName + "." + cfgType, cfgPath)
	fmt.Printf("key:[%s]\n", cfgParser.GetString("key"))
	fmt.Printf("not-exist-key:[%s]\n", cfgParser.GetString("not-exist-keys"))
}

func ENV_Demo() {
	// env already bind before configuration loaded
	var key string
	key = "env.k1"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k2"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k3"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))

	fmt.Printf("========> set env\n")
	_ = os.Setenv("PREFIX_ENV_K1", "20")
	_ = os.Setenv("PREFIX_ENV_K2", "20")
	//_ = os.Setenv("PREFIX_ENV_K3", "20")

	// ENV variable will not cache in viper, but always access data from system
	key = "env.k1"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k2"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k3"; fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
}

func deepNest_Demo() {
	key := "deep-nest.k11.k21.k31"
	fmt.Printf("%s:[%v]\n", key, viper.GetString(key))
}

// NOTE: 用来 mapping 的结构体字段，一定要首字母大写
// 不 public 出去，mapstructure 是没有办法访问的！
type mapStruct struct {
	IntKey    int    `mapstructure:"int-key"`
	StringKey string `mapstructure:"string-key"`
	IntSlice  []int  `mapstructure:"int-slice"`
	Sub       sub    `mapstructure:"sub"`
}

type sub struct {
	IntKey int `mapstructure:"int-key"`
}

func mapping_demo() {
	mapping := mapStruct{}
	err := viper.UnmarshalKey("mapping", &mapping)
	if err != nil {
		fmt.Printf("error when viper.Unmarshal() %s\n", err.Error())
		return
	}

	fmt.Printf("mapStruct:\n%+v\n", mapping)
}

```



## `Viper` struct
- `Viper` struct 代表了一个配置文件。一个 `Viper` 虽然能够多个不同的路径上搜索配置文件，但是最终只能解析一个配置文件。
  - 有多个配置文件需要进行解析时，可以通过 `Viper.New()` 来多创建几个 `Viper`


```go
// Viper is a prioritized configuration registry. It
// maintains a set of configuration sources, fetches
// values to populate those, and provides them according
// to the source's priority.
// The priority of the sources is the following（优先级排序）:
// 1. overrides
// 2. flags
// 3. env. variables
// 4. config file
// 5. key/value store
// 6. defaults
//
// For example, if values from the following sources were loaded:
//
//	Defaults : {
//		"secret": "",
//		"user": "default",
//		"endpoint": "https://localhost"
//	}
//	Config : {
//		"user": "root"
//		"secret": "defaultsecret"
//	}
//	Env : {
//		"secret": "somesecretkey"
//	}
//
// The resulting config will have the following values:
//
//	{
//		"secret": "somesecretkey",
//		"user": "root",
//		"endpoint": "https://localhost"
//	}
//
// Note: Vipers are not safe for concurrent Get() and Set() operations.
// 因为你在 package 内部，是不需要用 Viper.FuncXXX() 的，所有直接把主要 struct 跟 package 同名也没有太大问题
type Viper struct {
	.......

	// A set of paths to look for the config file in
	// 可以配置多个查找路径
	configPaths []string

	// Name of file to look for inside the path
	configName        string
	// Viper 可以在多个路径中查找相应的配置文件，当找到了之后，这个 configFile 就会记录这个配置文件的绝对路径 + 全名
	configFile        string
	configType        string
	configPermissions os.FileMode
	envPrefix         string

	.......

	// 分别存储了不同数据源的数据
	// 根据 Get() 时不同的查找顺序，最终达到了数据源优先级分类的效果
	// Q&A(DONE): 为什么都是 map[string]interface{} ？
	// 为了支持 nested 类型的配置项，详细的可以参考 Viper.searchMap()，Viper.searchIndexableWithPathPrefixes()
	// 而且底层的 decoder 库也是采用 map[string]interface{} 相互嵌套的方式来处理 nested 的
	// 每一层 map 的 key，仅仅是一节 key。比如：
	// k1.k2.k3
	// m2 := m1[k1].(map[string]interface{})
	// m3 := m2[k2].(map[string]interface{})
	// value := m3[k3]
	config         map[string]interface{}
	override       map[string]interface{}
	defaults       map[string]interface{}
	kvstore        map[string]interface{}
	pflags         map[string]FlagValue
	env            map[string][]string
	aliases        map[string]string
	typeByDefValue bool

	.........

	// 开源 decoder、encoder 的工厂对象
	encoderRegistry *encoding.EncoderRegistry
	decoderRegistry *encoding.DecoderRegistry
}

// New returns an initialized Viper instance.
func New() *Viper {
	v := new(Viper)
	v.keyDelim = "."
	v.configName = "config"
	v.configPermissions = os.FileMode(0o644)
	v.fs = afero.NewOsFs()
	v.config = make(map[string]interface{})
	v.override = make(map[string]interface{})
	v.defaults = make(map[string]interface{})
	v.kvstore = make(map[string]interface{})
	v.pflags = make(map[string]FlagValue)
	v.env = make(map[string][]string)
	v.aliases = make(map[string]string)
	v.typeByDefValue = false
	v.logger = jwwLogger{}

	v.resetEncoding()

	return v
}
```



## 怎么指定配置文件？
```go
func loadConfig() {
	// step 1: 指定文件名（不带文件类型的后缀）
	viper.SetConfigName(cfgName) // name of config file (without extension)

	// step 2: 指定文件类型
	viper.SetConfigType(cfgType) // REQUIRED if the config file does not have the extension in the name

	// step 3: 指定搜索路径（不需要包含文件名及后缀）。可以指定多个路径搜索，哪个路径先搜索到，就用那个
	viper.AddConfigPath(cfgPath) // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	// step 4: 将文件从硬盘中向本进程的内存中加载
	if err := viper.ReadInConfig(); err != nil {
		.......
	}
}
```

## Viper 的配置文件解析过程
- 因为 Viper 是支持配置项相互嵌套的，所以 `Viper.config` 这些 `map[string]interface{}` 实际上是相互嵌套的。就比如下面的例子：
```go
	// 要访问配置文件中的 k1.k2.k3，那么实际访问过程是：
	m2 := m1[k1].(map[string]interface{})
	m3 := m2[k2].(map[string]interface{})
	value := m3[k3]
```

具体过程如下：
```go
// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error { return v.ReadInConfig() }

func (v *Viper) ReadInConfig() error {
	......

	// step 1: 取出配置文件的绝对路径
	filename, err := v.getConfigFile()

	......

	// step 2: 读取配置文件的内容
	file, err := afero.ReadFile(v.fs, filename)

	.....

	config := make(map[string]interface{}) // 解析后的结果放进这里

	// step 3: 调用相应的 decoder 开始解析配置文件内容
	err = v.unmarshalReader(bytes.NewReader(file), config)

	........

	v.config = config
	return nil
}

// step 4: 调用相应的 decoder
// Unmarshal a Reader into a map.
// Should probably be an unexported function.
func unmarshalReader(in io.Reader, c map[string]interface{}) error {
	return v.unmarshalReader(in, c)
}

func (v *Viper) unmarshalReader(in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer) // 可能是网络 IO，也可能是 bytes.Reader，所以先套上一层 buf 再说
	buf.ReadFrom(in)

	switch format := strings.ToLower(v.getConfigType()); format {
	case "yaml", "yml", "json", "toml", "hcl", "tfvars", "ini", "properties", "props", "prop", "dotenv", "env":
		// 通过工厂的方式区 decode 相应的键值对
		// step 5: 通过 format 在 v.decoderRegistry 工厂中，调用相应的 decoder
		err := v.decoderRegistry.Decode(format, buf.Bytes(), c)
		......
	}

	......
}

// step 6: 具体 decoder 开始 decode
// Codec implements the encoding.Encoder and encoding.Decoder interfaces for YAML encoding.
type Codec struct{}

func (Codec) Decode(b []byte, v map[string]interface{}) error {
	return yaml.Unmarshal(b, &v)
}
```
注意：
- 当上面的解析过程结束之后，相应的解析结果将会存在 `Viper.config` 里面。
- `Viper.config` 并不会按照配置数据源优先级进行排列的


## 怎么读取配置文件中的配置项？

Viper 支持两种解析方式：
- 通过 Viper 提供的 API，通过相应的 key 去读取某一个 entry 的值：`viper.GetInt("key.k1.k2.k3")`
- 通过 mapstructure 的 tag 进行 `viper.UnmarshalKey("mapping-key", &mapping)`

但无论是上面的哪一种，最终都是通过 decoder 对配置文件进行解析得来的。

另外，配置数据源的优先级，是在读取配置的时候实现的。

### 第一种：通过 API 获取配置值
**step 1: 以 "nest.int-key" 为例**
```go
viper.GetInt("nest.int-key")

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return v.GetInt(key) }

func (v *Viper) GetInt(key string) int {
	return cast.ToInt(v.Get(key))
}
```

**step 2: 根据请求的 key 去 v.config 进行搜索**
```go
func (v *Viper) Get(key string) interface{} {
	lcaseKey := strings.ToLower(key) // key 
	val := v.find(lcaseKey, true)    // 根据配置数据源的优先级进行搜索
	if val == nil {
		return nil
	}

	......
	return val
}
```

**step 3: 根据配置数据源的优先级进行搜索**
```go
// Given a key, find the value.
//
// Viper will check to see if an alias exists first.
// Viper will then check in the following order:
// 通过这个函数来控制检查的优先顺序，从而达到了配置项优先级相互覆盖的效果
// flag, env, config file, key/value store.
// Lastly, if no value was found and flagDefault is true, and if the key
// corresponds to a flag, the flag's default value is returned.
//
// Note: this assumes a lower-cased key given.
func (v *Viper) find(lcaseKey string, flagDefault bool) interface{} {
	var (
		val    interface{}
		exists bool
		path   = strings.Split(lcaseKey, v.keyDelim) // 在这个配置文件中，key-value 的 path
		nested = len(path) > 1 // 是否 nested ？
	)

	......

	// if the requested key is an alias, then return the proper key
	lcaseKey = v.realKey(lcaseKey)
	path = strings.Split(lcaseKey, v.keyDelim)
	nested = len(path) > 1

	// 优先级：Set() override first
	val = v.searchMap(v.override, path)
	if val != nil {
		return val
	}
	.......

	// 优先级：PFlag override next
	flag, exists := v.pflags[lcaseKey]
	if exists && flag.HasChanged() {
		switch flag.ValueType() {
		case "int", "int8", "int16", "int32", "int64":
			return cast.ToInt(flag.ValueString())
		case "bool":
			return cast.ToBool(flag.ValueString())
		case "stringSlice", "stringArray":
			s := strings.TrimPrefix(flag.ValueString(), "[")
			s = strings.TrimSuffix(s, "]")
			res, _ := readAsCSV(s)
			return res
		case "intSlice":
			s := strings.TrimPrefix(flag.ValueString(), "[")
			s = strings.TrimSuffix(s, "]")
			res, _ := readAsCSV(s)
			return cast.ToIntSlice(res)
		case "stringToString":
			return stringToStringConv(flag.ValueString())
		default:
			return flag.ValueString()
		}
	}
	........

	// 优先级：Env override next
	..........
	envkeys, exists := v.env[lcaseKey]
	if exists {
		for _, envkey := range envkeys {
			if val, ok := v.getEnv(envkey); ok {
				return val
			}
		}
	}
	.......

	// 优先级：Config file next
	val = v.searchIndexableWithPathPrefixes(v.config, path)
	if val != nil {
		return val
	}
	.......

	// 优先级：K/V store next
	val = v.searchMap(v.kvstore, path)
	if val != nil {
		return val
	}
	......

	// 优先级：Default next
	val = v.searchMap(v.defaults, path)
	if val != nil {
		return val
	}
	.......

	if flagDefault {
		// last chance: if no value is found and a flag does exist for the key,
		// get the flag's default value even if the flag's value has not been set.
		if flag, exists := v.pflags[lcaseKey]; exists {
			switch flag.ValueType() {
			case "int", "int8", "int16", "int32", "int64":
				return cast.ToInt(flag.ValueString())
			case "bool":
				return cast.ToBool(flag.ValueString())
			case "stringSlice", "stringArray":
				s := strings.TrimPrefix(flag.ValueString(), "[")
				s = strings.TrimSuffix(s, "]")
				res, _ := readAsCSV(s)
				return res
			case "intSlice":
				s := strings.TrimPrefix(flag.ValueString(), "[")
				s = strings.TrimSuffix(s, "]")
				res, _ := readAsCSV(s)
				return cast.ToIntSlice(res)
			case "stringToString":
				return stringToStringConv(flag.ValueString())
			default:
				return flag.ValueString()
			}
		}
		// last item, no need to check shadowing
	}

	return nil
}
```

**step 4: 从配置文件的 map 中，进行 BFS 搜索**
```go
// 之所以要 BFS，是因为 Viper 支持配置项相互嵌套
// searchIndexableWithPathPrefixes recursively searches for a value for path in source map/slice.
//
// While searchMap() considers each path element as a single map key or slice index, this
// function searches for, and prioritizes, merged path elements.
// e.g., if in the source, "foo" is defined with a sub-key "bar", and "foo.bar"
// is also defined, this latter value is returned for path ["foo", "bar"].
//
// This should be useful only at config level (other maps may not contain dots
// in their keys).
//
// Note: This assumes that the path entries and map keys are lower cased.
// 整个搜索过程其实是 BFS
// 同一个 KEY 先做 BFS(map 加速了 BFS 的过程，直接变成了 O(1))
// 然后缩短 key 继续做 BFS
// 直到 key 为 0
func (v *Viper) searchIndexableWithPathPrefixes(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], v.keyDelim))
		//fmt.Printf("%d:%s, path%v\n", i, prefixKey, path)
		// key: deep-nest.k11.k21.k31
		// 4:deep-nest.k11.k21.k31, path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11.k21.k31 对应的 map)
		// 3:deep-nest.k11.k21,     path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11.k21 对应的 map)
		// 2:deep-nest.k11,         path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11 对应的 map)
		// 1:deep-nest,             path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，成功找到 deep-nest 对应的 map)
		// 新的 searchIndexableWithPathPrefixes()
		// 3:k11.k21.k31,           path[k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 k11.k21.k31 对应的 map)
		// 2:k11.k21,               path[k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 k11.k21 对应的 map)
		// 1:k11,                   path[k11 k21 k31](同一个 searchIndexableWithPathPrefixes，成功找到 k11 对应的 map)
		// 新的 searchIndexableWithPathPrefixes()
		// 2:k21.k31,               path[k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 k21.k31 对应的 map)
		// 1:k21,                   path[k21 k31](同一个 searchIndexableWithPathPrefixes，成功找到 k21 对应的 map)
		// 新的 searchIndexableWithPathPrefixes()
		// 1:k31, path[k31]         找到了 value
		var val interface{}
		switch sourceIndexable := source.(type) {
		........
		case map[string]interface{}:
			// 尝试进行搜索
			val = v.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		}
		if val != nil {
			return val
		}
		// 搜索失败，缩短 key，再次进行搜索
	}

	// not found
	return nil
}
```

**step 5: 是否能够进行下一层 map 的搜索**
```go
// searchMapWithPathPrefixes searches for a value for path in sourceMap
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (v *Viper) searchMapWithPathPrefixes(
	sourceMap map[string]interface{},
	prefixKey string,
	pathIndex int,
	path []string,
) interface{} {
	next, ok := sourceMap[prefixKey]
	if !ok {
		// prefixKey 不存在，直接退出
		// 例如这几个 case
		// 4:deep-nest.k11.k21.k31, path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11.k21.k31 对应的 map)
		// 3:deep-nest.k11.k21,     path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11.k21 对应的 map)
		// 2:deep-nest.k11,         path[deep-nest k11 k21 k31](同一个 searchIndexableWithPathPrefixes，但是找不到 deep-nest.k11 对应的 map)
		return nil
	}
	/*
	else {
		next 肯定存在，但是不确定这究竟是叶子节点，还是中间节点
	}
	*/


	// Fast path
	if pathIndex == len(path) {
		// 叶子节点，返回 value
		// Q&A(DONE): 为什么这样能够判断是叶子节点？
		// 1. 看下面的那个 switch-case，path 总是去掉前缀再给下一层的，
		//    当 pathIndex == len(path)，其实也就没有所谓的下一层了
		// 2. 再者，同一层的进来 searchMapWithPathPrefixes() 进行判断，path 总是相同的长度
		return next
	}

	// Nested case
	// next 是中间节点，继续下一层
	// 去掉前缀，再给下一层
	switch n := next.(type) {
	........
	case map[string]interface{}, []interface{}:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}
```


### 第二种：通过 mapstructure 映射配置值
```yaml
mapping:
  int-key: 10
  string-key: string-value
  int-slice:
    - 10
    - 20
    - 30
  sub:
    int-key: 40
```

**step 1: 正确创建结构体，并打上相应的 mapstructure tag**
```go
// NOTE: 用来 mapping 的结构体字段，一定要首字母大写
// 不 public 出去，mapstructure 是没有办法访问的！
type mapStruct struct {
	IntKey    int    `mapstructure:"int-key"`
	StringKey string `mapstructure:"string-key"`
	IntSlice  []int  `mapstructure:"int-slice"`
	Sub       sub    `mapstructure:"sub"`
}

type sub struct {
	IntKey int `mapstructure:"int-key"`
}

func mapping_demo() {
	mapping := mapStruct{}
	err := viper.UnmarshalKey("mapping", &mapping)
	if err != nil {
		fmt.Printf("error when viper.Unmarshal() %s\n", err.Error())
		return
	}

	fmt.Printf("mapStruct:\n%+v\n", mapping)
}
```

**step 2: 指定 key，并传入结构体，进行 mapping**
```go
viper.UnmarshalKey("mapping", &mapping)
```

**step 3: 通过 mapstructure 进行映射**
```go
// UnmarshalKey takes a single key and unmarshals it into a Struct.
func UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	return v.UnmarshalKey(key, rawVal, opts...)
}

func (v *Viper) UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	return decode(v.Get(key), defaultDecoderConfig(rawVal, opts...))
}

// A wrapper around mapstructure.Decode that mimics the WeakDecode functionality
func decode(input interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}
```








# cobra(v1.5.0)
- **command 控制的是：使用不同的入口函数；flag 跟 configuration-file 控制的是：不同变量的取值**
  - cobra 负责解析命令，cobra 将 flag 交由 pflag 处理，cobra 处理完 flag 之前，调用 viper 加载配置文件；配置文件的解析由 viper 负责完成
- 不同的 `cobra.Command` 其实也就意味着不同的开始函数。但是 command 并不改变任何变量的值
- `pflag.Flag` 跟 `Viper.Viper` 则会改变变量的值
- cobra 实际上是 viper 跟 pflag 的融合框架，但是目前 viper 跟 pflag 并不能很好的配合（viper 的部分 API 并不会 merge pflag 跟配置文件），这个问题需要等待 viper v2 才能比较好的解决
  - 所以利用 cobra 来构建 CLI 启动的 app，现在是需要自己完成 flag 跟配置文件之间的 merge 的

## How to run demo

```yaml
app:
  name: app-demo
  mode: debug
network:
  ip: 1.1.1.1
  port: 1111
  proto: tcp
log:
  name: log-name.log
  path: ./
  level: debug

```

```go
// https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/
// https://github.com/spf13/viper/discussions/1061
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 优先级：
// 1. flag
// 2. configuration
// 3. default value in flag

// test case:
// 1. go build -o app all_demo.go && ./app
// 2. go build -o app all_demo.go && ./app server
// 3. go build -o app all_demo.go && ./app server -i 192.168.0.1
// 4. go build -o app all_demo.go && ./app server -i 192.168.0.1 -p 1234 -c ./another_config.yaml
// 5. go build -o app all_demo.go && ./app server -c ./another_config.yaml        #(display flag default ip)
func main() {
	Execute()
	fmt.Printf("ipFlag[%s], portFlag[%d], configFileFlag[%s]\n", ipFlag.String(), portFlag, configFileFlag)
}

// cobra and flag ==============================================================
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// The error can then be caught at the execute function call
		fmt.Println("Execute error")
		os.Exit(1)
	}
}

var (
	configFileFlag string
	ipFlag         net.IP
	portFlag       int
)

const (
	defaultIP   = "127.0.0.1"
	defaultPort = 80
)

// update flag from viper
var flagUpdater = make([]func(), 0)

func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	// root command
	rootCmd.PersistentFlags().StringVarP(
		&configFileFlag, "config", "c",
		defaultCfgPath+defaultCfgFullName,
		"specifing path to configuration file")

	rootCmd.AddCommand(serverCmd)

	// sub command
	serverCmd.PersistentFlags().IPVarP(
		&ipFlag, "ip", "i", net.ParseIP(defaultIP),
		"server ip to listen on")

	serverCmd.PersistentFlags().IntVarP(
		&portFlag, "port", "p", defaultPort,
		"server port to listen on")

	// bind flag to configuration
	viper.BindPFlag("network.ip", serverCmd.PersistentFlags().Lookup("ip"))
	flagUpdater = append(flagUpdater, func() {
		ipFlag = net.ParseIP(viper.GetString("network.ip"))
	})

	viper.BindPFlag("network.port", serverCmd.PersistentFlags().Lookup("port"))
	flagUpdater = append(flagUpdater, func() {
		portFlag = viper.GetInt("network.port")
	})
}

// Command.Run, Command.RunE 都不设置的话，就会自动打印 help 信息
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "short description for app",
	Long:  `long description for app`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run app as a server",
	Long:  `run app as a server to provide service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command with arg:%v\n", cmd.CommandPath(), args)
	},
}

// configuration file ==========================================================
const (
	defaultCfgName     = "all_config"
	defaultCfgType     = "yaml"
	defaultCfgFullName = defaultCfgName + "." + defaultCfgType
	defaultCfgPath     = "./"
)

func loadConfig() {
	var file string
	if configFileFlag != "" {
		// specified configuration file by CLI flag
		viper.SetConfigFile(configFileFlag)
		file = configFileFlag
	} else {
		viper.SetConfigName(defaultCfgName) // name of config file (without extension)
		viper.SetConfigType(defaultCfgType) // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(defaultCfgPath) // path to look for the config file in
		file = defaultCfgFullName
	}

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s]", file)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s], but something error when parsing", file)
		}

		panic(info)
	}

	for _, updater := range flagUpdater {
		updater()
	}
}
```




## Why *cobra* package

- 标准库里面的 `flag` 确实也能够对命令行参数进行解析，但是并不支持子命令，以及不同层级的子命令参数管理
- 另外，cobra 是专精 command 的 package，而 flag 则是参数



## 流程梳理

### 注册 command、flag、配置文件

#### `Command` tree 构建
- 在 cobra 的设计上，一个 `Command` 其实就意味着一个程序入口，所以 `Command` 都需要指定一个函数入口的
- 其实 Command tree 的构建很简单，就是将不同的 `Command` 按照层级关系 add 进去就好了
```go
func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	.......

	// root command
	rootCmd.AddCommand(serverCmd)

	// sub command
	......
}

// Command.Run, Command.RunE 都不设置的话，就会自动打印 help 信息
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "short description for app",
	Long:  `long description for app`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run app as a server",
	Long:  `run app as a server to provide service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command with arg:%v\n", cmd.CommandPath(), args)
	},
}
```

```go
// Command is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follow:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	.........

	// Short is the short description shown in the 'help' output.
	Short string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	.......

	// Command 解析完之后，运行前可以使用的 hook
	// Q&A(DONE): 下面的 hook 带 E 后缀，跟不带有什么区别？
	// 这个 hook 能不能向外传递 error
	//
	// 下面这些函数在执行之前，已经处理完 flag 了
	// The *Run functions are executed in the following order:
	//   * PersistentPreRun() 在这里我们可以挂载 viper 加载的代码
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name.
	//
	PersistentPreRun func(cmd *Command, args []string)
	PersistentPreRunE func(cmd *Command, args []string) error
	PreRun func(cmd *Command, args []string)
	PreRunE func(cmd *Command, args []string) error
	Run func(cmd *Command, args []string)
	RunE func(cmd *Command, args []string) error
	PostRun func(cmd *Command, args []string)
	PostRunE func(cmd *Command, args []string) error
	PersistentPostRun func(cmd *Command, args []string)
	PersistentPostRunE func(cmd *Command, args []string) error

	// args is actual args parsed from flags. default is os.Args[1:]
	args []string
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer

	// NOTE:
	// cobra 直接利用 pflag package 的 FlagSet API
	// 这也导致了不同 FlagSet 的直接的割裂，没办法所有 FlagSet 保持逻辑上的一致。
	// 但是，也不会有很严重的问题。因为 Command.Execute() 会调用：
	// Command.ParseFlags() 通过 Command.mergePersistentFlags() 进行了修复
	//
	// flags is full set of flags.
	flags *flag.FlagSet
	// pflags contains persistent flags.（从自己这里开始一直向下传的）
	pflags *flag.FlagSet
	// lflags contains local flags. lflags = pflags + flags, see in Command.LocalFlags()
	lflags *flag.FlagSet
	// iflags contains inherited flags.（上游传下来的）
	iflags *flag.FlagSet
	// parentsPflags is all persistent flags of cmd's parents.
	parentsPflags *flag.FlagSet
	// globNormFunc is the global normalization function
	// that we can use on every pflag set and children commands

	........

	ctx context.Context // 更多的是为了让使用者拥有通过 cobra.Command.ctx 在不同 package 之间传递信息的能力, 而不是 cobra 内部使用

	// commands is the list of commands supported by this program.
	// 当前 Command 的所有 sub-Command 都会在这里
	commands []*Command
	// parent is a parent command for this command.
	parent *Command

	.......
}

```

#### 向 `Command` attach flag


```go
var (
	configFileFlag string
	ipFlag         net.IP
	portFlag       int
)

const (
	defaultIP   = "127.0.0.1"
	defaultPort = 80
)

func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	// root command
	rootCmd.PersistentFlags().StringVarP(
		&configFileFlag, "config", "c",
		defaultCfgPath+defaultCfgFullName,
		"specifing path to configuration file")

	rootCmd.AddCommand(serverCmd)

	// sub command
	serverCmd.PersistentFlags().IPVarP(
		&ipFlag, "ip", "i", net.ParseIP(defaultIP),
		"server ip to listen on")

	serverCmd.PersistentFlags().IntVarP(
		&portFlag, "port", "p", defaultPort,
		"server port to listen on")

	......
}
```

实际上这整个过程就是利用了 pflag package 的 `pflag.FlagSet` 来实现的

```go
// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}

	return c.flags
}


// PersistentFlags returns the persistent FlagSet specifically set in the current command.
// 新建一个 FlagSet，剩下的事情，就是 pflag 增加一个 Flag 了
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		// lazy 的方式进行初始化
		c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// LocalFlags returns the local FlagSet specifically set in the current command.
func (c *Command) LocalFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	if c.lflags == nil {
		c.lflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.lflags.SetOutput(c.flagErrorBuf)
	}
	c.lflags.SortFlags = c.Flags().SortFlags
	if c.globNormFunc != nil {
		c.lflags.SetNormalizeFunc(c.globNormFunc)
	}

	// addToLocal 只能每次都执行了，不然有些 Command.flags 不会同步到 Command.lflags 的
	addToLocal := func(f *flag.Flag) {
		if c.lflags.Lookup(f.Name) == nil && c.parentsPflags.Lookup(f.Name) == nil {
			c.lflags.AddFlag(f)
		}
	}
	c.Flags().VisitAll(addToLocal)
	c.PersistentFlags().VisitAll(addToLocal)
	return c.lflags
}
```

#### cobra 控制 viper 加载时机

- 通常命令行是可以指定使用哪一个配置文件的，这也就意味着：在命令行解析后、`Command` 运行前加载配置文件是比较适合的。通常我们会利用 `cobra.OnInitialize()` 或者是 `rootCmd.PersistentPreRunE` 来加载配置文件
- 为了利用 viper 本身跟 pflag 的集成能力。但是话说，挺多 bug 的，建议自己进行 merge。这里简单利用了 `flagUpdater` 在实际运行前进行 merge

```go
var (
	configFileFlag string
	ipFlag         net.IP
	portFlag       int
)

const (
	defaultIP   = "127.0.0.1"
	defaultPort = 80
)

// update flag from viper
var flagUpdater = make([]func(), 0)

func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	// root command
	rootCmd.PersistentFlags().StringVarP(
		&configFileFlag, "config", "c",
		defaultCfgPath+defaultCfgFullName,
		"specifing path to configuration file")

	rootCmd.AddCommand(serverCmd)

	// sub command
	serverCmd.PersistentFlags().IPVarP(
		&ipFlag, "ip", "i", net.ParseIP(defaultIP),
		"server ip to listen on")

	serverCmd.PersistentFlags().IntVarP(
		&portFlag, "port", "p", defaultPort,
		"server port to listen on")

	// bind flag to configuration
	viper.BindPFlag("network.ip", serverCmd.PersistentFlags().Lookup("ip"))
	flagUpdater = append(flagUpdater, func() {
		ipFlag = net.ParseIP(viper.GetString("network.ip"))
	})

	viper.BindPFlag("network.port", serverCmd.PersistentFlags().Lookup("port"))
	flagUpdater = append(flagUpdater, func() {
		portFlag = viper.GetInt("network.port")
	})
}

// configuration file ==========================================================
const (
	defaultCfgName     = "all_config"
	defaultCfgType     = "yaml"
	defaultCfgFullName = defaultCfgName + "." + defaultCfgType
	defaultCfgPath     = "./"
)

func loadConfig() {
	var file string
	if configFileFlag != "" {
		// specified configuration file by CLI flag
		viper.SetConfigFile(configFileFlag)
		file = configFileFlag
	} else {
		viper.SetConfigName(defaultCfgName) // name of config file (without extension)
		viper.SetConfigType(defaultCfgType) // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(defaultCfgPath) // path to look for the config file in
		file = defaultCfgFullName
	}

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s]", file)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s], but something error when parsing", file)
		}

		panic(info)
	}

	for _, updater := range flagUpdater {
		updater()
	}
}
```

### cobra 开始运行

处理顺序：
1. 从 root command 开始执行
2. 解析 flag
3. 运行 hook
4. 通过 hook 让 viper 加载配置文件
5. 执行使用者注册进来的 `XxxxxRun()` 函数

```go
/* ===== step 1 =====*/
func main() {
	Execute()
	fmt.Printf("ipFlag[%s], portFlag[%d], configFileFlag[%s]\n", ipFlag.String(), portFlag, configFileFlag)
}

// cobra and flag ==============================================================
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// The error can then be caught at the execute function call
		fmt.Println("Execute error")
		os.Exit(1)
	}
}

// Execute uses the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Command) Execute() error {
	_, err := c.ExecuteC()
	return err
}

// ExecuteC executes the command.
func (c *Command) ExecuteC() (cmd *Command, err error) {

	.......

	// 自动构建 help 等命令
	// initialize help at the last point to allow for user overriding
	c.InitDefaultHelpCmd()
	// initialize completion at the last point to allow for user overriding
	c.initDefaultCompletionCmd()

	args := c.args

	// Workaround FAIL with "go test -v" or "cobra.test -test.v", see #155
	if c.args == nil && filepath.Base(os.Args[0]) != "cobra.test" {
		args = os.Args[1:] // 默认情况下加载 os.Args[1:]
	}

	......

	// We have to pass global context to children command
	// if context is present on the parent command.
	if cmd.ctx == nil {
		cmd.ctx = c.ctx
	}

	err = cmd.execute(flags) // 就开始正式执行

	......

	return cmd, err
}
```


```go
/* ===== step 2 =====*/
// 解析 Flag，运行 hook
func (c *Command) execute(a []string) (err error) {

	// 自动构建 help 等命令
	// initialize help and version flag at the last point possible to allow for user
	// overriding
	c.InitDefaultHelpFlag()
	c.InitDefaultVersionFlag()

	/* ===== step 2 =====*/
	err = c.ParseFlags(a) // 解析 Flag

	......

	if !c.Runnable() {
		// 不想被运行的 Command，可以用这种方式，来打印 help 信息
		return flag.ErrHelp
	}

	/* ===== step 3 =====*/
	c.preRun() // 跑 hook

	argWoFlags := c.Flags().Args()
	if c.DisableFlagParsing {
		argWoFlags = a
	}

	if err := c.ValidateArgs(argWoFlags); err != nil {
		return err
	}

	/* ===== step 4 =====*/
	// 在这时候，其实加载了 viper
	// PersistentPreRunE() hook, 而且因为是 Persistent 所以要检查所有的 parent Command
	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			if err := p.PersistentPreRunE(c, argWoFlags); err != nil {
				return err
			}
			break
		} else if p.PersistentPreRun != nil {
			p.PersistentPreRun(c, argWoFlags)
			break
		}
	}
	if c.PreRunE != nil {
		if err := c.PreRunE(c, argWoFlags); err != nil {
			return err
		}
	} else if c.PreRun != nil {
		c.PreRun(c, argWoFlags)
	}

	if err := c.validateRequiredFlags(); err != nil {
		return err
	}
	if err := c.validateFlagGroups(); err != nil {
		return err
	}

	/* ===== step 5 =====*/
	// 这里其实就开始运行一个 app 的 main 函数了
	// 而且很显然，不同的 Command、sub-Command 是可以设置不同的入口函数的
	if c.RunE != nil {
		if err := c.RunE(c, argWoFlags); err != nil {
			return err
		}
	} else {
		c.Run(c, argWoFlags)
	}
	if c.PostRunE != nil {
		if err := c.PostRunE(c, argWoFlags); err != nil {
			return err
		}
	} else if c.PostRun != nil {
		c.PostRun(c, argWoFlags)
	}
	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPostRunE != nil {
			if err := p.PersistentPostRunE(c, argWoFlags); err != nil {
				return err
			}
			break
		} else if p.PersistentPostRun != nil {
			p.PersistentPostRun(c, argWoFlags)
			break
		}
	}

	return nil
}


```
