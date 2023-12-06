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
	return processAugmentedAssignMore(env, itArgs, names.AddAssign, names.Plus)
}

func additionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, names.Plus)
}

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, string(names.AmpersandId))
}

func andForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.And)
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func bitwiseAndAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.AndAssign, string(names.AmpersandId))
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.AndNotAssign, names.AndNot)
}

func bitwiseAndNotForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.AndNot)
}

func bitwiseOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.OrAssign, names.Pipe)
}

func bitwiseOrForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Pipe)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.XorAssign, names.Caret)
}

func bitwiseXOrForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Caret)
}

func callMethodForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, _ := itArgs.Next()
	methodId, ok := arg1.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}

	argsCode := compileToCodeSlice(env, itArgs)
	// returned value could be callable
	return callableWrapper{Renderer: compileToCode(env, arg0).Dot(string(methodId)).Call(argsCode...)}
}

func declareAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
}

func decrementForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryPostOperator(env, itArgs, names.Decrement)
}

func dereferenceOrMultiplyForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, string(names.StarId))
}

func divideAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.DivAssign, names.Slash)
}

func divideForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Slash)
}

func equalForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.Equal)
}

func extendSliceForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: compileToCode(env, arg0).Op(string(names.EllipsisId))}
}

func greaterForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processComparison(env, itArgs, names.Greater)
}

func greaterEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processComparison(env, itArgs, names.GreaterEqual)
}

func incrementForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryPostOperator(env, itArgs, names.Increment)
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
	// could be a generic function
	return callableWrapper{Renderer: slicingCode}
}

func leftShiftForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.LShift)
}

func leftShiftAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.LShiftAssign)
}

func lesserForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processComparison(env, itArgs, names.Lesser)
}

func lesserEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processComparison(env, itArgs, names.LesserEqual)
}

func moduloAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.ModAssign, names.Percent)
}

func moduloForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Percent)
}

func multiplyAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.MultAssign, string(names.StarId))
}

func notForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}
	return wrapper{Renderer: jen.Op(names.Not).Add(compileToCode(env, arg0))}
}

func notEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.NotEqual)
}

func orForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Or)
}

func receivingOrSendingForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	arg1, ok := itArgs.Next()
	if ok {
		return wrapper{Renderer: compileToCode(env, arg0).Op(names.Arrow).Add(compileToCode(env, arg1))}
	}

	targetCode := Renderer(extractType(arg0))
	if targetCode == nil {
		targetCode = compileToCode(env, arg0)
	}
	// returned value could be callable when receiving from channel
	return callableWrapper{Renderer: jen.Op(names.Arrow).Add(targetCode)}
}

func rightShiftForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.RShift)
}

func rightShiftAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.RShiftAssign)
}

func storeForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	slice0 := extractSliceIndexes(env, arg1)
	var slices [][]jen.Code
	types.ForEach(itArgs, func(elem types.Object) bool {
		sliceN := extractSliceIndexes(env, elem)
		slices = append(slices, sliceN)
		return true
	})

	lastIndex := len(slices) - 1
	slicingCode := compileToCode(env, arg0).Index(slice0...)
	for index := 0; index < lastIndex; index++ {
		slicingCode.Index(slices[index]...)
	}
	return wrapper{Renderer: slicingCode.Op(names.Equal).Add(slices[lastIndex][0])}
}

func substractAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.SubAssign, names.Minus)
}

func substractionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, names.Minus)
}
