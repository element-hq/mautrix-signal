// mautrix-signal - A Matrix-signal puppeting bridge.
// Copyright (C) 2023 Sumner Evans
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package libsignalgo

/*
#cgo LDFLAGS: -lsignal_ffi -ldl -lm
#include "./libsignal-ffi.h"
*/
import "C"
import (
	"context"
	"errors"
	"sync"
	"unsafe"
)

type WrappedStore[T any] struct {
	Store T
	Ctx   *CallbackContext
}

var nextUnsafePointer unsafe.Pointer
var stores = make(map[unsafe.Pointer]any)
var storesLock sync.RWMutex

func _putStore(store any) unsafe.Pointer {
	storesLock.Lock()
	defer storesLock.Unlock()
	nextUnsafePointer = unsafe.Add(nextUnsafePointer, 1)
	stores[nextUnsafePointer] = store
	return nextUnsafePointer
}

func _loadStore(storeCtx unsafe.Pointer) any {
	storesLock.RLock()
	defer storesLock.RUnlock()
	return stores[storeCtx]
}

type CallbackContext struct {
	Error  error
	Ctx    context.Context
	Unrefs []unsafe.Pointer
}

func NewCallbackContext(ctx context.Context) *CallbackContext {
	if ctx == nil {
		panic(errors.New("nil context passed to NewCallbackContext"))
	}
	return &CallbackContext{Ctx: ctx}
}

func (ctx *CallbackContext) Unref() {
	storesLock.Lock()
	for _, ptr := range ctx.Unrefs {
		delete(stores, ptr)
	}
	storesLock.Unlock()
}

func wrapStore[T any](ctx *CallbackContext, store T) unsafe.Pointer {
	wrappedStore := _putStore(&WrappedStore[T]{Store: store, Ctx: ctx})
	ctx.Unrefs = append(ctx.Unrefs, wrappedStore)
	return wrappedStore
}

func wrapStoreCallbackCustomReturn[T any](storeCtx unsafe.Pointer, callback func(store T, ctx context.Context) (int, error)) C.int {
	wrap := _loadStore(storeCtx).(*WrappedStore[T])
	retVal, err := callback(wrap.Store, wrap.Ctx.Ctx)
	if err != nil {
		wrap.Ctx.Error = err
	}
	return C.int(retVal)
}

func wrapStoreCallback[T any](storeCtx unsafe.Pointer, callback func(store T, ctx context.Context) error) C.int {
	wrap := _loadStore(storeCtx).(*WrappedStore[T])
	if err := callback(wrap.Store, wrap.Ctx.Ctx); err != nil {
		wrap.Ctx.Error = err
		return -1
	}
	return 0
}
