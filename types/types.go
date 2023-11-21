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
	"fmt"
	"io"
	"strconv"
)

type NoneType struct{}

func (n NoneType) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (n NoneType) Eval(env Environment) Object {
	return None
}

var None = NoneType{}

type Boolean bool

func (b Boolean) WriteTo(w io.Writer) (int64, error) {
	str := "false"
	if b {
		str = "true"
	}
	n, err := io.WriteString(w, str)
	return int64(n), err
}

func (b Boolean) Eval(env Environment) Object {
	return b
}

type Integer int64

func (i Integer) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, strconv.FormatInt(int64(i), 10))
	return int64(n), err
}

func (i Integer) Eval(env Environment) Object {
	return i
}

type Float float64

func (f Float) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, fmt.Sprint(float64(f)))
	return int64(n), err
}

func (f Float) Eval(env Environment) Object {
	return f
}

type Rune rune

func (r Rune) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(r))
	return int64(n), err
}

func (r Rune) Eval(env Environment) Object {
	return r
}

type String string

func (s String) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(s))
	return int64(n), err
}

func (s String) Eval(env Environment) Object {
	return s
}

func (s String) LoadInt(index int) Object {
	if index < 0 || index >= len(s) {
		return None
	}
	return s[index : index+1]
}

func (s String) Load(key Object) Object {
	switch casted := key.(type) {
	case Integer:
		return s.LoadInt(int(casted))
	case Float:
		return s.LoadInt(int(casted))
	case *List:
		max := len(s)
		start, end := extractIndex(casted.inner, max)
		if 0 <= start && start <= end && end <= max {
			return s[start:end]
		}
	}
	return None
}

func (s String) Size() int {
	return len(s)
}

type Identifier string

func (i Identifier) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (i Identifier) Eval(env Environment) Object {
	value, _ := env.LoadStr(string(i))
	return value
}

type NativeAppliable struct {
	NoneType
	inner func(Environment, Iterator) Object
}

func (n NativeAppliable) Apply(env Environment, it Iterable) Object {
	it2 := it.Iter()
	defer it2.Close()
	return n.inner(env, it2)
}

func MakeNativeAppliable(f func(Environment, Iterator) Object) NativeAppliable {
	return NativeAppliable{inner: f}
}
