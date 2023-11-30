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
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func addAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AddAssign, names.Plus)
}

func additionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, names.Plus)
}

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, string(names.AmpersandId))
}

func andForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.And)
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func bitwiseAndAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AndAssign, string(names.AmpersandId))
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AndNotAssign, names.AndNot)
}

func bitwiseAndNotForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.AndNot)
}

func bitwiseOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.OrAssign, names.Pipe)
}

func bitwiseOrForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Pipe)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.XorAssign, names.Caret)
}

func bitwiseXOrForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Caret)
}

func declareAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
}

func dereferenceOrMultiplyForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, string(names.StarId))
}

func divideAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.DivAssign, names.Slash)
}

func divideForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Slash)
}

func equalForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.Equal)
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

func moduloAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.ModAssign, names.Percent)
}

func moduloForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Percent)
}

func multiplyAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.MultAssign, string(names.StarId))
}

func notEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.NotEqual)
}

func orForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryMoreOperator(env, itArgs, names.Or)
}

func substractAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.SubAssign, names.Minus)
}

func substractionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, names.Minus)
}
