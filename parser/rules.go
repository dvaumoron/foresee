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
		parseEllipsis, parseTilde, parseAddressing, parseDereference, parseNot, parseArrowChanType, parseChanArrowType, parseChanType,
		parseArrayOrSliceType, parseMapType, parseFuncType, parseGenericType, parseDotField,
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
func handleWord(word string) types.Object {
	for _, parser := range wordParsers {
		if node, match := parser(word); match {
			return node
		}
	}
	return types.Identifier(word)
}

// empty string are handled as None, otherwise call handleWord
func handleSubWord(word string) types.Object {
	if word = strings.TrimSpace(word); word == "" {
		return types.None
	}
	return handleWord(word)
}

// always return a list with the ListId header
// (manage melting with string literal or nested part)
func handleTypeList(word string) types.Object {
	if word == "" {
		return types.NewList(names.ListId)
	}

	if res, ok := splitListSep(word, ',', names.ListId); ok {
		return res
	}
	return types.NewList(names.ListId, handleWord(word))
}

func handleChanType(word string, prefix string, typeId types.Identifier) (types.Object, bool) {
	lastIndex := len(word) - 1
	if !strings.HasPrefix(word, prefix) || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(typeId, handleSubWord(word[len(prefix):lastIndex])), true
}

// handle "&value" as (& value)
func parseAddressing(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '&' || len(word) == 1 ||
		word == names.And || word == names.AndAssign ||
		word == names.AndNot || word == names.AndNotAssign {
		return nil, false
	}
	return types.NewList(names.AmpersandId, handleSubWord(word[1:])), true
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

// handle "<-chan[type]" as (<-chan type)
func parseArrowChanType(word string) (types.Object, bool) {
	return handleChanType(word, "<-chan[", names.ArrowChanId)
}

// handle "chan<-[type]" as (chan<- type)
func parseChanArrowType(word string) (types.Object, bool) {
	return handleChanType(word, "chan<-[", names.ChanArrowId)
}

// handle "chan[type]" as (chan type)
func parseChanType(word string) (types.Object, bool) {
	return handleChanType(word, "chan[", names.ChanId)
}

// handle "*a" as (* a)
func parseDereference(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '*' || len(word) == 1 || word == names.MultAssign {
		return nil, false
	}
	return types.NewList(names.StarId, handleSubWord(word[1:])), true
}

// handle "a.b.c" as (get a b c)
// (manage melting with string literal or nested part)
func parseDotField(word string) (types.Object, bool) {
	if word == names.Dot {
		return nil, false
	}
	return splitListSep(word, '.', names.GetId)
}

// handle "...type" as (... type)
func parseEllipsis(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if !strings.HasPrefix(word, string(names.EllipsisId)) || len(word) == 3 {
		return nil, false
	}
	return types.NewList(names.EllipsisId, handleSubWord(word[3:])), true
}

func parseFalse(word string) (types.Object, bool) {
	return types.Boolean(false), word == "false"
}

func parseFloat(word string) (types.Object, bool) {
	f, err := strconv.ParseFloat(word, 64)
	return types.Float(f), err == nil
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

// handle "type[typeList]" as (gen type typeList)
// typeList format is "t1,t2" as (list t1 t2) where t1 and t2 can be any node (including "name:type" format)
func parseGenericType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, '[')
	lastIndex := len(word) - 1
	// (no need to test ':' (this rule is applied after parseList))
	if index == -1 || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(names.GenId, handleSubWord(word[:index]), handleTypeList(word[index+1:lastIndex])), true
}

func parseInt(word string) (types.Object, bool) {
	i, err := strconv.ParseInt(word, 10, 64)
	return types.Integer(i), err == nil
}

// handle "a:b:c" as (list a b c)
// (manage melting with string literal or nested part)
func parseList(word string) (types.Object, bool) {
	// exception for ":="
	if word == names.DeclareAssign {
		return nil, false
	}
	return splitListSep(word, ':', names.ListId)
}

// handle "$type" as (lit type)
// mark a type in order to use it as literal
func parseLiteral(word string) (types.Object, bool) {
	if word[0] != '$' {
		return nil, false
	}
	return types.NewList(names.LitId, handleSubWord(word[1:])), true
}

// handle "map[t1]t2" as (map t1 t2)
func parseMapType(word string) (types.Object, bool) {
	index := strings.IndexByte(word, ']')
	if !strings.HasPrefix(word, "map[") || index == -1 {
		return nil, false
	}
	return types.NewList(names.MapId, handleSubWord(word[4:index]), handleSubWord(word[index+1:])), true
}

func parseNone(word string) (types.Object, bool) {
	return types.None, word == "None"
}

// handle "!b" as (! b)
func parseNot(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '!' || len(word) == 1 || word == names.NotEqual {
		return nil, false
	}
	return types.NewList(names.NotId, handleSubWord(word[1:])), true
}

func parseRune(word string) (types.Object, bool) {
	lastIndex := len(word) - 1
	if word[0] != '\'' || word[lastIndex] != '\'' {
		return nil, false
	}

	first := false
	var extracted rune
	for _, char := range word[1:lastIndex] {
		if char == '\\' {
			first = false
			continue
		}
		extracted = char
		if first {
			break
		}

		switch char {
		case 'a':
			extracted = '\a'
		case 'b':
			extracted = '\b'
		case 'f':
			extracted = '\f'
		case 'n':
			extracted = '\n'
		case 'r':
			extracted = '\r'
		case 't':
			extracted = '\t'
		case 'v':
			extracted = '\v'
		}
		break
	}
	return types.Rune(extracted), true
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

// handle "~type" as (~ type)
func parseTilde(word string) (types.Object, bool) {
	// test len to keep the basic identifier case
	if word[0] != '~' || len(word) == 1 {
		return nil, false
	}
	return types.NewList(names.TildeId, handleSubWord(word[1:])), true
}

func parseTrue(word string) (types.Object, bool) {
	return types.Boolean(true), word == "true"
}

// handle ",a" as (quote a)
func parseUnquote(word string) (types.Object, bool) {
	if word[0] != ',' {
		return nil, false
	}
	return types.NewList(names.UnquoteId, handleSubWord(word[1:])), true
}
