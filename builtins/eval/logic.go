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

var (
	errComparableType = errors.New("wait comparable value")
	errOrderableType  = errors.New("wait orderable value")
)

type comparator struct {
	compareInt    func(int64, int64) bool
	compareFloat  func(float64, float64) bool
	compareString func(string, string) bool
}

type ordered interface {
	number | string
}

func greaterEqual[O ordered](value0 O, value1 O) bool {
	return value0 >= value1
}

func greaterThan[O ordered](value0 O, value1 O) bool {
	return value0 > value1
}

func lessEqual[O ordered](value0 O, value1 O) bool {
	return value0 <= value1
}

func lessThan[O ordered](value0 O, value1 O) bool {
	return value0 < value1
}

var (
	greaterEqualComparator = comparator{compareInt: greaterEqual[int64], compareFloat: greaterEqual[float64], compareString: greaterEqual[string]}
	greaterThanComparator  = comparator{compareInt: greaterThan[int64], compareFloat: greaterThan[float64], compareString: greaterThan[string]}
	lessEqualComparator    = comparator{compareInt: lessEqual[int64], compareFloat: lessEqual[float64], compareString: lessEqual[string]}
	lessThanComparator     = comparator{compareInt: lessThan[int64], compareFloat: lessThan[float64], compareString: lessThan[string]}
)

func compareForm(env types.Environment, itArgs types.Iterator, c comparator) types.Object {
	arg0, res := itArgs.Next()
	if !res {
		return types.Boolean(false)
	}

	res = false
	previousValue := arg0.Eval(env)
	types.ForEach(itArgs, func(currentArg types.Object) bool {
		currentValue := currentArg.Eval(env)
		// change the variable that will be returned in the caller
		res = compare(previousValue, currentValue, c)
		previousValue = currentValue
		return res
	})
	return types.Boolean(res)
}

func compare(value0 types.Object, value1 types.Object, c comparator) bool {
	switch casted0 := value0.(type) {
	case types.Integer:
		switch casted1 := value1.(type) {
		case types.Integer:
			return c.compareInt(int64(casted0), int64(casted1))
		case types.Float:
			return c.compareFloat(float64(casted0), float64(casted1))
		}
	case types.Float:
		switch casted1 := value1.(type) {
		case types.Integer:
			return c.compareFloat(float64(casted0), float64(casted1))
		case types.Float:
			return c.compareFloat(float64(casted0), float64(casted1))
		}
	case types.String:
		casted1, ok := value1.(types.String)
		if !ok {
			break
		}

		return c.compareString(string(casted0), string(casted1))
	}
	panic(errOrderableType)
}

func boolOperatorForm(env types.Environment, itArgs types.Iterator, defaultB bool) types.Object {
	allBool := true
	res := types.Boolean(defaultB)
	var temp types.Boolean
	types.ForEach(itArgs, func(arg types.Object) bool {
		temp, allBool = arg.Eval(env).(types.Boolean)
		if temp != res {
			res = temp

			return false
		}

		return allBool
	})

	if !allBool {
		panic(errBooleanType)
	}

	return res
}

func equals(value0 types.Object, value1 types.Object) bool {
	switch casted0 := value0.(type) {
	case types.NoneType:
		_, ok := value1.(types.NoneType)
		return ok
	case types.Boolean:
		casted1, ok := value1.(types.Boolean)
		return ok && (casted0 == casted1)
	case types.Integer:
		switch casted1 := value1.(type) {
		case types.Integer:
			return casted0 == casted1
		case types.Float:
			return types.Float(casted0) == casted1
		}
	case types.Float:
		switch casted1 := value1.(type) {
		case types.Integer:
			return casted0 == types.Float(casted1)
		case types.Float:
			return casted0 == casted1
		}
	case types.String:
		casted1, ok := value1.(types.String)
		if !ok {
			break
		}

		return casted0 == casted1
	default:
		// TODO other type (func, struct, etc.)
	}
	panic(errComparableType)
}
