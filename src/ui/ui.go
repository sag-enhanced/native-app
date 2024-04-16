package ui

import "github.com/sag-enhanced/native-app/src/options"

type UII interface {
	Run()
	Eval(code string)
	SetBindHandler(handler bindHandler)
	Quit()
}

type bindHandler func(method string, callId int, params string) error

func NewUI(opt *options.Options) UII {
	if opt.UI == options.PlaywrightUI {
		return createPlaywrightUII(opt)
	} else {
		return createWebviewUII(opt)
	}
}
