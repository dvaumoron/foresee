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

func Parse(str string) (*types.List, error) {
	nodes, err := splitIndentToSyntax(str)
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

func appendClosingParenthesis(chars []rune) []rune {
	return append(chars, ')')
}

func appendNothing(chars []rune) []rune {
	return chars
}

func splitIndentToSyntax(str string) ([]split.Node, error) {
	chars := []rune{}
	closePreviousLine := appendNothing
	indentStack := stack.New[int]()
	indentStack.Push(0)
	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index, char := 0, rune(0)
			for index, char = range line {
				switch char {
				case ' ':
					continue
				case '\t':
					return nil, errTab
				}

				if top := indentStack.Peek(); top < index {
					indentStack.Push(index)
				} else {
					chars = closePreviousLine(chars)
					if top > index {
						indentStack.Pop()
						for top = indentStack.Peek(); top > index; top = indentStack.Peek() {
							chars = append(chars, ')')
							indentStack.Pop()
						}
						if top < index {
							return nil, errIndent
						}
						chars = append(chars, ')')
					}
				}
				chars = append(chars, '(')
				break

			}

			for _, char := range line[index:] {
				if char == '#' {
					break
				}
				chars = append(chars, char)
			}
			closePreviousLine = appendClosingParenthesis
		}
	}
	for range indentStack.Size() {
		chars = append(chars, ')')
	}

	return split.SmartSplit(chars)
}
