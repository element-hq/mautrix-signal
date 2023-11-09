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

package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mau.fi/util/configupgrade"
	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/bridge"
	"maunium.net/go/mautrix/bridge/bridgeconfig"
	"maunium.net/go/mautrix/bridge/commands"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"

	"github.com/element-hq/mautrix-signal/config"
	"github.com/element-hq/mautrix-signal/database"
	"github.com/element-hq/mautrix-signal/msgconv/matrixfmt"
	"github.com/element-hq/mautrix-signal/msgconv/signalfmt"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow"
	"github.com/element-hq/mautrix-signal/pkg/signalmeow/store"
)

const activeUserMetricsIntervalSec = 60

//go:embed example-config.yaml
var ExampleConfig string

// Information to find out exactly which commit the bridge was built from.
// These are filled at build time with the -X linker flag.
var (
	Tag       = "unknown"
	Commit    = "unknown"
	BuildTime = "unknown"
)

type SignalBridge struct {
	bridge.Bridge

	Config    *config.Config
	DB        *database.Database
	Metrics   *MetricsHandler
	MeowStore *store.StoreContainer

	provisioning *ProvisioningAPI

	puppetActivity             *PuppetActivity
	activePuppetMetricLoopDone chan bool
	activePuppetMetricRequest  chan bool
	lastBlockingNotification   int64

	usersByMXID     map[id.UserID]*User
	usersBySignalID map[uuid.UUID]*User
	usersLock       sync.Mutex

	managementRooms     map[id.RoomID]*User
	managementRoomsLock sync.Mutex

	portalsByMXID map[id.RoomID]*Portal
	portalsByID   map[database.PortalKey]*Portal
	portalsLock   sync.Mutex

	puppets             map[uuid.UUID]*Puppet
	puppetsByCustomMXID map[id.UserID]*Puppet
	puppetsLock         sync.Mutex

	disappearingMessagesManager *DisappearingMessagesManager
}

var _ bridge.ChildOverride = (*SignalBridge)(nil)

func (br *SignalBridge) GetExampleConfig() string {
	return ExampleConfig
}

func (br *SignalBridge) GetConfigPtr() interface{} {
	br.Config = &config.Config{
		BaseConfig: &br.Bridge.Config,
	}
	br.Config.BaseConfig.Bridge = &br.Config.Bridge
	return br.Config
}

func (br *SignalBridge) Init() {
	br.CommandProcessor = commands.NewProcessor(&br.Bridge)
	br.RegisterCommands()

	signalmeow.SetLogger(br.ZLog.With().Str("component", "signalmeow").Logger())

	br.DB = database.New(br.Bridge.DB)
	br.MeowStore = store.NewStore(br.Bridge.DB, dbutil.ZeroLogger(br.ZLog.With().Str("db_section", "signalmeow").Logger()))

	ss := br.Config.Bridge.Provisioning.SharedSecret
	if len(ss) > 0 && ss != "disable" {
		br.provisioning = &ProvisioningAPI{bridge: br, log: br.ZLog.With().Str("component", "provisioning").Logger()}
	}
	br.disappearingMessagesManager = &DisappearingMessagesManager{
		DB:     br.DB,
		Log:    br.ZLog.With().Str("component", "disappearing messages").Logger(),
		Bridge: br,
	}

	br.Metrics = NewMetricsHandler(br.Config.Metrics.Listen, br.Log.Sub("Metrics"), br.DB, br.puppetActivity)
	br.MatrixHandler.TrackEventDuration = br.Metrics.TrackMatrixEvent

	signalFormatParams = &signalfmt.FormatParams{
		GetUserInfo: func(u uuid.UUID) signalfmt.UserInfo {
			puppet := br.GetPuppetBySignalID(u)
			if puppet == nil {
				return signalfmt.UserInfo{}
			}
			user := br.GetUserBySignalID(u)
			if user != nil {
				return signalfmt.UserInfo{
					MXID: user.MXID,
					Name: puppet.Name,
				}
			}
			return signalfmt.UserInfo{
				MXID: puppet.MXID,
				Name: puppet.Name,
			}
		},
	}
	matrixFormatParams = &matrixfmt.HTMLParser{
		GetUUIDFromMXID: func(userID id.UserID) uuid.UUID {
			parsed, ok := br.ParsePuppetMXID(userID)
			if ok {
				return parsed
			}
			user := br.GetUserByMXIDIfExists(userID)
			if user != nil && user.SignalID != uuid.Nil {
				return user.SignalID
			}
			return uuid.Nil
		},
	}
}

