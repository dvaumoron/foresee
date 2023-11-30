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

func compileToCodeSlice(env types.Environment, iterable types.Iterable) []jen.Code {
	var codes []jen.Code
	types.ForEach(iterable, func(elem types.Object) bool {
		codes = append(codes, compileToCode(env, elem))
		return true
	})
	return codes
}

func compileToCode(env types.Environment, object types.Object) Renderer {
	return handleBasicType(object, true, func(object types.Object) Renderer {
		switch casted := object.Eval(env).(type) {
		case appliableWrapper:
			return casted.Renderer
		case literalWrapper:
			return casted.Renderer
		case wrapper:
			return casted.Renderer
		default:
			return handleBasicType(casted, false, emptyCode)
		}
	})
}

func handleBasicType(object types.Object, noneToNil bool, defaultCase func(types.Object) Renderer) Renderer {
	switch casted := object.(type) {
	case types.NoneType:
		if noneToNil {
			return jen.Nil()
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
	return defaultCase(object)
}

func emptyCode(object types.Object) Renderer {
	return jen.Empty()
}

func extractSliceIndexes(env types.Environment, object types.Object) []jen.Code {
	casted, _ := object.(*types.List)
	itCasted := casted.Iter()
	defer itCasted.Close()

	arg0, _ := itCasted.Next()
	// detect slice (could be a classic function/operator call)
	if header, _ := arg0.(types.Identifier); header == names.ListId {
		return compileToCodeSlice(env, itCasted)
	}
	return []jen.Code{compileToCode(env, object)}
}

// handle *type, []type, map[t1]t2 format and func or chan types  (and their combinations like [][]*type)
// TODO manage anonymous struct ?
func extractType(object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		switch casted.Size() {
		case 2:
			switch op, _ := casted.LoadInt(0).(types.Identifier); op {
			case names.ArrowChanId:
				return jen.Op(names.Arrow).Chan().Add(extractType(casted.LoadInt(1)))
			case names.ChanArrowId:
				return jen.Chan().Op(names.Arrow).Add(extractType(casted.LoadInt(1)))
			case names.ChanId:
				return jen.Chan().Add(extractType(casted.LoadInt(1)))
			case names.LoadId:
				return jen.Index().Add(extractType(casted.LoadInt(1)))
			case names.FuncId:
				params, ok := casted.LoadInt(1).(*types.List)
				if !ok {
					return nil
				}

				itType := params.Iter() // no need to close (done in ForEach)
				itType.Next()           // skip ListId

				var typeIds []jen.Code
				types.ForEach(itType, func(elem types.Object) bool {
					typeIds = append(typeIds, extractType(elem))
					return true
				})
				return jen.Func().Params(typeIds...)
			case names.AmpersandId, names.EllipsisId, names.StarId:
				return jen.Op(string(op)).Add(extractType(casted.LoadInt(1)))
			}
		case 3:
			switch op, _ := casted.LoadInt(0).(types.Identifier); op {
			case names.LoadId:
				// manage [size]type
				switch castedSize := casted.LoadInt(1).(type) {
				case types.Integer:
					return jen.Index(jen.Lit(int(castedSize))).Add(extractType(casted.LoadInt(2)))
				case types.Identifier:
					// size should be "..." (or it's not a type representation and that should be handled elsewhere)
					if castedSize == names.EllipsisId {
						return jen.Index(jen.Op(string(names.EllipsisId))).Add(extractType(casted.LoadInt(2)))
					}
				}
			case names.MapId:
				// manage map[t1]t2
				return jen.Op(string(op)).Add(extractType(casted.LoadInt(1))).Add(extractType(casted.LoadInt(2)))
			case names.FuncId:
				param, ok := casted.LoadInt(1).(*types.List)
				if !ok {
					return nil
				}

				itType := param.Iter() // no need to close (done in ForEach)
				itType.Next()          // skip ListId

				var typeIds []jen.Code
				types.ForEach(itType, func(elem types.Object) bool {
					typeIds = append(typeIds, extractType(elem))
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
					outputTypeIds = append(outputTypeIds, extractType(elem))
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

// handle "a" as a,  "(* a)" as *a and ([] a b c) as a[b][c]
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
	switch op, _ := list.LoadInt(0).(types.Identifier); op {
	case names.LoadId:
		it := list.Iter()
		defer it.Close()

		it.Next() // skip LoadId
		id, _ := it.Next()
		index, ok := it.Next()
		if !ok {
			return nil
		}

		castedId, _ := id.(types.Identifier)
		code := jen.Id(string(castedId)).Index(compileToCode(env, index)) // can not be slicing
		types.ForEach(it, func(elem types.Object) bool {
			code.Index(compileToCode(env, elem)) // can not be slicing
			return true
		})
		return code
	case names.StarId:
		if list.Size() > 1 {
			return jen.Op(string(op)).Add(extractAssignTarget(env, list.LoadInt(1)))
		}
	}
	return nil
}
