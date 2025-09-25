package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var cfp string

func init() {
	flag.StringVar(&cfp, "f", "./config.yml", "config file")
}

// TCP local forward
func main() {
	flag.Parse()
	if err := cmd(); err != nil {
		slog.Error("Startup error: " + err.Error())
		os.Exit(1)
	}
}

func cmd() error {
	file, err := os.Open(cfp)
	if err != nil {
		return fmt.Errorf("open config file error: %w", err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read config error: %w", err)
	}
	var cfs Configs
	err = yaml.Unmarshal(bytes, &cfs)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}
	// finish signal
	finish := make(chan os.Signal)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go startAllListener(cfs)
	<-finish
	slog.Info("shutdown done")
	return nil
}

func startAllListener(configs Configs) {
	for _, config := range configs {
		go startForward(config)
	}
}

func startForward(config ListenConfig) {
	bindLocalAddr := fmt.Sprintf(":%d", config.ListenerPort)
	listener, err := net.Listen("tcp", bindLocalAddr)
	if err != nil {
		slog.Error("Listen local port on " + bindLocalAddr + ", error: " + err.Error())
		return
	}
	slog.Info("Start listener on: " + bindLocalAddr)
	defer listener.Close()
	reqs := 0
	idx := reqs % len(config.Forwards)
	forward := config.Forwards[idx]
	for {
		conn, err := listener.Accept()
		if err != nil {
			if err == io.EOF {
				return
			}
			slog.Error("Accept local port on " + bindLocalAddr + ", error: " + err.Error())
			continue
		}
		go handeConnection(conn, forward)
	}
}

func handeConnection(conn net.Conn, forward string) {
	defer conn.Close()
	target, err := net.Dial("tcp", forward)
	if err != nil {
		slog.Error("Dial target on " + forward + ", error: " + err.Error())
		return
	}
	defer target.Close()
	slog.Info(fmt.Sprintf("%s >>> %s", conn.RemoteAddr(), target.RemoteAddr()))
	// 设置超时时间（如果有需要）
	// conn.SetReadDeadline(time.Now().Add(time.Minute * 3))
	go io.Copy(conn, target)
	io.Copy(target, conn)
	slog.Info(fmt.Sprintf("%v <--> %v Connection closed", conn.RemoteAddr(), target.RemoteAddr()))
}