func (br *SignalBridge) logLostPortals(ctx context.Context) {
	exists, err := br.DB.TableExists(ctx, "lost_portals")
	if err != nil {
		br.ZLog.Err(err).Msg("Failed to check if lost_portals table exists")
	} else if !exists {
		return
	}
	lostPortals, err := br.DB.LostPortal.GetAll(ctx)
	if err != nil {
		br.ZLog.Err(err).Msg("Failed to get lost portals")
		return
	} else if len(lostPortals) == 0 {
		return
	}
	lostCountByReceiver := make(map[string]int)
	for _, lost := range lostPortals {
		lostCountByReceiver[lost.Receiver]++
	}
	br.ZLog.Warn().
		Any("count_by_receiver", lostCountByReceiver).
		Msg("Some portals were discarded due to the receiver not being logged into the bridge anymore. " +
			"Use `!signal cleanup-lost-portals` to remove them from the database. " +
			"Alternatively, you can re-insert the data into the portal table with the appropriate receiver column to restore the portals.")
}

func (br *SignalBridge) Start() {
	go br.logLostPortals(context.TODO())
	err := br.MeowStore.Upgrade(context.TODO())
	if err != nil {
		br.Log.Fatalln("Failed to upgrade signalmeow database: %v", err)
		os.Exit(15)
	}
	if br.provisioning != nil {
		br.Log.Debugln("Initializing provisioning API")
		br.provisioning.Init()
	}
	go br.StartUsers()
	br.updateActivePuppetMetricNow()
	go br.loopActivePuppetMetric()
	if br.Config.Metrics.Enabled {
		go br.Metrics.Start()
	}
	go br.disappearingMessagesManager.StartDisappearingLoop(context.TODO())
}

func (br *SignalBridge) Stop() {
	br.Metrics.Stop()
	for _, user := range br.usersByMXID {
		br.Log.Debugln("Disconnecting", user.MXID)
		user.Disconnect()
	}
	close(br.activePuppetMetricLoopDone)
}

func (br *SignalBridge) GetIPortal(mxid id.RoomID) bridge.Portal {
	p := br.GetPortalByMXID(mxid)
	if p == nil {
		return nil
	}
	return p
}

func (br *SignalBridge) GetIUser(mxid id.UserID, create bool) bridge.User {
	p := br.GetUserByMXID(mxid)
	if p == nil {
		return nil
	}
	return p
}

func (br *SignalBridge) IsGhost(mxid id.UserID) bool {
	_, isGhost := br.ParsePuppetMXID(mxid)
	return isGhost
}

func (br *SignalBridge) GetIGhost(mxid id.UserID) bridge.Ghost {
	p := br.GetPuppetByMXID(mxid)
	if p == nil {
		return nil
	}
	return p
}

func (br *SignalBridge) CreatePrivatePortal(roomID id.RoomID, brInviter bridge.User, brGhost bridge.Ghost) {
	inviter := brInviter.(*User)
	puppet := brGhost.(*Puppet)

	log := br.ZLog.With().
		Str("action", "create private portal").
		Str("target_room_id", roomID.String()).
		Str("inviter_mxid", brInviter.GetMXID().String()).
		Str("invitee_uuid", puppet.SignalID.String()).
		Logger()
	log.Debug().Msg("Creating private chat portal")

	key := database.NewPortalKey(puppet.SignalID.String(), inviter.SignalID)
	portal := br.GetPortalByChatID(key)
	ctx := log.WithContext(context.TODO())

	if len(portal.MXID) == 0 {
		br.createPrivatePortalFromInvite(ctx, roomID, inviter, puppet, portal)
		return
	}
	log.Debug().
		Str("existing_room_id", portal.MXID.String()).
		Msg("Existing private chat portal found, trying to invite user")

	ok := portal.ensureUserInvited(ctx, inviter)
	if !ok {
		log.Warn().Msg("Failed to invite user to existing private chat portal. Redirecting portal to new room")
		br.createPrivatePortalFromInvite(ctx, roomID, inviter, puppet, portal)
		return
	}
	intent := puppet.DefaultIntent()
	errorMessage := fmt.Sprintf("You already have a private chat portal with me at [%[1]s](https://matrix.to/#/%[1]s)", portal.MXID)
	errorContent := format.RenderMarkdown(errorMessage, true, false)
	_, _ = intent.SendMessageEvent(ctx, roomID, event.EventMessage, errorContent)
	log.Debug().Msg("Leaving ghost from private chat room after accepting invite because we already have a chat with the user")
	_, _ = intent.LeaveRoom(ctx, roomID)
}

