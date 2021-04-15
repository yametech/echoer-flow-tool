package common

import "flag"

const (
	DefaultNamespace = "verthandi"
	Pipeline         = "pipeline"
)

var (
	EchoerAddr = "http://127.0.0.1:8080"
)

func init() {
	flag.StringVar(&EchoerAddr, "echoer", "http://127.0.0.1:8080", "-echoer http://127.0.0.1:8080")

}
