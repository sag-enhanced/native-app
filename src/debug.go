package app

import (
	"fmt"
	"strings"
)

func (app *App) InstallDebugger(session string) {
	// XSS protection although probably not needed
	// would need to be self-XSS, manually entering the wrong id and remotejs
	// is giving you full access to the console anyway
	if strings.Contains(session, "script") || strings.ContainsAny(session, "<>\"'/") {
		return
	}
	js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", session)
	app.webview.Init(js)
}