func (br *SignalBridge) createPrivatePortalFromInvite(ctx context.Context, roomID id.RoomID, inviter *User, puppet *Puppet, portal *Portal) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Creating private portal from invite")

	// Check if room is already encrypted
	var existingEncryption event.EncryptionEventContent
	var encryptionEnabled bool
	err := portal.MainIntent().StateEvent(ctx, roomID, event.StateEncryption, "", &existingEncryption)
	if err != nil {
		log.Err(err).Msg("Failed to check if encryption is enabled in private chat room")
	} else {
		encryptionEnabled = existingEncryption.Algorithm == id.AlgorithmMegolmV1
	}
	portal.MXID = roomID
	br.portalsLock.Lock()
	br.portalsByMXID[portal.MXID] = portal
	br.portalsLock.Unlock()
	intent := puppet.DefaultIntent()

	if br.Config.Bridge.Encryption.Default || encryptionEnabled {
		log.Debug().Msg("Adding bridge bot to new private chat portal as encryption is enabled")
		_, err = intent.InviteUser(ctx, roomID, &mautrix.ReqInviteUser{UserID: br.Bot.UserID})
		if err != nil {
			log.Err(err).Msg("Failed to invite bridge bot to enable e2be")
		}
		err = br.Bot.EnsureJoined(ctx, roomID)
		if err != nil {
			log.Err(err).Msg("Failed to join as bridge bot to enable e2be")
		}
		if !encryptionEnabled {
			_, err = intent.SendStateEvent(ctx, roomID, event.StateEncryption, "", portal.getEncryptionEventContent())
			if err != nil {
				log.Err(err).Msg("Failed to enable e2be")
			}
		}
		br.AS.StateStore.SetMembership(ctx, roomID, inviter.MXID, event.MembershipJoin)
		br.AS.StateStore.SetMembership(ctx, roomID, puppet.MXID, event.MembershipJoin)
		br.AS.StateStore.SetMembership(ctx, roomID, br.Bot.UserID, event.MembershipJoin)
		portal.Encrypted = true
	}
	portal.UpdateDMInfo(ctx, true)
	_, _ = intent.SendNotice(ctx, roomID, "Private chat portal created")
	log.Info().Msg("Created private chat portal after invite")
}

func (br *SignalBridge) UpdateActivePuppetMetric() {
	defer func() {
		if recover() != nil {
			br.ZLog.Warn().Msg("Attempted to update active puppet metrics after bridge has stopped")
		}
	}()
	br.activePuppetMetricRequest <- true
}

func (br *SignalBridge) updateActivePuppetMetricNow() {
	br.ZLog.Debug().Msg("Updating active puppet count")
	br.updateActivePuppetMetric(br.ZLog)
}

func (br *SignalBridge) updateActivePuppetMetric(log *zerolog.Logger) {
	activeUsers, err := br.DB.Puppet.GetRecentlyActiveCount(
		context.TODO(),
		br.Config.Bridge.Limits.MinPuppetActivityDays,
		br.Config.Bridge.Limits.PuppetInactivityDays,
	)
	if err != nil {
		log.Warn().Msg("Failed to scan number of active puppets")
		return
	}

	if br.Config.Bridge.Limits.BlockOnLimitReached {
		blocked := br.Config.Bridge.Limits.MaxPuppetLimit < activeUsers
		if br.puppetActivity.isBlocked != blocked {
			br.puppetActivity.isBlocked = blocked
			br.notifyBridgeBlocked(blocked)
		}
	}
	log.Debug().Msgf("Current active puppet count is %d", activeUsers)
	br.puppetActivity.currentUserCount = activeUsers

	if br.Config.Metrics.Enabled {
		br.Metrics.updatePuppetActivity()
	}
}

