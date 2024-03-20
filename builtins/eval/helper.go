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

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

var (
	errAssignableType = errors.New("wait assignable type")
	errIndexableType  = errors.New("wait indexable type")
	errPairSize       = errors.New("wait at least 2 elements")
	errSelectableType = errors.New("wait indexable type")
	errTripleSize     = errors.New("wait at least 3 elements")
	errUnarySize      = errors.New("wait 1 argument")
	errUnknownField   = errors.New("field or method unknown")
)

type evalIterator struct {
	types.NoneType
	inner types.Iterator
	env   types.Environment
}

func (e evalIterator) Iter() types.Iterator {
	return e
}

func (e evalIterator) Next() (types.Object, bool) {
	value, ok := e.inner.Next()
	return value.Eval(e.env), ok
}

func (e evalIterator) Close() {
	e.inner.Close()
}

func makeEvalIterator(it types.Iterator, env types.Environment) evalIterator {
	return evalIterator{inner: it, env: env}
}

// handle "a" as a,  "(* a)" as a, "([] a b c)" as a[b][c] and "(get a b c)" as a.b.c
func buildAssignFunc(env types.Environment, object types.Object) func(types.Object) {
	switch casted := object.(type) {
	case types.Identifier:
		return func(value types.Object) {
			env.StoreStr(string(casted), value)
		}
	case *types.List:
		return buildAssignFuncFromList(env, casted)
	}
	return nil
}

func buildAssignFuncFromList(env types.Environment, list *types.List) func(types.Object) {
	size := list.Size()
	switch op, _ := list.LoadInt(0).(types.Identifier); op {
	case names.GetId:
		if size < 3 {
			panic(errPairSize)
		}

		current, ok := list.LoadInt(1).Eval(env).(types.StringLoadable)
		if !ok {
			panic(errSelectableType)
		}

		index, ok := list.LoadInt(2).(types.Identifier)
		if !ok {
			panic(errIdentifierType)
		}

		for i := 3; i < size; i++ {
			temp, _ := current.LoadStr(string(index))
			current, ok = temp.(types.StringLoadable)
			if !ok {
				panic(errSelectableType)
			}

			index, ok = list.LoadInt(i).(types.Identifier)
			if !ok {
				panic(errIdentifierType)
			}
		}

		storable, ok := current.(types.Environment)
		if !ok {
			panic(errAssignableType)
		}

		return func(value types.Object) {
			storable.StoreStr(string(index), value)
		}
	case names.ListId:
		id, ok := list.LoadInt(1).(types.Identifier)
		if !ok {
			panic(errIdentifierType)
		}

		return func(value types.Object) {
			env.StoreStr(string(id), value)
		}
	case names.Load:
		if size < 3 {
			panic(errPairSize)
		}

		current, ok := list.LoadInt(1).Eval(env).(types.Loadable)
		if !ok {
			panic(errIndexableType)
		}

		index := list.LoadInt(2).Eval(env)
		for i := 3; i < size; i++ {
			current, ok = current.Load(index).(types.Loadable)
			if !ok {
				panic(errIndexableType)
			}

			index = list.LoadInt(i).Eval(env)
		}

		storable, ok := current.(types.Storable)
		if !ok {
			panic(errAssignableType)
		}

		return func(value types.Object) {
			storable.Store(index, value)
		}
	case names.StarId:
		if size != 1 {
			panic(errUnarySize)
		}

		return buildAssignFunc(env, list.LoadInt(1))
	}

	return nil
}
