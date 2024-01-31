// mautrix-signal - A Matrix-signal puppeting bridge.
// Copyright (C) 2023 Scott Weber
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

package signalmeow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mau.fi/util/exerrors"
	"go.mau.fi/util/random"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"

	"github.com/element-hq/mautrix-signal/pkg/libsignalgo"
	signalpb "github.com/element-hq/mautrix-signal/pkg/signalmeow/protobuf"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow/store"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow/types"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow/web"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow/wspb"
)

type ConfirmDeviceResponse struct {
	ACI      uuid.UUID `json:"uuid"`
	PNI      uuid.UUID `json:"pni,omitempty"`
	DeviceID int       `json:"deviceId"`
}

type ProvisioningState int

const (
	StateProvisioningError ProvisioningState = iota
	StateProvisioningURLReceived
	StateProvisioningDataReceived
	StateProvisioningPreKeysRegistered
)

func (s ProvisioningState) String() string {
	switch s {
	case StateProvisioningError:
		return "StateProvisioningError"
	case StateProvisioningURLReceived:
		return "StateProvisioningURLReceived"
	case StateProvisioningDataReceived:
		return "StateProvisioningDataReceived"
	case StateProvisioningPreKeysRegistered:
		return "StateProvisioningPreKeysRegistered"
	default:
		return fmt.Sprintf("ProvisioningState(%d)", s)
	}
}

// Enum for the provisioningUrl, ProvisioningMessage, and error
type ProvisioningResponse struct {
	State            ProvisioningState
	ProvisioningURL  string
	ProvisioningData *store.DeviceData
	Err              error
}

