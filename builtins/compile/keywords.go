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
	return wrapper{Renderer: compileToCode(env, arg0).Assert(extractType(env, arg1))}
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
	return wrapper{Renderer: jen.Case(condCodes...).Add(instructionCodes...)}
}

func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Const())
}

func defaultForm(env types.Environment, itArgs types.Iterator) types.Object {
	intructionCodes := compileToCodeSlice(env, itArgs)
	return wrapper{Renderer: jen.Default().Add(intructionCodes...)}
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
	env.StoreStr(hiddenImportsName, types.MakeBaseEnvironment())

	codes := compileToCodeSlice(env, itArgs)

	packageName, _ := env.LoadStr(hiddenPackageName)
	packageNameId, _ := packageName.(types.Identifier)

	jenFile := jen.NewFile(string(packageNameId))
	jenFile.Add(codes...)

	imports, _ := env.LoadStr(hiddenImportsName)
	castedImport, _ := imports.(types.BaseEnvironment)
	types.ForEach(castedImport, func(importDesc types.Object) bool {
		casted, _ := importDesc.(*types.List)
		name, _ := casted.LoadInt(0).(types.String)
		path, _ := casted.LoadInt(1).(types.String)
		if name == "_" {
			jenFile.Anon(string(path))
		} else {
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
			if i := len(condCodes); i > 1 {
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
	if nameCode := extractNameWithGenericDef(env, arg0); nameCode == nil {
		casted, ok := arg0.(*types.List)
		if !ok {
			return wrappedErrorComment
		}

		var receiverCode *jen.Statement
		if casted.Size() > 1 {
			receiverId, _ := casted.LoadInt(0).(types.Identifier)
			receiverCode = jen.Id(string(receiverId)).Add(extractType(env, casted.LoadInt(1)))
		} else {
			receiverCode = extractType(env, casted.LoadInt(0))
		}

		arg1, _ := itArgs.Next()
		methodId, _ := arg1.(types.Identifier) // no need to handle generic on method
		funcCode.Parens(receiverCode).Id(string(methodId))
	} else {
		funcCode.Add(nameCode)
	}

	params, _ := itArgs.Next()
	paramCodes, ok := extractParameter(env, params)
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

func genTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return literalWrapper{Renderer: extractGenType(env, arg0, arg1)}
}

func getForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, _ := itArgs.Next()
	fieldId, ok := arg1.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}

	getCode := extractQualified(env, arg0, arg1)
	if getCode == nil {
		getCode = compileToCode(env, arg0).Dot(string(fieldId))
	}

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
	castedImport, _ := imports.(types.BaseEnvironment)
	types.ForEach(itArgs, func(importDesc types.Object) bool {
		switch casted := importDesc.(type) {
		case *types.List:
			if casted.Size() > 1 {
				packageId, _ := casted.LoadInt(1).(types.Identifier)
				castedImport.StoreStr(string(packageId), casted.LoadInt(1))
			} else {
				path, _ := casted.LoadInt(0).(types.String)
				castedImport.StoreStr(extractPackageName(path), path)
			}
			return true
		case types.Identifier:
			path, _ := itArgs.Next()
			castedImport.StoreStr(string(casted), path)
		case types.String:
			castedImport.StoreStr(extractPackageName(casted), casted)
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
	paramCodes, ok := extractParameter(env, arg0)
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

func literalForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return literalWrapper{Renderer: extractType(env, arg0)}
}

func mapTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return literalWrapper{Renderer: jen.Map(extractType(env, arg0)).Add(extractType(env, arg1))}
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

func sliceOrArrayTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if ok {
		typeCode := extractArrayType(env, arg0, arg1)
		if typeCode == nil {
			return wrappedErrorComment
		}
		return literalWrapper{Renderer: typeCode}
	}

	typeCode := extractType(env, arg0)
	if typeCode == nil {
		return wrappedErrorComment
	}
	return literalWrapper{Renderer: jen.Index().Add(typeCode)}
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

	typeCode := jen.Type().Add(extractNameWithGenericDef(env, arg0))

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
					case names.GenId:
						defCode = extractGenType(env, casted.LoadInt(1), casted.LoadInt(2))
					case names.GetId:
						// qualified name of another interface
						defCode = extractQualified(env, casted.LoadInt(1), casted.LoadInt(2))
					case names.TildeId:
						defCode = jen.Op(string(names.TildeId)).Add(extractType(env, casted.LoadInt(1)))
					default:
						// method description
						paramTypes, _ := casted.LoadInt(1).(*types.List)
						var paramCodes []jen.Code
						// handle type or name:type
						types.ForEach(paramTypes, func(elem types.Object) bool {
							typeCode := extractType(env, elem)
							if typeCode == nil {
								casted, _ := elem.(*types.List)
								paramId, _ := casted.LoadInt(1).(types.Identifier)
								typeCode = jen.Id(string(paramId)).Add(extractType(env, casted.LoadInt(2)))
							}
							paramCodes = append(paramCodes, typeCode)
							return true
						})

						defCode = jen.Id(string(casted2)).Params(paramCodes...)
						if typeCode := extractType(env, casted.LoadInt(2)); typeCode != nil {
							defCode.Add(typeCode)
						}
					}
				case *types.List:
					// land here with syntaxic sugar
					switch header, _ := casted2.LoadInt(0).(types.Identifier); header {
					case names.GenId:
						defCode = extractGenType(env, casted.LoadInt(1), casted.LoadInt(2))
					case names.GetId:
						// qualified name of another interface
						defCode = extractQualified(env, casted2.LoadInt(1), casted2.LoadInt(2))
					case names.TildeId:
						first := true
						types.ForEach(casted, func(elem types.Object) bool {
							casted3, _ := elem.(*types.List)
							if typeCode := extractType(env, casted3.LoadInt(1)); first {
								first = false
								defCode = jen.Op(string(names.TildeId)).Add(typeCode)
							} else {
								defCode.Op(names.Pipe).Op(string(names.TildeId)).Add(typeCode)
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
				castedSize := casted.Size()
				if castedSize == 1 {
					// nested type
					defCodes = append(defCodes, extractType(env, casted.LoadInt(0)))
					return true
				}

				fieldId, _ := casted.LoadInt(0).(types.Identifier)
				defCode := jen.Id(string(fieldId)).Add(extractType(env, casted.LoadInt(1)))
				if castedSize > 2 {
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
		typeCode.Add(extractTypeFromList(env, oldType))
	default:
		return wrappedErrorComment
	}
	return wrapper{Renderer: typeCode.Line()}
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processDef(env, itArgs, jen.Var())
}
