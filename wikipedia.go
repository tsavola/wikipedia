// Copyright (c) 2023 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wikipedia

import (
	"compress/bzip2"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	wikiparse "github.com/dustin/go-wikiparse"
)

const indexMapCapacity = 23456789

type indexOffset struct {
	streamOffset int64
	pageID       int
}

var ErrNoArticle = errors.New("article not found")

type MultistreamDump struct {
	index   map[string]indexOffset
	content io.ReaderAt
}

func NewMultistreamDump(index io.Reader, content io.ReaderAt) (*MultistreamDump, error) {
	r := wikiparse.NewIndexReader(bzip2.NewReader(index))
	m := make(map[string]indexOffset, indexMapCapacity)

	for {
		e, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		m[e.ArticleName] = indexOffset{e.StreamOffset, e.PageOffset}
	}

	if len(m) > indexMapCapacity {
		fmt.Fprintf(os.Stderr, "wikipedia: index size %d exceeds initial allocation capacity %d\n", len(m), indexMapCapacity)
	}

	return &MultistreamDump{m, content}, nil
}

func (dump *MultistreamDump) ReadArticle(name string) (string, error) {
	entry, found := dump.index[name]
	if !found {
		return "", fmt.Errorf("%w: %s", ErrNoArticle, name)
	}

	decoder := xml.NewDecoder(bzip2.NewReader(io.NewSectionReader(dump.content, entry.streamOffset, math.MaxInt64-entry.streamOffset)))

	if _, err := decoder.Token(); err != nil {
		return "", err
	}

	var page wikiparse.Page

	for {
		if err := decoder.Decode(&page); err != nil {
			return "", err
		}

		if page.ID == uint64(entry.pageID) {
			return page.Revisions[0].Text, nil
		}

		page = wikiparse.Page{}
	}
}
