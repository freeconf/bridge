package slack

import "testing"
import "github.com/freeconf/gconf/device"
import "github.com/freeconf/gconf/meta"
import "github.com/freeconf/gconf/nodes"
import "github.com/freeconf/gconf/node"
import "github.com/freeconf/gconf/c2"
import "time"
import "os"

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
		c2.AssertEqual(t, "c", actual.Channel)
		c2.AssertEqual(t, msg, actual.Text)
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
