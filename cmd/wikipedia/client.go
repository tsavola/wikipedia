// Copyright (c) 2023 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"

	"import.name/pan"

	. "import.name/pan/mustcheck"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s article\n", os.Args[0])
		os.Exit(2)
	}
	article := os.Args[1]
	if len(article) > 255 {
		fmt.Fprintf(os.Stderr, "%s: article name is too long\n", os.Args[0])
		os.Exit(2)
	}

	var fail bool

	err := pan.Recover(func() {
		conn := Must(net.Dial("tcp", "localhost:11314"))
		defer conn.Close()

		Must(conn.Write(append([]byte{uint8(len(article))}, article...)))
		buf := Must(io.ReadAll(conn))

		out := os.Stdout
		if bytes.HasPrefix(buf, []byte("Error: ")) {
			out = os.Stderr
			fail = true
		}
		fmt.Fprintf(out, "%s\n", buf)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	if fail {
		os.Exit(1)
	}
}
