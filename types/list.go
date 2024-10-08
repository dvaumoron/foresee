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

import (
	"io"
	"iter"
	"slices"
)

type List struct {
	inner []Object
}

func (l *List) Add(value Object) *List {
	l.inner = append(l.inner, value)
	return l
}

func (l *List) AddAll(it iter.Seq[Object]) *List {
	for value := range it {
		l.Add(value)
	}
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

// No panic with nil receiver
func (l *List) Iter() iter.Seq[Object] {
	if l == nil {
		return slices.Values(([]Object)(nil))
	}

	return slices.Values(l.inner)
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
	next, stop := Pull(l.Iter())
	defer stop()

	firstElem, _ := next() // None is not an Appliable
	appliable, ok := firstElem.Eval(env).(Appliable)
	if !ok {
		return None
	}
	return appliable.Apply(env, Push(next))
}

func NewList(objects ...Object) *List {
	return &List{inner: objects}
}
