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
	"github.com/dvaumoron/foresee/types"
)

var errIndent = errors.New("identation not consistent")

type stack[T any] struct {
	inner []T
}

func (s *stack[T]) push(e T) {
	s.inner = append(s.inner, e)
}

func (s *stack[T]) peek() T {
	return s.inner[len(s.inner)-1]
}

func (s *stack[T]) pop() T {
	last := len(s.inner) - 1
	res := s.inner[last]
	s.inner = s.inner[:last]
	return res
}

func newStack[T any]() *stack[T] {
	return &stack[T]{}
}

func Parse(str string) (*types.List, error) {
	tokens, err := splitIndentToSyntax(str)
	if err != nil {
		return nil, err
	}

	listStack := newStack[*types.List]()
	res := types.NewList(names.FileId)
	listStack.push(res)
	manageOpen(listStack)
	for _, token := range tokens {
		handleToken(token, listStack)
	}

	return res, err
}

func handleToken(token string, listStack *stack[*types.List]) {
	switch token {
	case "(":
		manageOpen(listStack)
	case ")":
		listStack.pop()
	default:
		listStack.peek().Add(handleWord(token))
	}
}

func manageOpen(listStack *stack[*types.List]) {
	current := types.NewList()
	listStack.peek().Add(current)
	listStack.push(current)
}

func sendChar(chars chan<- rune, line string) {
	for _, char := range line {
		chars <- char
	}
	close(chars)
}

func splitIndentToSyntax(str string) ([]string, error) {
	indentStack := newStack[int]()
	indentStack.push(0)

	var splitted []string
	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index := 0
			var char rune
			for index, char = range line {
				if !unicode.IsSpace(char) {
					if top := indentStack.peek(); top < index {
						indentStack.push(index)
					} else {
						splitted = append(splitted, ")")
						if top > index {
							indentStack.pop()
							for top = indentStack.peek(); top > index; top = indentStack.peek() {
								splitted = append(splitted, ")")
								indentStack.pop()
							}
							if top < index {
								return nil, errIndent
							}
							splitted = append(splitted, ")")
						}
					}
					splitted = append(splitted, "(")
					break
				}
			}

			tokens, err := splitTokens(line[index:])
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, tokens...)
		}
	}
	return splitted, nil
}

func splitTokens(line string) ([]string, error) {
	chars := make(chan rune)
	go sendChar(chars, line)

	var err error
	var buffer []rune
	var splitted []string
	for char := range chars {
		switch {
		case unicode.IsSpace(char):
			splitted, buffer = appendBuffer(splitted, buffer)
		case char == '(', char == ')':
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = append(splitted, string(char))
		case char == '"', char == '\'':
			buffer, err = readUntil(buffer, chars, char)
			if err != nil {
				return nil, err
			}
		case char == '[':
			buffer, err = readSub(buffer, chars, '[', ']')
			if err != nil {
				return nil, err
			}
		case char == '{':
			buffer, err = readSub(buffer, chars, '{', '}')
			if err != nil {
				return nil, err
			}
		case char == '<':
			subBuffer, err := readSub(nil, chars, '<', '>')
			if err != nil {
				if err == errParsingParent { // to handle "<", "<=" and "<-"
					subBuffer[0] = '|' // replace '<' (avoid infinite recursion)
					subTokens, err := splitTokens(string(subBuffer))
					if err != nil {
						return nil, err
					}

					buffer = append(buffer, '<')               // add back '<'
					for _, subChar := range subTokens[0][1:] { // skip the placeholder
						buffer = append(buffer, subChar)
					}

					splitted, buffer = appendBuffer(splitted, buffer)
					splitted = append(splitted, subTokens[1:]...)

					continue
				}

				return nil, err
			}
			buffer = append(buffer, subBuffer...)
		case char == ']', char == '}':
			return nil, errParsingWrongClosing
		case char == '#':
			for range chars { // read the rest of the line (avoid go routine leak)
			}
		default:
			buffer = append(buffer, char)
		}
	}
	splitted, _ = appendBuffer(splitted, buffer)
	return splitted, nil
}
