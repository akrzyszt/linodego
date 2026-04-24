package unit

import (
	"context"
	"testing"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
)

const (
	testLogsDestinationID = 12345
	testLogStreamID       = 456
)

// ---- Logs Destination tests ----

func TestCreateLogsDestination(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_get")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPost("monitor/streams/destinations", fixtureData)

	path := "audit-logs"
	opts := linodego.LogsDestinationCreateOptions{
		Label: "my-logs-destination",
		Type:  linodego.LogsDestinationTypeAkamaiObjectStorage,
		Details: linodego.LogsDestinationDetailsCreateOptions{
			AccessKeyID: "123",
			BucketName:  "primary-bucket",
			Host:        "primary-bucket-1.us-east-12.linodeobjects.com",
			Path:        &path,
		},
	}

	dest, err := base.Client.CreateLogsDestination(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, dest)
	assert.Equal(t, testLogsDestinationID, dest.ID)
	assert.Equal(t, "OBJ_logs_destination", dest.Label)
	assert.Equal(t, linodego.LogsDestinationStatusActive, dest.Status)
	assert.Equal(t, linodego.LogsDestinationTypeAkamaiObjectStorage, dest.Type)
	assert.Equal(t, "123", string(dest.Details.AccessKeyID))
	assert.Equal(t, "primary-bucket", dest.Details.BucketName)
	assert.Equal(t, "primary-bucket-1.us-iad-12.linodeobjects.com", dest.Details.Host)
	assert.Equal(t, "audit-logs", dest.Details.Path)
	assert.NotNil(t, dest.Created)
	assert.NotNil(t, dest.Updated)
}

func TestCreateLogsDestination_NoPath(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_get")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPost("monitor/streams/destinations", fixtureData)

	opts := linodego.LogsDestinationCreateOptions{
		Label: "my-logs-destination",
		Type:  linodego.LogsDestinationTypeAkamaiObjectStorage,
		Details: linodego.LogsDestinationDetailsCreateOptions{
			AccessKeyID:     "1ABCD23EFG4HIJKLMNO5",
			AccessKeySecret: "1aB2CD3e4fgHi5JK6lmnop7qR8STU9VxYzabcdefHh",
			BucketName:      "primary-bucket",
			Host:            "primary-bucket-1.us-east-12.linodeobjects.com",
			// Path intentionally omitted
		},
	}

	dest, err := base.Client.CreateLogsDestination(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, dest)
	assert.Equal(t, testLogsDestinationID, dest.ID)
}

func TestGetLogsDestination(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_get")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/destinations/12345", fixtureData)

	dest, err := base.Client.GetLogsDestination(context.Background(), testLogsDestinationID)
	assert.NoError(t, err)
	assert.NotNil(t, dest)
	assert.Equal(t, testLogsDestinationID, dest.ID)
	assert.Equal(t, "OBJ_logs_destination", dest.Label)
	assert.Equal(t, linodego.LogsDestinationStatusActive, dest.Status)
	assert.Equal(t, linodego.LogsDestinationTypeAkamaiObjectStorage, dest.Type)
	assert.Equal(t, "John Q. Linode", dest.CreatedBy)
	assert.Equal(t, "Jane Q. Linode", dest.UpdatedBy)
	assert.Equal(t, 1, dest.Version)
	assert.Equal(t, "123", string(dest.Details.AccessKeyID))
	assert.Equal(t, "primary-bucket", dest.Details.BucketName)
	assert.Equal(t, "primary-bucket-1.us-iad-12.linodeobjects.com", dest.Details.Host)
	assert.Equal(t, "audit-logs", dest.Details.Path)
	assert.NotNil(t, dest.Created)
	assert.NotNil(t, dest.Updated)
}

func TestListLogsDestinations(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_list")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/destinations", fixtureData)

	dests, err := base.Client.ListLogsDestinations(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, dests, 1)
	assert.Equal(t, testLogsDestinationID, dests[0].ID)
	assert.Equal(t, "OBJ_logs_destination", dests[0].Label)
	assert.Equal(t, linodego.LogsDestinationStatusActive, dests[0].Status)
	assert.Equal(t, linodego.LogsDestinationTypeAkamaiObjectStorage, dests[0].Type)
	assert.Equal(t, "123", string(dests[0].Details.AccessKeyID))
}

func TestUpdateLogsDestination(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_get")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPut("monitor/streams/destinations/12345", fixtureData)

	opts := linodego.LogsDestinationUpdateOptions{
		Label: "my-logs-destination-renamed",
	}

	dest, err := base.Client.UpdateLogsDestination(context.Background(), testLogsDestinationID, opts)
	assert.NoError(t, err)
	assert.NotNil(t, dest)
}

func TestDeleteLogsDestination(t *testing.T) {
	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockDelete("monitor/streams/destinations/12345", nil)

	err := base.Client.DeleteLogsDestination(context.Background(), testLogsDestinationID)
	assert.NoError(t, err)
}

