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

func appendSeparator(splitted []Node) []Node {
	return append(splitted, separatorNode{})
}

func appendNothing(splitted []Node) []Node {
	return splitted
}

func consumeString(chars <-chan rune, delim rune) StringNode {
	var buffer []rune
	for char := range chars {
		switch char {
		case delim:
			return StringNode(buffer)
		case '\\':
			buffer = append(buffer, char, <-chars)
		default:
			buffer = append(buffer, char)
		}
	}
	panic(errParsingString)
}

func SmartSplit(chars <-chan rune) ([]Node, error) {
	var buffer []rune
	var splitted []Node
	appender := appendSeparator
	for char := range chars {
		switch {
		case unicode.IsSpace(char):
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = appender(splitted)
			appender = appendNothing
		case char == '"', char == '\'':
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = append(splitted, consumeString(chars, char))
			appender = appendSeparator
		case char == '(':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, ')', ParenthesisKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == '[':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, ']', SquareBracketsKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == '{':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, '}', CurlyBracesKind)
			if err != nil {
				return nil, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == ')', char == ']', char == '}':
			return nil, errParsingUnexpectedClosing
		default:
			buffer = append(buffer, char)
			appender = appendSeparator
		}
	}

	splitted, _ = appendBuffer(splitted, buffer)
	return splitted, nil
}

func splitSub(chars <-chan rune, delim rune, kind Kind) (listNode, error) {
	var buffer []rune
	var splitted []Node
	appender := appendSeparator
	for char := range chars {
		switch {
		case char == delim:
			splitted, _ = appendBuffer(splitted, buffer)
			return listNode{nodes: splitted, kind: kind}, nil
		case unicode.IsSpace(char):
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = appender(splitted)
			appender = appendNothing
		case char == '"', char == '\'':
			splitted, buffer = appendBuffer(splitted, buffer)
			splitted = append(splitted, consumeString(chars, char))
			appender = appendSeparator
		case char == '(':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, ')', ParenthesisKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == '[':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, ']', SquareBracketsKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == '{':
			splitted, buffer = appendBuffer(splitted, buffer)
			sub, err := splitSub(chars, '}', CurlyBracesKind)
			if err != nil {
				return listNode{}, err
			}
			splitted = append(splitted, sub)
			appender = appendSeparator
		case char == ')', char == ']', char == '}':
			return listNode{}, errParsingWrongClosing
		default:
			buffer = append(buffer, char)
			appender = appendSeparator
		}
	}
	return listNode{}, errParsingClosing
}
