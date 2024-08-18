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
	"bufio"
	"errors"
	"io"
	"iter"
	"slices"
	"strings"

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/parser/split"
	"github.com/dvaumoron/foresee/parser/stack"
	"github.com/dvaumoron/foresee/types"
)

var (
	errIndent = errors.New("identation not consistent")
	errNode   = errors.New("unhandled node")
	errTab    = errors.New("tabulation not allowed in indentation")
)

func Parse(reader io.Reader) (*types.List, error) {
	var err error
	nodes := slices.Collect(splitIndentToSyntax(reader, func(innerErr error) {
		err = innerErr
	}))
	if err != nil {
		return nil, err
	}

	res := types.NewList(names.FileId)
	if err := processNodes(nodes, res); err != nil {
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

func yieldClosingParenthesis(yield func(rune) bool) bool {
	return yield(')')
}

func yieldNothing(yield func(rune) bool) bool {
	return true
}

func indentToSyntax(reader io.Reader, registerError func(error)) iter.Seq[rune] {
	closePreviousLine := yieldNothing
	indentStack := stack.New[int]()
	indentStack.Push(0)

	scanner := bufio.NewScanner(reader)
	return func(yield func(rune) bool) {
		for scanner.Scan() {
			line := scanner.Text()
			if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
				index, char := 0, rune(0)
				for index, char = range line {
					switch char {
					case ' ':
						continue
					case '\t':
						registerError(errTab)
						return
					}

					if top := indentStack.Peek(); top < index {
						indentStack.Push(index)
					} else {
						if !closePreviousLine(yield) {
							return
						}
						if top > index {
							indentStack.Pop()
							for top = indentStack.Peek(); top > index; top = indentStack.Peek() {
								if !yield(')') {
									return
								}
								indentStack.Pop()
							}
							if top < index {
								registerError(errIndent)
								return
							}
							if !yield(')') {
								return
							}
						}
					}
					if !yield('(') {
						return
					}
					break
				}

				for _, char := range line[index:] {
					if char == '#' {
						break
					}
					if !yield(char) {
						return
					}
				}
				closePreviousLine = yieldClosingParenthesis
			}
		}

		if err := scanner.Err(); err != nil {
			registerError(err)
			return
		}

		for range indentStack.Size() {
			if !yield(')') {
				return
			}
		}
	}
}

func splitIndentToSyntax(reader io.Reader, registerError func(error)) iter.Seq[split.Node] {
	return split.SmartSplit(indentToSyntax(reader, registerError), registerError)
}
