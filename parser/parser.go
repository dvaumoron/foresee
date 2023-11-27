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

var (
	errIndent  = errors.New("identation not consistent")
	errUnended = errors.New("unended string")
)

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
	indentStack := newStack[int]()
	indentStack.push(0)
	listStack := newStack[*types.List]()
	res := types.NewList(names.FileId)
	listStack.push(res)
	manageOpen(listStack)
	var err error
LineLoop:
	for _, line := range strings.Split(str, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed[0] != '#' {
			index := 0
			var char rune
			for index, char = range line {
				if !unicode.IsSpace(char) {
					if top := indentStack.peek(); top < index {
						indentStack.push(index)
						manageOpen(listStack)
					} else if top == index {
						listStack.pop()
						manageOpen(listStack)
					} else {
						indentStack.pop()
						listStack.pop()
						for top = indentStack.peek(); top > index; top = indentStack.peek() {
							indentStack.pop()
							listStack.pop()
						}
						if top < index {
							err = errIndent
							break LineLoop
						}
						listStack.pop()
						manageOpen(listStack)
					}
					break
				}
			}
			words := make(chan string)
			done := make(chan types.NoneType)
			go handleWord(words, listStack, done)
			chars := make(chan rune)
			go sendChar(chars, line[index:])
			var buildingWord []rune
			for char := range chars {
				switch {
				case unicode.IsSpace(char):
					buildingWord = sendReset(words, buildingWord)
				case char == '(', char == ')':
					buildingWord = sendReset(words, buildingWord)
					words <- string(char)
				case char == '"', char == '\'':
					buildingWord, err = readUntil(buildingWord, chars, char)
					if err != nil {
						break LineLoop
					}
				case char == '#':
					finishLine(words, buildingWord, done)
					continue LineLoop
				default:
					buildingWord = append(buildingWord, char)
				}
			}
			finishLine(words, buildingWord, done)
		}
	}
	return res, err
}

func manageOpen(listStack *stack[*types.List]) {
	current := types.NewList()
	listStack.peek().Add(current)
	listStack.push(current)
}

func handleWord(words <-chan string, listStack *stack[*types.List], done chan<- types.NoneType) {
	for word := range words {
		switch word {
		case "(":
			manageOpen(listStack)
		case ")":
			listStack.pop()
		default:
			listStack.peek().Add(HandleClassicWord(word))
		}
	}
	done <- types.None
}

func sendChar(chars chan<- rune, line string) {
	for _, char := range line {
		chars <- char
	}
	close(chars)
}

func sendReset(words chan<- string, buildingWord []rune) []rune {
	if len(buildingWord) != 0 {
		words <- string(buildingWord)
		// doesn't realloc memmory
		buildingWord = buildingWord[:0]
	}
	return buildingWord
}

func readUntil(buildingWord []rune, chars <-chan rune, delim rune) ([]rune, error) {
	unended := true
	buildingWord = append(buildingWord, delim)
CharLoop:
	for char := range chars {
		buildingWord = append(buildingWord, char)
		switch char {
		case delim:
			unended = false
			break CharLoop
		case '\\':
			buildingWord = append(buildingWord, <-chars)
		}
	}
	if unended {
		return nil, errUnended
	}
	return buildingWord, nil
}

func finishLine(words chan<- string, buildingWord []rune, done <-chan types.NoneType) {
	sendReset(words, buildingWord)
	close(words)
	<-done
}
