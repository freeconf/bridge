package prombridge

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/freeconf/gconf/device"
	"github.com/freeconf/gconf/meta"
	"github.com/freeconf/gconf/node"
	"github.com/freeconf/gconf/nodes"
	"github.com/freeconf/gconf/restconf"
	"github.com/freeconf/gconf/val"
)

type Bridge struct {
	options       Options
	device        device.Device
	RenderMetrics RenderMetrics
	localServer   *http.Server
	Modules       Modules
}

type Modules struct {
	Ignore []string
}

func NewBridge(d device.Device) *Bridge {
	return &Bridge{
		device: d,
	}
}

type RenderMetrics struct {
	Duration time.Duration
	Count    int64
}

func (b *Bridge) Apply(options Options) error {
	if options.Port == "" {
		bwsr, err := b.device.Browser("restconf")
		if err != nil {
			return err
		}
		if b == nil {
			return errors.New("no internal browser found and port not configured")
		}
		server, valid := bwsr.Root().Peek(nil).(*restconf.Server)
		if !valid {
			return errors.New("expected to find *restconf.Server when peeking at path 'restconf'")
		}
		var existingHandler = server.UnhandledRequestHandler
		server.UnhandledRequestHandler = func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/metrics" {
				b.ServeHTTP(w, r)
			} else if existingHandler != nil {
				existingHandler(w, r)
			}
		}
	} else {
		if b.localServer != nil {
			b.localServer.Close()
		}
		demux := http.NewServeMux()
		demux.Handle("/metrics", b)
		b.localServer = &http.Server{
			Handler: demux,
			Addr:    options.Port,
		}
		go func() {
			if err := b.localServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("trouble with server server %s", err)
			}
		}()
	}
	b.options = options
	return nil
}

type Options struct {
	Port           string // ":2112"
	UseLocalServer bool
}

func (b *Bridge) Options() Options {
	return b.options
}

func (b *Bridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.generate(w)
}

func (b *Bridge) generate(out io.Writer) error {
	e := newExporter(out)
	t0 := time.Now()

Modules:
	for name := range b.device.Modules() {
		for _, ignore := range b.Modules.Ignore {
			if ignore == name {
				continue Modules
			}
		}
		bwsr, err := b.device.Browser(name)
		if err != nil {
			return err
		}
		sel := bwsr.Root().Constrain("content=nonconfig").InsertInto(e.node(cleanName(name)))
		if sel.LastErr != nil {
			return sel.LastErr
		}
	}
	e.out.Flush()
	b.RenderMetrics = RenderMetrics{
		Duration: time.Now().Sub(t0),
		Count:    e.count,
	}
	fmt.Printf("m = %v\n", b.RenderMetrics)
	return nil
}

type exporter struct {
	out   *bufio.Writer
	count int64
}

var metricExposition *template.Template

func init() {
	metricExposition = template.Must(template.New("metric").Parse(`# HELP {{.Name}} {{.Desc}}
# TYPE {{.Name}} {{.Type}}
{{.Name}} {{.Value}}
`))
}

func newExporter(out io.Writer) *exporter {
	return &exporter{
		out: bufio.NewWriter(out),
	}
}

var invalidChars = regexp.MustCompile("[-]")

func cleanName(ident string) string {
	return invalidChars.ReplaceAllString(ident, "_")
}

func (e *exporter) node(prefix string) node.Node {
	return &nodes.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			id := fmt.Sprintf("%s_%s", prefix, cleanName(r.Meta.Ident()))
			promType := "gauge"
			if ext := r.Meta.Extensions().Get("metric"); ext != nil {
				promType = ext.Arguments()[0]
			}
			vars := struct {
				Desc  string
				Type  string
				Name  string
				Value interface{}
			}{
				Desc:  strings.TrimSpace(r.Meta.(meta.Describable).Description()),
				Type:  promType,
				Name:  id,
				Value: hnd.Val.Value(),
			}
			e.count++
			if err := metricExposition.Execute(e.out, vars); err != nil {
				return err
			}
			return nil
		},
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if !r.New {
				return nil, nil
			}
			id := fmt.Sprintf("%s_%s", prefix, cleanName(r.Meta.Ident()))
			return e.node(id), nil
		},
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			if !r.New {
				return nil, nil, nil
			}
			id := fmt.Sprintf("%s_%d", prefix, r.Row)
			return e.node(id), nil, nil
		},
	}
}

func (e *exporter) close() error {
	return e.out.Flush()
}
