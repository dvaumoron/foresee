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

func assertForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: compileToCode(env, arg0).Assert(extractType(arg1))}
}

func blockForm(env types.Environment, itArgs types.Iterator) types.Object {
	intructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Block(intructionCodes...)}
}

func breakForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processLabellable(env, itArgs, jen.Break())
}

func caseForm(env types.Environment, itArgs types.Iterator) types.Object {
	var condCodes []jen.Code
	if arg0, ok := itArgs.Next(); ok {
		condCodes = extractValueOrMultiple(env, arg0)
	}
	if len(condCodes) == 0 {
		return wrappedErrorComment
	}

	instructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Case(condCodes...).Op(names.Colon).Add(instructionCodes...)}
}

func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Const())
}

func defaultForm(env types.Environment, itArgs types.Iterator) types.Object {
	intructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Default().Op(names.Colon).Add(intructionCodes...)}
}

func deferForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Defer().Add(compileToCode(env, arg0))}
}

func continueForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processLabellable(env, itArgs, jen.Continue())
}

func fallthroughForm(env types.Environment, itArgs types.Iterator) types.Object {
	return wrapper{Renderer: jen.Fallthrough()}
}

func fileForm(env types.Environment, itArgs types.Iterator) types.Object {
	// init default value
	env.StoreStr(hiddenPackageName, mainId)
	env.StoreStr(hiddenImportsName, types.NewList())

	codes := compileToCodeSlice(env, itArgs)

	packageName, _ := env.LoadStr(hiddenPackageName)
	packageNameId, _ := packageName.(types.Identifier)

	jenFile := jen.NewFile(string(packageNameId))
	jenFile.Add(codes...)

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
			jenFile.ImportName(string(path), "")
		default:
			jenFile.ImportAlias(string(path), string(name))
		}
		return true
	})
	return wrapper{Renderer: jenFile}
}

func forForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	var condCodes []jen.Code
	switch casted := arg0.(type) {
	case types.Identifier:
		condCodes = []jen.Code{jen.Id(string(casted))}
	case types.Boolean:
		condCodes = []jen.Code{jen.Lit(bool(casted))}
	case types.NoneType:
		// let the condition empty
	case *types.List:
		if casted.Size() != 0 {
			condCodes = extractSingleOrMultiple(env, casted)
			if i := len(condCodes); i != 0 {
				for ; i < 3; i++ {
					condCodes = append(condCodes, jen.Empty())
				}
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
	if nameCode := extractNameWithGenericDef(arg0); nameCode == nil {
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
		methodId, _ := arg1.(types.Identifier) // no need to handle generic on method
		funcCode.Parens(receiverCode).Id(string(methodId))
	} else {
		funcCode.Add(nameCode)
	}

	params, _ := itArgs.Next()
	paramCodes, ok := extractParameter(params)
	if !ok {
		return wrappedErrorComment
	}

	funcCode.Params(paramCodes...)

	argN, _ := itArgs.Next()
	returnCode, instructionCodes := extractReturnType(env, argN)
	if returnCode != nil {
		funcCode.Add(returnCode)
	}

	instructionCodesTemp := compileToCodeSlice(env, itArgs)
	instructionCodes = append(instructionCodes, instructionCodesTemp...)
	return wrapper{Renderer: funcCode.Block(instructionCodes...).Line()}
}

func getForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, _ := itArgs.Next()
	fieldId, ok := arg1.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}

	getCode := compileToCode(env, arg0).Dot(string(fieldId))
	types.ForEach(itArgs, func(elem types.Object) bool {
		fieldId, _ := elem.(types.Identifier)
		getCode.Dot(string(fieldId))
		return true
	})

	// returned value could be callable
	return callableWrapper{Renderer: getCode}
}

func goForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Go().Add(compileToCode(env, arg0))}
}

func gotoForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	labelId, ok := arg0.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Goto().Id(string(labelId))}
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
			condCodes = extractSingleOrMultiple(env, casted)
		}
	}
	if len(condCodes) == 0 {
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

func labelForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	labelId, ok := arg0.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Id(string(labelId)).Op(":")}
}

func lambdaForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	paramCodes, ok := extractParameter(arg0)
	if !ok {
		return wrappedErrorComment
	}

	funcCode := jen.Func().Params(paramCodes...)

	arg1, _ := itArgs.Next()
	returnCode, instructionCodes := extractReturnType(env, arg1)
	if returnCode != nil {
		funcCode.Add(returnCode)
	}

	instructionCodesTemp := compileToCodeSlice(env, itArgs)
	instructionCodes = append(instructionCodes, instructionCodesTemp...)
	return callableWrapper{Renderer: funcCode.Block(instructionCodes...)}
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

