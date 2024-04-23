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

package stack

type Stack[T any] struct {
	inner []T
}

func (s *Stack[T]) Push(e T) {
	s.inner = append(s.inner, e)
}

func (s *Stack[T]) Peek() T {
	return s.inner[len(s.inner)-1]
}

func (s *Stack[T]) Pop() T {
	last := len(s.inner) - 1
	res := s.inner[last]
	s.inner = s.inner[:last]
	return res
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}