func PerformProvisioning(ctx context.Context, deviceStore store.DeviceStore, deviceName string) chan ProvisioningResponse {
	log := zerolog.Ctx(ctx).With().Str("action", "perform provisioning").Logger()
	c := make(chan ProvisioningResponse)
	go func() {
		defer close(c)

		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		ws, resp, err := web.OpenWebsocket(ctx, web.WebsocketProvisioningPath)
		if err != nil {
			log.Err(err).Any("resp", resp).Msg("error opening provisioning websocket")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}
		defer ws.Close(websocket.StatusInternalError, "Websocket StatusInternalError")
		provisioningCipher := NewProvisioningCipher()

		provisioningURL, err := startProvisioning(ctx, ws, provisioningCipher)
		if err != nil {
			log.Err(err).Msg("startProvisioning error")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}
		c <- ProvisioningResponse{State: StateProvisioningURLReceived, ProvisioningURL: provisioningURL, Err: err}

		provisioningMessage, err := continueProvisioning(ctx, ws, provisioningCipher)
		if err != nil {
			log.Err(err).Msg("continueProvisioning error")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}
		ws.Close(websocket.StatusNormalClosure, "")

		aciPublicKey := exerrors.Must(libsignalgo.DeserializePublicKey(provisioningMessage.GetAciIdentityKeyPublic()))
		aciPrivateKey := exerrors.Must(libsignalgo.DeserializePrivateKey(provisioningMessage.GetAciIdentityKeyPrivate()))
		aciIdentityKeyPair := exerrors.Must(libsignalgo.NewIdentityKeyPair(aciPublicKey, aciPrivateKey))
		pniPublicKey := exerrors.Must(libsignalgo.DeserializePublicKey(provisioningMessage.GetPniIdentityKeyPublic()))
		pniPrivateKey := exerrors.Must(libsignalgo.DeserializePrivateKey(provisioningMessage.GetPniIdentityKeyPrivate()))
		pniIdentityKeyPair := exerrors.Must(libsignalgo.NewIdentityKeyPair(pniPublicKey, pniPrivateKey))
		profileKey := libsignalgo.ProfileKey(provisioningMessage.GetProfileKey())

		username := *provisioningMessage.Number
		password := random.String(22)
		code := provisioningMessage.ProvisioningCode
		registrationId := mrand.Intn(16383) + 1
		pniRegistrationId := mrand.Intn(16383) + 1
		aciSignedPreKey := GenerateSignedPreKey(1, types.UUIDKindACI, aciIdentityKeyPair)
		pniSignedPreKey := GenerateSignedPreKey(2, types.UUIDKindPNI, pniIdentityKeyPair)
		aciPQLastResortPreKeys := GenerateKyberPreKeys(1, 1, types.UUIDKindACI, aciIdentityKeyPair)
		pniPQLastResortPreKeys := GenerateKyberPreKeys(1, 1, types.UUIDKindPNI, pniIdentityKeyPair)
		aciPQLastResortPreKey := aciPQLastResortPreKeys[0]
		pniPQLastResortPreKey := pniPQLastResortPreKeys[0]
		deviceResponse, err := confirmDevice(
			ctx,
			username,
			password,
			*code,
			registrationId,
			pniRegistrationId,
			aciSignedPreKey,
			pniSignedPreKey,
			aciPQLastResortPreKey,
			pniPQLastResortPreKey,
			aciIdentityKeyPair,
			deviceName,
		)
		if err != nil {
			log.Err(err).Msg("confirmDevice error")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}

		deviceId := 1
		if deviceResponse.DeviceID != 0 {
			deviceId = deviceResponse.DeviceID
		}

		data := &store.DeviceData{
			ACIIdentityKeyPair: aciIdentityKeyPair,
			PNIIdentityKeyPair: pniIdentityKeyPair,
			RegistrationID:     registrationId,
			PNIRegistrationID:  pniRegistrationId,
			ACI:                deviceResponse.ACI,
			PNI:                deviceResponse.PNI,
			DeviceID:           deviceId,
			Number:             *provisioningMessage.Number,
			Password:           password,
		}

		// Store the provisioning data
		err = deviceStore.PutDevice(ctx, data)
		if err != nil {
			log.Err(err).Msg("error storing new device")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}

		device, err := deviceStore.DeviceByACI(ctx, data.ACI)
		if err != nil {
			log.Err(err).Msg("error retrieving new device")
			c <- ProvisioningResponse{State: StateProvisioningError, Err: err}
			return
		}

		// In case this is an existing device, we gotta clear out keys
		device.ClearDeviceKeys(ctx)

		// Store identity keys?
		address, err := libsignalgo.NewUUIDAddress(device.ACI, uint(device.DeviceID))
		if err != nil {
			c <- ProvisioningResponse{
				State: StateProvisioningError,
				Err:   fmt.Errorf("error creating new address: %w", err),
			}
			return
		}
		_, err = device.IdentityStore.SaveIdentityKey(ctx, address, device.ACIIdentityKeyPair.GetIdentityKey())
		if err != nil {
			c <- ProvisioningResponse{
				State: StateProvisioningError,
				Err:   fmt.Errorf("error saving identity key: %w", err),
			}
			return
		}

		// Store signed prekeys (now that we have a device)
		device.PreKeyStoreExtras.SaveSignedPreKey(ctx, types.UUIDKindACI, aciSignedPreKey, true)
		device.PreKeyStoreExtras.SaveSignedPreKey(ctx, types.UUIDKindPNI, pniSignedPreKey, true)
		device.PreKeyStoreExtras.SaveKyberPreKey(ctx, types.UUIDKindACI, aciPQLastResortPreKey, true)
		device.PreKeyStoreExtras.SaveKyberPreKey(ctx, types.UUIDKindPNI, pniPQLastResortPreKey, true)

		// Store our profile key
		err = device.ProfileKeyStore.StoreProfileKey(ctx, data.ACI, profileKey)
		if err != nil {
			c <- ProvisioningResponse{
				State: StateProvisioningError,
				Err:   fmt.Errorf("error storing profile key: %w", err),
			}
			return
		}

		// Return the provisioning data
		c <- ProvisioningResponse{State: StateProvisioningDataReceived, ProvisioningData: data}

		// Generate, store, and register prekeys
		// TODO hacky client construction
		cli := &Client{Store: device}
		err = cli.GenerateAndRegisterPreKeys(ctx, types.UUIDKindACI)
		if err != nil {
			c <- ProvisioningResponse{
				State: StateProvisioningError,
				Err:   fmt.Errorf("error generating and registering ACI prekeys: %w", err),
			}
			return
		}
		err = cli.GenerateAndRegisterPreKeys(ctx, types.UUIDKindPNI)
		if err != nil {
			c <- ProvisioningResponse{
				State: StateProvisioningError,
				Err:   fmt.Errorf("error generating and registering PNI prekeys: %w", err),
			}
			return
		}

		c <- ProvisioningResponse{State: StateProvisioningPreKeysRegistered}
	}()
	return c
}

// Returns the provisioningUrl and an error
func startProvisioning(ctx context.Context, ws *websocket.Conn, provisioningCipher *ProvisioningCipher) (string, error) {
	log := zerolog.Ctx(ctx).With().Str("action", "start provisioning").Logger()
	pubKey := provisioningCipher.GetPublicKey()

	msg := &signalpb.WebSocketMessage{}
	err := wspb.Read(ctx, ws, msg)
	if err != nil {
		log.Err(err).Msg("error reading websocket message")
		return "", err
	}

	// Ensure the message is a request and has a valid verb and path
	if msg.GetType() != signalpb.WebSocketMessage_REQUEST || msg.GetRequest().GetVerb() != http.MethodPut || msg.GetRequest().GetPath() != "/v1/address" {
		return "", fmt.Errorf("unexpected websocket message: %v", msg)
	}

	var provisioningBody signalpb.ProvisioningUuid
	err = proto.Unmarshal(msg.GetRequest().GetBody(), &provisioningBody)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal provisioning UUID: %w", err)
	}

	provisioningURL := (&url.URL{
		Scheme: "sgnl",
		Host:   "linkdevice",
		RawQuery: url.Values{
			"uuid":    []string{provisioningBody.GetUuid()},
			"pub_key": []string{base64.StdEncoding.EncodeToString(exerrors.Must(pubKey.Serialize()))},
		}.Encode(),
	}).String()

	// Create and send response
	response := web.CreateWSResponse(ctx, msg.GetRequest().GetId(), 200)
	err = wspb.Write(ctx, ws, response)
	if err != nil {
		log.Err(err).Msg("error writing websocket message")
		return "", err
	}
	return provisioningURL, nil
}

