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
	return processAugmentedAssign(env, itArgs, names.AddAssign)
}

func additionForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, names.Plus)
}

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryOperator(env, itArgs, string(names.AmpersandId))
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func bitwiseAndAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AndAssign)
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.AndNotAssign)
}

func bitwiseAndNotForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.AndNot)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAugmentedAssign(env, itArgs, names.XorAssign)
}

func bitwiseXOrForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processBinaryOperator(env, itArgs, names.Caret)
}

func declareAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
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
