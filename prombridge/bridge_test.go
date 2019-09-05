package prombridge

import (
	"bytes"
	"flag"
	"testing"

	"github.com/freeconf/gconf/c2"

	"github.com/freeconf/gconf/meta/yang"
	"github.com/freeconf/gconf/node"
	"github.com/freeconf/gconf/nodes"
	"github.com/freeconf/gconf/val"
)

var updateFlag = flag.Bool("update", false, "update golden files instead of verifying against them")

const testModule = `module x {
	prefix "x";
	namespace "x";
	revision 0000-00-00;

	extension metric {
		argument "type";
	}
	
	leaf c {
		description "int32 counter";
		type int32;
		config false;
		x:metric "counter";
	}

	leaf g {
		description "int32 gauge";
		type int32;
		config false;
	}

	container y {
		config false;

		leaf z { 
			description "float gauge";
			type decimal64;
		}
	}

	leaf m {
		description "should not show as it is a configurable";
		type int32;
	}

	list f {
		config false;

		leaf g {			
			type int32;
		}
	}
}`

func TestBridge(t *testing.T) {
	m := yang.RequireModuleFromString(nil, testModule)
	n := &nodes.Basic{}
	n.OnField = func(r node.FieldRequest, hnd *node.ValueHandle) error {
		hnd.Val = val.Int32(99)
		return nil
	}
	n.OnChild = func(r node.ChildRequest) (node.Node, error) {
		return n, nil
	}
	n.OnNext = func(r node.ListRequest) (node.Node, []val.Value, error) {
		if r.Row >= 2 {
			return nil, nil, nil
		}
		return n, nil, nil
	}
	bwsr := node.NewBrowser(m, n)
	var actual bytes.Buffer
	x := newExporter(&actual)

	if err := bwsr.Root().Constrain("content=nonconfig").InsertInto(x.node("x")).LastErr; err != nil {
		t.Fatal(err)
	}
	x.close()
	c2.Gold(t, *updateFlag, actual.Bytes(), "./gold/bridge1.txt")
}

func TestClean(t *testing.T) {
	c2.AssertEqual(t, "prom_bridge", cleanName("prom-bridge"))
}
