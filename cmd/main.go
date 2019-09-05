package main

import (
	"flag"

	"github.com/freeconf/gconf/meta/yang"

	"github.com/freeconf/bridge/slack"
	"github.com/freeconf/bridge/prombridge"
	"github.com/freeconf/gconf/c2"
	"github.com/freeconf/gconf/device"
	"github.com/freeconf/gconf/restconf"
)

var startup = flag.String("startup", "startup.json", "start-up configuration file.")
var verbose = flag.Bool("verbose", false, "verbose")

func main() {
	flag.Parse()
	c2.DebugLog(*verbose)

	yangPath := yang.YangPath()
	d := device.New(yangPath)

	c := slack.NewClient(d)
	chkErr(d.Add("slack-client", slack.Manage(c)))

	b := prombridge.NewBridge(d)
	chkErr(d.Add("prom-bridge", prombridge.Manage(b)))

	restconf.NewServer(d)
	chkErr(d.ApplyStartupConfigFile(*startup))

	// wait for cntrl-c...
	select {}
}

func chkErr(err error) {
	if err != nil {
		panic(err)
	}
}