func TestListLogsDestinationHistory(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_destinations_list")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/destinations/12345/history", fixtureData)

	history, err := base.Client.ListLogsDestinationHistory(context.Background(), testLogsDestinationID, nil)
	assert.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, testLogsDestinationID, history[0].ID)
	assert.Equal(t, "OBJ_logs_destination", history[0].Label)
	assert.Equal(t, linodego.LogsDestinationStatusActive, history[0].Status)
	assert.Equal(t, linodego.LogsDestinationTypeAkamaiObjectStorage, history[0].Type)
	assert.Equal(t, 1, history[0].Version)
	assert.Equal(t, "123", string(history[0].Details.AccessKeyID))
	assert.Equal(t, "primary-bucket", history[0].Details.BucketName)
	assert.NotNil(t, history[0].Created)
	assert.NotNil(t, history[0].Updated)
}

// ---- Log Stream tests ----

func TestCreateLogStream(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPost("monitor/streams", fixtureData)

	status := linodego.StreamStatusActive
	createOpts := linodego.StreamCreateOptions{
		Destinations: []int{testLogsDestinationID},
		Label:        "AuditLog-config",
		Type:         linodego.StreamTypeAuditLogs,
		Status:       &status,
	}

	stream, err := base.Client.CreateLogStream(context.Background(), createOpts)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, testLogStreamID, stream.ID)
	assert.Equal(t, "AuditLog-config", stream.Label)
	assert.Equal(t, linodego.StreamStatusActive, stream.Status)
	assert.Equal(t, linodego.StreamTypeAuditLogs, stream.Type)
	assert.NotNil(t, stream.Created)
	assert.NotNil(t, stream.Updated)
	assert.Len(t, stream.Destinations, 1)
	assert.Equal(t, testLogsDestinationID, stream.Destinations[0].ID)
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
	assert.Equal(t, testLogStreamID, streams[0].ID)
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

	stream, err := base.Client.GetLogStream(context.Background(), testLogStreamID)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, testLogStreamID, stream.ID)
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
		Destinations: []int{testLogsDestinationID},
		Label:        &label,
		Type:         &streamType,
		Status:       &status,
	}

	stream, err := base.Client.UpdateLogStream(context.Background(), testLogStreamID, updateOpts)
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, testLogStreamID, stream.ID)
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

	streams, err := base.Client.ListLogStreamHistory(context.Background(), testLogStreamID, nil)
	assert.NoError(t, err)
	assert.Len(t, streams, 2)
	assert.Equal(t, "AuditLog-config", streams[0].Label)
	assert.Equal(t, 1, streams[0].Version)
	assert.Equal(t, "AuditLog-config-updated", streams[1].Label)
	assert.Equal(t, 2, streams[1].Version)
	assert.Equal(t, linodego.StreamStatusInactive, streams[1].Status)
}

func TestGetLogStream_DestinationDetails(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockGet("monitor/streams/456", fixtureData)

	stream, err := base.Client.GetLogStream(context.Background(), testLogStreamID)
	assert.NoError(t, err)
	assert.Len(t, stream.Destinations, 1)

	dest := stream.Destinations[0]
	assert.Equal(t, testLogsDestinationID, dest.ID)
	assert.Equal(t, "OBJ_logs_destination", dest.Label)
	assert.Equal(t, linodego.StreamDestinationTypeAkamaiObjectStorage, dest.Type)
	assert.Equal(t, 123, dest.Details.AccessKeyID)
	assert.Equal(t, "primary-bucket", dest.Details.BucketName)
	assert.Equal(t, "primary-bucket-1.us-iad-12.linodeobjects.com", dest.Details.Host)
	assert.Equal(t, "audit-logs", dest.Details.Path)
}

func TestUpdateLogStream_DestinationsOnly(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPut("monitor/streams/456", fixtureData)

	stream, err := base.Client.UpdateLogStream(context.Background(), testLogStreamID, linodego.StreamUpdateOptions{
		Destinations: []int{testLogsDestinationID},
	})
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Len(t, stream.Destinations, 1)
	assert.Equal(t, testLogsDestinationID, stream.Destinations[0].ID)
}

func TestUpdateLogStream_LabelAndStatus(t *testing.T) {
	fixtureData, err := fixtures.GetFixture("monitor_log_stream")
	assert.NoError(t, err)

	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockPut("monitor/streams/456", fixtureData)

	label := "AuditLog-config"
	status := linodego.StreamStatusActive
	stream, err := base.Client.UpdateLogStream(context.Background(), testLogStreamID, linodego.StreamUpdateOptions{
		Label:  &label,
		Status: &status,
	})
	assert.NoError(t, err)
	assert.Equal(t, "AuditLog-config", stream.Label)
	assert.Equal(t, linodego.StreamStatusActive, stream.Status)
}

func TestDeleteLogStream(t *testing.T) {
	var base ClientBaseCase
	base.SetUp(t)
	defer base.TearDown(t)

	base.MockDelete("monitor/streams/456", nil)

	err := base.Client.DeleteLogStream(context.Background(), testLogStreamID)
	assert.NoError(t, err)
}
