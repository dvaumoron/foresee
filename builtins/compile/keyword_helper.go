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

func extractNameWithGenericDef(object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		if header, _ := casted.LoadInt(0).(types.Identifier); header == names.LoadId {
			itCasted := casted.Iter()
			defer itCasted.Close()

			itCasted.Next() // skip LoadId

			name, _ := itCasted.Next()
			nameId, ok := name.(types.Identifier)
			if !ok {
				return nil
			}

			noError := true
			var genericCodes []jen.Code
			types.ForEach(itCasted, func(elem types.Object) bool {
				var genDesc *types.List
				// assume it's in "name:type" format
				if genDesc, noError = elem.(*types.List); noError {
					genId, _ := genDesc.LoadInt(1).(types.Identifier)
					genericCodes = append(genericCodes, jen.Id(string(genId)).Add(extractType(genDesc.LoadInt(2))))
				}
				return noError
			})

			if noError {
				return jen.Id(string(nameId)).Types(genericCodes...)
			}
		}
	}
	return nil
}

func extractParameter(object types.Object) []jen.Code {
	paramList, ok := object.(*types.List)
	if !ok {
		return nil
	}

	var paramCodes []jen.Code
	types.ForEach(paramList, func(elem types.Object) bool {
		// assume it's in "name:type" format (type should be inferred if not declared)
		paramDesc, _ := elem.(*types.List)
		varId, _ := paramDesc.LoadInt(1).(types.Identifier)
		paramCodes = append(paramCodes, jen.Id(string(varId)).Add(extractType(paramDesc.LoadInt(2))))
		return true
	})
	return paramCodes
}

func extractReturnType(env types.Environment, object types.Object) (jen.Code, []jen.Code) {
	var returnCode jen.Code
	var instructionCodes []jen.Code
	switch casted := object.(type) {
	case types.NoneType:
		// optional marker for missing return type
	case types.Identifier:
		returnCode = jen.Id(string(casted))
	case *types.List:
		if head, _ := casted.LoadInt(0).(types.Identifier); head == names.ListId {
			typeCodes := extractTypes(casted)
			returnCode = jen.Parens(jen.List(typeCodes...))
		} else {
			if returnCode = extractType(object); returnCode == nil {
				// can not extract type, so object is the first instruction of the code block
				instructionCodes = []jen.Code{compileToCode(env, object)}
			}
		}
	}
	return returnCode, instructionCodes
}

func extractSingleOrMultiple(env types.Environment, list *types.List) []jen.Code {
	switch list.LoadInt(0).(type) {
	case types.Identifier:
		return []jen.Code{compileToCode(env, list)}
	case *types.List:
		return compileToCodeSlice(env, list)
	}
	return nil
}

func extractValueOrMultiple(env types.Environment, object types.Object) []jen.Code {
	condCode := handleBasicType(object, true, func(object types.Object) Renderer {
		return nil
	})
	if condCode != nil {
		return []jen.Code{condCode}
	}

	// object is not a basic type
	if list, ok := object.(*types.List); ok {
		return extractSingleOrMultiple(env, list)
	}
	return nil
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
				if typeCode := extractType(casted.LoadInt(2)); typeCode != nil {
					defCode.Add(typeCode)
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

// labellableCode is not cloned (must generate a new one on each call)
func processLabellable(env types.Environment, itArgs types.Iterator, labellableCode *jen.Statement) types.Object {
	if arg0, ok := itArgs.Next(); ok {
		labelId, ok := arg0.(types.Identifier)
		if !ok {
			return wrappedErrorComment
		}
		labellableCode.Id(string(labelId))
	}
	return wrapper{Renderer: labellableCode}
}
