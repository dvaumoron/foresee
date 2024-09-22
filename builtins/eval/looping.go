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

	"github.com/dvaumoron/foresee/types"
)

const (
	breakKind       loopMarkerKind = iota
	continueKind    loopMarkerKind = iota
	fallthroughKind loopMarkerKind = iota
)

type loopMarkerKind int

type loopMarker struct {
	types.NoneType
	kind  loopMarkerKind
	label string
}

func processLabellable(itArgs iter.Seq[types.Object], kind loopMarkerKind) types.Object {
	ok1 := false
	var arg0 types.Object = types.None
	for arg := range itArgs {
		arg0, ok1 = arg, true
		break
	}

	labelId, ok2 := arg0.(types.Identifier)
	if ok1 && !ok2 {
		panic(errIdentifierType)
	}

	return loopMarker{kind: kind, label: string(labelId)}
}
