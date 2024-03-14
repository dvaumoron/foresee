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

package types

import "io"

type List struct {
	inner []Object
}

func (l *List) Add(value Object) *List {
	l.inner = append(l.inner, value)
	return l
}

// If the action func return false that break the loop.
func ForEach(it Iterable, action func(Object) bool) {
	it2 := it.Iter()
	defer it2.Close()
	var value Object
	for ok := true; ok; {
		if value, ok = it2.Next(); ok {
			ok = action(value)
		}
	}
}

func (l *List) AddAll(it Iterable) *List {
	ForEach(it, func(value Object) bool {
		l.Add(value)
		return true
	})
	return l
}

func convertToInt(arg Object, init int) int {
	casted, ok := arg.(Integer)
	if !ok {
		return init
	}
	return int(casted)
}

func extractIndex(args []Object, max int) (int, int) {
	switch len(args) {
	case 0:
		return 0, max
	case 1:
		return convertToInt(args[0], 0), max
	}
	return convertToInt(args[0], 0), convertToInt(args[1], max)
}

// No panic with nil receiver
func (l *List) LoadInt(index int) Object {
	if l == nil || index < 0 || index >= len(l.inner) {
		return None
	}
	return l.inner[index]
}

// No panic with nil receiver
func (l *List) Load(key Object) Object {
	switch casted := key.(type) {
	case Integer:
		return l.LoadInt(int(casted))
	case Float:
		return l.LoadInt(int(casted))
	case *List:
		if l == nil {
			return &List{}
		}

		max := len(l.inner)
		start, end := extractIndex(casted.inner, max)
		if 0 > start || start > end || end > max {
			return &List{}
		}
		return &List{inner: l.inner[start:end]}
	}
	return None
}

// No panic with nil receiver
func (l *List) Store(key Object, value Object) {
	if integer, ok := key.(Integer); ok {
		index := int(integer)
		if index >= 0 && index < l.Size() {
			l.inner[index] = value
		}
	}
}

func (l *List) Size() int {
	if l == nil {
		return 0
	}
	return len(l.inner)
}

type listIterator struct {
	NoneType
	list    *List
	current int
}

func (it *listIterator) Iter() Iterator {
	return it
}

func (it *listIterator) Next() (Object, bool) {
	if it.list == nil {
		return None, false
	}

	inner := it.list.inner
	current := it.current
	if current >= len(inner) {
		return None, false
	}

	it.current++
	return inner[current], true
}

func (*listIterator) Close() {
}

// No panic with nil receiver
func (l *List) Iter() Iterator {
	return &listIterator{list: l}
}

func (l *List) Render(w io.Writer) error {
	for _, value := range l.inner {
		if err := value.Render(w); err != nil {
			return err
		}
	}
	return nil
}

func (l *List) Eval(env Environment) Object {
	it := l.Iter()
	defer it.Close()

	firstElem, _ := it.Next() // None is not an Appliable
	appliable, ok := firstElem.Eval(env).(Appliable)
	if !ok {
		return None
	}
	return appliable.Apply(env, it)
}

func NewList(objects ...Object) *List {
	return &List{inner: objects}
}
