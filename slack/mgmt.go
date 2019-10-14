package slack

import (
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/val"
)

func Manage(c *Client) node.Node {
	options := c.Options()
	return &nodeutil.Extend{
		Base: nodeutil.ReflectChild(&options),
		OnChild: func(p node.Node, r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "sub":
				if r.New {
					c.subs = make(map[string]*Subscription)
				}
				if len(c.subs) > 0 || r.New {
					return manageSubs(c.subs), nil
				}
			default:
				return p.Child(r)
			}
			return nil, nil
		},
		OnNotify: func(p node.Node, r node.NotifyRequest) (node.NotifyCloser, error) {
			switch r.Meta.Ident() {
			case "err":
				sub := c.OnError(func(err error, sub *Subscription) {
					msg := map[string]interface{}{
						"subId": sub.Id,
						"msg":   err.Error(),
					}
					r.Send(nodeutil.ReflectChild(msg))
				})
				return sub.Close, nil
			}
			return nil, nil
		},
		OnEndEdit: func(p node.Node, r node.NodeRequest) error {
			return c.Apply(options)
		},
	}
}

func manageSubs(subs map[string]*Subscription) node.Node {
	index := node.NewIndex(subs)
	return &nodeutil.Basic{
		Peekable: subs,
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			var sub *Subscription
			key := r.Key
			if key != nil {
				id := key[0].String()
				if r.New {
					sub = &Subscription{Id: id}
					subs[sub.Id] = sub
				} else {
					sub = subs[id]
				}
			} else if r.Row < index.Len() {
				if v := index.NextKey(r.Row); v != node.NO_VALUE {
					id := v.String()
					sub = subs[id]
					var err error
					key, err = node.NewValues(r.Meta.KeyMeta(), id)
					if err != nil {
						return nil, nil, err
					}
				}
			}
			if sub != nil {
				return manageSub(sub), key, nil
			}
			return nil, nil, nil
		},
	}
}

func manageSub(s *Subscription) node.Node {
	return &nodeutil.Extend{
		Base: nodeutil.ReflectChild(s),
		OnField: func(p node.Node, r node.FieldRequest, hnd *node.ValueHandle) error {
			switch r.Meta.Ident() {
			case "active":
				hnd.Val = val.Bool(s.Active())
			default:
				return p.Field(r, hnd)
			}
			return nil
		},
	}
}
