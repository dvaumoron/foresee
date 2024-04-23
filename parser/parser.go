/*
 *
 * Copyright 2023 foresee authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, v. 2.0.
 *
 */

package parser

import (
	"errors"
	"strings"
	"unicode"

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/parser/split"
	"github.com/dvaumoron/foresee/parser/stack"
	"github.com/dvaumoron/foresee/types"
)

var errIndent = errors.New("identation not consistent")

type splitResult struct {
	nodes []split.Node
	err   error
}

func Parse(str string) (*types.List, error) {
	splitRes := splitIndentToSyntax(str)
	if splitRes.err != nil {
		return nil, splitRes.err
	}

	last := len(splitRes.nodes)
	res := types.NewList(names.FileId)
	for i := 0; i < last; {
		object, consumed := handleSlice(splitRes.nodes[i:min(i+2, last)])
		res.Add(object)
		i += consumed
	}

	return res, nil
}

func handleSlice(nodes []split.Node) (types.Object, int) {
	// TODO
	return types.None, 1
}

func splitIndentToSyntax(str string) splitResult {
	indentStack := stack.New[int]()
	indentStack.Push(0)

	chars := make(chan rune)
	resChan := make(chan splitResult)
	go func() {
		nodes, err := split.SmartSplit(chars)
		resChan <- splitResult{nodes: nodes, err: err}
	}()

	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index := 0
			var char rune
			for index, char = range line {
				if !unicode.IsSpace(char) {
					if top := indentStack.Peek(); top < index {
						indentStack.Push(index)
					} else {
						chars <- ')'
						if top > index {
							indentStack.Pop()
							for top = indentStack.Peek(); top > index; top = indentStack.Peek() {
								chars <- ')'
								indentStack.Pop()
							}
							if top < index {
								return splitResult{err: errIndent}
							}
							chars <- ')'
						}
					}
					chars <- '('
					break
				}
			}

			for _, char := range line[index:] {
				if char == '#' {
					break
				}
				chars <- char
			}
		}
	}
	close(chars)

	return <-resChan
}
