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

func addAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.AddAssign, names.Plus)
}

func additionForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, names.Plus)
}

func addressOrBitwiseAndForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, string(names.AmpersandId))
}

func andForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.And)
}

func assignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func bitwiseAndAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.AndAssign, string(names.AmpersandId))
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.AndNotAssign, names.AndNot)
}

func bitwiseAndNotForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.AndNot)
}

func bitwiseOrAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.OrAssign, names.Pipe)
}

func bitwiseOrForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Pipe)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.XorAssign, names.Caret)
}

func bitwiseXOrForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Caret)
}

func callMethodForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, _ := next()
	methodId, ok := arg1.(types.Identifier)
	if !ok {
		return wrappedErrorComment
	}

	argsCode := compileToCodeSlice(env, types.Push(next))
	// returned value could be callable
	return callableWrapper{Renderer: compileToCode(env, arg0).Dot(string(methodId)).Call(argsCode...)}
}

func declareAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
}

func decrementForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryPostOperator(env, itArgs, names.Decrement)
}

func dereferenceOrMultiplyForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, string(names.StarId))
}

func divideAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.DivAssign, names.Slash)
}

func divideForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Slash)
}

func equalForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryOperator(env, itArgs, names.Equal)
}

func extendSliceForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	for arg0 := range itArgs {
		return wrapper{Renderer: compileToCode(env, arg0).Op(string(names.EllipsisId))}
	}
	return wrappedErrorComment
}

func greaterForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processComparison(env, itArgs, names.Greater)
}

func greaterEqualForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processComparison(env, itArgs, names.GreaterEqual)
}

func incrementForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryPostOperator(env, itArgs, names.Increment)
}

func indexOrSliceForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	slice0 := extractSliceIndexes(env, arg1)
	slicingCode := compileToCode(env, arg0).Index(slice0...)
	for elem := range types.Push(next) {
		sliceN := extractSliceIndexes(env, elem)
		slicingCode.Index(sliceN...)
	}
	// returned value could be callable
	return callableWrapper{Renderer: slicingCode}
}

func leftShiftForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryOperator(env, itArgs, names.LShift)
}

func leftShiftAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssign(env, itArgs, names.LShiftAssign)
}

func lesserForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processComparison(env, itArgs, names.Lesser)
}

func lesserEqualForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processComparison(env, itArgs, names.LesserEqual)
}

func moduloAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.ModAssign, names.Percent)
}

func moduloForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Percent)
}

func multiplyAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.MultAssign, string(names.StarId))
}

func notForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	for arg0 := range itArgs {
		return wrapper{Renderer: jen.Op(string(names.NotId)).Add(compileToCode(env, arg0))}
	}
	return wrappedErrorComment
}

func notEqualForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryOperator(env, itArgs, names.NotEqual)
}

func orForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Or)
}

func receivingOrSendingForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	arg1, ok := next()
	if ok {
		return wrapper{Renderer: compileToCode(env, arg0).Op(names.Arrow).Add(compileToCode(env, arg1))}
	}

	targetCode := Renderer(extractType(env, arg0))
	if targetCode == (*jen.Statement)(nil) {
		targetCode = compileToCode(env, arg0)
	}
	// returned value could be callable when receiving from channel
	return callableWrapper{Renderer: jen.Op(names.Arrow).Add(targetCode)}
}

func rightShiftForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processBinaryOperator(env, itArgs, names.RShift)
}

func rightShiftAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssign(env, itArgs, names.RShiftAssign)
}

func storeForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg0, _ := next()
	arg1, ok := next()
	if !ok {
		return wrappedErrorComment
	}

	slice0 := extractSliceIndexes(env, arg1)
	var slices [][]jen.Code
	for elem := range types.Push(next) {
		slices = append(slices, extractSliceIndexes(env, elem))
	}

	lastIndex := len(slices) - 1
	slicingCode := compileToCode(env, arg0).Index(slice0...)
	for index := 0; index < lastIndex; index++ {
		slicingCode.Index(slices[index]...)
	}
	return wrapper{Renderer: slicingCode.Op(names.Equal).Add(slices[lastIndex][0])}
}

func substractAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processAugmentedAssignMore(env, itArgs, names.SubAssign, names.Minus)
}

func substractionForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryOrBinaryMoreOperator(env, itArgs, names.Minus)
}
