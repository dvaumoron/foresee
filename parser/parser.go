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

var (
	errIndent = errors.New("identation not consistent")
	errNode   = errors.New("unhandled node")
)

type splitResult struct {
	nodes []split.Node
	err   error
}

func Parse(str string) (*types.List, error) {
	splitRes := splitIndentToSyntax(str)
	if splitRes.err != nil {
		return nil, splitRes.err
	}

	res := types.NewList(names.FileId)
	if err := processNodes(splitRes.nodes, res); err != nil {
		return nil, err
	}
	return res, nil
}

func processNodes(nodes []split.Node, list *types.List) error {
	for i, last := 0, len(nodes); i < last; {
		switch object, consumed := handleSlice(nodes[i:]); consumed {
		case -1: // separator marker
			i += 1
		case 0:
			return errNode
		default:
			list.Add(object)
			i += consumed
		}
	}
	return nil
}

func sendClosingParenthesis(chars chan<- rune) {
	chars <- ')'
}

func sendNothing(_ chan<- rune) {
}

func splitIndentToSyntax(str string) splitResult {
	chars, resChan := make(chan rune), make(chan splitResult)
	go func() {
		nodes, err := split.SmartSplit(chars)
		for range chars { // emptying non consumed
		}

		resChan <- splitResult{nodes: nodes, err: err}
	}()

	closePreviousLine := sendNothing
	indentStack := stack.New[int]()
	indentStack.Push(0)
	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index := 0
			var char rune
			for index, char = range line {
				if !unicode.IsSpace(char) {
					if top := indentStack.Peek(); top < index {
						indentStack.Push(index)
					} else {
						closePreviousLine(chars)
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
			closePreviousLine = sendClosingParenthesis
		}
	}
	for range indentStack.Size() {
		chars <- ')'
	}
	close(chars)

	return <-resChan
}
