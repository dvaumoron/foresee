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
		parseTrue, parseFalse, parseNone, parseString, parseRune, parseInt, parseFloat, parseUnquote,
		// handle "&type", "*type", "[n]type", "map[t1]t2", "func[typeList]typeList2"
		// as (& type), (* type), ([] n? type), (map t1 t2), (func typeList typeList2)
		// typeList format is "t1,t2" as (list t1 t2)
		parseAddressing, parseDereference, parseArrayOrSliceType, parseMapType, parseFuncType,
		// handle "<-chan[type]", "chan<-[type]", "chan[type]" "a:b" as (<-chan type), (chan<- type), (chan type), (list a b)
		parseArrowChanType, parseChanArrowType, parseChanType, parseList,
	}
}

func HandleClassicWord(word string) types.Object {
	var res types.Object
	continueLoop := true
	if res, continueLoop = nativeRules(word); continueLoop {
		args := types.NewList(types.String(word))
		types.ForEach(CustomRules, func(object types.Object) bool {
			if rule, ok := object.(types.Appliable); ok {
				// The Apply must return None if it fails.
				node := rule.Apply(BuiltinsCopy, args)
				if _, continueLoop = node.(types.NoneType); !continueLoop {
					res = node
				}
			}
			return continueLoop
		})
	}
	return res
}

// An identifier and true is returned when no rule match
func nativeRules(word string) (types.Object, bool) {
	for _, parser := range wordParsers {
		if node, ok := parser(word); ok {
			return node, false
		}
	}
	return types.Identifier(word), true
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

func parseList(word string) (types.Object, bool) {
	if word == names.Var {
		return nil, false
	}
	return parseListSep(word, ':')
}

// manage melting with string literal
func parseListSep(word string, sep rune) (types.Object, bool) {
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
		case sep:
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
		nodeList.Add(handleSubWord(word[startIndex:splitIndex]))
		startIndex = splitIndex + 1
	}
	nodeList.Add(handleSubWord(word[startIndex:]))
	return nodeList, true
}

func handleSubWord(word string) types.Object {
	if word == "" {
		return types.None
	}
	return HandleClassicWord(word)
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
	nodeList.Add(handleSubWord(word[1:]))
	return nodeList, true
}

func parseAddressing(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '&' || len(word) == 1 || word == names.AndEqual || word == names.NotAndEqual {
		return nil, false
	}
	nodeList := types.NewList(names.AmpersandId)
	nodeList.Add(handleSubWord(word[1:]))
	return nodeList, true
}

func parseDereference(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '*' || len(word) == 1 || word == "*=" {
		return nil, false
	}
	nodeList := types.NewList(names.StarId)
	nodeList.Add(handleSubWord(word[1:]))
	return nodeList, true
}

func parseArrayOrSliceType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, ']')
	// test len to keep the basic identifier case
	if word[0] != '[' || index == -1 || len(word) == 2 || word == string(names.StoreId) {
		return nil, false
	}
	nodeList := types.NewList(names.LoadId)
	sizeNode := handleSubWord(word[1:index])
	if _, ok := sizeNode.(types.NoneType); !ok {
		nodeList.Add(sizeNode)
	}
	nodeList.Add(handleSubWord(word[index+1:]))
	return nodeList, true
}

func parseMapType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, ']')
	if !strings.HasPrefix(word, "map[") || index == -1 {
		return nil, false
	}
	nodeList := types.NewList(names.MapId)
	nodeList.Add(handleSubWord(word[4:index])) // can be *type
	nodeList.Add(handleSubWord(word[index+1:]))
	return nodeList, true
}

func parseArrowChanType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("<-chan[")) || word[lastIndex] != ']' {
		return nil, false
	}
	nodeList := types.NewList(names.ArrowChanId)
	nodeList.Add(handleSubWord(word[7:lastIndex]))
	return nodeList, true
}

func parseChanArrowType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("chan<-[")) || word[lastIndex] != ']' {
		return nil, false
	}
	nodeList := types.NewList(names.ChanArrowId)
	nodeList.Add(handleSubWord(word[7:lastIndex]))
	return nodeList, true
}

func parseChanType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("chan[")) || word[lastIndex] != ']' {
		return nil, false
	}
	nodeList := types.NewList(names.ChanId)
	nodeList.Add(handleSubWord(word[5:lastIndex]))
	return nodeList, true
}

func parseFuncType(word string) (types.Object, bool) {
	if !strings.HasPrefix(word, string("func[")) || strings.IndexByte(word, ']') == -1 {
		return nil, false
	}

	// search of the corresponding closing square bracket
	index, count := 5, 1
IndexLoop:
	for wordLen := len(word); index < wordLen; index++ {
		switch word[index] {
		case '[':
			count++
		case ']':
			count--
			if count == 0 {
				break IndexLoop
			}
		}
	}

	if count != 0 {
		// incorrect syntax
		return nil, false
	}

	nodeList := types.NewList(names.FuncId)
	nodeList.Add(handleTypeList(word[5:index]))
	nodeList.Add(handleTypeList(word[index+1:]))
	return nodeList, true
}

func handleTypeList(word string) types.Object {
	nodeList := types.NewList(names.ListId)
	if word == "" {
		return nodeList
	}

	if res, ok := parseListSep(word, ','); ok {
		return res
	}

	nodeList.Add(HandleClassicWord(word))
	return nodeList
}
