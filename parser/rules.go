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
	wordParsers []ConvertString

	// an empty environment to execute custom rules
	BuiltinsCopy types.Environment = types.MakeBaseEnvironment()
)

type ConvertString = func(string) (types.Object, bool)

// needed to prevent a cycle in the initialisation
func init() {
	wordParsers = []ConvertString{
		parseTrue, parseFalse, parseNone, parseString, parseRune, parseInt, parseFloat, parseUnquote, parseLiteral, parseList,
		parseArrowChanType, parseChanArrowType, parseChanType, parseGenericType, parseArrayOrSliceType, parseMapType, parseFuncType,
		parseEllipsis, parseTilde, parseAddressing, parseDereference, parseDotField,
	}
}

// Must be called before the parsing of files to affect them
func AddCustomRule(rule types.Appliable) {
	// TODO use a mutex ? (will macro or parsing run concurrently ???)
	wordParsers = append(wordParsers, func(word string) (types.Object, bool) {
		args := types.NewList(types.String(word))
		node := rule.Apply(BuiltinsCopy, args)
		// The Apply must return None if it fails.
		_, isNone := node.(types.NoneType)
		return node, !isNone
	})
}

// try to apply parsing rule in order (including custom rules),
// fallback to an identifier when nothing matches
func HandleClassicWord(word string) types.Object {
	for _, parser := range wordParsers {
		if node, match := parser(word); match {
			return node
		}
	}
	return types.Identifier(word)
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

// handle "a:b:c" as (list a b c)
// (manage melting with string literal or nested part)
func parseList(word string) (types.Object, bool) {
	// exception for ":=
	if word == names.DeclareAssign {
		return nil, false
	}
	return parseListSep(word, ':', names.ListId)
}

// manage melting with string literal or nested part
func parseListSep(word string, sep rune, kindId types.Identifier) (types.Object, bool) {
	chars := make(chan rune)
	go sendChar(chars, word)

	index := 0
	var indexes []int
	var waiteds []rune
	for char := range chars {
		switch char {
		case '"', '\'':
			index = consumeString(chars, index, char)
		case '<':
			waiteds = append(waiteds, '>')
		case '[':
			waiteds = append(waiteds, ']')
		case '{':
			waiteds = append(waiteds, '}')
		case '>', ']', '}':
			opennedLen := len(waiteds)
			if opennedLen == 0 {
				return nil, true
			}

			lastIndex := opennedLen - 1
			if waiteds[lastIndex] != char {
				return nil, false // non matching close
			}
			waiteds = waiteds[:lastIndex] // pop closed pair
		case sep:
			if len(waiteds) == 0 {
				indexes = append(indexes, index)
			}
		}
		index++
	}
	if len(indexes) == 0 || len(waiteds) != 0 {
		return nil, false
	}

	nodeList := types.NewList(kindId)
	startIndex := 0
	for _, splitIndex := range indexes {
		nodeList.Add(handleSubWord(word[startIndex:splitIndex]))
		startIndex = splitIndex + 1
	}
	nodeList.Add(handleSubWord(word[startIndex:]))
	return nodeList, true
}

func consumeString(chars <-chan rune, index int, delim rune) int {
	for char := range chars {
		index++
		switch char {
		case delim:
			// no need of unended string detection (already tested during word splitting)
			break
		case '\\':
			<-chars
			index++
		}
	}
	return index
}

// empty string are handled as None, otherwise call HandleClassicWord
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

// handle ",a" as (quote a)
func parseUnquote(word string) (types.Object, bool) {
	if word[0] != ',' {
		return nil, false
	}
	return types.NewList(names.UnquoteId, handleSubWord(word[1:])), true
}

// handle "$type" as (lit type)
// mark a type in order to use it as literal
func parseLiteral(word string) (types.Object, bool) {
	if word[0] != '$' {
		return nil, false
	}
	return types.NewList(names.LitId, handleSubWord(word[1:])), true
}

// handle "...type" as (... type)
func parseEllipsis(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if !strings.HasPrefix(word, string(names.EllipsisId)) || len(word) == 3 {
		return nil, false
	}
	return types.NewList(names.EllipsisId, handleSubWord(word[3:])), true
}

// handle "~type" as (~ type)
func parseTilde(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '~' || len(word) == 1 {
		return nil, false
	}
	return types.NewList(names.TildeId, handleSubWord(word[1:])), true
}

// handle "&type" as (& type)
func parseAddressing(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '&' || len(word) == 1 || word == names.AndAssign || word == names.NotAndAssign {
		return nil, false
	}
	return types.NewList(names.AmpersandId, handleSubWord(word[1:])), true
}

// handle "*type" as (* type)
func parseDereference(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '*' || len(word) == 1 || word == "*=" {
		return nil, false
	}
	return types.NewList(names.StarId, handleSubWord(word[1:])), true
}

// handle "[n]type" or "[]type" as (slice n type) or (slice type)
func parseArrayOrSliceType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, ']')
	// test len to keep the basic identifier case
	if word[0] != '[' || index == -1 || len(word) == 2 || word == names.Store {
		return nil, false
	}
	nodeList := types.NewList(names.SliceId)
	sizeNode := handleSubWord(word[1:index])
	if _, ok := sizeNode.(types.NoneType); !ok {
		nodeList.Add(sizeNode)
	}
	nodeList.Add(handleSubWord(word[index+1:]))
	return nodeList, true
}

