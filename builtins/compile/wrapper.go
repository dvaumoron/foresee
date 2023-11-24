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

package compile

import (
	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/types"
)

var _ types.Appliable = appliableWrapper{}

type appliableWrapper struct {
	types.Renderer
}

func (w appliableWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w appliableWrapper) Apply(env types.Environment, itArgs types.Iterable) types.Object {
	if casted, ok := w.Renderer.(*jen.Statement); ok {
		argsCode := compileToCodeSlice(env, itArgs)
		return appliableWrapper{Renderer: casted.Clone().Call(argsCode...)}
	}
	return w
}

// all unknown Identifier eval return a wrapped jen.Id
// (the wrapper is a function call form appliable)
type compileEnvironment struct {
	types.Environment
}

func (c compileEnvironment) LoadStr(key string) (types.Object, bool) {
	res, ok := c.Environment.LoadStr(key)
	if !ok {
		res = appliableWrapper{Renderer: jen.Id(key)}
	}
	return res, true
}

func (c compileEnvironment) Load(key types.Object) types.Object {
	return types.Load(c, key)
}
