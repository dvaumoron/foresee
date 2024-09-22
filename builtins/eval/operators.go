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
	"iter"

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func addressOrBitwiseAndForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	// TODO improve adressing and dereferencing eval with boxing and unboxing behavior
	return processUnaryOrBinaryMoreFunc(env, itArgs, evalFirstForm, bitwiseAndFunc)
}

func andForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return boolOperatorForm(env, itArgs, true)
}

func assignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	arg0, _ := itArgs.Next()
	values := types.NewList().AddAll(evalIterator(itArgs, env))
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
			if assignFunc == nil {
				panic(errAssignableType)
			}

			assignFunc(values.LoadInt(index))
			index++

			return true
		})
	}

	return types.None
}

func bitwiseAndAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.AmpersandId))
}

func bitwiseAndFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseAndOperator)
}

func bitwiseAndNotAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.AndNot))
}

func bitwiseAndNotFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseAndNotOperator)
}

func bitwiseOrAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Pipe)
}

func bitwiseOrFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseOrOperator)
}

func bitwiseXOrAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Caret)
}

func bitwiseXOrFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, bitwiseXOrOperator)
}

func callMethodForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		panic(errPairSize)
	}

	d, ok := arg0.Eval(env).(dynamicObject)
	if !ok {
		panic(errObjectType)
	}

	methodId, ok := arg1.(types.Identifier)
	if !ok {
		panic(errIdentifierType)
	}

	method, ok := d.objectType.methods[string(methodId)]
	if !ok {
		panic(errUnknownField)
	}

	augmentedItArgs := types.NewList(d).AddAll(itArgs).Iter()
	defer augmentedItArgs.Close()

	return method.Apply(env, augmentedItArgs)
}

func decrementForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceUnaryOperatorForm(env, itArgs, names.Minus)
}

func dereferenceOrMultiplyForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return processUnaryOrBinaryMoreFunc(env, itArgs, evalFirstForm, productFunc)
}

func divideSetForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Slash)
}

func extendSliceForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	// TODO

	return types.None
}

func equalFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	return types.Boolean(ok && equals(arg0.Eval(env), arg1.Eval(env)))
}

func greaterEqualForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return compareForm(env, itArgs, greaterEqualComparator)
}

func greaterForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return compareForm(env, itArgs, greaterThanComparator)
}

func incrementForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceUnaryOperatorForm(env, itArgs, names.Plus)
}

func indexOrSliceForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	res, _ := itArgs.Next()
	types.ForEach(itArgs, func(elem types.Object) bool {
		loadable, ok := res.(types.Loadable)
		if !ok {
			panic(errIndexableType)
		}

		res = loadable.Load(elem.Eval(env))

		return true
	})

	return res
}

func leftShiftAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.LShift)
}

func leftShiftFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, leftShiftOperator)
}

func lesserForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return compareForm(env, itArgs, lessThanComparator)
}

func lesserEqualForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return compareForm(env, itArgs, lessEqualComparator)
}

func minusSetForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Minus)
}

func notEqualFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	return types.Boolean(ok && !equals(arg0.Eval(env), arg1.Eval(env)))
}

func notFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	arg, _ := itArgs.Next()

	return types.Boolean(!extractBoolean(arg.Eval(env)))
}

func orForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return boolOperatorForm(env, itArgs, false)
}

func productFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return cumulNumber(evalIterator(itArgs, env), productKind)
}

func productSetForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, string(names.StarId))
}

func receivingOrSendingForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	// TODO channel wrapper type

	return types.None
}

func remainderSetForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Percent)
}

func rightShiftAssignForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.RShift)
}

func rightShiftFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return intOperatorFunc(env, itArgs, rightShiftOperator)
}

func storeForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	args := types.NewList().AddAll(itArgs)
	lastIndex := args.Size() - 1
	if lastIndex < 2 {
		panic(errTripleSize)
	}

	current, ok := args.LoadInt(0).Eval(env).(types.Loadable)
	if !ok {
		panic(errIndexableType)
	}

	index := args.LoadInt(1).Eval(env)
	for i := 2; i < lastIndex; i++ {
		current, ok = current.Load(index).(types.Loadable)
		if !ok {
			panic(errIndexableType)
		}

		index = args.LoadInt(i).Eval(env)
	}

	storable, ok := current.(types.Storable)
	if !ok {
		panic(errAssignableType)
	}

	storable.Store(index, args.LoadInt(lastIndex).Eval(env))

	return types.None
}

func sumFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	args := types.NewList().AddAll(evalIterator(itArgs, env))
	if _, isString := args.LoadInt(0).(types.String); isString {
		return concatStrings(args)
	}

	return cumulNumber(args, sumKind)
}

func sumSetForm(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	return inplaceOperatorForm(env, itArgs, names.Plus)
}
