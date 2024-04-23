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
	"github.com/dvaumoron/foresee/parser/stack"
	"github.com/dvaumoron/foresee/types"
)

var errIndent = errors.New("identation not consistent")

func Parse(str string) (*types.List, error) {
	tokens, err := splitIndentToSyntax(str)
	if err != nil {
		return nil, err
	}

	listStack := stack.New[*types.List]()
	res := types.NewList(names.FileId)
	listStack.Push(res)
	manageOpen(listStack)
	for _, token := range tokens {
		handleToken(token, listStack)
	}

	return res, err
}

func handleToken(token string, listStack *stack.Stack[*types.List]) {
	switch token {
	case "(":
		manageOpen(listStack)
	case ")":
		listStack.Pop()
	default:
		listStack.Peek().Add(handleWord(token))
	}
}

func manageOpen(listStack *stack.Stack[*types.List]) {
	current := types.NewList()
	listStack.Peek().Add(current)
	listStack.Push(current)
}

func sendChar(chars chan<- rune, line string) {
	for _, char := range line {
		chars <- char
	}
	close(chars)
}

func splitIndentToSyntax(str string) ([]string, error) {
	indentStack := stack.New[int]()
	indentStack.Push(0)

	var splitted []string
	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index := 0
			var char rune
			for index, char = range line {
				if !unicode.IsSpace(char) {
					if top := indentStack.Peek(); top < index {
						indentStack.Push(index)
					} else {
						splitted = append(splitted, ")")
						if top > index {
							indentStack.Pop()
							for top = indentStack.Peek(); top > index; top = indentStack.Peek() {
								splitted = append(splitted, ")")
								indentStack.Pop()
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
