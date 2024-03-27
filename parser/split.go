/*
 *
 * Copyright 2024 gosince authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package parser

import (
	"errors"

	"github.com/dvaumoron/foresee/types"
)

var (
	errParsingClosing      = errors.New("parsing failure : wait closing separator")
	errParsingParent       = errors.New("parsing failure : unexpected parenthesis")
	errParsingString       = errors.New("parsing failure : unended string")
	errParsingWrongClosing = errors.New("parsing failure : wait another closing separator")
)

func appendBuffer(splitted []string, buffer []rune) ([]string, []rune) {
	if len(buffer) != 0 {
		splitted = append(splitted, string(buffer))
		buffer = buffer[:0]
	}
	return splitted, buffer
}
func readSub(buffer []rune, chars <-chan rune, startDelim rune, endDelim rune) ([]rune, error) {
	var err error
	buffer = append(buffer, startDelim)
	for char := range chars {
		switch {
		case char == endDelim:
			buffer = append(buffer, char)
			return buffer, nil
		case char == '<':
			buffer, err = readSub(buffer, chars, '<', '>')
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
		case char == '>', char == ']', char == '}':
			return nil, errParsingWrongClosing
		case char == '(', char == ')':
			// return buffer with parenthesis to handle special case like "<", "<=" or "<-"
			return append(buffer, char), errParsingParent
		default:
			buffer = append(buffer, char)
		}
	}
	return nil, errParsingClosing
}

func readUntil(buffer []rune, chars <-chan rune, delim rune) ([]rune, error) {
	buffer = append(buffer, delim)
	for char := range chars {
		buffer = append(buffer, char)
		switch char {
		case delim:
			return buffer, nil
		case '\\':
			buffer = append(buffer, <-chars)
		}
	}
	return nil, errParsingString
}

// manage melting with string literal or nested part
func splitListSep(word string, sep rune, kindId types.Identifier) (types.Object, bool) {
	chars := make(chan rune)
	go sendChar(chars, word)

	index := 0
	var err error
	var buffer []rune
	nodeList := types.NewList(kindId)
	for char := range chars {
		switch {
		case char == '"', char == '\'':
			buffer, err = readUntil(buffer, chars, char)
			if err != nil {
				return nil, false
			}
		case char == '<':
			buffer, err = readSub(buffer, chars, '<', '>')
			if err != nil {
				return nil, false
			}
		case char == '[':
			buffer, err = readSub(buffer, chars, '[', ']')
			if err != nil {
				return nil, false
			}
		case char == '{':
			buffer, err = readSub(buffer, chars, '{', '}')
			if err != nil {
				return nil, false
			}
		case char == '>', char == ']', char == '}':
			return nil, false
		case char == sep:
			nodeList.Add(handleSubWord(string(buffer)))
			buffer = buffer[:0]
		default:
			buffer = append(buffer, char)
		}
		index++
	}
	if nodeList.Size() == 1 {
		return nil, false
	}
	nodeList.Add(handleSubWord(string(buffer)))

	return nodeList, true
}
