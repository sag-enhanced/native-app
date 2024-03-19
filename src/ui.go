package app

type UII interface {
	run()
	eval(code string)
	quit()
}
