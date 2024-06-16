package ui

import (
	"fmt"

	"github.com/sag-enhanced/native-app/src/options"
)

func getScripts(options *options.Options) []string {
	scripts := []string{}

	// arbitrary redirect protection
	origin := options.GetRealmOrigin()
	js := fmt.Sprintf("if([%q, 'id.sage.party'].indexOf(location.origin)===-1)location.href=%q", origin, origin)
	scripts = append(scripts, js)

	// expose current URL
	js = fmt.Sprintf("window.saged=window.saged||[];window.sage('setUrl',window.saged.push([()=>{},console.error.bind(console)]),JSON.stringify([location.href, %q]))", options.CurrentUrlSecret)
	scripts = append(scripts, js)

	// inject remotejs agent
	// we inject it here because we cant trust the main js to work (why else would we be debugging it?)
	if options.RemotejsSession != "" {
		js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", options.RemotejsSession)
		scripts = append(scripts, js)
	}

	return scripts
}
