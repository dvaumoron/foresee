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
	codes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Block(codes...)}
}

func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Const())
}

// baseCode is not cloned (must generate a new one on each call)
func processDef(env types.Environment, itArgs types.Iterator, baseCode *jen.Statement) types.Object {
	// K is supposed to be "const" or "var" (passed as baseCode)
	arg0, _ := itArgs.Next()
	switch casted := arg0.(type) {
	case types.Identifier:
		// detect "K name value"
		arg1, ok := itArgs.Next()
		if !ok {
			// no declared type and no value
			return wrappedErrorComment
		}
		return wrapper{Renderer: baseCode.Id(string(casted)).Op(names.Assign).Add(compileToCode(env, arg1))}
	case *types.List:
		switch casted2 := casted.LoadInt(0).(type) {
		case types.Identifier:
			// detect "K name:type value"
			if casted2 == names.ListId {
				constId, _ := casted.LoadInt(1).(types.Identifier)
				defCode := baseCode.Id(string(constId))
				if typeId := extractType(casted.LoadInt(2)); typeId != nil {
					defCode.Add(typeId)
				}
				arg1, ok := itArgs.Next()
				if ok {
					defCode.Op(names.Assign).Add(compileToCode(env, arg1))
				}
				return wrapper{Renderer: defCode}
			}

			if casted.Size() > 1 {
				// "K (name value)"
				valueCode := compileToCode(env, casted.LoadInt(1))
				defCodes := []jen.Code{jen.Id(string(casted2)).Op(names.Assign).Add(valueCode)}

				// following lines
				defCodes = processDefLines(env, itArgs, defCodes)

				return wrapper{Renderer: baseCode.Defs(defCodes...)}
			}
		case *types.List:
			// "K (name:type value)"
			constId, _ := casted2.LoadInt(1).(types.Identifier)
			typeCode := extractType(casted2.LoadInt(2))
			valueCode := compileToCode(env, casted.LoadInt(1))
			defCodes := []jen.Code{jen.Id(string(constId)).Add(typeCode).Op(names.Assign).Add(valueCode)}

			// following lines
			defCodes = processDefLines(env, itArgs, defCodes)

			return wrapper{Renderer: baseCode.Defs(defCodes...)}
		}
	}
	return wrappedErrorComment
}

// handle "(name value)" or "(name:type)" or "(name:type value)"
func processDefLines(env types.Environment, itArgs types.Iterable, defCodes []jen.Code) []jen.Code {
	types.ForEach(itArgs, func(elem types.Object) bool {
		defDesc, _ := elem.(*types.List)
		var defCode *jen.Statement
		switch casted := defDesc.LoadInt(0).(type) {
		case types.Identifier:
			defCode = jen.Id(string(casted))
		case *types.List:
			constId, _ := casted.LoadInt(1).(types.Identifier)
			defCode = jen.Id(string(constId)).Add(extractType(casted.LoadInt(2)))
		}
		if defDesc.Size() > 1 {
			defCode.Op(names.Assign).Add(compileToCode(env, defDesc.LoadInt(1)))
		}
		defCodes = append(defCodes, defCode)
		return true
	})
	return defCodes
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

func funcForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	funcCode := jen.Func()
	switch casted := arg0.(type) {
	case types.Identifier:
		funcCode.Id(string(casted))
	case *types.List:
		var receiverCode *jen.Statement
		if castedReceiver, _ := casted.LoadInt(0).(*types.List); castedReceiver.Size() > 1 {
			receiverId, _ := castedReceiver.LoadInt(0).(types.Identifier)
			receiverCode = jen.Id(string(receiverId)).Add(extractType(castedReceiver.LoadInt(1)))
		} else {
			receiverCode = extractType(castedReceiver.LoadInt(0))
		}

		arg1, _ := itArgs.Next()
		methodId, _ := arg1.(types.Identifier)
		funcCode.Parens(receiverCode).Id(string(methodId))
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

func returnForm(env types.Environment, itArgs types.Iterator) types.Object {
	codes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Return(codes...)}
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Var())
}
