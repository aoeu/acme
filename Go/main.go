// Copyright (c) 2016 David R. Jenni. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Go is a wrapper around several Go tools for use inside of Acme.

Usage:
	Go <cmd>

The following commands are available:
	Go addtags <tags>		# Adds the given tags to the selected struct fields.
	Go cle <scope>			# Shows possible targets of the function call under the cursor.
	Go clr <scope>			# Shows possible callers of the function under the cursor.
	Go cs <scope>			# Shows the path from the callgraph root to the function under the cursor.
	Go def					# Shows the declaration for the identifier under the cursor.
	Go desc					# Describes the declaration for the syntax under the cursor.
	Go doc					# Shows the documentation for the entity under the cursor.
	Go err <scope>			# Shows possible values of the error variable under the cursor.
	Go ex <name>			# Extracts the selected statements to a new function/method with name <name>.
	Go fstruct				# Fills a struct literal with default values.
	Go fswitch				# Fills a (type) switch statement with case statements.
	Go fv					# Shows the free variables of the selected snippet.
	Go impl <recv> <iface>	# Generates method stubs with receiver <recv> for implementing the interface <iface> and inserts them at the location of the cursor.
	Go impls <scope>		# Shows the `implements` relation for the type or method under the cursor.
	Go peers <scope>		# Shows send/receive corresponding to the selected channel op.
	Go pto <scope>			# Shows variables the selected pointer may point to.
	Go refs					# Shows all refs to the entity denoted by identifier under the cursor.
	Go rmtags <tags>		# Removes the given tags from the selected struct fields.
	Go rn <name>			# Renames the entity under the cursor with <name>.
	Go share				# Uploads the selected snippet to play.golang.org and prints the URL.
	Go what					# Shows basic information about the selected syntax node.

<scope> is a comma-separated list of packages the analysis should be limited to, this parameter is optional.

The following tools are used:
	golang.org/x/tools/cmd/gorename
	golang.org/x/tools/cmd/guru
	github.com/godoctor/godoctor
	github.com/zmb3/gogetdoc
	github.com/josharian/impl
	github.com/fatih/gomodifytags
	github.com/davidrjenni/reftools/cmd/fillstruct
	github.com/davidrjenni/reftools/cmd/fillswitch
*/
package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
)

const usage = `Usage: Go <cmd>

Commands:
	addtags	adds tags to the selected struct fields
	cle		shows possible targets of the selected function call
	clr		shows possible callers of the selected function
	cs		shows the path from the callgraph root to the selected function
	def		shows declaration of selected identifier
	desc	describes the selected syntax: definition, methods, etc.
	doc		shows documentation for items in Go source code
	err		shows possible values of the selected error variable
	ex		extracts statements to a new function/method
	fstruct	fills a struct literal with default values
	fv		shows declaration of selected identifier
	impl	generate method stubs for implementing an interface
	impls	shows the 'implements' relation for the selected type or method
	peers	shows send/receive corresponding to selected channel op
	pto		shows variables the selected pointer may point to
	rmtags	removes tags from the selected struct fields
	rn		renames the selected identifier
	refs	shows all refs to the entity denoted by selected identifier
	share	uploads the selected code to play.golang.org
	what	shows basic information about the selected syntax node
`

var cmds = map[string]func(selection, []string){
	"addtags": addTags,
	"cle":     callees,
	"clr":     callers,
	"cs":      callstack,
	"def":     definition,
	"desc":    describe,
	"doc":     godoc,
	"err":     whicherrs,
	"ex":      extract,
	"fstruct": fillstruct,
	"fswitch": fillswitch,
	"fv":      freevars,
	"impl":    impl,
	"impls":   implements,
	"peers":   peers,
	"pto":     pointsto,
	"refs":    referrers,
	"rmtags":  rmTags,
	"rn":      rename,
	"share":   share,
	"what":    what,
}

func main() {
	log.SetPrefix("")
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal(usage)
	}

	s, err := readSelection()
	if err != nil {
		log.Fatalf("cannot read selection: %v\n", err)
	}

	f, ok := cmds[os.Args[1]]
	if !ok {
		log.Fatal(usage)
	}
	f(s, os.Args[2:])
}

func run(cmd string, args ...string) string {
	return runWithStdin(nil, cmd, args...)
}

func runWithStdin(stdin io.Reader, cmd string, args ...string) string {
	var buf bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = &buf
	c.Stdin = stdin
	if err := c.Run(); err != nil {
		log.Fatalf("%s failed: %v\n", cmd, err)
	}
	return buf.String()
}
