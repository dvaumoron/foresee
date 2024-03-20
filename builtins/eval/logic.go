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
	compareInt    func(types.Integer, types.Integer) bool
	compareFloat  func(types.Float, types.Float) bool
	compareString func(types.String, types.String) bool
}

type ordered interface {
	number | types.String
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
	greaterEqualComparator = comparator{compareInt: greaterEqual[types.Integer], compareFloat: greaterEqual[types.Float], compareString: greaterEqual[types.String]}
	greaterThanComparator  = comparator{compareInt: greaterThan[types.Integer], compareFloat: greaterThan[types.Float], compareString: greaterThan[types.String]}
	lessEqualComparator    = comparator{compareInt: lessEqual[types.Integer], compareFloat: lessEqual[types.Float], compareString: lessEqual[types.String]}
	lessThanComparator     = comparator{compareInt: lessThan[types.Integer], compareFloat: lessThan[types.Float], compareString: lessThan[types.String]}
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
			return c.compareInt(casted0, casted1)
		case types.Float:
			return c.compareFloat(types.Float(casted0), casted1)
		}
	case types.Float:
		switch casted1 := value1.(type) {
		case types.Integer:
			return c.compareFloat(casted0, types.Float(casted1))
		case types.Float:
			return c.compareFloat(casted0, casted1)
		}
	case types.String:
		casted1, ok := value1.(types.String)
		if !ok {
			break
		}

		return c.compareString(casted0, casted1)
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
