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
	"strconv"
)

type NativeFunc = func(Environment, iter.Seq[Object]) Object

type NoneType struct{}

func (n NoneType) Render(w io.Writer) error {
	_, err := io.WriteString(w, "nil")
	return err
}

func (n NoneType) Eval(env Environment) Object {
	return None
}

var None = NoneType{}

type Boolean bool

func (b Boolean) Render(w io.Writer) error {
	str := "false"
	if b {
		str = "true"
	}
	_, err := io.WriteString(w, str)
	return err
}

func (b Boolean) Eval(env Environment) Object {
	return b
}

type Integer int64

func (i Integer) Render(w io.Writer) error {
	_, err := io.WriteString(w, strconv.FormatInt(int64(i), 10))
	return err
}

func (i Integer) Eval(env Environment) Object {
	return i
}

type Float float64

func (f Float) Render(w io.Writer) error {
	_, err := io.WriteString(w, strconv.FormatFloat(float64(f), 'g', -1, 64))
	return err
}

func (f Float) Eval(env Environment) Object {
	return f
}

type Rune rune

func (r Rune) Render(w io.Writer) error {
	_, err := io.WriteString(w, strconv.QuoteRune(rune(r)))
	return err
}

func (r Rune) Eval(env Environment) Object {
	return r
}

type String string

func (s String) Render(w io.Writer) error {
	_, err := io.WriteString(w, strconv.Quote(string(s)))
	return err
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

func (i Identifier) Render(w io.Writer) error {
	_, err := io.WriteString(w, string(i))
	return err
}

func (i Identifier) Eval(env Environment) Object {
	value, _ := env.LoadStr(string(i))
	return value
}

type NativeAppliable struct {
	NoneType
	inner NativeFunc
}

func (n NativeAppliable) Apply(env Environment, it Iterable) Object {
	return n.inner(env, it.Iter())
}

func MakeNativeAppliable(f NativeFunc) NativeAppliable {
	return NativeAppliable{inner: f}
}
