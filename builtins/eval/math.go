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
	"github.com/dvaumoron/foresee/types"
)

type cumulCarac struct {
	init       int64
	cumulInt   func(int64, int64) int64
	cumulFloat func(float64, float64) float64
}

type number interface {
	int64 | float64
}

func addNumber[N number](a, b N) N {
	return a + b
}

func multNumber[N number](a, b N) N {
	return a * b
}

var (
	sumCarac     = cumulCarac{init: 0, cumulInt: addNumber[int64], cumulFloat: addNumber[float64]}
	productCarac = cumulCarac{init: 1, cumulInt: multNumber[int64], cumulFloat: multNumber[float64]}
)

func cumulFunc(env types.Environment, itArgs types.Iterator, carac cumulCarac) types.Object {
	cumulI := carac.init
	cumulF := float64(cumulI)
	allValidType, hasFloat := true, false
	types.ForEach(itArgs, func(arg types.Object) bool {
		switch casted := arg.Eval(env).(type) {
		case types.Integer:
			cumulI = carac.cumulInt(cumulI, int64(casted))
		case types.Float:
			hasFloat = true
			cumulF = carac.cumulFloat(cumulF, float64(casted))
		default:
			allValidType = false
		}
		return allValidType
	})

	if !allValidType {
		return types.None
	}

	if hasFloat {
		return types.Float(carac.cumulFloat(float64(cumulI), cumulF))
	}

	return types.Integer(cumulI)
}

func minusFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	switch casted := arg0.Eval(env).(type) {
	case types.Integer:
		switch casted2 := sumFunc(env, itArgs).(type) {
		case types.Integer:
			return types.Integer(casted - casted2)
		case types.Float:
			return types.Float(float64(casted) - float64(casted2))
		}
	case types.Float:
		switch casted2 := sumFunc(env, itArgs).(type) {
		case types.Integer:
			return types.Float(float64(casted) - float64(casted2))
		case types.Float:
			return types.Float(casted - casted2)
		}
	}
	return types.None
}

func divideFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	switch casted := arg0.Eval(env).(type) {
	case types.Integer:
		return divideObject(float64(casted), productFunc(env, itArgs))
	case types.Float:
		return divideObject(float64(casted), productFunc(env, itArgs))
	}
	return types.None
}

func divideObject(a float64, b types.Object) types.Object {
	switch casted := b.(type) {
	case types.Integer:
		if casted != 0 {
			return types.Float(a / float64(casted))
		}
	case types.Float:
		if casted != 0 {
			return types.Float(a / float64(casted))
		}
	}
	return types.None
}

func remainderFunc(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	a, allInt := arg0.Eval(env).(types.Integer)
	if !allInt {
		return types.None
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
		return types.None
	}

	return types.Integer(res)
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

func intOperatorFunc(env types.Environment, itArgs types.Iterator, intOperator func(int64, int64) int64) types.Object {
	arg0, _ := itArgs.Next()
	a, allInt := arg0.Eval(env).(types.Integer)
	if !allInt {
		return types.None
	}

	res := int64(a)
	var b types.Integer
	types.ForEach(itArgs, func(arg types.Object) bool {
		b, allInt = arg.Eval(env).(types.Integer)
		res = intOperator(res, int64(b))
		return allInt
	})

	if !allInt {
		return types.None
	}

	return types.Integer(res)
}
