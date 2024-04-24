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
	"github.com/dvaumoron/foresee/parser/split"
	"github.com/dvaumoron/foresee/types"
)

var (
	sliceParsers []SliceParser

	// an empty environment to execute custom rules
	BuiltinsCopy types.Environment = types.MakeBaseEnvironment()
)

type SliceParser = func([]split.Node) (types.Object, int)

// needed to prevent a cycle in the initialisation
func init() {
	sliceParsers = []SliceParser{
		parseTrue, parseFalse, parseNone, parseString, parseRune, parseInt, parseFloat, parseUnquote, parseLiteral, parseList,
		parseEllipsis, parseTilde, parseAddressing, parseDereference, parseNot, parseArrowChanType, parseChanArrowType, parseChanType,
		parseArrayOrSliceType, parseMapType, parseFuncType, parseGenericType, parseDotField,
	}
}

// Must be called before the parsing of files to affect them
func AddCustomRule(rule types.Appliable) {
	// TODO use a mutex ? (will macro or parsing run concurrently ???)
	sliceParsers = append(sliceParsers, func(sliced []split.Node) (types.Object, int) {
		args := types.NewList(types.String("todo"))
		node := rule.Apply(BuiltinsCopy, args)
		// The Apply must return None if it fails.
		if _, isNone := node.(types.NoneType); isNone {
			return types.None, 0
		}
		return types.None, 1
	})
}

// try to apply parsing rule in order (including custom rules),
// fallback to an identifier when nothing matches
func handleSlice(sliced []split.Node) (types.Object, int) {
	for _, parser := range sliceParsers {
		if node, consumed := parser(sliced); consumed != 0 {
			return node, consumed
		}
	}

	switch k, s, l := sliced[0].Cast(); k {
	case split.StringKind:
		if s != "" {
			return types.Identifier(s), 1
		}
	case split.ParenthesisKind:
		res := types.NewList()
		if processNodes(l, res) == nil {
			return res, 1
		}
	}
	return types.None, 0
}

// empty string are handled as None, otherwise call handleWord
func handleSubWord(node split.Node) types.Object {
	if _, s, _ := node.Cast(); s == "" {
		return types.None
	}
	o, _ := handleSlice([]split.Node{node})
	return o
}

// always return a list with the ListId header
// (manage melting with string literal or nested part)
func handleTypeList(node split.Node) types.Object {
	switch k, s, l := node.Cast(); k {
	case split.StringKind:
		if s != "" {
			return types.NewList(names.ListId, handleSubWord(node))
		}
	case split.ParenthesisKind, split.SquareBracketsKind, split.CurlyBracesKind:
		res := types.NewList()
		if processNodes(l, res) == nil {
			return res
		}
	}
	return types.NewList(names.ListId)
}

func handleChanType(sliced []split.Node, typeId types.Identifier) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	if s != string(typeId) || len(sliced) < 2 {
		return nil, 0
	}

	k, _, l := sliced[1].Cast()
	if k != split.SquareBracketsKind {
		return nil, 0
	}

	res := types.NewList(typeId)
	if processNodes(l, res) == nil {
		return res, 1
	}
	return nil, 0
}

// handle "&value" as (& value)
func parseAddressing(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	// test len to keep the basic identifier case
	if s[0] != '&' || len(s) == 1 ||
		s == names.And || s == names.AndAssign ||
		s == names.AndNot || s == names.AndNotAssign {
		return nil, 0
	}
	return types.NewList(names.AmpersandId, handleSubWord(s[1:])), 1
}

// handle "[n]type" or "[]type" as (slice n type) or (slice type)
func parseArrayOrSliceType(sliced []split.Node) (types.Object, int) {
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
func parseArrowChanType(sliced []split.Node) (types.Object, int) {
	return handleChanType(sliced, names.ArrowChanId)
}

// handle "chan<-[type]" as (chan<- type)
func parseChanArrowType(sliced []split.Node) (types.Object, int) {
	return handleChanType(sliced, names.ChanArrowId)
}

// handle "chan[type]" as (chan type)
func parseChanType(sliced []split.Node) (types.Object, int) {
	return handleChanType(word, names.ChanId)
}

// handle "*a" as (* a)
func parseDereference(sliced []split.Node) (types.Object, int) {
	// test len to keep the basic identifier case
	if _, s, _ := sliced[0].Cast(); s[0] == '*' && len(s) != 1 && s != names.MultAssign {
		return types.NewList(names.StarId, handleSubWord(s[1:])), 1
	}
	return nil, 0
}

// handle "a.b.c" as (get a b c)
// (manage melting with string literal or nested part)
func parseDotField(sliced []split.Node) (types.Object, int) {
	if word == names.Dot {
		return nil, false
	}
	return splitListSep(word, '.', names.GetId)
}

// handle "...type" as (... type)
func parseEllipsis(sliced []split.Node) (types.Object, int) {
	// test len to keep the basic identifier case
	if _, s, _ := sliced[0].Cast(); strings.HasPrefix(s, string(names.EllipsisId)) && len(s) != 3 {
		return types.NewList(names.EllipsisId, handleSubWord(s[3:])), 1
	}
	return nil, 0
}

func parseFalse(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s == "false" {
		return types.Boolean(false), 1
	}
	return nil, 0
}

func parseFloat(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return types.Float(f), 1
	}
	return nil, 0
}

