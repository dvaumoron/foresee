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
	"errors"

	"github.com/dvaumoron/foresee/types"
)

var errZeroDivision = errors.New("division by zero")

type cumulCarac struct {
	init       types.Integer
	cumulInt   func(types.Integer, types.Integer) types.Integer
	cumulFloat func(types.Float, types.Float) types.Float
}

type number interface {
	types.Integer | types.Float
}

func addNumberOperator[N number](a, b N) N {
	return a + b
}

func bitwiseAndNotOperator(a, b int64) int64 {
	return a &^ b
}

func bitwiseAndOperator(a, b int64) int64 {
	return a & b
}

func bitwiseOrOperator(a, b int64) int64 {
	return a | b
}

func bitwiseXOrOperator(a, b int64) int64 {
	return a ^ b
}

func leftShiftOperator(a, b int64) int64 {
	return a << b
}

func multNumberOperator[N number](a, b N) N {
	return a * b
}

func rightShiftOperator(a, b int64) int64 {
	return a >> b
}

var (
	sumCarac     = cumulCarac{init: 0, cumulInt: addNumberOperator[types.Integer], cumulFloat: addNumberOperator[types.Float]}
	productCarac = cumulCarac{init: 1, cumulInt: multNumberOperator[types.Integer], cumulFloat: multNumberOperator[types.Float]}
)

func cumulFunc(args types.Iterable, carac cumulCarac) types.Object {
	cumulI := carac.init
	cumulF := types.Float(cumulI)
	allNumericType, hasFloat := true, false
	types.ForEach(args, func(arg types.Object) bool {
		switch casted := arg.(type) {
		case types.Integer:
			cumulI = carac.cumulInt(cumulI, casted)
		case types.Float:
			hasFloat = true
			cumulF = carac.cumulFloat(cumulF, casted)
		default:
			allNumericType = false
		}
		return allNumericType
	})

	if !allNumericType {
		panic(errNumericType)
	}

	if hasFloat {
		return carac.cumulFloat(types.Float(cumulI), cumulF)
	}

	return cumulI
}

func minusFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	arg1, ok := itArgs.Next()
	if !ok {
		// unary version
		switch casted := arg0.Eval(env).(type) {
		case types.Integer:
			return -casted
		case types.Float:
			return -casted
		}
		panic(errNumericType)
	}

	itArgs = types.NewList(arg1).AddAll(itArgs).Iter()
	defer itArgs.Close()

	switch casted := arg0.Eval(env).(type) {
	case types.Integer:
		switch casted2 := sumFunc(env, itArgs).(type) {
		case types.Integer:
			return casted - casted2
		case types.Float:
			return types.Float(casted) - casted2
		}
	case types.Float:
		switch casted2 := sumFunc(env, itArgs).(type) {
		case types.Integer:
			return casted - types.Float(casted2)
		case types.Float:
			return casted - casted2
		}
	}
	panic(errNumericType)
}

func divideFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	switch casted := arg0.Eval(env).(type) {
	case types.Integer:
		return divideObject(types.Float(casted), productFunc(env, itArgs))
	case types.Float:
		return divideObject(casted, productFunc(env, itArgs))
	}
	panic(errNumericType)
}

func divideObject(a types.Float, b types.Object) types.Object {
	switch casted := b.(type) {
	case types.Integer:
		return divideFloat(a, types.Float(casted))
	case types.Float:
		return divideFloat(a, casted)
	}
	panic(errNumericType)
}

func divideFloat(a types.Float, b types.Float) types.Object {
	if b == 0 {
		panic(errZeroDivision)
	}
	return a / b
}

func remainderFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	a, allInt := arg0.Eval(env).(types.Integer)
	if !allInt {
		panic(errIntegerType)
	}

	res := int64(a)
	var b types.Integer
	types.ForEach(itArgs, func(arg types.Object) bool {
		b, allInt = arg.Eval(env).(types.Integer)
		if allInt = allInt && b != 0; allInt {
			res %= int64(b)
		}
		return allInt
	})

	if !allInt {
		panic(errIntegerType)
	}

	return types.Integer(res)
}

func intOperatorFunc(env types.Environment, itArgs types.Iterator, intOperator func(int64, int64) int64) types.Object {
	arg0, _ := itArgs.Next()
	a, allInt := arg0.Eval(env).(types.Integer)
	if !allInt {
		panic(errIntegerType)
	}

	res := int64(a)
	var b types.Integer
	types.ForEach(itArgs, func(arg types.Object) bool {
		b, allInt = arg.Eval(env).(types.Integer)
		res = intOperator(res, int64(b))
		return allInt
	})

	if !allInt {
		panic(errIntegerType)
	}

	return types.Integer(res)
}
