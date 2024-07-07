package ui

import (
	"fmt"

	"github.com/sag-enhanced/native-app/src/options"
)

func getScripts(options *options.Options) []string {
	scripts := []string{}

	// arbitrary redirect protection
	origin := options.GetRealmOrigin()
	js := fmt.Sprintf("if([%q,'https://id.sage.party'].indexOf(location.origin)===-1)location.href=%q", origin, origin)
	scripts = append(scripts, js)

	// expose current URL
	js = fmt.Sprintf("window.saged=window.saged||[];window.sage('setUrl',window.saged.push([()=>{},console.error.bind(console)]),JSON.stringify([location.href, %q]))", options.CurrentUrlSecret)
	scripts = append(scripts, js)

	return scripts
}
