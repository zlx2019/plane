package main

type Configs []ListenConfig
type ListenConfig struct {
	ListenerPort int      `yaml:"listenerPort"`
	Forwards     []string `yaml:"forwards"`
}
