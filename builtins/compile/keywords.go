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
	return appliableWrapper{Renderer: jen.Block(codes...)}
}

// TODO manage multiline
func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	value, ok := compileToCode(env, itArgs)
	switch casted := arg0.(type) {
	case types.Identifier:
		if !ok {
			// no declared type and no value
			return wrappedErrorComment
		}
		return appliableWrapper{Renderer: jen.Const().Id(string(casted)).Op(names.Assign).Add(value)}
	case *types.List:
		// expect a:b
		if firstId, _ := casted.LoadInt(0).(types.Identifier); firstId != names.ListId {
			return wrappedErrorComment
		}

		constId, _ := casted.LoadInt(1).(types.Identifier)
		constCode := jen.Const().Id(string(constId))
		if typeId := extractType(casted.LoadInt(2)); typeId != nil {
			constCode.Add(typeId)
		}
		if ok {
			constCode.Op(names.Assign).Add(value)
		}
		return appliableWrapper{Renderer: constCode}
	}
	return wrappedErrorComment
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
	return appliableWrapper{Renderer: jenFile}
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
	if ok {
		return wrappedErrorComment
	}

	var paramCodes []jen.Code
	types.ForEach(params, func(elem types.Object) bool {
		// assume it's in a:b format
		paramDesc, _ := elem.(*types.List)
		varId, _ := paramDesc.LoadInt(1).(types.Identifier)
		paramCodes = append(paramCodes, jen.Id(string(varId)).Add(extractType(paramDesc.LoadInt(2))))
		return true
	})
	funcCode.Params(paramCodes...)

	var codes []jen.Code
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
				codes = []jen.Code{extractCode(argN.Eval(env))}
			} else {
				funcCode.Add(typeCode)
			}
		}
	}

	codes = append(codes, compileToCodeSlice(env, itArgs)...)
	return appliableWrapper{Renderer: funcCode.Block(codes...)}
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
	return appliableWrapper{Renderer: jen.Return(codes...)}
}
