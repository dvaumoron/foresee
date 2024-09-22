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
	"iter"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func extractNameWithGenericDef(env types.Environment, object types.Object) *jen.Statement {
	switch casted := object.(type) {
	case types.Identifier:
		return jen.Id(string(casted))
	case *types.List:
		if header, _ := casted.LoadInt(0).(types.Identifier); header == names.GenId {
			nameId, ok := casted.LoadInt(1).(types.Identifier)
			if !ok {
				return nil
			}

			genDefList, ok := casted.LoadInt(2).(*types.List)
			if !ok {
				return nil
			}

			next, stop := types.Pull(genDefList.Iter())
			defer stop()

			next() // skip ListId

			if genericCodes, noError := innerExtractParameter(env, types.Push(next)); noError {
				return jen.Id(string(nameId)).Types(genericCodes...)
			}
		}
	}
	return nil
}

func extractPackageName(path types.String) string {
	casted := string(path)
	splitIndex := strings.LastIndexByte(casted, '/') + 1 // 0 when not found, so splitting do nothing
	return casted[splitIndex:]
}

func extractParameter(env types.Environment, object types.Object) ([]jen.Code, bool) {
	paramIterable, ok := object.(types.Iterable)
	if !ok {
		return nil, false
	}
	return innerExtractParameter(env, paramIterable.Iter())
}

func innerExtractParameter(env types.Environment, paramIt iter.Seq[types.Object]) ([]jen.Code, bool) {
	ok := true
	var paramCodes []jen.Code
	for elem := range paramIt {
		var paramDesc *types.List
		// assume it's in "name:type" format (type should be inferred if not declared)
		if paramDesc, ok = elem.(*types.List); !ok {
			break
		}

		varId, _ := paramDesc.LoadInt(1).(types.Identifier)
		paramCodes = append(paramCodes, jen.Id(string(varId)).Add(extractType(env, paramDesc.LoadInt(2))))
	}
	return paramCodes, ok
}

func extractReturnType(env types.Environment, object types.Object) (*jen.Statement, []jen.Code) {
	switch casted := object.(type) {
	case types.NoneType:
		return nil, nil
	case types.Identifier:
		return jen.Id(string(casted)), nil
	case *types.List:
		if head, _ := casted.LoadInt(0).(types.Identifier); head == names.ListId {
			if typeCodes, ok := extractTypes(env, casted); ok {
				return jen.Parens(jen.List(typeCodes...)), nil
			}
		} else {
			if returnCode := extractType(env, object); returnCode != nil {
				return returnCode, nil
			}
			// can not extract type, so object is the first instruction of the code block
			return nil, []jen.Code{compileToCode(env, object)}
		}
	}
	return nil, nil
}

func extractSingleOrMultiple(env types.Environment, list *types.List) []jen.Code {
	switch list.LoadInt(0).(type) {
	case types.Identifier:
		return []jen.Code{compileToCode(env, list)}
	case *types.List:
		return compileToCodeSlice(env, list.Iter())
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
func processDef(env types.Environment, itArgs iter.Seq[types.Object], baseCode *jen.Statement) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	// K is supposed to be "const" or "var" (passed as baseCode)
	arg0, _ := next()
	switch casted := arg0.(type) {
	case types.Identifier:
		// detect "K name value"
		// no declared type, need a value
		if arg1, ok := next(); ok {
			return wrapper{Renderer: baseCode.Id(string(casted)).Op(names.Assign).Add(compileToCode(env, arg1))}
		}
	case *types.List:
		// detect "K name:type" and an optional value
		if header, _ := casted.LoadInt(0).(types.Identifier); header == names.ListId {
			nameId, _ := casted.LoadInt(1).(types.Identifier)
			baseCode.Id(string(nameId))
			if typeCode := extractType(env, casted.LoadInt(2)); typeCode != nil {
				baseCode.Add(typeCode)
			}
			if arg1, ok := next(); ok {
				baseCode.Op(names.Assign).Add(compileToCode(env, arg1))
			}
			return wrapper{Renderer: baseCode}
		}

		defCodes := []jen.Code{processDefLine(env, casted)}

		// following lines (optional)
		defCodes = processDefLines(env, types.Push(next), defCodes)
		return wrapper{Renderer: baseCode.Defs(defCodes...)}
	}
	return wrappedErrorComment
}

// handle multiple "(name value)" or "(name:type)" or "(name:type value)"
func processDefLines(env types.Environment, itArgs iter.Seq[types.Object], defCodes []jen.Code) []jen.Code {
	for elem := range itArgs {
		defDesc, _ := elem.(*types.List)
		defCodes = append(defCodes, processDefLine(env, defDesc))
	}
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
		defCode = jen.Id(string(nameId)).Add(extractType(env, casted.LoadInt(2)))
	}
	if defDesc.Size() > 1 {
		defCode.Op(names.Assign).Add(compileToCode(env, defDesc.LoadInt(1)))
	}
	return defCode
}

// labellableCode is not cloned (must generate a new one on each call)
func processLabellable(_ types.Environment, itArgs iter.Seq[types.Object], labellableCode *jen.Statement) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	if arg0, ok := next(); ok {
		labelId, ok := arg0.(types.Identifier)
		if !ok {
			return wrappedErrorComment
		}
		labellableCode.Id(string(labelId))
	}
	return wrapper{Renderer: labellableCode}
}
