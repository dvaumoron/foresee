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

package compile

import (
	"math"

	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

const (
	// user can not directly use this kind of id (# start a comment)
	hiddenImportsName = "#imports"
	hiddenPackageName = "#package"

	mainId types.Identifier = "main"
)

var (
	wrappedErrorComment = appliableWrapper{Renderer: jen.Comment("line with error, can't generate correct go code")}

	Builtins = initBuitins()
)

func Compile(l *types.List) types.Object {
	return l.Eval(compileEnvironment{Environment: types.MakeLocalEnvironment(Builtins)})
}

func initBuitins() types.BaseEnvironment {
	base := types.MakeBaseEnvironment()
	base.StoreStr(names.Assign, types.MakeNativeAppliable(assignForm))
	base.StoreStr(names.Block, types.MakeNativeAppliable(blockForm))
	base.StoreStr(names.Const, types.MakeNativeAppliable(constForm))
	base.StoreStr(string(names.FileId), types.MakeNativeAppliable(fileForm))
	base.StoreStr(names.Import, types.MakeNativeAppliable(importForm))
	base.StoreStr(names.Package, types.MakeNativeAppliable(packageForm))
	base.StoreStr(names.Var, types.MakeNativeAppliable(varForm))

	// TODO
	return base
}

func compileToCodeSlice(env types.Environment, instructions types.Iterable) []jen.Code {
	var codes []jen.Code
	types.ForEach(instructions, func(instruction types.Object) bool {
		codes = append(codes, extractCode(instruction.Eval(env)))
		return true
	})
	return codes
}

func compileToCode(env types.Environment, instructions types.Iterator) (jen.Code, bool) {
	instruction, ok := instructions.Next()
	return extractCode(instruction.Eval(env)), ok
}

// manage wrapper and literals
func extractCode(object types.Object) jen.Code {
	switch casted := object.(type) {
	case appliableWrapper:
		if code, ok := casted.Renderer.(jen.Code); ok {
			return code
		}
	case types.Boolean:
		return jen.Lit(bool(casted))
	case types.Integer:
		if math.MinInt <= casted && casted <= math.MaxInt {
			return jen.Lit(int(casted))
		}
		return jen.Lit(int64(casted))
	case types.Float:
		return jen.Lit(float64(casted))
	case types.Rune:
		return jen.LitRune(rune(casted))
	case types.String:
		return jen.Lit(string(casted))
	case types.Identifier:
		return jen.Id(string(casted))
	}
	return jen.Empty()
}

// handle *type, []type, map[t1]t2 format and func or chan types  (and their combinations like [][]*type)
// TODO manage anonymous struct ?
func extractTypeId(object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		switch casted.Size() {
		case 2:
			switch op, _ := casted.LoadInt(0).(types.Identifier); op {
			case "":
				// not an identifier
				return nil
			case names.ArrowChanId:
				return jen.Op(names.Arrow).Chan().Add(extractTypeId(casted.LoadInt(1)))
			case names.ChanArrowId:
				return jen.Chan().Op(names.Arrow).Add(extractTypeId(casted.LoadInt(1)))
			case names.ChanId:
				return jen.Chan().Add(extractTypeId(casted.LoadInt(1)))
			case names.LoadId:
				return jen.Index().Add(extractTypeId(casted.LoadInt(1)))
			case names.FuncId:
				params, ok := casted.LoadInt(1).(*types.List)
				if !ok {
					return nil
				}

				itType := params.Iter() // no need to close (done in ForEach)
				itType.Next()           // skip ListId

				var typeIds []jen.Code
				types.ForEach(itType, func(elem types.Object) bool {
					typeIds = append(typeIds, extractTypeId(elem))
					return true
				})
				return jen.Func().Params(typeIds...)
			default:
				return jen.Op(string(op)).Add(extractTypeId(casted.LoadInt(1)))
			}
		case 3:
			switch op, _ := casted.LoadInt(0).(types.Identifier); op {
			case names.LoadId:
				// manage [size]type
				switch castedSize := casted.LoadInt(1).(type) {
				case types.Integer:
					return jen.Index(jen.Lit(int(castedSize))).Add(extractTypeId(casted.LoadInt(2)))
				case types.Identifier:
					// size is ...
					return jen.Index(jen.Op(string(castedSize))).Add(extractTypeId(casted.LoadInt(2)))
				}
			case names.MapId:
				// manage map[t1]t2
				return jen.Op(string(op)).Add(extractTypeId(casted.LoadInt(1))).Add(extractTypeId(casted.LoadInt(2)))
			case names.FuncId:
				param, ok := casted.LoadInt(1).(*types.List)
				if !ok {
					return nil
				}

				itType := param.Iter() // no need to close (done in ForEach)
				itType.Next()          // skip ListId

				var typeIds []jen.Code
				types.ForEach(itType, func(elem types.Object) bool {
					typeIds = append(typeIds, extractTypeId(elem))
					return true
				})
				funcCode := jen.Func().Params(typeIds...)

				returns, ok := casted.LoadInt(2).(*types.List)
				if !ok {
					return nil
				}

				itType = returns.Iter() // no need to close (done in ForEach)
				itType.Next()           // skip ListId

				var outputTypeIds []jen.Code
				types.ForEach(itType, func(elem types.Object) bool {
					outputTypeIds = append(outputTypeIds, extractTypeId(elem))
					return true
				})

				switch len(outputTypeIds) {
				case 0:
					// no return
					return funcCode
				case 1:
					// single return type
					return funcCode.Add(outputTypeIds[0])
				default:
					return funcCode.Parens(jen.List(outputTypeIds...))
				}
			}
		}
	}
	return nil
}

// handle (* a) as *a and ([] a b c) as a[b][c]
func extractAssignTarget(env types.Environment, object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		return extractAssignTargetFromList(env, casted)
	}
	return nil
}

func extractAssignTargetFromList(env types.Environment, list *types.List) *jen.Statement {
	if list.Size() > 1 {
		switch op, _ := list.LoadInt(0).(types.Identifier); op {
		case names.LoadId:
			it := list.Iter()
			id, _ := it.Next()
			castedId, _ := id.(types.Identifier)
			code := jen.Id(string(castedId))
			types.ForEach(it, func(elem types.Object) bool {
				code.Index(extractCode(elem.Eval(env)))
				return true
			})
			return code
		case names.StarId:
			return jen.Op(string(op)).Add(extractAssignTarget(env, list.LoadInt(1)))
		}
	}
	return nil
}
