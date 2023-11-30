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

func addAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AddAssign)
}

func processAugmentedAssign(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	targetCode := extractAssignTarget(env, arg0)
	if !ok || targetCode == nil {
		return wrappedErrorComment
	}
	return wrapper{Renderer: targetCode.Op(op).Add(compileToCode(env, arg1))}
}

func additionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, names.Plus)
}

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, string(names.AmpersandId))
}

func processUnaryOrBinaryOperator(env types.Environment, itArgs types.Iterator, op string) types.Object {
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
		// adressing, usable to build a literal
		return literalWrapper{Renderer: jen.Op(op).Add(targetCode)}
	}

	binaryCode := compileToCode(env, arg0)
	for _, code := range valueCodesTemp {
		binaryCode.Op(op).Add(code)
	}
	return wrapper{Renderer: binaryCode}
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func declareAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
}

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

func dereferenceOrMultiplyForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, string(names.StarId))
}

func divideAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.SlashAssign)
}

func divideForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.Slash)
}

func processBinaryOperator(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

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

func indexOrSliceForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	slice0 := extractSliceIndexes(env, arg1)
	slicingCode := compileToCode(env, arg0).Index(slice0...)
	types.ForEach(itArgs, func(elem types.Object) bool {
		sliceN := extractSliceIndexes(env, elem)
		slicingCode.Index(sliceN...)
		return true
	})
	return wrapper{Renderer: slicingCode}
}

func multiplyAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.StarAssign)
}

func substractAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.MinusAssign)
}

func substractionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, names.Minus)
}
