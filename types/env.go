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

import "time"

type BaseEnvironment struct {
	NoneType
	objects map[string]Object
}

func (b BaseEnvironment) LoadStr(key string) (Object, bool) {
	res, ok := b.objects[key]
	if !ok {
		return None, false
	}
	return res, true
}

func Load(env StringLoadable, key Object) Object {
	switch casted := key.(type) {
	case Rune:
		res, _ := env.LoadStr(string(casted))
		return res
	case String:
		res, _ := env.LoadStr(string(casted))
		return res
	}
	return None
}

func (b BaseEnvironment) Load(key Object) Object {
	return Load(b, key)
}

func (b BaseEnvironment) Store(key Object, value Object) {
	switch casted := key.(type) {
	case Rune:
		b.objects[string(casted)] = value
	case String:
		b.objects[string(casted)] = value
	}
}

func (b BaseEnvironment) StoreStr(key string, value Object) {
	b.objects[key] = value
}

func (b BaseEnvironment) Delete(key Object) {
	switch casted := key.(type) {
	case Rune:
		delete(b.objects, string(casted))
	case String:
		delete(b.objects, string(casted))
	}
}

func (b BaseEnvironment) DeleteStr(key string) {
	delete(b.objects, key)
}

func (b BaseEnvironment) CopyTo(other Environment) {
	for key, value := range b.objects {
		other.StoreStr(key, value)
	}
}

func (b BaseEnvironment) Size() int {
	return len(b.objects)
}

type chanIterator struct {
	NoneType
	receiver    <-chan Object
	closeSender chan<- NoneType
}

func (it *chanIterator) Iter() Iterator {
	return it
}

func (it *chanIterator) Next() (Object, bool) {
	value, ok := <-it.receiver
	if !ok {
		return None, false
	}
	return value, true
}

func (it *chanIterator) Close() {
	if it.closeSender != nil {
		ticker := time.NewTicker(time.Microsecond)
		defer ticker.Stop()
		for {
			select {
			case it.closeSender <- None:
				close(it.closeSender)
				it.closeSender = nil
				return
			case <-ticker.C:
				if it.closeSender == nil {
					return
				}
			}
		}
	}
}

func (b BaseEnvironment) Iter() Iterator {
	objectChannel := make(chan Object)
	closeChannel := make(chan NoneType)
	go sendMapValue(b.objects, objectChannel, closeChannel)
	return &chanIterator{receiver: objectChannel, closeSender: closeChannel}
}

func sendMapValue(objects map[string]Object, transmitter chan<- Object, shouldClose <-chan NoneType) {
ForLoop:
	for key, value := range objects {
		select {
		case transmitter <- NewList(String(key), value):
		case <-shouldClose:
			break ForLoop
		}
	}
	close(transmitter)
}

func MakeBaseEnvironment() BaseEnvironment {
	return BaseEnvironment{objects: map[string]Object{}}
}

type LocalEnvironment struct {
	BaseEnvironment
	parent Environment
}

func (l LocalEnvironment) LoadStr(key string) (Object, bool) {
	res, ok := l.BaseEnvironment.LoadStr(key)
	if ok {
		return res, true
	}
	return l.parent.LoadStr(key)
}

func (l LocalEnvironment) Load(key Object) Object {
	return Load(l, key)
}

func MakeLocalEnvironment(env Environment) LocalEnvironment {
	return LocalEnvironment{BaseEnvironment: MakeBaseEnvironment(), parent: env}
}
