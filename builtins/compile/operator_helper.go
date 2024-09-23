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

	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func extractSliceIndexes(env types.Environment, object types.Object) []jen.Code {
	if casted, ok := object.(*types.List); ok {
		next, stop := types.Pull(casted.Iter())
		defer stop()

		arg0, _ := next()
		// detect slice (could be a classic function/operator call)
		if header, _ := arg0.(types.Identifier); header == names.ListId {
			return compileToCodeSlice(env, types.Push(next))
		}
	}
	return []jen.Code{compileToCode(env, object)}
}

func processAssign(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	values := compileToCodeSlice(env, types.Push(next))
	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(op).Add(values[0])}
	case *types.List:
		if id := extractAssignTargetFromList(env, casted); id != nil {
			return wrapper{Renderer: id.Op(op).Add(values[0])}
		}

		var ids []jen.Code
		for elem := range casted.Iter() {
			ids = append(ids, extractAssignTarget(env, elem))
		}
		return wrapper{Renderer: jen.List(ids...).Op(op).List(values...)}
	}
	return wrappedErrorComment
}

func processAugmentedAssign(env types.Environment, itArgs iter.Seq[types.Object], opAssign string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	targetCode := extractAssignTarget(env, arg0)
	if !ok || targetCode == nil {
		return wrappedErrorComment
	}
	return wrapper{Renderer: targetCode.Op(opAssign).Add(compileToCode(env, arg1))}
}

func processAugmentedAssignMore(env types.Environment, itArgs iter.Seq[types.Object], opAssign string, op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	targetCode := extractAssignTarget(env, arg0)
	if !ok || targetCode == nil {
		return wrappedErrorComment
	}

	targetCode.Op(opAssign).Add(compileToCode(env, arg1))
	for elem := range types.Push(next) {
		targetCode.Op(op).Add(compileToCode(env, elem))
	}
	return wrapper{Renderer: targetCode}
}

func processBinaryMoreOperator(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	binaryCode := compileToCode(env, arg0).Op(op).Add(compileToCode(env, arg1))
	for _, code := range compileToCodeSlice(env, types.Push(next)) {
		binaryCode.Op(op).Add(code)
	}
	return wrapper{Renderer: binaryCode}
}

func processBinaryOperator(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: compileToCode(env, arg0).Op(op).Add(compileToCode(env, arg1))}
}

func processComparison(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	argCode := jen.Code(compileToCode(env, arg1))
	binaryCode := compileToCode(env, arg0).Op(op).Add(argCode)
	for _, currentCode := range compileToCodeSlice(env, types.Push(next)) {
		binaryCode.Op(names.And).Add(argCode).Op(op).Add(currentCode)
		argCode = currentCode
	}
	return wrapper{Renderer: binaryCode}
}

func processUnaryPostOperator(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	for arg0 := range itArgs {
		return wrapper{Renderer: compileToCode(env, arg0).Op(op)}
	}
	return wrappedErrorComment
}

func processUnaryOrBinaryMoreOperator(env types.Environment, itArgs iter.Seq[types.Object], op string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	valueCodesTemp := compileToCodeSlice(env, types.Push(next))
	if len(valueCodesTemp) == 0 {
		targetCode := Renderer(extractType(env, arg0))
		if targetCode == (*jen.Statement)(nil) {
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
