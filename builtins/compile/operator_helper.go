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
	"github.com/dvaumoron/foresee/types"
)

func processAssign(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(op).Add(values[0])}
	case *types.List:
		if id := extractAssignTargetFromList(env, casted); id != nil {
			return wrapper{Renderer: id.Op(op).Add(values[0])}
		}

		var ids []jen.Code
		types.ForEach(casted, func(elem types.Object) bool {
			ids = append(ids, extractAssignTarget(env, elem))
			return true
		})
		return wrapper{Renderer: jen.List(ids...).Op(op).List(values...)}
	}
	return wrappedErrorComment
}

func processAugmentedAssign(env types.Environment, itArgs types.Iterator, opAssign string, op string) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	targetCode := extractAssignTarget(env, arg0)
	if !ok || targetCode == nil {
		return wrappedErrorComment
	}

	targetCode.Op(opAssign).Add(compileToCode(env, arg1))
	types.ForEach(itArgs, func(elem types.Object) bool {
		targetCode.Op(op).Add(compileToCode(env, elem))
		return true
	})
	return wrapper{Renderer: targetCode}
}

func processBinaryMoreOperator(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	binaryCode := compileToCode(env, arg0).Op(op).Add(compileToCode(env, arg1))
	for _, code := range compileToCodeSlice(env, itArgs) {
		binaryCode.Op(op).Add(code)
	}
	return wrapper{Renderer: binaryCode}
}

func processBinaryOperator(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: compileToCode(env, arg0).Op(op).Add(compileToCode(env, arg1))}
}

func processUnaryOrBinaryMoreOperator(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	valueCodesTemp := compileToCodeSlice(env, itArgs)
	if len(valueCodesTemp) == 0 {
		targetCode := Renderer(extractType(arg0))
		if targetCode == nil {
			targetCode = compileToCode(env, arg0)
		}
		// usable to build a literal when adressing
		return literalWrapper{Renderer: jen.Op(op).Add(targetCode)}
	}

	binaryCode := compileToCode(env, arg0)
	for _, code := range valueCodesTemp {
		binaryCode.Op(op).Add(code)
	}
	return wrapper{Renderer: binaryCode}
}
