// Copyright (c) 2023 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
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
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s article\n", os.Args[0])
		os.Exit(2)
	}
	article := os.Args[1]

	err := pan.Recover(func() {
		content := Must(os.Open(contentFilename))
		defer content.Close()

		index := Must(os.Open(indexFilename))
		defer index.Close()

		dump := Must(wikipedia.NewMultistreamDump(index, content))
		text := Must(dump.ReadArticle(article))
		fmt.Println(text)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}
