package slack

import (
	"flag"
	"strings"
	"testing"

	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/source"

	"github.com/freeconf/yang/fc"

	"github.com/freeconf/restconf/device"
)

var update = flag.Bool("update", false, "update gold files, do not compare with them")

func TestMgmt(t *testing.T) {
	ypath := source.Dir("../yang")
	d := device.New(ypath)
	c := NewClient(d)
	if err := d.Add("slack-client", Manage(c)); err != nil {
		t.Error(err)
	}
	cfg := `{
		"slack-client" : {
			"sub" : [{
				"id" : "x",
				"path" : "err",
				"module" : "slack-client"
			}]
		}
	}`
	if err := d.ApplyStartupConfig(strings.NewReader(cfg)); err != nil {
		t.Error(err)
	}
	b, err := d.Browser("slack-client")
	if err != nil {
		t.Error(err)
	}
	fc.AssertEqual(t, 1, len(c.subs))
	actual, err := nodeutil.WritePrettyJSON(b.Root())
	if err != nil {
		t.Error(err)
	}
	fc.Gold(t, *update, []byte(actual), "gold/mgmt.json")
}
