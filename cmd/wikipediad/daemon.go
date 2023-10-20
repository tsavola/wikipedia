// Copyright (c) 2023 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/tsavola/wikipedia"
	"import.name/pan"

	. "import.name/pan/mustcheck"
)

const (
	filenamePrefix  = "/home/user/Downloads/enwiki-20231001-pages-articles-multistream"
	indexFilename   = filenamePrefix + "-index.txt.bz2"
	contentFilename = filenamePrefix + ".xml.bz2"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		os.Exit(2)
	}

	err := pan.Recover(func() {
		content := Must(os.Open(contentFilename))
		defer content.Close()

		index := Must(os.Open(indexFilename))
		defer index.Close()

		listener := Must(net.Listen("tcp", "localhost:11314"))
		defer listener.Close()

		dump := Must(wikipedia.NewMultistreamDump(index, content))

		for {
			go handle(dump, Must(listener.Accept()))
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func handle(dump *wikipedia.MultistreamDump, conn net.Conn) {
	defer conn.Close()

	err := pan.Recover(func() {
		buf := make([]uint8, 1, 256)
		Must(io.ReadFull(conn, buf))
		buf = buf[:int(buf[0])]
		Must(io.ReadFull(conn, buf))

		text, err := dump.ReadArticle(string(buf))
		if err != nil {
			text = fmt.Sprintf("Error: %v", err)
		}
		Must(conn.Write([]byte(text)))
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", conn.RemoteAddr(), err)
	}
}
