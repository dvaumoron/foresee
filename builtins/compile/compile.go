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

// user can not directly use this kind of id (# start a comment)
const (
	hiddenImportsName = "#imports"
	hiddenPackageName = "#package"

	mainId types.Identifier = "main"
)

var (
	wrappedErrorComment = wrapper{Renderer: jen.Comment("line with error, can't generate correct go code")}

	Builtins = initBuitins()
)

func Compile(l *types.List) types.Object {
	return l.Eval(compileEnvironment{Environment: types.MakeLocalEnvironment(Builtins)})
}

func initBuitins() types.BaseEnvironment {
	base := types.MakeBaseEnvironment()
	base.StoreStr(names.Assign, types.MakeNativeAppliable(assignForm))
	base.StoreStr(names.Block, types.MakeNativeAppliable(blockForm))
	base.StoreStr(string(names.FileId), types.MakeNativeAppliable(fileForm))
	base.StoreStr(names.Package, types.MakeNativeAppliable(packageForm))
	base.StoreStr(names.Var, types.MakeNativeAppliable(varForm))

	// TODO
	return base
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(names.Assign).List(values...)}
	case *types.List:
		var ids []jen.Code
		types.ForEach(casted, func(elem types.Object) bool {
			if id, ok := elem.(types.Identifier); ok {
				ids = append(ids, jen.Id(string(id)))
			}
			return true
		})
		return wrapper{Renderer: jen.List(ids...).Op(names.Assign).List(values...)}
	}
	return wrappedErrorComment
}

func blockForm(env types.Environment, itArgs types.Iterator) types.Object {
	codes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Block(codes...)}
}

func fileForm(env types.Environment, itArgs types.Iterator) types.Object {
	// init default value
	env.StoreStr(hiddenPackageName, mainId)
	env.StoreStr(hiddenImportsName, types.NewList())

	codes := compileToCodeSlice(env, itArgs)

	packageName, _ := env.LoadStr(hiddenPackageName)
	packageNameId, _ := packageName.(types.Identifier)

	jenFile := jen.NewFile(string(packageNameId))

	imports, _ := env.LoadStr(hiddenImportsName)
	importList, _ := imports.(*types.List)
	types.ForEach(importList, func(importDesc types.Object) bool {
		casted, _ := importDesc.(*types.List)
		name, _ := casted.LoadInt(0).(types.String)
		path, _ := casted.LoadInt(1).(types.String)
		switch name {
		case "_":
			jenFile.Anon(string(path))
		case "":
			jenFile.ImportName(string(path), string(name))
		default:
			jenFile.ImportAlias(string(path), string(name))
		}
		return true
	})

	jenFile.Add(codes...)
	return wrapper{Renderer: jenFile}
}

func packageForm(env types.Environment, itArgs types.Iterator) types.Object {
	packageName, _ := itArgs.Next()
	switch casted := packageName.(type) {
	case types.Identifier:
		env.StoreStr(hiddenPackageName, casted)
	case types.String:
		env.StoreStr(hiddenPackageName, types.Identifier(casted))
	}
	return types.None
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	if len(values) == 0 {
		if list, ok := arg0.(*types.List); ok && list.Size() > 2 {
			varId, _ := list.LoadInt(1).(types.Identifier)
			typeStmt := extractComplexId(list.LoadInt(2))
			return wrapper{Renderer: jen.Var().Id(string(varId)).Add(typeStmt)}
		}
		return wrappedErrorComment
	}

	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(names.Var).List(values...)}
	case *types.List:
		// test if it's a:b instead of (a:b c:d)
		if firstId, _ := casted.LoadInt(0).(types.Identifier); firstId == names.ListId {
			if casted.Size() < 2 {
				return wrappedErrorComment
			}

			varId, _ := casted.LoadInt(1).(types.Identifier)
			varCode := jen.Id(string(varId)).Op(names.Var)
			if typeStmt := extractComplexId(casted.LoadInt(2)); typeStmt == nil {
				varCode.List(values...)
			} else {
				varCode.Add(typeStmt).Call(values...)
			}
			return wrapper{Renderer: varCode}
		}

		var varIds []jen.Code
		var typeIds []*jen.Statement
		types.ForEach(casted, func(varDesc types.Object) bool {
			switch casted2 := varDesc.(type) {
			case types.Identifier:
				varIds = append(varIds, jen.Id(string(casted2)))
				typeIds = append(typeIds, nil)
			case *types.List:
				// assume it's in a:b format
				varId, _ := casted2.LoadInt(1).(types.Identifier)
				varIds = append(varIds, jen.Id(string(varId)))
				typeIds = append(typeIds, extractComplexId(casted2.LoadInt(2)))
			}
			return true
		})

		// add cast calls
		for index, typeId := range typeIds {
			if typeId != nil {
				values[index] = typeId.Call(values[index])
			}
		}
		return wrapper{Renderer: jen.List(varIds...).Op(names.Var).List(values...)}
	}
	return wrappedErrorComment
}

func compileToCodeSlice(env types.Environment, instructions types.Iterable) []jen.Code {
	var codes []jen.Code
	types.ForEach(instructions, func(instruction types.Object) bool {
		codes = append(codes, extractCode(instruction.Eval(env)))
		return true
	})
	return codes
}

// manage wrapper and literals
func extractCode(object types.Object) jen.Code {
	switch casted := object.(type) {
	case wrapper:
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
	}
	return jen.Empty()
}

// handle *a, []a and &a format
func extractComplexId(object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		if casted.Size() > 1 {
			op, _ := casted.LoadInt(0).(types.Identifier)
			return jen.Op(string(op)).Add(extractComplexId(casted.LoadInt(1)))
		}
	}
	return nil
}