// handle "map[t1]t2" as (map t1 t2)
func parseMapType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, ']')
	if !strings.HasPrefix(word, "map[") || index == -1 {
		return nil, false
	}
	return types.NewList(names.MapId, handleSubWord(word[4:index]), handleSubWord(word[index+1:])), true
}

// handle "<-chan[type]" as (<-chan type)
func parseArrowChanType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("<-chan[")) || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(names.ArrowChanId, handleSubWord(word[7:lastIndex])), true
}

// handle "chan<-[type]" as (chan<- type)
func parseChanArrowType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("chan<-[")) || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(names.ChanArrowId, handleSubWord(word[7:lastIndex])), true
}

// handle "chan[type]" as (chan type)
func parseChanType(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, string("chan[")) || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(names.ChanId, handleSubWord(word[5:lastIndex])), true
}

// handle "func[typeList]typeList2" as (func typeList typeList2),
// typeList format is "t1,t2" as (list t1 t2)
func parseFuncType(word string) (types.Object, bool) {
	if !strings.HasPrefix(word, string("func[")) || strings.IndexByte(word, ']') == -1 {
		return nil, false
	}

	// search of the corresponding closing square bracket (without string handling)
	index, count := 5, 1
	for wordLen := len(word); index < wordLen; index++ {
		switch word[index] {
		case '[':
			count++
		case ']':
			count--
			if count == 0 {
				break
			}
		}
	}

	if count != 0 {
		// incorrect syntax
		return nil, false
	}
	return types.NewList(names.FuncId, handleTypeList(word[5:index]), handleTypeList(word[index+1:])), true
}

// always return a list with the ListId header
// (manage melting with string literal or nested part)
func handleTypeList(word string) types.Object {
	if word == "" {
		return types.NewList(names.ListId)
	}

	if res, ok := parseListSep(word, ',', names.ListId); ok {
		return res
	}
	return types.NewList(names.ListId, HandleClassicWord(word))
}

// handle "type<typeList>" as (gen type typeList)
// typeList format is "t1,t2" as (list t1 t2) where t1 and t2 can be any node (including "name:type" format)
func parseGenericType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, '<')
	lastIndex := len(word) - 1
	// (no need to test ':' (this rule is applied after parseList))
	if index == -1 || lastIndex == -1 || word[lastIndex] != '>' {
		return nil, false
	}
	return types.NewList(names.GenId, handleSubWord(word[:index]), handleTypeList(word[index+1:lastIndex])), true
}

// handle "a.b.c" as (get a b c)
// (manage melting with string literal or nested part)
func parseDotField(word string) (types.Object, bool) {
	if word == names.Dot {
		return nil, false
	}
	return parseListSep(word, '.', names.GetId)
}
