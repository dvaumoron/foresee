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
	"strconv"
	"strings"

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

var (
	CustomRules = types.NewList()

	wordParsers []ConvertString

	// an empty environment to execute custom rules
	BuiltinsCopy types.Environment = types.MakeBaseEnvironment()
)

type ConvertString func(string) (types.Object, bool)

// needed to prevent a cycle in the initialisation
func init() {
	wordParsers = []ConvertString{
		parseTrue, parseFalse, parseNone, parseUnquote, parseList,
		parseAddressing, parseDereference, parseSliceType,
		parseString, parseRune, parseInt, parseFloat,
	}
}

func HandleClassicWord(word string, nodeList *types.List) {
	if nativeRules(word, nodeList) {
		args := types.NewList(types.String(word))
		continueLoop := true
		types.ForEach(CustomRules, func(object types.Object) bool {
			rule, ok := object.(types.Appliable)
			if ok {
				// The Apply must return None if it fails.
				node := rule.Apply(BuiltinsCopy, args)
				if _, continueLoop = node.(types.NoneType); !continueLoop {
					nodeList.Add(node)
				}
			}
			return continueLoop
		})
		if continueLoop {
			nodeList.Add(types.Identifier(word))
		}
	}
}

// a true is returned when no rule match
func nativeRules(word string, nodeList *types.List) bool {
	for _, parser := range wordParsers {
		if node, ok := parser(word); ok {
			nodeList.Add(node)
			return false
		}
	}
	return true
}

func parseTrue(word string) (types.Object, bool) {
	return types.Boolean(true), word == "true"
}

func parseFalse(word string) (types.Object, bool) {
	return types.Boolean(false), word == "false"
}

func parseNone(word string) (types.Object, bool) {
	return types.None, word == "None"
}

func parseString(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if word[0] != '"' || word[lastIndex] != '"' {
		return nil, false
	}
	escape := false
	extracted := make([]rune, 0, lastIndex)
	for _, char := range word[1:lastIndex] {
		if escape {
			escape = false
			if char == '\'' {
				extracted = append(extracted, '\'')
			} else {
				extracted = append(extracted, '\\', char)
			}
		} else {
			switch char {
			case '"':
				return nil, false
			case '\\':
				escape = true
			default:
				extracted = append(extracted, char)
			}
		}
	}
	return types.String(extracted), true
}

func parseRune(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if word[0] != '\'' || word[lastIndex] != '\'' {
		return nil, false
	}
	var extracted rune
	for _, char := range word[1:lastIndex] {
		if char == '\\' {
			continue
		}
		extracted = char
		break
	}
	return types.Rune(extracted), true
}

// manage melting with string literal
func parseList(word string) (types.Object, bool) {
	if word == names.Var {
		return nil, false
	}
	chars := make(chan rune)
	go sendChar(chars, word)
	index := 0
	var indexes []int
	for char := range chars {
		switch char {
		case '"', '\'':
			delim := char
		InnerCharLoop:
			for char := range chars {
				index++
				switch char {
				case delim:
					// no need of unended string detection,
					// this have already been tested in the word splitting part
					break InnerCharLoop
				case '\\':
					<-chars
					index++
				}
			}
		case ':':
			indexes = append(indexes, index)
		}
		index++
	}
	if len(indexes) == 0 {
		return nil, false
	}
	nodeList := types.NewList(names.ListId)
	startIndex := 0
	for _, splitIndex := range indexes {
		handleSubWord(word[startIndex:splitIndex], nodeList)
		startIndex = splitIndex + 1
	}
	handleSubWord(word[startIndex:], nodeList)
	return nodeList, true
}

func handleSubWord(word string, nodeList *types.List) {
	if word == "" {
		nodeList.Add(types.None)
	} else {
		HandleClassicWord(word, nodeList)
	}
}

func parseInt(word string) (types.Object, bool) {
	i, err := strconv.ParseInt(word, 10, 64)
	return types.Integer(i), err == nil
}

func parseFloat(word string) (types.Object, bool) {
	f, err := strconv.ParseFloat(word, 64)
	return types.Float(f), err == nil
}

func parseUnquote(word string) (types.Object, bool) {
	if word[0] != ',' {
		return nil, false
	}
	nodeList := types.NewList(types.Identifier(names.UnquoteId))
	handleSubWord(word[1:], nodeList)
	return nodeList, true
}

func parseAddressing(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '&' || len(word) == 1 || word == "&=" || word == "&^=" {
		return nil, false
	}
	nodeList := types.NewList(names.AmpersandId)
	handleSubWord(word[1:], nodeList)
	return nodeList, true
}

func parseDereference(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '*' || len(word) == 1 || word == "*=" {
		return nil, false
	}
	nodeList := types.NewList(names.StarId)
	handleSubWord(word[1:], nodeList)
	return nodeList, true
}

func parseSliceType(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if !strings.HasPrefix(word, string(names.LoadId)) || len(word) == 2 || word == string(names.StoreId) {
		return nil, false
	}
	nodeList := types.NewList(names.LoadId)
	handleSubWord(word[1:], nodeList)
	return nodeList, true
}
