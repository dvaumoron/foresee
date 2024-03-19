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

func extractRenderString(o types.Object) string {
	var builder strings.Builder
	o.Render(&builder)
	return builder.String()
}
