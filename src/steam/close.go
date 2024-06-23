package steam

import (
	"fmt"
	"time"

	"github.com/sag-enhanced/native-app/src/options"
	"github.com/shirou/gopsutil/v3/process"
)

func CloseSteam(options *options.Options) error {
	killed := int32(0)
	for {
		var proc *process.Process
		var err error
		if proc, err = findSteamProcess(); err != nil {
			break
		}
		if proc.Pid != killed {
			// new process found (this can happen if we close steam while its still bootstrapping)
			if options.Verbose {
				fmt.Println("Steam running, shutting it down...")
			}

			RunSteamWithArguments(options, "-shutdown")
			killed = proc.Pid
		}
		if options.Verbose {
			fmt.Println("Waiting for Steam to shut down...", proc.Pid)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
