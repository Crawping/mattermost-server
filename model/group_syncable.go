// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"net/http"
)

type GroupSyncableType int

const (
	GroupSyncableTypeTeam GroupSyncableType = iota
	GroupSyncableTypeChannel
)

var GroupSyncableTypes = []GroupSyncableType{GroupSyncableTypeTeam, GroupSyncableTypeChannel}

func (gst GroupSyncableType) String() string {
	// Order matters. Keep in sync with iotas.
	return [...]string{"Team", "Channel"}[gst]
}

type GroupSyncable struct {
	GroupId string `json:"group_id"`

	// SyncableId represents the Id of the model that is being synced with the group, for example a ChannelId or
	// TeamId.
	SyncableId string `db:"-" json:"-"`

	CanLeave bool              `json:"can_leave"`
	AutoAdd  bool              `json:"auto_add"`
	CreateAt int64             `json:"create_at"`
	DeleteAt int64             `json:"delete_at"`
	UpdateAt int64             `json:"update_at"`
	Type     GroupSyncableType `db:"-" json:"-"`

	// Values joined in from the associated team and/or channel
	ChannelDisplayName string `db:"-" json:"-"`
	TeamDisplayName    string `db:"-" json:"-"`
	TeamType           string `db:"-" json:"-"`
	ChannelType        string `db:"-" json:"-"`
	TeamID             string `db:"-" json:"-"`
}

func (syncable *GroupSyncable) IsValid() *AppError {
	if !IsValidId(syncable.GroupId) {
		return NewAppError("GroupSyncable.SyncableIsValid", "model.group_syncable.group_id.app_error", nil, "", http.StatusBadRequest)
	}
	if !IsValidId(syncable.SyncableId) {
		return NewAppError("GroupSyncable.SyncableIsValid", "model.group_syncable.syncable_id.app_error", nil, "", http.StatusBadRequest)
	}
	if syncable.AutoAdd == false && syncable.CanLeave == false {
		return NewAppError("GroupSyncable.SyncableIsValid", "model.group_syncable.invalid_state", nil, "", http.StatusBadRequest)
	}
	return nil
}

func (syncable *GroupSyncable) UnmarshalJSON(b []byte) error {
	var kvp map[string]interface{}
	err := json.Unmarshal(b, &kvp)
	if err != nil {
		return err
	}
	for key, value := range kvp {
		switch key {
		case "team_id":
			syncable.SyncableId = value.(string)
			syncable.Type = GroupSyncableTypeTeam
		case "channel_id":
			syncable.SyncableId = value.(string)
			syncable.Type = GroupSyncableTypeChannel
		case "group_id":
			syncable.GroupId = value.(string)
		case "can_leave":
			syncable.CanLeave = value.(bool)
		case "auto_add":
			syncable.AutoAdd = value.(bool)
		default:
		}
	}
	return nil
}

func (syncable *GroupSyncable) MarshalJSON() ([]byte, error) {
	type Alias GroupSyncable

	switch syncable.Type {
	case GroupSyncableTypeTeam:
		return json.Marshal(&struct {
			TeamID          string `json:"team_id"`
			TeamDisplayName string `json:"team_display_name"`
			TeamType        string `json:"team_type"`
			*Alias
		}{
			TeamDisplayName: syncable.TeamDisplayName,
			TeamType:        syncable.TeamType,
			TeamID:          syncable.SyncableId,
			Alias:           (*Alias)(syncable),
		})
	case GroupSyncableTypeChannel:
		return json.Marshal(&struct {
			ChannelID          string `json:"channel_id"`
			ChannelDisplayName string `json:"channel_display_name"`
			ChannelType        string `json:"channel_type"`

			TeamID          string `json:"team_id"`
			TeamDisplayName string `json:"team_display_name"`
			TeamType        string `json:"team_type"`

			*Alias
		}{
			ChannelID:          syncable.SyncableId,
			ChannelDisplayName: syncable.ChannelDisplayName,
			ChannelType:        syncable.ChannelType,

			TeamID:          syncable.TeamID,
			TeamDisplayName: syncable.TeamDisplayName,
			TeamType:        syncable.TeamType,

			Alias: (*Alias)(syncable),
		})
	default:
		return nil, &json.MarshalerError{}
	}
}

type GroupSyncablePatch struct {
	CanLeave *bool `json:"can_leave"`
	AutoAdd  *bool `json:"auto_add"`
}

func (syncable *GroupSyncable) Patch(patch *GroupSyncablePatch) {
	if patch.CanLeave != nil {
		syncable.CanLeave = *patch.CanLeave
	}
	if patch.AutoAdd != nil {
		syncable.AutoAdd = *patch.AutoAdd
	}
}

type UserTeamIDPair struct {
	UserID string
	TeamID string
}

type UserChannelIDPair struct {
	UserID    string
	ChannelID string
}