// handle "func[typeList]typeList2" as (func typeList typeList2),
// typeList format is "t1,t2" as (list t1 t2)
func parseFuncType(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s != string(names.FuncId) || len(sliced) < 3 {
		return nil, 0
	}
	if k, _, _ := sliced[1].Cast(); k != split.SquareBracketsKind {
		return nil, 0
	}
	return types.NewList(names.FuncId, handleTypeList(sliced[1]), handleTypeList(sliced[2])), 3
}

// handle "type[typeList]" as (gen type typeList)
// typeList format is "t1,t2" as (list t1 t2) where t1 and t2 can be any node (including "name:type" format)
func parseGenericType(sliced []split.Node) (types.Object, int) {
	index := strings.IndexByte(word, '[')
	lastIndex := len(word) - 1
	// (no need to test ':' (this rule is applied after parseList))
	if index == -1 || word[lastIndex] != ']' {
		return nil, false
	}
	return types.NewList(names.GenId, handleSubWord(word[:index]), handleTypeList(word[index+1:lastIndex])), true
}

func parseInt(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return types.Integer(i), 1
	}
	return nil, 0
}

// handle "a:b:c" as (list a b c)
// (manage melting with string literal or nested part)
func parseList(sliced []split.Node) (types.Object, int) {
	// exception for ":="
	if _, s, _ := sliced[0].Cast(); s != names.DeclareAssign {
		return splitListSep(s, ':', names.ListId)
	}
	return nil, 0
}

// handle "$type" as (lit type)
// mark a type in order to use it as literal
func parseLiteral(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s[0] == '$' {
		return types.NewList(names.LitId, handleSubWord(s[1:])), 1
	}
	return nil, 0
}

// handle "map[t1]t2" as (map t1 t2)
func parseMapType(sliced []split.Node) (types.Object, int) {
	index := strings.IndexByte(word, ']')
	if !strings.HasPrefix(word, "map[") || index == -1 {
		return nil, false
	}
	return types.NewList(names.MapId, handleSubWord(word[4:index]), handleSubWord(word[index+1:])), true
}

func parseNone(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s == "None" {
		return types.None, 1
	}
	return nil, 0
}

// handle "!b" as (! b)
func parseNot(sliced []split.Node) (types.Object, int) {
	// test len to keep the basic identifier case
	if _, s, _ := sliced[0].Cast(); s[0] == '!' && len(s) != 1 && s != names.NotEqual {
		return types.NewList(names.NotId, handleSubWord(s[1:])), 1
	}
	return nil, 0
}

func parseRune(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	lastIndex := len(s) - 1
	if s[0] != '\'' || s[lastIndex] != '\'' {
		return nil, 0
	}

	first := false
	var extracted rune
	for _, char := range s[1:lastIndex] {
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
	return types.Rune(extracted), 1
}

func parseString(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	lastIndex := len(s) - 1
	if s[0] != '"' || s[lastIndex] != '"' {
		return nil, 0
	}

	escape := false
	extracted := make([]rune, 0, lastIndex)
	for _, char := range s[1:lastIndex] {
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
				return nil, 0
			case '\\':
				escape = true
			default:
				extracted = append(extracted, char)
			}
		}
	}
	return types.String(extracted), 1
}

// handle "~type" as (~ type)
func parseTilde(sliced []split.Node) (types.Object, int) {
	// test len to keep the basic identifier case
	if _, s, _ := sliced[0].Cast(); s[0] == '~' && len(s) != 1 {
		return types.NewList(names.TildeId, handleSubWord(s[1:])), 1
	}
	return nil, 0
}

func parseTrue(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s == "true" {
		return types.Boolean(true), 1
	}
	return nil, 0
}

// handle ",a" as (quote a)
func parseUnquote(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s[0] == ',' {
		return types.NewList(names.UnquoteId, handleSubWord(s[1:])), 1
	}
	return nil, 0
}
