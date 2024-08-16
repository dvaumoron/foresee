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

package split

import (
	"errors"
	"unicode"
)

const (
	StringKind Kind = iota
	ParenthesisKind
	SquareBracketsKind
	CurlyBracesKind
	SeparatorKind
)

var (
	errParsingClosing           = errors.New("parsing failure : wait closing separator")
	errParsingString            = errors.New("parsing failure : unended string")
	errParsingUnexpectedClosing = errors.New("parsing failure : unexpected closing separator")
	errParsingWrongClosing      = errors.New("parsing failure : wait another closing separator")
)

type Kind uint8

type Node interface {
	Cast() (Kind, string, []Node)
}

type listNode struct {
	nodes []Node
	kind  Kind
}

func (l listNode) Cast() (Kind, string, []Node) {
	return l.kind, "", l.nodes
}

type separatorNode struct{}

func (s separatorNode) Cast() (Kind, string, []Node) {
	return SeparatorKind, "", nil
}

type StringNode string

func (s StringNode) Cast() (Kind, string, []Node) {
	return StringKind, string(s), nil
}

func appendBuffer(splitted []Node, buffer []rune) ([]Node, []rune) {
	if len(buffer) != 0 {
		splitted = append(splitted, StringNode(buffer))
		buffer = buffer[:0]
	}
	return splitted, buffer
}

// append a separatorNode if the previous was not one
func appendSeparator(splitted []Node) []Node {
	if last := len(splitted) - 1; last >= 0 {
		if k, _, _ := splitted[last].Cast(); k != SeparatorKind {
			return append(splitted, separatorNode{})
		}
	}
	return splitted
}

func consumeString(nextChar func() (rune, bool), delim rune) StringNode {
	buffer := []rune{delim}
	for {
		char, ok := nextChar()
		if !ok {
			panic(errParsingString)
		}

		switch char {
		case delim:
			return StringNode(append(buffer, delim))
		case '\\':
			char2, ok := nextChar()
			if !ok {
				panic(errParsingString)
			}

			buffer = append(buffer, char, char2)
		default:
			buffer = append(buffer, char)
		}
	}
}

func SmartSplit(chars []rune) ([]Node, error) {
	nextChar := initNextChar(chars)

	var buffer []rune
	var splitted []Node
	for {
		char, ok := nextChar()
		if !ok {
			break
		}

		switch {
		case unicode.IsSpace(char):
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = appendSeparator(splitted)
		case char == '"', char == '\'':
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = append(splitted, consumeString(nextChar, char))
		case char == '(':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, ')', ParenthesisKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
		case char == '[':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, ']', SquareBracketsKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
		case char == '{':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, '}', CurlyBracesKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
		case char == ')', char == ']', char == '}':
			return nil, errParsingUnexpectedClosing
		default:
			buffer = append(buffer, char)
		}
	}

	splitted, _ = appendBuffer(splitted, buffer)
	return splitted, nil
}

func splitSub(nextChar func() (rune, bool), delim rune, kind Kind) (listNode, error) {
	var buffer []rune
	var splitted []Node
	for {
		char, ok := nextChar()
		if !ok {
			break
		}

		switch {
		case char == delim:
			splitted, _ = appendBuffer(splitted, buffer)
			return listNode{nodes: splitted, kind: kind}, nil
		case unicode.IsSpace(char):
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = appendSeparator(splitted)
		case char == '"', char == '\'':
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = append(splitted, consumeString(nextChar, char))
		case char == '(':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, ')', ParenthesisKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
		case char == '[':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, ']', SquareBracketsKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
		case char == '{':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(nextChar, '}', CurlyBracesKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
		case char == ')', char == ']', char == '}':
			return listNode{}, errParsingWrongClosing
		default:
			buffer = append(buffer, char)
		}
	}
	return listNode{}, errParsingClosing
}

func initNextChar(chars []rune) func() (rune, bool) {
	i := 0
	charLen := len(chars)
	return func() (rune, bool) {
		if i >= charLen {
			return 0, false
		}

		char := chars[i]
		i++
		return char, true
	}
}
