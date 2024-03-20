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

const (
	Break    loopMarkerKind = iota
	Continue loopMarkerKind = iota
)

type loopMarkerKind int

type loopMarker struct {
	kind  loopMarkerKind
	label string
}
