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
	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func blockForm(env types.Environment, itArgs types.Iterator) types.Object {
	intructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Block(intructionCodes...)}
}

func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Const())
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
		name, _ := casted.LoadInt(0).(types.Identifier)
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

func forForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	var condCodes []jen.Code
	switch casted := arg0.(type) {
	case types.Identifier:
		condCodes = []jen.Code{jen.Id(string(casted))}
	case types.Boolean:
		if !casted {
			condCodes = []jen.Code{jen.False()}
		}
	case types.NoneType:
		// let the condition empty
	case *types.List:
		if casted.Size() != 0 {
			switch casted.LoadInt(0).(type) {
			case types.Identifier:
				// could be a range or a condition
				condCodes = []jen.Code{compileToCode(env, casted)}
			case *types.List:
				condCodes = compileToCodeSlice(env, casted)
				for i := len(condCodes); i < 3; i++ {
					condCodes = append(condCodes, jen.Empty())
				}
			default:
				return wrappedErrorComment
			}
		}
	default:
		return wrappedErrorComment
	}

	instructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.For(condCodes...).Block(instructionCodes...)}
}

func funcForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	funcCode := jen.Func()
	if nameCode := extractNameWithGeneric(arg0); nameCode == nil {
		casted, ok := arg0.(*types.List)
		if !ok {
			return wrappedErrorComment
		}

		var receiverCode *jen.Statement
		if casted.Size() > 1 {
			receiverId, _ := casted.LoadInt(0).(types.Identifier)
			receiverCode = jen.Id(string(receiverId)).Add(extractType(casted.LoadInt(1)))
		} else {
			receiverCode = extractType(casted.LoadInt(0))
		}

		arg1, _ := itArgs.Next()
		methodId, _ := arg1.(types.Identifier)
		funcCode.Parens(receiverCode).Id(string(methodId))
	} else {
		funcCode.Add(nameCode)
	}

	argN, _ := itArgs.Next()
	params, ok := argN.(*types.List)
	if !ok {
		return wrappedErrorComment
	}

	var paramCodes []jen.Code
	types.ForEach(params, func(elem types.Object) bool {
		// assume it's in "name:type" format (type should be inferred if not declared)
		paramDesc, _ := elem.(*types.List)
		varId, _ := paramDesc.LoadInt(1).(types.Identifier)
		paramCodes = append(paramCodes, jen.Id(string(varId)).Add(extractType(paramDesc.LoadInt(2))))
		return true
	})
	funcCode.Params(paramCodes...)

	var instructionCodes []jen.Code
	argN, _ = itArgs.Next()
	switch casted3 := argN.(type) {
	case types.Identifier:
		funcCode.Id(string(casted3))
	case *types.List:
		if head, _ := casted3.LoadInt(0).(types.Identifier); head == names.ListId {
			var typeCodes []jen.Code
			types.ForEach(casted3, func(elem types.Object) bool {
				typeCodes = append(typeCodes, extractType(elem))
				return true
			})
			funcCode.Parens(jen.List(typeCodes...))
		} else {
			if typeCode := extractType(argN); typeCode == nil {
				// can not extract type, so argN is the first instruction of the code block
				instructionCodes = []jen.Code{compileToCode(env, argN)}
			} else {
				funcCode.Add(typeCode)
			}
		}
	}

	instructionCodesTemp := compileToCodeSlice(env, itArgs)
	instructionCodes = append(instructionCodes, instructionCodesTemp...)
	return wrapper{Renderer: funcCode.Block(instructionCodes...)}
}

func ifForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, _ := itArgs.Next()
	instruction1, ok := arg1.(*types.List)
	if !ok {
		return wrappedErrorComment
	}

	var condCodes []jen.Code
	switch casted := arg0.(type) {
	case types.Identifier:
		condCodes = []jen.Code{jen.Id(string(casted))}
	case types.Boolean:
		condCodes = []jen.Code{jen.Lit(bool(casted))}
	case *types.List:
		if casted.Size() != 0 {
			switch casted.LoadInt(0).(type) {
			case types.Identifier:
				// normal condition
				condCodes = []jen.Code{compileToCode(env, casted)}
			case *types.List:
				// start with an instruction
				condCodes = compileToCodeSlice(env, casted)
			default:
				return wrappedErrorComment
			}
		}
	default:
		return wrappedErrorComment
	}

	ifCode := jen.If(condCodes...)
	if header, _ := instruction1.LoadInt(0).(types.Identifier); header == names.Block {
		ifCode.Add(compileToCode(env, arg1))
	} else {
		ifCode.Block(compileToCode(env, arg1))
	}

	if arg2, ok := itArgs.Next(); ok {
		instruction2, ok := arg2.(*types.List)
		if !ok {
			return wrappedErrorComment
		}

		ifCode.Else()
		if header, _ := instruction2.LoadInt(0).(types.Identifier); header == names.Block {
			ifCode.Add(compileToCode(env, arg2))
		} else {
			ifCode.Block(compileToCode(env, arg2))
		}
	}
	return wrapper{Renderer: ifCode}
}

func importForm(env types.Environment, itArgs types.Iterator) types.Object {
	imports, _ := env.LoadStr(hiddenImportsName)
	importList, _ := imports.(*types.List)
	types.ForEach(itArgs, func(importDesc types.Object) bool {
		switch casted := importDesc.(type) {
		case *types.List:
			if casted.Size() > 1 {
				importList.Add(casted)
			} else {
				importList.Add(types.NewList(types.Identifier(""), casted.LoadInt(0)))
			}
			return true
		case types.Identifier:
			path, _ := itArgs.Next()
			importList.Add(types.NewList(casted, path))
		case types.String:
			importList.Add(types.NewList(types.Identifier(""), casted))
		}
		return false // onliner cases so break
	})
	return types.None
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

func rangeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Range().Add(compileToCode(env, arg0))}
}

func returnForm(env types.Environment, itArgs types.Iterator) types.Object {
	valueCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Return(valueCodes...)}
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Var())
}
