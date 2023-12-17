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
#cgo LDFLAGS: -lsignal_ffi -ldl
#include "./libsignal-ffi.h"

typedef const SignalProtocolAddress const_address;

typedef const SignalSenderKeyRecord const_sender_key_record;
typedef const uint8_t const_uuid_bytes[16];

extern int signal_load_sender_key_callback(void *store_ctx, SignalSenderKeyRecord**, const_address*, const_uuid_bytes*, void *ctx);
extern int signal_store_sender_key_callback(void *store_ctx, const_address*, const_uuid_bytes*, const_sender_key_record*, void *ctx);
*/
import "C"
import (
	"context"
	"unsafe"

	"github.com/google/uuid"
	gopointer "github.com/mattn/go-pointer"
)

type SenderKeyStore interface {
	LoadSenderKey(sender Address, distributionID uuid.UUID, ctx context.Context) (*SenderKeyRecord, error)
	StoreSenderKey(sender Address, distributionID uuid.UUID, record *SenderKeyRecord, ctx context.Context) error
}

func wrapSenderKeyStore(store SenderKeyStore) *C.SignalSenderKeyStore {
	// TODO this is probably a memory leak since I'm never getting rid of the
	// saved pointer.
	return &C.SignalSenderKeyStore{
		ctx:              gopointer.Save(store),
		load_sender_key:  C.SignalLoadSenderKey(C.signal_load_sender_key_callback),
		store_sender_key: C.SignalStoreSenderKey(C.signal_store_sender_key_callback),
	}
}

//export signal_load_sender_key_callback
func signal_load_sender_key_callback(storeCtx unsafe.Pointer, recordp **C.SignalSenderKeyRecord, address *C.const_address, distributionIDBytes *C.const_uuid_bytes, ctxPtr unsafe.Pointer) C.int {
	return wrapStoreCallback(storeCtx, ctxPtr, func(store SenderKeyStore, ctx context.Context) error {
		distributionID := uuid.UUID(*(*[16]byte)(unsafe.Pointer(distributionIDBytes)))
		record, err := store.LoadSenderKey(
			Address{ptr: (*C.SignalProtocolAddress)(unsafe.Pointer(address))},
			distributionID,
			ctx,
		)
		if err == nil && record != nil {
			*recordp = record.ptr
		}
		return err
	})
}

//export signal_store_sender_key_callback
func signal_store_sender_key_callback(storeCtx unsafe.Pointer, address *C.const_address, distributionIDBytes *C.const_uuid_bytes, senderKeyRecord *C.const_sender_key_record, ctxPtr unsafe.Pointer) C.int {
	return wrapStoreCallback(storeCtx, ctxPtr, func(store SenderKeyStore, ctx context.Context) error {
		distributionID := uuid.UUID(*(*[16]byte)(unsafe.Pointer(distributionIDBytes)))
		record := SenderKeyRecord{ptr: (*C.SignalSenderKeyRecord)(unsafe.Pointer(senderKeyRecord))}
		cloned, err := record.Clone()
		if err != nil {
			return err
		}

		return store.StoreSenderKey(
			Address{ptr: (*C.SignalProtocolAddress)(unsafe.Pointer(address))},
			distributionID,
			cloned,
			ctx,
		)
	})
}
