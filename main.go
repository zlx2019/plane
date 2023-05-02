/**
  @author: Zero
  @date: 2023/4/30 14:25:03
  @desc: 程序入口

**/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 配置文件路径,通过命令行指定,默认值为当前目录的config.json文件
var filePath string

// Main
func main() {
	// 解析命令行参数
	flag.Parse()
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGKILL)
	// 读取配置文件
	file, err := os.Open(filePath)
	if err != nil {
		exit("open config file err: " + err.Error())
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		exit("read config file err: " + err.Error())
	}
	var configs []Config
	err = json.Unmarshal(bytes, &configs)
	if err != nil {
		exit("parse config file err: " + err.Error())
	}
	//TODO 业务处理
	go doRun(configs)
	<-stopChan
	// 关闭程序
	log.Println("Server Stop Bye~")
}

// 程序错误退出
func exit(message string) {
	log.Println(message)
	os.Exit(0)
}

// 初始化,读取命令行参数,获取配置文件位置
func init() {
	// go run -f ./xxx/xxx/config.json
	flag.StringVar(&filePath, "f", "./config.json", "config file")
}

// Config 配置实体,用于映射config.json配置文件
type Config struct {
	// 要代理的端口
	ListenerPort int16 `json:"listener_port"`
	// 代理要访问的具体服务
	Forward []string `json:"forward"`
}

// 运行服务
func doRun(configs []Config) {
	//将配置文件内所有配置的代理服务运行起来
	for _, config := range configs {
		go StartProxy(config)
	}
}

// StartProxy 开启代理服务
func StartProxy(config Config) {
	// 当前服务已处理请求次数
	reqCount := 0
	// 取模,获取本次要负载均衡的服务索引
	Index := reqCount % len(config.Forward)
	// 获取本次要访问的目标服务
	forward := config.Forward[Index]

	// 开启代理服务
	proxyAddress := fmt.Sprintf("0.0.0.0:%v", config.ListenerPort)
	listener, err := net.Listen("tcp", proxyAddress)
	if err != nil {
		log.Printf("Proxy Start Port On %v Err:\n %s\n", config.ListenerPort, err.Error())
		return
	}
	defer listener.Close()
	log.Println("Start Proxy On ", listener.Addr().String())
	// 处理每一次连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// 连接处理,并且传入本次负载均衡要访问的目标服务地址
		go connProxy(conn, forward)
	}
}

// 客户端与真实的服务端代理实现
// 主要交换两端的请求数据与响应数据
func connProxy(conn net.Conn, forward string) {
	defer conn.Close()
	log.Printf("loadbalancer %s \n", forward)
	// 与真实要访问的tcp服务建立连接
	proxyClient, err := net.Dial("tcp", forward)
	// 建立连接失败
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyClient.Close()
	// 设置客户端读取数据超时时间(连接有效时间5s),超过后断开连接(防止HTTP协议卡死)
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	// TODO 将客户端和目标服务端的通信数据进行交换
	// 将客户端的请求数据流拷贝到目标服务端
	go io.Copy(conn, proxyClient)
	// 再将目标服务端响应的数据流拷贝给客户端
	io.Copy(proxyClient, conn)
	defer log.Printf("%s 断开连接~", conn.RemoteAddr().String())
}
