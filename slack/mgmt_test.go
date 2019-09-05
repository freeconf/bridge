package slack

import (
	"flag"
	"strings"
	"testing"

	"github.com/freeconf/gconf/nodes"

	"github.com/freeconf/gconf/c2"

	"github.com/freeconf/gconf/device"
	"github.com/freeconf/gconf/meta"
)

var update = flag.Bool("update", false, "update gold files, do not compare with them")

func TestMgmt(t *testing.T) {
	ypath := &meta.FileStreamSource{Root: "../yang"}
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
	c2.AssertEqual(t, 1, len(c.subs))
	actual, err := nodes.WritePrettyJSON(b.Root())
	if err != nil {
		t.Error(err)
	}
	c2.Gold(t, *update, []byte(actual), "gold/mgmt.json")
}
