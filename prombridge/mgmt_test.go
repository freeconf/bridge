package prombridge

import (
	"bytes"
	"testing"

	"github.com/freeconf/gconf/device"
	"github.com/freeconf/gconf/meta"
)

func TestBridgeMgmt(t *testing.T) {
	ypath := &meta.FileStreamSource{Root: "../yang"}
	d := device.New(ypath)
	b := NewBridge(d)
	if err := d.Add("prom-bridge", Manage(b)); err != nil {
		t.Fatal(err)
	}
	var actual bytes.Buffer
	if err := b.generate(&actual); err != nil {
		t.Fatal(err)
	}
	t.Log(actual.String())
}
