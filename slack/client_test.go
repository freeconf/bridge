package slack

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/source"
)

func TestClient(t *testing.T) {
	mstr := `
		module m {
			revision 0;
			notification n {
				leaf x {
					type string;
				}
			}
		}`
	send := make(chan string)
	d := device.New(source.Named("m", strings.NewReader(mstr)))
	n := &nodeutil.Basic{
		OnNotify: func(r node.NotifyRequest) (node.NotifyCloser, error) {
			go func() {
				r.Send(nodeutil.ReadJSON(<-send))
			}()
			nop := func() error {
				return nil
			}
			return nop, nil
		},
	}
	if err := d.Add("m", n); err != nil {
		t.Fatal(err)
	}
	c := NewClient(d)
	c.OnError(func(err error, s *Subscription) {
		t.Fatal(err)
	})

	t.Run("emulator", func(t *testing.T) {
		e := NewEmulator()
		c.conn = e
		err := c.AddSubscription(&Subscription{
			Module:  "m",
			Channel: "c",
			Path:    "n",
			Enable:  true,
		})
		if err != nil {
			t.Fatal(err)
		}
		msg := `{"x":"hi"}`
		go func() {
			send <- msg
		}()
		actual := <-e.msgs
		fc.AssertEqual(t, "c", actual.Channel)
		fc.AssertEqual(t, msg, actual.Text)
	})

	t.Run("real", func(t *testing.T) {
		token := os.Getenv("FC_SLACK_TOKEN")
		if token == "" {
			t.Skipped()
			return
		}
		o := c.Options()
		o.Debug = true
		o.ApiToken = token
		if err := c.Apply(o); err != nil {
			t.Fatal(err)
		}
		err := c.AddSubscription(&Subscription{
			Module:  "m",
			Channel: "#api",
			Path:    "n",
			Enable:  true,
		})
		if err != nil {
			t.Fatal(err)
		}
		msg := `{"x":"hi"}`
		send <- msg
		<-time.After(1 * time.Second)
	})
}
