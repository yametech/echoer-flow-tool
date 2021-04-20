package common

import "flag"

const (
	DefaultNamespace  = "verthandi"
	Pipeline          = "pipeline"
	Stage             = "stage"
	Step              = "step"
	DefaultServerName = "pipeline"
)

var (
	EchoerAddr = "http://10.200.65.192:8080"
)

func init() {
	flag.StringVar(&EchoerAddr, "echoer", "http://127.0.0.1:8080", "-echoer=http://10.200.65.192:8080")
}
