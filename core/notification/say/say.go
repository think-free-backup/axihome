package say

import "github.com/think-free/jsonrpc/common/tools"

type SayNotification struct {
}

func New() *SayNotification {

	say := &SayNotification{}
	return say
}

func (say *SayNotification) Send(device string, message string) {

	jsontools.GetUrl("http://" + device + ":3333/say/" + message)
}
