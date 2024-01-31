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

package libsignalgo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/element-hq/mautrix-signal/pkg/libsignalgo"
)

// From PublicAPITests.swift:testAddress
func TestAddress(t *testing.T) {
	setupLogging()

	addr, err := libsignalgo.NewPhoneAddress("addr1", 5)
	assert.NoError(t, err)

	name, err := addr.Name()
	assert.NoError(t, err)
	assert.Equal(t, "addr1", name)

	deviceID, err := addr.DeviceID()
	assert.NoError(t, err)
	assert.Equal(t, uint(5), deviceID)
}