func selectForm(env types.Environment, itArgs types.Iterator) types.Object {
	caseCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Select().Block(caseCodes...)}
}

func switchForm(env types.Environment, itArgs types.Iterator) types.Object {
	var condCodes []jen.Code
	if arg0, ok := itArgs.Next(); ok {
		condCodes = extractValueOrMultiple(env, arg0)
	}

	caseCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Switch(condCodes...).Block(caseCodes...)}
}

func typeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	typeCode := jen.Type().Add(extractNameWithGenericDef(arg0))

	switch oldType := arg1.(type) {
	case types.Identifier:
		switch oldType {
		case names.Interface:
			var defCodes []jen.Code
			types.ForEach(itArgs, func(elem types.Object) bool {
				casted, _ := elem.(*types.List)
				var defCode *jen.Statement
				switch casted2 := casted.LoadInt(0).(type) {
				case types.Identifier:
					switch casted2 {
					case names.Dot, names.GetId:
						// qualified name of another interface
						defCode = extractQualified(casted)
					case names.LoadId:
						if genTypes, ok := casted.LoadInt(2).(*types.List); ok {
							// qualified type with generic parameter
							typeCodes, _ := extractTypes(genTypes)
							defCode = extractNameOrQualified(casted.LoadInt(1)).Types(typeCodes...)
						}
					case names.TildeId:
						defCode = jen.Op(string(names.TildeId)).Add(extractType(casted.LoadInt(1)))
					default:
						// method description
						paramTypes, _ := casted.LoadInt(1).(*types.List)
						// do not skip first (!= extractTypes)
						var paramCodes []jen.Code
						types.ForEach(paramTypes, func(elem types.Object) bool {
							paramCodes = append(paramCodes, extractType(elem))
							return true
						})

						defCode = jen.Id(string(casted2)).Params(paramCodes...)
						if typeCode := extractType(casted.LoadInt(2)); typeCode != nil {
							defCode.Add(typeCode)
						}
					}
				case *types.List:
					// land here with syntaxic sugar
					switch header, _ := casted2.LoadInt(0).(types.Identifier); header {
					case names.Dot, names.GetId:
						// qualified name of another interface
						defCode = extractQualified(casted2)
					case names.LoadId:
						if genTypes, ok := casted2.LoadInt(2).(*types.List); ok {
							// qualified type with generic parameter
							typeCodes, _ := extractTypes(genTypes)
							defCode = extractNameOrQualified(casted2.LoadInt(1)).Types(typeCodes...)
						}
					case names.TildeId:
						first := true
						types.ForEach(casted, func(elem types.Object) bool {
							casted3, _ := elem.(*types.List)
							if first {
								first = false
								defCode = jen.Op(string(names.TildeId)).Add(extractType(casted3.LoadInt(1)))
							} else {
								defCode.Op(names.Pipe).Op(string(names.TildeId)).Add(extractType(casted3.LoadInt(1)))
							}
							return true // handle several by line
						})
					}
				}

				if defCode != nil {
					defCodes = append(defCodes, defCode)
				}
				return true
			})
			typeCode.Interface(defCodes...)
		case names.Struct:
			var defCodes []jen.Code
			types.ForEach(itArgs, func(elem types.Object) bool {
				casted, _ := elem.(*types.List)
				fieldId, _ := casted.LoadInt(0).(types.Identifier)
				defCode := jen.Id(string(fieldId)).Add(extractType(casted.LoadInt(1)))
				if casted.Size() > 2 {
					itemList, _ := elem.(*types.List)
					items := map[string]string{}
					types.ForEach(itemList, func(item types.Object) bool {
						casted, _ := elem.(*types.List)
						key := casted.LoadInt(1).(types.String)
						value := casted.LoadInt(2).(types.String)
						items[string(key)] = string(value)
						return true
					})
					defCode.Tag(items)
				}
				defCodes = append(defCodes, defCode)
				return true
			})
			typeCode.Struct(defCodes...)
		default:
			typeCode.Id(string(oldType))
		}
	case *types.List:
		typeCode.Add(extractTypeFromList(oldType))
	default:
		return wrappedErrorComment
	}
	return wrapper{Renderer: typeCode.Line()}
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Var())
}
