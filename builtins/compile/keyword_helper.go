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
			nameId, ok := casted.LoadInt(1).(types.Identifier)
			if !ok {
				return nil
			}

			genDefList, ok := casted.LoadInt(2).(*types.List)
			if !ok {
				return nil
			}

			itGenDef := genDefList.Iter()
			defer itGenDef.Close()

			itGenDef.Next() // skip ListId

			if genericCodes, noError := innerExtractParameter(itGenDef); noError {
				return jen.Id(string(nameId)).Types(genericCodes...)
			}
		}
	}
	return nil
}

func extractParameter(object types.Object) ([]jen.Code, bool) {
	paramIterable, ok := object.(types.Iterable)
	if !ok {
		return nil, false
	}
	return innerExtractParameter(paramIterable)
}

func innerExtractParameter(paramIterable types.Iterable) ([]jen.Code, bool) {
	noError := true
	var paramCodes []jen.Code
	types.ForEach(paramIterable, func(elem types.Object) bool {
		var paramDesc *types.List
		// assume it's in "name:type" format (type should be inferred if not declared)
		if paramDesc, noError = elem.(*types.List); noError {
			varId, _ := paramDesc.LoadInt(1).(types.Identifier)
			paramCodes = append(paramCodes, jen.Id(string(varId)).Add(extractType(paramDesc.LoadInt(2))))
		}
		return noError
	})
	return paramCodes, noError
}

func extractReturnType(env types.Environment, object types.Object) (*jen.Statement, []jen.Code) {
	var returnCode *jen.Statement
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
		// no declared type, need a value
		if arg1, ok := itArgs.Next(); ok {
			return wrapper{Renderer: baseCode.Id(string(casted)).Op(names.Assign).Add(compileToCode(env, arg1))}
		}
	case *types.List:
		// detect "K name:type" and an optional value
		if header, _ := casted.LoadInt(0).(types.Identifier); header == names.ListId {
			nameId, _ := casted.LoadInt(1).(types.Identifier)
			baseCode.Id(string(nameId))
			if typeCode := extractType(casted.LoadInt(2)); typeCode != nil {
				baseCode.Add(typeCode)
			}
			if arg1, ok := itArgs.Next(); ok {
				baseCode.Op(names.Assign).Add(compileToCode(env, arg1))
			}
			return wrapper{Renderer: baseCode}
		}

		defCodes := []jen.Code{processDefLine(env, casted)}

		// following lines (optional)
		defCodes = processDefLines(env, itArgs, defCodes)
		return wrapper{Renderer: baseCode.Defs(defCodes...)}
	}
	return wrappedErrorComment
}

// handle multiple "(name value)" or "(name:type)" or "(name:type value)"
func processDefLines(env types.Environment, itArgs types.Iterable, defCodes []jen.Code) []jen.Code {
	types.ForEach(itArgs, func(elem types.Object) bool {
		defDesc, _ := elem.(*types.List)
		defCodes = append(defCodes, processDefLine(env, defDesc))
		return true
	})
	return defCodes
}

// handle "(name value)" or "(name:type)" or "(name:type value)"
func processDefLine(env types.Environment, defDesc *types.List) jen.Code {
	var defCode *jen.Statement
	switch casted := defDesc.LoadInt(0).(type) {
	case types.Identifier:
		defCode = jen.Id(string(casted))
	case *types.List:
		nameId, _ := casted.LoadInt(1).(types.Identifier)
		defCode = jen.Id(string(nameId)).Add(extractType(casted.LoadInt(2)))
	}
	if defDesc.Size() > 1 {
		defCode.Op(names.Assign).Add(compileToCode(env, defDesc.LoadInt(1)))
	}
	return defCode
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
