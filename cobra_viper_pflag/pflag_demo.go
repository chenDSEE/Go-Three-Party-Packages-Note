package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"time"
)

func main() {
	fmt.Printf("======> app start\n")

	//pflagParse_Demo()
	//shorthandFlag_Demo()
	//CLIFlagSyntax_Demo()
	//BoolCLIFlagSyntax_Demo()
	//multiFlagValue_Demo()
	//count_Demo()
	//time_Demo()
	//network_Demo()
	//slice_Demo()
	bytes_Demo()

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
	pflag.BoolVarP(&a,"bool-a", "a", false, "bool type flag for a")
	pflag.BoolVarP(&b,"bool-b", "b", false, "bool type flag for b")
	pflag.BoolVarP(&c,"bool-c", "c", false, "bool type flag for c")
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

	t := pflag.DurationP("time", "t", 1 * time.Second, "time in second")
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