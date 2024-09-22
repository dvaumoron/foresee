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

package debug

import (
	"iter"
	"strings"

	"github.com/dvaumoron/foresee/types"
)

var (
	_ types.Appliable   = debugWrapper{}
	_ types.Environment = DebugEnvironment{}
)

// Augment Indentifier with an Apply which display list
type debugWrapper struct {
	types.Identifier
}

func (w debugWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w debugWrapper) Apply(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	var buffer strings.Builder
	buffer.WriteByte('(')
	buffer.WriteString(string(w.Identifier))
	for elem := range itArgs {
		buffer.WriteByte(' ')
		elem.Eval(env).Render(&buffer)
	}
	buffer.WriteByte(')')
	return debugWrapper{Identifier: types.Identifier(buffer.String())}
}

// all Identifier eval return a wrapped Identifier
// (the wrapper is a display list appliable)
type DebugEnvironment struct {
	types.NoneType
}

func (DebugEnvironment) LoadStr(key string) (types.Object, bool) {
	return debugWrapper{Identifier: types.Identifier(key)}, true
}

func (d DebugEnvironment) Load(key types.Object) types.Object {
	return types.Load(d, key)
}

func (DebugEnvironment) CopyTo(_ types.Environment) {
}

func (DebugEnvironment) Delete(_ types.Object) {
}

func (DebugEnvironment) DeleteStr(_ string) {
}

func (DebugEnvironment) Store(_ types.Object, _ types.Object) {
}

func (DebugEnvironment) StoreStr(_ string, _ types.Object) {
}
