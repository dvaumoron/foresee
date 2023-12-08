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
		case callableWrapper:
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

// handle *type, []type, map[t1]t2, t1[t2,t3] format and func or chan types  (and their combinations like [][]*type)
// TODO manage anonymous struct ?
func extractType(env types.Environment, object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		return extractTypeFromList(env, casted)
	}
	return nil
}

func extractTypeFromList(env types.Environment, casted *types.List) *jen.Statement {
	switch casted.Size() {
	case 2:
		switch op, _ := casted.LoadInt(0).(types.Identifier); op {
		case names.ArrowChanId:
			return buildParameterizedType(env, jen.Op(names.Arrow).Chan(), casted)
		case names.ChanArrowId:
			return buildParameterizedType(env, jen.Chan().Op(names.Arrow), casted)
		case names.ChanId:
			return buildParameterizedType(env, jen.Chan(), casted)
		case names.SliceId:
			return buildParameterizedType(env, jen.Index(), casted)
		case names.FuncId:
			params, ok := casted.LoadInt(1).(*types.List)
			if !ok {
				return nil
			}

			if typeCodes, ok := extractTypes(env, params); ok {
				return jen.Func().Params(typeCodes...)
			}
		case names.EllipsisId, names.StarId, names.TildeId:
			return buildParameterizedType(env, jen.Op(string(op)), casted)
		}
	case 3:
		switch op, _ := casted.LoadInt(0).(types.Identifier); op {
		case names.FuncId:
			params, ok := casted.LoadInt(1).(*types.List)
			returns, ok2 := casted.LoadInt(2).(*types.List)
			if !(ok && ok2) {
				return nil
			}

			typeCodes, ok := extractTypes(env, params)
			if !ok {
				return nil
			}

			outputTypeIds, ok := extractTypes(env, returns)
			if !ok {
				return nil
			}

			funcCode := jen.Func().Params(typeCodes...)
			switch len(outputTypeIds) {
			case 0:
				// no return
				return funcCode
			case 1:
				// single return type
				return funcCode.Add(outputTypeIds[0])
			}
			return funcCode.Parens(jen.List(outputTypeIds...))
		case names.GenId:
			return extractGenType(env, casted.LoadInt(1), casted.LoadInt(2))
		case names.GetId:
			return extractQualified(env, casted.LoadInt(1), casted.LoadInt(2))
		case names.MapId:
			// manage map[t1]t2
			return jen.Map(extractType(env, casted.LoadInt(1))).Add(extractType(env, casted.LoadInt(2)))
		case names.SliceId:
			return extractArrayType(env, casted.LoadInt(1), casted.LoadInt(2))
		}
	}
	return nil
}

func extractNameOrQualified(env types.Environment, object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		if header, _ := casted.LoadInt(0).(types.Identifier); header == names.Dot || header == names.GetId {
			return extractQualified(env, casted.LoadInt(1), casted.LoadInt(2))
		}
	}
	return nil
}

func buildParameterizedType(env types.Environment, base *jen.Statement, list *types.List) *jen.Statement {
	typeCode := extractType(env, list.LoadInt(1))
	if typeCode != nil {
		return base.Add(typeCode)
	}
	return nil
}

func extractQualified(env types.Environment, arg0 types.Object, arg1 types.Object) *jen.Statement {
	packagePath := ""
	nameId, _ := arg1.(types.Identifier)
	switch casted := arg0.(type) {
	case types.Identifier:
		imports, _ := env.LoadStr(hiddenImportsName)
		castedImport, _ := imports.(types.BaseEnvironment)
		path, ok := castedImport.LoadStr(string(casted))
		if !ok {
			return nil // not a type
		}

		castedPath, _ := path.(types.String)
		packagePath = string(castedPath)
	case types.String:
		packagePath = string(casted)
	default:
		return nil // not a type
	}
	return jen.Qual(string(packagePath), string(nameId))
}

func extractArrayType(env types.Environment, arg0 types.Object, arg1 types.Object) *jen.Statement {
	switch casted := arg0.(type) {
	case types.Integer:
		return jen.Index(jen.Lit(int(casted))).Add(extractType(env, arg1))
	case types.Identifier:
		if casted == names.EllipsisId {
			// array type with automatic count
			return jen.Index(jen.Op(string(names.EllipsisId))).Add(extractType(env, arg1))
		}
	}
	return nil
}

func extractGenType(env types.Environment, arg0 types.Object, arg1 types.Object) *jen.Statement {
	typeCode := extractType(env, arg0)
	genTypes, _ := arg1.(*types.List)
	// type with generic parameter
	if typeCodes, ok := extractTypes(env, genTypes); typeCode != nil && ok {
		return typeCode.Types(typeCodes...)
	}
	return nil
}

// skip first elem (should be ListId)
func extractTypes(env types.Environment, typeIterable types.Iterable) ([]jen.Code, bool) {
	itType := typeIterable.Iter() // no need to close (done in ForEach)
	itType.Next()                 // skip ListId

	noError := false
	var typeCodes []jen.Code
	types.ForEach(itType, func(elem types.Object) bool {
		typeCode := extractType(env, elem)
		if noError = typeCode != nil; noError {
			typeCodes = append(typeCodes, typeCode)
		}
		return noError
	})
	return typeCodes, noError
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
	case names.Load:
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