func (br *SignalBridge) loopActivePuppetMetric() {
	log := br.ZLog.With().Str("module", "mau.active_puppet_metric").Logger()

	ticker := time.Tick(activeUserMetricsIntervalSec * time.Second)
	for {
		select {
		case <-ticker:
			log.Info().Msg("Executing periodic active puppet metric check")
			br.updateActivePuppetMetric(&log)
		case <-br.activePuppetMetricRequest:
			br.updateActivePuppetMetricNow()
		case <-br.activePuppetMetricLoopDone:
			close(br.activePuppetMetricRequest)
			return
		}
	}
}

func (br *SignalBridge) notifyBridgeBlocked(isBlocked bool) {
	var msg string
	if isBlocked {
		msg = br.Config.Bridge.Limits.BlockBeginsNotification
		nextNotification := br.lastBlockingNotification + int64(br.Config.Bridge.Limits.BlockNotificationIntervalSeconds)
		// We're only checking if the block is active, since the unblock notification will not be resent and we want it ASAP
		if now := time.Now().Unix(); nextNotification > now {
			return
		} else {
			br.lastBlockingNotification = now
		}
	} else {
		msg = br.Config.Bridge.Limits.BlockEndsNotification
	}

	admins := make([]id.UserID, 0, len(br.Config.Bridge.Permissions))
	for key, permissionLevel := range br.Config.Bridge.Permissions {
		if permissionLevel == bridgeconfig.PermissionLevelAdmin {
			// Only explicit MXIDs are notified, not wildcards or domains
			if _, _, err := id.UserID(key).ParseAndValidate(); err == nil {
				admins = append(admins, id.UserID(key))
			}
		}
	}
	if len(admins) == 0 {
		br.ZLog.Debug().Msg("No bridge admins to notify about the bridge being blocked")
		return
	}

	allAdmins := string(admins[0])
	for _, adminMXID := range admins[1:] {
		allAdmins += "," + string(adminMXID)
	}
	br.ZLog.Debug().Msgf("Notifying bridge admins (%s) about bridge being blocked", allAdmins)
	ctx := br.ZLog.WithContext(context.TODO())
	for _, adminMXID := range admins {
		admin := br.GetUserByMXID(adminMXID)
		if admin == nil {
			continue
		}

		roomID := admin.GetManagementRoomID()
		if roomID == "" {
			resp, err := br.Bot.CreateRoom(ctx, &mautrix.ReqCreateRoom{
				Name:     "Signal Bridge notice room",
				IsDirect: true,
				Invite:   []id.UserID{adminMXID},
			})
			if err != nil {
				br.ZLog.Warn().Err(err).Msg("Failed to create notice room")
				continue
			}
			roomID = resp.RoomID
			admin.SetManagementRoom(roomID)
		}

		// \u26a0 is a warning sign
		br.Bot.SendNotice(ctx, roomID, "\u26a0 "+msg)
	}
}

func main() {
	br := &SignalBridge{
		usersByMXID:     make(map[id.UserID]*User),
		usersBySignalID: make(map[uuid.UUID]*User),

		managementRooms: make(map[id.RoomID]*User),

		portalsByMXID: make(map[id.RoomID]*Portal),
		portalsByID:   make(map[database.PortalKey]*Portal),

		puppets:             make(map[uuid.UUID]*Puppet),
		puppetsByCustomMXID: make(map[id.UserID]*Puppet),

		puppetActivity: &PuppetActivity{
			currentUserCount: 0,
			isBlocked:        false,
		},
		activePuppetMetricLoopDone: make(chan bool),
		activePuppetMetricRequest:  make(chan bool),
	}
	br.Bridge = bridge.Bridge{
		Name:              "mautrix-signal",
		URL:               "https://github.com/element-hq/mautrix-signal",
		Description:       "A Matrix-Signal puppeting bridge.",
		Version:           "0.4.99-mod-1",
		ProtocolName:      "Signal",
		BeeperServiceName: "signal",
		BeeperNetworkName: "signal",

		CryptoPickleKey: "mautrix.bridge.e2ee",

		ConfigUpgrader: &configupgrade.StructUpgrader{
			SimpleUpgrader: configupgrade.SimpleUpgrader(config.DoUpgrade),
			Blocks:         config.SpacedBlocks,
			Base:           ExampleConfig,
		},

		Child: br,
	}
	br.InitVersion(Tag, Commit, BuildTime)

	br.Main()
}
