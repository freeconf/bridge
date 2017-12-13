package slack

import "testing"
import "github.com/freeconf/c2g/device"
import "github.com/freeconf/c2g/meta"
import "github.com/freeconf/c2g/nodes"
import "github.com/freeconf/c2g/node"
import "github.com/freeconf/c2g/c2"

func TestClient(t *testing.T) {
	mstr := func(r string) (string, error) {
		return `module m {
			revision 0;
			notification n {
				leaf x {
					type string;
				}
			}
		}
		`, nil
	}
	send := make(chan string)
	d := device.New(&meta.StringSource{Streamer: mstr})
	n := &nodes.Basic{
		OnNotify: func(r node.NotifyRequest) (node.NotifyCloser, error) {
			go func() {
				r.Send(nodes.ReadJSON(<-send))
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
	c2.AssertEqual(t, "c", actual.Channel)
	c2.AssertEqual(t, msg, actual.Text)
}
