package unit

import (
	"context"
	"testing"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
)

func TestCreateLogStream(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPost("monitor/streams", fixtureData)

	status := linodego.StreamStatusActive
	createOpts := linodego.StreamCreateOptions{
		Destinations: []int{12345},
		Label:        "AuditLog-config",
		Type:         linodego.StreamTypeAuditLogs,
		Status:       &status,
	}

	stream, err := base.Client.CreateLogStream(context.Background(), createOpts)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, 456, stream.ID)
	assert.Equal(t, "AuditLog-config", stream.Label)
	assert.Equal(t, linodego.StreamStatusActive, stream.Status)
	assert.Equal(t, linodego.StreamTypeAuditLogs, stream.Type)
	assert.NotNil(t, stream.Created)
	assert.NotNil(t, stream.Updated)
	assert.Len(t, stream.Destinations, 1)
	assert.Equal(t, 12345, stream.Destinations[0].ID)
	assert.Equal(t, linodego.StreamDestinationTypeAkamaiObjectStorage, stream.Destinations[0].Type)
}

func TestListLogStreams(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_streams_list")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams", fixtureData)

	streams, err := base.Client.ListLogStreams(context.Background(), &linodego.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, streams, 1)
	assert.Equal(t, 456, streams[0].ID)
	assert.Equal(t, "AuditLog-config", streams[0].Label)
	assert.Equal(t, linodego.StreamStatusActive, streams[0].Status)
	assert.Equal(t, 1, streams[0].Version)
}

func TestGetLogStream(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/456", fixtureData)

	stream, err := base.Client.GetLogStream(context.Background(), 456)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, 456, stream.ID)
	assert.Equal(t, "AuditLog-config", stream.Label)
	assert.Equal(t, linodego.StreamStatusActive, stream.Status)
	assert.NotNil(t, stream.Created)
	assert.NotNil(t, stream.Updated)
}

func TestUpdateLogStream(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPut("monitor/streams/456", fixtureData)

	label := "AuditLog-config"
	streamType := linodego.StreamTypeAuditLogs
	status := linodego.StreamStatusActive
	updateOpts := linodego.StreamUpdateOptions{
		Destinations: []int{12345},
		Label:        &label,
		Type:         &streamType,
		Status:       &status,
	}

	stream, err := base.Client.UpdateLogStream(context.Background(), 456, updateOpts)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, 456, stream.ID)
	assert.Equal(t, "AuditLog-config", stream.Label)
	assert.Equal(t, linodego.StreamStatusActive, stream.Status)
	assert.Equal(t, 1, stream.Version)
}

func TestListLogStreamHistory(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_streams_history")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/456/history", fixtureData)

	streams, err := base.Client.ListLogStreamHistory(context.Background(), 456, nil)
	assert.NoError(t, err)
	assert.Len(t, streams, 2)
	assert.Equal(t, "AuditLog-config", streams[0].Label)
	assert.Equal(t, 1, streams[0].Version)
	assert.Equal(t, "AuditLog-config-updated", streams[1].Label)
	assert.Equal(t, 2, streams[1].Version)
	assert.Equal(t, linodego.StreamStatusInactive, streams[1].Status)
}

func TestDeleteLogStream(t *testing.T) {
	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockDelete("monitor/streams/456", nil)

	err := base.Client.DeleteLogStream(context.Background(), 456)
	assert.NoError(t, err)
}
