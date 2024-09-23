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

import "iter"

func nilSafe(next func() (Object, bool)) func() (Object, bool) {
	return func() (Object, bool) {
		value, ok := next()
		if value == nil {
			value = None
		}
		return value, ok
	}
}

func Pull(it iter.Seq[Object]) (func() (Object, bool), func()) {
	next, stop := iter.Pull(it)
	return nilSafe(next), stop
}

func Push(next func() (Object, bool)) iter.Seq[Object] {
	return func(yield func(Object) bool) {
		for {
			value, ok := next()
			if !ok || !yield(value) {
				break
			}
		}
	}
}
