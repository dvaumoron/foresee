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
	"iter"
	"strings"

	"github.com/dvaumoron/foresee/types"
)

func extractBoolean(o types.Object) bool {
	switch casted := o.(type) {
	case types.NoneType:
		return false
	case types.Boolean:
		return bool(casted)
	case types.Integer:
		return casted != 0
	case types.Float:
		return casted != 0
	case types.Sizable:
		return casted.Size() != 0
	}

	return true
}

func extractFloat(o types.Object) float64 {
	switch casted := o.(type) {
	case types.Integer:
		return float64(casted)
	case types.Float:
		return float64(casted)
	}
	panic(errNumericType)
}

func extractInteger(o types.Object) int64 {
	switch casted := o.(type) {
	case types.Integer:
		return int64(casted)
	case types.Float:
		return int64(casted)
	}
	panic(errNumericType)
}

func extractRenderString(o types.Object) string {
	var builder strings.Builder
	o.Render(&builder)
	return builder.String()
}

func floatConvFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	for arg := range itArgs {
		return types.Float(extractFloat(arg.Eval(env)))
	}
	return types.Float(0)
}

func intConvFunc(env types.Environment, itArgs iter.Seq[types.Object]) types.Object {
	for arg := range itArgs {
		return types.Integer(extractInteger(arg.Eval(env)))
	}
	return types.Integer(0)
}
