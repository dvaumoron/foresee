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
		skipSeparator, parseTrue, parseFalse, parseNone, parseString, parseRune, parseInt, parseFloat, parseUnquote, parseLiteral, parseList,
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
	size := len(sliced)
	if size == 0 {
		return types.None, 0
	}

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
	case split.SquareBracketsKind:
		if len(l) == 0 {
			if size > 1 {
				if _, s, _ := sliced[0].Cast(); s == "=" {
					return names.StoreId, 2
				}
			}

			return names.LoadId, 1
		}
	case split.CurlyBracesKind:
		if len(l) == 0 {
			return types.Identifier("{}"), 1
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
func handleTypeList(node split.Node) types.Object {
	switch k, s, l := node.Cast(); k {
	case split.StringKind:
		if s != "" {
			return types.NewList(names.ListId, handleSubWord(node))
		}
	case split.ParenthesisKind, split.SquareBracketsKind, split.CurlyBracesKind:
		res := types.NewList(names.ListId)
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
		return res, 2
	}
	return nil, 0
}

func lenWithoutLastSeparator(l []split.Node) int {
	res := len(l)
	if res == 0 {
		return 0
	}

	if k, _, _ := l[res-1].Cast(); k == split.SeparatorKind {
		return res - 1
	}
	return res
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

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[1:])}, sliced[1:]...))
	return types.NewList(names.AmpersandId, object), consumed
}

// handle "[n]type" or "[]type" as (slice n type) or (slice type)
func parseArrayOrSliceType(sliced []split.Node) (types.Object, int) {
	k, _, l := sliced[0].Cast()
	if k != split.SquareBracketsKind || len(sliced) < 2 {
		return nil, 0
	}

	k, s, _ := sliced[1].Cast()
	if k != split.StringKind || s == "=" {
		return nil, 0
	}

	nodeList := types.NewList(names.SliceId)
	sizeNode, consumed := handleSlice(l)
	if consumed != lenWithoutLastSeparator(l) {
		return nil, 0
	}

	if consumed != 0 {
		nodeList.Add(sizeNode)
	}
	object, consumed := handleSlice(sliced[1:])
	nodeList.Add(object)
	return nodeList, consumed + 1
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
	return handleChanType(sliced, names.ChanId)
}

// handle "*a" as (* a)
func parseDereference(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	// test len to keep the basic identifier case
	if s[0] != '*' || len(s) == 1 || s == names.MultAssign {
		return nil, 0
	}

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[1:])}, sliced[1:]...))
	return types.NewList(names.StarId, object), consumed
}

// handle "a.b.c" as (get a b c)
// (manage melting with string literal or nested part)
func parseDotField(sliced []split.Node) (types.Object, int) {
	if word == names.Dot {
		return nil, 0
	}
	return splitListSep(word, '.', names.GetId)
}

// handle "...type" as (... type)
func parseEllipsis(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	// test len to keep the basic identifier case
	if !strings.HasPrefix(s, string(names.EllipsisId)) || len(s) == 3 {
		return nil, 0
	}

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[3:])}, sliced[1:]...))
	return types.NewList(names.EllipsisId, object), consumed
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
// typeList format is "t1 t2" as (list t1 t2)
// typeList2 format could be "t1" or "(t1 t2)" as (list t1) or (list t1 t2)
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
// typeList format is "t1 t2" as (list t1 t2) where t1 and t2 can be any node (including "name:type" format)
func parseGenericType(sliced []split.Node) (types.Object, int) {
	if len(sliced) < 2 {
		return nil, 0
	}

	k, _, _ := sliced[1].Cast()
	if k != split.SquareBracketsKind {
		return nil, 0
	}
	return types.NewList(names.GenId, handleSubWord(sliced[0]), handleTypeList(sliced[1])), 2
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
	_, s, _ := sliced[0].Cast()
	if s[0] != '$' {
		return nil, 0
	}

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[1:])}, sliced[1:]...))
	return types.NewList(names.LitId, object), consumed
}

// handle "map[t1]t2" as (map t1 t2)
func parseMapType(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s != string(names.MapId) || len(sliced) < 3 {
		return nil, 0
	}

	k, _, l := sliced[1].Cast()
	if k != split.SquareBracketsKind {
		return nil, 0
	}

	t1, consumed := handleSlice(l)
	if consumed != lenWithoutLastSeparator(l) || consumed == 0 {
		return nil, 0
	}

	t2, consumed := handleSlice(sliced[2:])
	return types.NewList(names.MapId, t1, t2), consumed + 2
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
		return types.NewList(names.NotId, handleSubWord(split.StringNode(s[1:]))), 1
	}
	return nil, 0
}

func parseRune(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	lastIndex := len(s) - 1
	if s[0] != '\'' || s[lastIndex] != '\'' {
		return nil, 0
	}

	extracted, _, _, _ := strconv.UnquoteChar(s[1:], '\'')
	return types.Rune(extracted), 1
}

func parseString(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	if s[0] != '"' || s[len(s)-1] != '"' {
		return nil, 0
	}

	extracted, _ := strconv.Unquote(s)
	return types.String(extracted), 1
}

// handle "~type" as (~ type)
func parseTilde(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	// test len to keep the basic identifier case
	if s[0] != '~' || len(s) == 1 {
		return nil, 0
	}

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[1:])}, sliced[1:]...))
	return types.NewList(names.TildeId, object), consumed
}

func parseTrue(sliced []split.Node) (types.Object, int) {
	if _, s, _ := sliced[0].Cast(); s == "true" {
		return types.Boolean(true), 1
	}
	return nil, 0
}

// handle ",a" as (quote a)
func parseUnquote(sliced []split.Node) (types.Object, int) {
	_, s, _ := sliced[0].Cast()
	if s[0] != ',' {
		return nil, 0
	}

	object, consumed := handleSlice(append([]split.Node{split.StringNode(s[1:])}, sliced[1:]...))
	return types.NewList(names.UnquoteId, object), consumed
}

func skipSeparator(sliced []split.Node) (types.Object, int) {
	if k, _, _ := sliced[0].Cast(); k == split.SeparatorKind {
		return nil, -1
	}
	return nil, 0
}
