package prombridge

import (
	"time"

	"github.com/freeconf/gconf/node"
	"github.com/freeconf/gconf/nodes"
	"github.com/freeconf/gconf/val"
)

func Manage(b *Bridge) node.Node {
	return &nodes.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "service":
				return serviceNode(b), nil
			case "modules":
				return nodes.ReflectChild(&b.Modules), nil
			case "render":
				return renderMetrics(b.RenderMetrics), nil
			}
			return nil, nil
		},
	}
}

func renderMetrics(m RenderMetrics) node.Node {
	return &nodes.Extend{
		Base: nodes.ReflectChild(&m),
		OnField: func(p node.Node, r node.FieldRequest, hnd *node.ValueHandle) error {
			switch r.Meta.Ident() {
			case "duration":
				hnd.Val = val.Int64(m.Duration / time.Millisecond)
			default:
				return p.Field(r, hnd)
			}
			return nil
		},
	}
}

func serviceNode(b *Bridge) node.Node {
	options := b.Options()
	return &nodes.Extend{
		Base: nodes.ReflectChild(&options),
		OnEndEdit: func(p node.Node, r node.NodeRequest) error {
			if err := p.EndEdit(r); err != nil {
				return err
			}
			return b.Apply(options)
		},
	}
}
