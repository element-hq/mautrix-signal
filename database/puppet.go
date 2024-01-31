// mautrix-signal - A Matrix-signal puppeting bridge.
// Copyright (C) 2023 Scott Weber, Tulir Asokan
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

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix/id"
)

const (
	puppetBaseSelect = `
        SELECT uuid, number, name, name_quality, avatar_path, avatar_hash, avatar_url, name_set, avatar_set,
               contact_info_set, is_registered, custom_mxid, access_token, first_activity_ts, last_activity_ts
        FROM puppet
	`
	getPuppetBySignalIDQuery   = puppetBaseSelect + `WHERE uuid=$1`
	getPuppetByNumberQuery     = puppetBaseSelect + `WHERE number=$1`
	getPuppetByCustomMXIDQuery = puppetBaseSelect + `WHERE custom_mxid=$1`
	getPuppetsWithCustomMXID   = puppetBaseSelect + `WHERE custom_mxid<>''`
	updatePuppetQuery          = `
		UPDATE puppet SET
			number=$2, name=$3, name_quality=$4, avatar_path=$5, avatar_hash=$6, avatar_url=$7,
			name_set=$8, avatar_set=$9, contact_info_set=$10, is_registered=$11,
			custom_mxid=$12, access_token=$13
		WHERE uuid=$1
	`
	insertPuppetQuery = `
		INSERT INTO puppet (
			uuid, number, name, name_quality, avatar_path, avatar_hash, avatar_url,
			name_set, avatar_set, contact_info_set, is_registered,
			custom_mxid, access_token
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	oneDayMs = 24 * 60 * 60 * 1000
)

type PuppetQuery struct {
	*dbutil.QueryHelper[*Puppet]
}

type Puppet struct {
	qh *dbutil.QueryHelper[*Puppet]

	SignalID    uuid.UUID
	Number      string
	Name        string
	NameQuality int
	AvatarPath  string
	AvatarHash  string
	AvatarURL   id.ContentURI
	NameSet     bool
	AvatarSet   bool

	IsRegistered bool

	CustomMXID     id.UserID
	AccessToken    string
	ContactInfoSet bool

	FirstActivityTs int64
	LastActivityTs  int64
}

func newPuppet(qh *dbutil.QueryHelper[*Puppet]) *Puppet {
	return &Puppet{qh: qh}
}

func (pq *PuppetQuery) GetBySignalID(ctx context.Context, signalID uuid.UUID) (*Puppet, error) {
	return pq.QueryOne(ctx, getPuppetBySignalIDQuery, signalID)
}

func (pq *PuppetQuery) GetByNumber(ctx context.Context, number string) (*Puppet, error) {
	return pq.QueryOne(ctx, getPuppetByNumberQuery, number)
}

func (pq *PuppetQuery) GetByCustomMXID(ctx context.Context, mxid id.UserID) (*Puppet, error) {
	return pq.QueryOne(ctx, getPuppetByCustomMXIDQuery, mxid)
}

func (pq *PuppetQuery) GetAllWithCustomMXID(ctx context.Context) ([]*Puppet, error) {
	return pq.QueryMany(ctx, getPuppetsWithCustomMXID)
}

func (p *Puppet) Scan(row dbutil.Scannable) (*Puppet, error) {
	var number, customMXID sql.NullString
	var firstActivityTs, lastActivityTs sql.NullInt64
	err := row.Scan(
		&p.SignalID,
		&number,
		&p.Name,
		&p.NameQuality,
		&p.AvatarPath,
		&p.AvatarHash,
		&p.AvatarURL,
		&p.NameSet,
		&p.AvatarSet,
		&p.ContactInfoSet,
		&p.IsRegistered,
		&customMXID,
		&p.AccessToken,
		&firstActivityTs,
		&lastActivityTs,
	)
	if err != nil {
		return nil, nil
	}
	p.Number = number.String
	p.CustomMXID = id.UserID(customMXID.String)
	p.FirstActivityTs = firstActivityTs.Int64
	p.LastActivityTs = lastActivityTs.Int64
	return p, nil
}

func (p *Puppet) sqlVariables() []any {
	return []any{
		p.SignalID,
		dbutil.StrPtr(p.Number),
		p.Name,
		p.NameQuality,
		p.AvatarPath,
		p.AvatarHash,
		&p.AvatarURL,
		p.NameSet,
		p.AvatarSet,
		p.ContactInfoSet,
		p.IsRegistered,
		dbutil.StrPtr(p.CustomMXID),
		p.AccessToken,
	}
}

func (p *Puppet) Insert(ctx context.Context) error {
	return p.qh.Exec(ctx, insertPuppetQuery, p.sqlVariables()...)
}

func (p *Puppet) Update(ctx context.Context) error {
	return p.qh.Exec(ctx, updatePuppetQuery, p.sqlVariables()...)
}

func (p *Puppet) UpdateActivityTs(ctx context.Context, activityTs int64) {
	if p.LastActivityTs > activityTs {
		return
	}
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Updating activity time for %s to %d", p.SignalID, activityTs)
	p.LastActivityTs = activityTs
	err := p.qh.Exec(ctx, "UPDATE puppet SET last_activity_ts=$1 WHERE uuid=$2", p.LastActivityTs, p.SignalID)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to update last_activity_ts for %s", p.SignalID)
	}

	if p.FirstActivityTs == 0 {
		p.FirstActivityTs = activityTs
		err = p.qh.Exec(ctx, "UPDATE puppet SET first_activity_ts=$1 WHERE uuid=$2 AND first_activity_ts is NULL", p.FirstActivityTs, p.SignalID)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to update first_activity_ts for %s", p.SignalID)
		}
	}
}

func (pq *PuppetQuery) GetRecentlyActiveCount(ctx context.Context, minActivityDays uint, maxActivityDays *uint) (count uint, err error) {
	q := "SELECT COUNT(*) FROM puppet WHERE (last_activity_ts - first_activity_ts) > $1"
	var row *sql.Row
	lastActivityTs := oneDayMs * minActivityDays
	if maxActivityDays == nil {
		row = pq.GetDB().QueryRow(ctx, q, lastActivityTs)
	} else {
		q += " AND ($2 - last_activity_ts) <= $3"
		row = pq.GetDB().QueryRow(ctx, q, lastActivityTs, time.Now().UnixMilli(), oneDayMs*(*maxActivityDays))
	}
	err = row.Scan(&count)
	return
}
