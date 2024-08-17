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
	"iter"
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
	// errParsingClosing           = errors.New("parsing failure : wait closing separator")
	// errParsingString            = errors.New("parsing failure : unended string")
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

func yieldBuffer(yield func(Node) bool, buffer []rune) ([]rune, bool) {
	if len(buffer) == 0 {
		return buffer, true
	}

	ok := yield(StringNode(buffer))
	return buffer[:0], ok
}

func yieldSeparator(yield func(Node) bool) bool {
	return yield(separatorNode{})
}

func yieldNothing(yield func(Node) bool) bool {
	return true
}

func consumeString(yieldChar *func(rune) bool, delim rune, yield func(Node) bool) {
	previousYieldChar := *yieldChar

	var directAppender func(char rune) bool

	buffer := []rune{delim}
	stoppableAppender := func(char rune) bool {
		switch char {
		case delim:
			*yieldChar = previousYieldChar
			return yield(StringNode(append(buffer, delim)))
		case '\\':
			buffer = append(buffer, char)
			*yieldChar = directAppender
		default:
			buffer = append(buffer, char)
		}
		return true
	}

	directAppender = func(char rune) bool {
		buffer = append(buffer, char)
		*yieldChar = stoppableAppender
		return true
	}

	*yieldChar = stoppableAppender
}

func SmartSplit(chars iter.Seq[rune], registerError func(error)) iter.Seq[Node] {
	var buffer []rune
	yielder, ok := yieldSeparator, true
	return func(yield func(Node) bool) {
		var yieldChar func(char rune) bool
		highYieldChar := func(char rune) bool {
			switch {
			case unicode.IsSpace(char):
				buffer, ok = yieldBuffer(yield, buffer)
				if !ok {
					return false
				}

				if !yielder(yield) {
					return false
				}
				yielder = yieldNothing
			case char == '"', char == '\'':
				buffer, ok = yieldBuffer(yield, buffer)
				if !ok {
					return false
				}

				consumeString(&yieldChar, char, yield)
				yielder = yieldSeparator
			case char == '(':
				buffer, ok = yieldBuffer(yield, buffer)
				if !ok {
					return false
				}

				splitSub(&yieldChar, ')', ParenthesisKind, yield, registerError)
				yielder = yieldSeparator
			case char == '[':
				buffer, ok = yieldBuffer(yield, buffer)
				if !ok {
					return false
				}

				splitSub(&yieldChar, ']', SquareBracketsKind, yield, registerError)
				yielder = yieldSeparator
			case char == '{':
				buffer, ok = yieldBuffer(yield, buffer)
				if !ok {
					return false
				}

				splitSub(&yieldChar, '}', CurlyBracesKind, yield, registerError)
				yielder = yieldSeparator
			case char == ')', char == ']', char == '}':
				registerError(errParsingUnexpectedClosing)
				return false
			default:
				buffer = append(buffer, char)
				yielder = yieldSeparator
			}
			return true
		}

		yieldChar = highYieldChar
		for char := range chars {
			if !yieldChar(char) {
				return
			}
		}

		yieldBuffer(yield, buffer)
	}
}

func splitSub(yieldChar *func(rune) bool, delim rune, kind Kind, yield func(Node) bool, registerError func(error)) {
	previousYieldChar := *yieldChar

	var splitted []Node
	localYield := func(node Node) bool {
		splitted = append(splitted, node)
		return true
	}

	var buffer []rune
	yielder := yieldSeparator
	*yieldChar = func(char rune) bool {
		switch {
		case char == delim:
			yieldBuffer(localYield, buffer)
			*yieldChar = previousYieldChar
			return yield(listNode{nodes: splitted, kind: kind})
		case unicode.IsSpace(char):
			buffer, _ = yieldBuffer(localYield, buffer)
			yielder(localYield)
			yielder = yieldNothing
		case char == '"', char == '\'':
			buffer, _ = yieldBuffer(localYield, buffer)
			consumeString(yieldChar, char, localYield)
			yielder = yieldSeparator
		case char == '(':
			buffer, _ = yieldBuffer(localYield, buffer)
			splitSub(yieldChar, ')', ParenthesisKind, localYield, registerError)
			yielder = yieldSeparator
		case char == '[':
			buffer, _ = yieldBuffer(localYield, buffer)
			splitSub(yieldChar, ']', SquareBracketsKind, localYield, registerError)
			yielder = yieldSeparator
		case char == '{':
			buffer, _ = yieldBuffer(localYield, buffer)
			splitSub(yieldChar, '}', CurlyBracesKind, localYield, registerError)
			yielder = yieldSeparator
		case char == ')', char == ']', char == '}':
			registerError(errParsingWrongClosing)
			return false
		default:
			buffer = append(buffer, char)
		}
		return true
	}
}
