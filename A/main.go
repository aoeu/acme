package main // import "github.com/aoeu/acme/A"
import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"9fans.net/go/acme"
)

var usageTemplate = `usage: {{.}} [ files ]

{{.}} is run from a command line interface or script to edit the
specified files from the acme text editor, waiting until each 
file has been written to disk and closed.

If an existing instance of acme is already running, new windows
will be opened for any of the files specified to {{.}} that are not already
not open in acme. For each file, {{.}} will wait until each relevant window 
is closed before cleanly exiting.

example:  {{.}} /tmp/out.txt /tmp/foo.txt

git config --global core.editor {{.}}
<make changes to a git repostiory>
git commit
<edit commit message in acme, Put and Del window>

`

func usage() {
	var t *template.Template
	var err error
	if t, err = template.New("usage").Parse(usageTemplate); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := t.Execute(os.Stdout, os.Args[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	flag.PrintDefaults()
	os.Exit(2)
}

var existingWindows map[string]int

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		usage()
	}
	flag.Parse()
	var err error
	existingWindows, err = windowIDsByName()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not determine existing, open acme windows: %v\n", err)
		os.Exit(1)
	}
	wg := &sync.WaitGroup{}
	for _, a := range flag.Args() {
		f, err := filepath.Abs(a)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not determine absolute path of %v : %v\n", a, err)
			os.Exit(1)
		}
		w, err := open(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open window in acme for %v : %v\n", f, err)
			os.Exit(1)
		}
		wg.Add(1)
		go writeEvents(w, wg)
	}
	wg.Wait()
}

func windowIDsByName() (map[string]int, error) {
	m := make(map[string]int, 0)
	w, err := acme.Windows()
	if err != nil {
		return m, err
	}
	for _, wi := range w {
		m[wi.Name] = wi.ID
	}
	return m, nil
}

func open(filepath string) (w *acme.Win, err error) {
	if id, ok := existingWindows[filepath]; ok {
		if w, err = acme.Open(id, nil); err != nil {
			return w, err
		}
	} else if w, err = acme.New(); err == nil {
		w.Name(filepath)
		w.Write("tag", []byte("Put "))
		w.Ctl("get")
	}
	return w, err
}

func writeEvents(w *acme.Win, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case e, ok := <-w.EventChan():
			if !ok {
				return
			}
			w.WriteEvent(e)
			if string(e.Text) == "Del" {
				return
			}
		}

	}
}
