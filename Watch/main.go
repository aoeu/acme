// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was originally taken from:
// https://github.com/rsc/rsc/blob/e1cb635eaa1bed9efc1ec21c8537c46161abfcb8/cmd/Watch/main.go
//
// This code has been modified due to the problem that this program used syscalls specific to macOS,
// as explained by bradfitz:
// 	https://groups.google.com/forum/#!msg/golang-dev/x18NTIq4g6A/zM1i76ikBAAJ
//

// Watch runs a command each time files in the current directory change.
//
// Usage:
//
//	Watch cmd [args...]
//
// Example:
//
// Watch wc -l main.go
//
// Watch opens a new acme window named for the current directory
// with a suffix of /+watch. The window shows the execution of the given
// command. Each time a file in that directory changes, Watch reexecutes
// the command and updates the window.
package main // import "github.com/aoeu/acme/Watch"

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"9fans.net/go/acme"
	"github.com/rjeczalik/notify"
)

var args []string
var win *acme.Win
var needrun = make(chan bool, 1)

// TODO(aoeu): Rename struct to something more relevant now that Kqueue syscall usage is removed.
var watchList struct {
	fd   int
	dir  *os.File
	m    map[string]*os.File
	name map[int]string
}

// watch watches changes made on file-system by the text editor
// when saving a file. Usually, either InCloseWrite or InMovedTo (when swapping
// with a temporary file) event is created, so only those events are monitored.
// This function is lifted from https://github.com/rjeczalik/notify/blob/master/example_inotify_test.go#L20
func watch(path string) {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening for inotify-specific events within a
	// current working directory. Dispatch each InCloseWrite and InMovedTo
	// events separately to c.
	//
	// func Watch(path string, c chan<- EventInfo, events ...Event) error
	//
	if err := notify.Watch(path, c, notify.InCloseWrite, notify.InMovedTo); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		// Block until an event is received.
		switch ei := <-c; ei.Event() {
		case notify.InCloseWrite:
			// log.Println("Editing of", ei.Path(), "file is done.")
			needrun <- true
		case notify.InMovedTo:
			// log.Println("File", ei.Path(), "was swapped/moved into the watched directory.")
			needrun <- true
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: Watch cmd args...\n")
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args = flag.Args()
	if len(args) == 0 {
		usage()
	}

	var err error
	win, err = acme.New()
	if err != nil {
		log.Fatal(err)
	}
	pwd, _ := os.Getwd()
	win.Name(pwd + "/+watch")
	win.Ctl("clean")
	win.Fprintf("tag", "Get ")
	needrun <- true
	go events()
	go runner()

	watchList.fd = 777
	watchList.m = make(map[string]*os.File)
	watchList.name = make(map[int]string)

	dir, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	watchList.dir = dir
	watch(".")
	readdir := true

	for {
		if readdir {
			watchList.dir.Seek(0, 0)
			names, err := watchList.dir.Readdirnames(-1)
			if err != nil {
				log.Fatalf("readdir: %v", err)
			}
			for _, name := range names {
				if watchList.m[name] != nil {
					continue
				}
				f, err := os.Open(name)
				if err != nil {
					continue
				}
				watchList.m[name] = f
				fd := int(f.Fd())
				watchList.name[fd] = name
				watch(name)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func events() {
	for e := range win.EventChan() {
		switch e.C2 {
		case 'x', 'X': // execute
			if string(e.Text) == "Get" {
				select {
				case needrun <- true:
				default:
				}
				continue
			}
			if string(e.Text) == "Del" {
				win.Ctl("delete")
			}
		}
		win.WriteEvent(e)
	}
	os.Exit(0)
}

var run struct {
	sync.Mutex
	id int
}

func runner() {
	var lastcmd *exec.Cmd
	for _ = range needrun {
		run.Lock()
		run.id++
		id := run.id
		run.Unlock()
		if lastcmd != nil {
			lastcmd.Process.Kill()
		}
		lastcmd = nil
		cmd := exec.Command(args[0], args[1:]...)
		r, w, err := os.Pipe()
		if err != nil {
			log.Fatal(err)
		}
		win.Addr(",")
		win.Write("data", nil)
		win.Ctl("clean")
		win.Fprintf("body", "$ %s\n", strings.Join(args, " "))
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Start(); err != nil {
			r.Close()
			w.Close()
			win.Fprintf("body", "%s: %s\n", strings.Join(args, " "), err)
			continue
		}
		lastcmd = cmd
		w.Close()
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := r.Read(buf)
				if err != nil {
					break
				}
				run.Lock()
				if id == run.id {
					win.Write("body", buf[:n])
				}
				run.Unlock()
			}
			if err := cmd.Wait(); err != nil {
				run.Lock()
				if id == run.id {
					win.Fprintf("body", "%s: %s\n", strings.Join(args, " "), err)
				}
				run.Unlock()
			}
			win.Fprintf("body", "$\n")
			win.Fprintf("addr", "#0")
			win.Ctl("dot=addr")
			win.Ctl("show")
			win.Ctl("clean")
		}()
	}
}
