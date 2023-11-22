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

// user can not directly use this kind of id (# start a comment)
const (
	hiddenImportsName = "#imports"
	hiddenPackageName = "#package"
)

var Builtins = initBuitins()

func initBuitins() types.BaseEnvironment {
	base := types.MakeBaseEnvironment()
	base.StoreStr(names.Block, types.MakeNativeAppliable(blocForm))
	base.StoreStr(string(names.FileId), types.MakeNativeAppliable(fileForm))
	base.StoreStr(names.Package, types.MakeNativeAppliable(packageForm))

	// TODO
	return base
}

func blocForm(env types.Environment, itArgs types.Iterator) types.Object {
	codes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Block(codes...)}
}

func fileForm(env types.Environment, itArgs types.Iterator) types.Object {
	// init default value
	env.StoreStr(hiddenPackageName, types.String("main"))
	env.StoreStr(hiddenImportsName, types.NewList())

	codes := compileToCodeSlice(env, itArgs)

	packageName, _ := env.LoadStr(hiddenPackageName)
	packageNameStr, _ := packageName.(types.String)

	jenFile := jen.NewFile(string(packageNameStr))

	imports, ok := env.LoadStr(hiddenImportsName)
	if ok {
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
	}

	jenFile.Add(codes...)
	return wrapper{Renderer: jenFile}
}

func packageForm(env types.Environment, itArgs types.Iterator) types.Object {
	packageName, _ := itArgs.Next()
	switch casted := packageName.(type) {
	case types.Identifier:
		env.StoreStr(hiddenPackageName, types.String(casted))
	case types.String:
		env.StoreStr(hiddenPackageName, casted)
	}
	return types.None
}

func compileToCodeSlice(env types.Environment, instructions types.Iterable) []jen.Code {
	var codes []jen.Code
	types.ForEach(instructions, func(instruction types.Object) bool {
		evalued := instruction.Eval(env)
		if wrapped, ok := evalued.(wrapper); ok {
			if code, ok := wrapped.Renderer.(jen.Code); ok {
				codes = append(codes, code)
			}
		}
		return true
	})
	return codes
}
