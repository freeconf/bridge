package slack

import (
	"container/list"

	"github.com/freeconf/gconf/c2"
	"github.com/freeconf/gconf/device"
	"github.com/freeconf/gconf/node"
	"github.com/freeconf/gconf/nodes"
	nslack "github.com/nlopes/slack"
)

type Client struct {
	dev          device.Device
	options      Options
	subs         map[string]*Subscription
	errListeners *list.List
	conn         Connection
}

func NewClient(d device.Device) *Client {
	return &Client{
		dev:          d,
		errListeners: list.New(),
		subs:         make(map[string]*Subscription),
	}
}

func (c *Client) OnError(l ErrListener) c2.Subscription {
	return c2.NewSubscription(c.errListeners, c.errListeners.PushBack(l))
}

type Msg struct {
	Channel string
	Text    string
}

type Connection interface {
	Send(msg Msg) error
}

type Slack struct {
	api *nslack.Client
}

func NewSlack(c *nslack.Client) *Slack {
	return &Slack{api: c}
}

type Emulator struct {
	msgs chan Msg
}

func NewEmulator() *Emulator {
	return &Emulator{
		msgs: make(chan Msg),
	}
}

func (e *Emulator) Send(msg Msg) error {
	e.msgs <- msg
	return nil
}

func (s *Slack) Send(msg Msg) error {
	_, _, err := s.api.PostMessage(msg.Channel, msg.Text, nslack.PostMessageParameters{})
	return err
}

type ErrListener func(err error, r *Subscription)

type Subscription struct {
	Id      string
	Enable  bool
	Channel string
	Device  string
	Module  string
	Path    string
	Counter uint32
	sub     node.NotifyCloser
}

func (s *Subscription) Active() bool {
	return s.sub != nil
}

type Options struct {
	ApiToken  string
	UserToken string
	Debug     bool
	Emulate   bool
}

func (b *Client) Options() Options {
	return b.options
}

func (b *Client) AddSubscription(r *Subscription) error {
	b.subs[r.Id] = r
	return b.updateSubscription(r)
}

func (b *Client) updateSubscription(r *Subscription) error {
	if r.sub != nil {
		// don't think we care if unsubscribe doesn't work
		r.sub()
	}
	if b.conn == nil || !r.Enable {
		return nil
	}
	bwsr, err := b.dev.Browser(r.Module)
	if err != nil {
		return err
	}
	sel := bwsr.Root().Find(r.Path)
	if sel.IsNil() {
		return c2.NewErr(r.Path + " not found in module " + r.Module)
	}
	r.sub, err = sel.Notifications(b.stream(r))
	if err != nil {
		return err
	}
	return nil
}

func (b *Client) onErr(r *Subscription, err error) {
	for p := b.errListeners.Front(); p != nil; p = p.Next() {
		p.Value.(ErrListener)(err, r)
	}
}

func (b *Client) stream(r *Subscription) node.NotifyStream {
	return func(msg node.Selection) {
		txt, err := nodes.WriteJSON(msg)
		if err != nil {
			b.onErr(r, err)
			return
		}
		err = b.conn.Send(Msg{
			Channel: r.Channel,
			Text:    txt,
		})
		if err != nil {
			b.onErr(r, err)
			return
		}
		r.Counter++
	}
}

func (b *Client) updateSubscriptions() {
	for _, r := range b.subs {
		b.updateSubscription(r)
	}
}

func (b *Client) Apply(options Options) error {
	b.options = options
	if options.Emulate {
		b.conn = NewEmulator()
	} else {
		api := nslack.New(b.options.ApiToken)
		api.SetDebug(b.options.Debug)
		b.conn = NewSlack(api)
	}

	b.updateSubscriptions()
	return nil
}
