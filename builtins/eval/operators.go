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

package eval

import (
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreFunc(env, itArgs, evalFirstOp, bitwiseAndFunc)
}

func andForm(env types.Environment, itArgs types.Iterator) types.Object {
	return boolOperatorForm(env, itArgs, true)
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	values := types.NewList().AddAll(makeEvalIterator(itArgs, env))
	switch casted := arg0.(type) {
	case types.Identifier:
		env.StoreStr(string(casted), values.LoadInt(0))
	case *types.List:
		if assignFunc := buildAssignFuncFromList(env, casted); assignFunc != nil {
			assignFunc(values.LoadInt(0))

			break
		}

		index := 0
		types.ForEach(casted, func(elem types.Object) bool {
			assignFunc := buildAssignFunc(env, elem)
			ok := assignFunc != nil
			if ok {
				assignFunc(values.LoadInt(index))
				index++
			}

			return ok
		})
	}

	return types.None
}

func bitwiseAndAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.AmpersandId))
}

func bitwiseAndFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseAndOperator)
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.AndNot))
}

func bitwiseAndNotFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseAndNotOperator)
}

func bitwiseOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Pipe)
}

func bitwiseOrFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseOrOperator)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Caret)
}

func bitwiseXOrFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseXOrOperator)
}

func callMethodForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func decrementForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceUnaryOperatorForm(env, itArgs, names.Minus)
}

func dereferenceOrMultiplyForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processUnaryOrBinaryMoreFunc(env, itArgs, evalFirstOp, productFunc)
}

func divideSetForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Slash)
}

func extendSliceForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func equalFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	return types.Boolean(ok && equals(arg0.Eval(env), arg1.Eval(env)))
}

func greaterEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func greaterForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func incrementForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceUnaryOperatorForm(env, itArgs, names.Plus)
}

func indexOrSliceForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func leftShiftAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func leftShiftForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func lesserForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func lesserEqualForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func minusSetForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Minus)
}

func notEqualFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	return types.Boolean(ok && !equals(arg0.Eval(env), arg1.Eval(env)))
}

func notFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg, _ := itArgs.Next()

	return types.Boolean(!extractBoolean(arg.Eval(env)))
}

func orForm(env types.Environment, itArgs types.Iterator) types.Object {
	return boolOperatorForm(env, itArgs, false)
}

func productFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return cumulFunc(env, itArgs, productCarac)
}

func productSetForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.StarId))
}

func receivingOrSendingForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func remainderSetForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Percent)
}

func rightShiftAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func rightShiftForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func storeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func sumFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return cumulFunc(env, itArgs, sumCarac)
}

func sumSetForm(env types.Environment, itArgs types.Iterator) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Plus)
}