func continueProvisioning(ctx context.Context, ws *websocket.Conn, provisioningCipher *ProvisioningCipher) (*signalpb.ProvisionMessage, error) {
	log := zerolog.Ctx(ctx).With().Str("action", "continue provisioning").Logger()
	envelope := &signalpb.ProvisionEnvelope{}
	msg := &signalpb.WebSocketMessage{}
	err := wspb.Read(ctx, ws, msg)
	if err != nil {
		log.Err(err).Msg("error reading websocket message")
		return nil, err
	}

	// Wait for provisioning message in a request, then send a response
	if *msg.Type == signalpb.WebSocketMessage_REQUEST &&
		*msg.Request.Verb == http.MethodPut &&
		*msg.Request.Path == "/v1/message" {

		err = proto.Unmarshal(msg.Request.Body, envelope)
		if err != nil {
			return nil, err
		}

		response := web.CreateWSResponse(ctx, *msg.Request.Id, 200)
		err = wspb.Write(ctx, ws, response)
		if err != nil {
			log.Err(err).Msg("error writing websocket message")
			return nil, err
		}
	} else {
		err = fmt.Errorf("invalid provisioning message, type: %v, verb: %v, path: %v", *msg.Type, *msg.Request.Verb, *msg.Request.Path)
		log.Err(err).Msg("problem reading websocket message")
		return nil, err
	}
	provisioningMessage, err := provisioningCipher.Decrypt(envelope)
	return provisioningMessage, err
}

func confirmDevice(
	ctx context.Context,
	username string,
	password string,
	code string,
	registrationId int,
	pniRegistrationId int,
	aciSignedPreKey *libsignalgo.SignedPreKeyRecord,
	pniSignedPreKey *libsignalgo.SignedPreKeyRecord,
	aciPQLastResortPreKey *libsignalgo.KyberPreKeyRecord,
	pniPQLastResortPreKey *libsignalgo.KyberPreKeyRecord,
	aciIdentityKeyPair *libsignalgo.IdentityKeyPair,
	deviceName string,
) (*ConfirmDeviceResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("action", "confirm device").Logger()
	ctx = log.WithContext(ctx)
	encryptedDeviceName, err := EncryptDeviceName(deviceName, aciIdentityKeyPair.GetPublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt device name: %w", err)
	}

	ws, resp, err := web.OpenWebsocket(ctx, web.WebsocketPath)
	if err != nil {
		log.Err(err).Any("resp", resp).Msg("error opening websocket")
		return nil, err
	}
	defer ws.Close(websocket.StatusInternalError, "Websocket StatusInternalError")

	aciSignedPreKeyJson := SignedPreKeyToJSON(aciSignedPreKey)
	pniSignedPreKeyJson := SignedPreKeyToJSON(pniSignedPreKey)

	aciPQLastResortPreKeyJson := KyberPreKeyToJSON(aciPQLastResortPreKey)
	pniPQLastResortPreKeyJson := KyberPreKeyToJSON(pniPQLastResortPreKey)

	data := map[string]any{
		"verificationCode": code,
		"accountAttributes": map[string]any{
			"fetchesMessages":   true,
			"name":              encryptedDeviceName,
			"registrationId":    registrationId,
			"pniRegistrationId": pniRegistrationId,
			"capabilities": map[string]any{
				"pni": true,
			},
		},
		"aciSignedPreKey":       aciSignedPreKeyJson,
		"pniSignedPreKey":       pniSignedPreKeyJson,
		"aciPqLastResortPreKey": aciPQLastResortPreKeyJson,
		"pniPqLastResortPreKey": pniPQLastResortPreKeyJson,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create and send request TODO: Use SignalWebsocket
	request := web.CreateWSRequest(http.MethodPut, "/v1/devices/link", jsonBytes, &username, &password)
	one := uint64(1)
	request.Id = &one
	msg_type := signalpb.WebSocketMessage_REQUEST
	message := &signalpb.WebSocketMessage{
		Type:    &msg_type,
		Request: request,
	}
	err = wspb.Write(ctx, ws, message)
	if err != nil {
		return nil, fmt.Errorf("failed on write protobuf data to websocket: %w", err)
	}

	receivedMsg := &signalpb.WebSocketMessage{}
	err = wspb.Read(ctx, ws, receivedMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to read from websocket after devices call: %w", err)
	}

	status := int(*receivedMsg.Response.Status)
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("non-200 status code (%d) from devices response: %s", status, *receivedMsg.Response.Message)
	}

	// unmarshal JSON response into ConfirmDeviceResponse
	deviceResp := ConfirmDeviceResponse{}
	err = json.Unmarshal(receivedMsg.Response.Body, &deviceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &deviceResp, nil
}
