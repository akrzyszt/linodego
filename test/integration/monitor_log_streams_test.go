package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	aclpLogsCapability = "aclp_logs"
)

// requireACLPLogsStreamTests skips the test if RUN_ACLP_LOGS_STREAM_TESTS is not set.
// Call this before creating a test client so the env check short-circuits early.
func requireACLPLogsStreamTests(t *testing.T) {
	t.Helper()
	val := os.Getenv("RUN_ACLP_LOGS_STREAM_TESTS")
	if val != "yes" && val != "true" {
		t.Skipf("RUN_ACLP_LOGS_STREAM_TESTS must be set to 'yes' or 'true' to run stream tests")
	}
}

// requireACLPLogsCapability skips the test if the aclp_logs capability is not enabled.
// Call this after creating a test client.
func requireACLPLogsCapability(t *testing.T, client *linodego.Client) {
	t.Helper()
	account, err := client.GetAccount(context.Background())
	require.NoError(t, err)
	for _, cap := range account.Capabilities {
		if cap == aclpLogsCapability {
			return
		}
	}
	t.Skipf("aclp_logs capability not enabled for this account")
}

// setupObjectStorageResources creates an Object Storage key and bucket for use with log destinations.
// The teardown function empties the bucket, then deletes both the bucket and the key.
func setupObjectStorageResources(
	t *testing.T,
	client *linodego.Client,
) (*linodego.ObjectStorageKey, *linodego.ObjectStorageBucket, func()) {
	t.Helper()
	ctx := context.Background()

	key, err := client.CreateObjectStorageKey(ctx, linodego.ObjectStorageKeyCreateOptions{
		Label: fmt.Sprintf("go-test-aclp-key-%d", time.Now().UnixNano()),
	})
	require.NoError(t, err)

	corsEnabled := false
	bucket, err := client.CreateObjectStorageBucket(ctx, linodego.ObjectStorageBucketCreateOptions{
		Region:      "us-southeast",
		Label:       fmt.Sprintf("go-aclp-%d", time.Now().UnixNano()),
		ACL:         linodego.ACLPrivate,
		CorsEnabled: &corsEnabled,
	})
	if err != nil {
		_ = client.DeleteObjectStorageKey(ctx, key.ID)
		require.NoError(t, err)
	}

	teardown := func() {
		emptyObjectStorageBucket(ctx, client, bucket)
		_ = client.DeleteObjectStorageBucket(ctx, bucket.Region, bucket.Label)
		_ = client.DeleteObjectStorageKey(ctx, key.ID)
	}

	return key, bucket, teardown
}

// emptyObjectStorageBucket deletes all objects in a bucket via presigned DELETE URLs.
// Errors are silently ignored since this is best-effort cleanup.
func emptyObjectStorageBucket(ctx context.Context, client *linodego.Client, bucket *linodego.ObjectStorageBucket) {
	expiresIn := 300
	contents, err := client.ListObjectStorageBucketContents(ctx, bucket.Region, bucket.Label, nil)
	if err != nil || contents == nil {
		return
	}
	for _, obj := range contents.Data {
		urlResp, err := client.CreateObjectStorageObjectURL(ctx, bucket.Region, bucket.Label, linodego.ObjectStorageObjectURLCreateOptions{
			Name:      obj.Name,
			Method:    "DELETE",
			ExpiresIn: &expiresIn,
		})
		if err != nil {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlResp.URL, nil)
		if err != nil {
			continue
		}
		if resp, err := http.DefaultClient.Do(req); err == nil {
			resp.Body.Close()
		}
	}
}

// setupLogDestination creates an ACLP logs destination backed by the given OBJ key and bucket.
func setupLogDestination(
	t *testing.T,
	client *linodego.Client,
	key *linodego.ObjectStorageKey,
	bucket *linodego.ObjectStorageBucket,
) (*linodego.Destination, func(), error) {
	t.Helper()
	ctx := context.Background()

	dest, err := client.CreateLogDestination(ctx, linodego.DestinationCreateOptions{
		Label: fmt.Sprintf("go-test-aclp-dest-%d", time.Now().UnixNano()),
		Type:  linodego.DestinationTypeAkamaiObjectStorage,
		Details: linodego.DestinationCreateDetails{
			AccessKeyID:     key.AccessKey,
			AccessKeySecret: key.SecretKey,
			BucketName:      bucket.Label,
			Host:            bucket.Hostname,
		},
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("creating log destination: %w", err)
	}

	teardown := func() {
		_ = client.DeleteLogDestination(context.Background(), dest.ID)
	}

	return dest, teardown, nil
}

// setupLogStream creates a log stream with the given destination IDs.
func setupLogStream(
	t *testing.T,
	client *linodego.Client,
	destIDs []int,
) (*linodego.Stream, func(), error) {
	t.Helper()

	stream, err := client.CreateLogStream(context.Background(), linodego.StreamCreateOptions{
		Label:        fmt.Sprintf("go-test-log-stream-%d", time.Now().UnixNano()),
		Type:         linodego.StreamTypeAuditLogs,
		Destinations: destIDs,
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("creating log stream: %w", err)
	}

	teardown := func() {
		_ = client.DeleteLogStream(context.Background(), stream.ID)
	}

	return stream, teardown, nil
}

// waitForLogStreamProvisioned polls until the stream leaves provisioning state.
// Stream provisioning can take up to 60 minutes; timeoutSeconds should be set accordingly.
func waitForLogStreamProvisioned(
	ctx context.Context,
	client *linodego.Client,
	streamID int,
	pollIntervalSeconds int,
	timeoutSeconds int,
) (*linodego.Stream, error) {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	for {
		stream, err := client.GetLogStream(ctx, streamID)
		if err != nil {
			return nil, err
		}
		if stream.Status == linodego.StreamStatusActive || stream.Status == linodego.StreamStatusInactive {
			return stream, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for log stream %d to leave provisioning state", streamID)
		}
		time.Sleep(time.Duration(pollIntervalSeconds) * time.Second)
	}
}

func TestLogStream_Create_InvalidDestination(t *testing.T) {
	requireACLPLogsStreamTests(t)

	client, teardown := createTestClient(t, "fixtures/TestLogStream_Create_InvalidDestination")
	defer teardown()
	requireACLPLogsCapability(t, client)

	// Only one stream is allowed per account — skip if one already exists.
	existing, err := client.ListLogStreams(context.Background(), nil)
	require.NoError(t, err)
	if len(existing) > 0 {
		ids := make([]int, len(existing))
		for i, s := range existing {
			ids[i] = s.ID
		}
		t.Skipf("existing stream(s) on account (IDs: %v); only one stream allowed per account", ids)
	}

	_, err = client.CreateLogStream(context.Background(), linodego.StreamCreateOptions{
		Label:        fmt.Sprintf("go-test-invalid-dest-%d", time.Now().UnixNano()),
		Type:         linodego.StreamTypeAuditLogs,
		Destinations: []int{999999999},
	})
	require.Error(t, err)

	apiErr, ok := err.(*linodego.Error)
	require.True(t, ok, "expected *linodego.Error")
	assert.Equal(t, 400, apiErr.Code)
}

func TestLogStream_List(t *testing.T) {
	requireACLPLogsStreamTests(t)

	client, teardown := createTestClient(t, "fixtures/TestLogStream_List")
	defer teardown()
	requireACLPLogsCapability(t, client)

	objKey, objBucket, objTeardown := setupObjectStorageResources(t, client)
	defer objTeardown()
	dest, destTeardown, err := setupLogDestination(t, client, objKey, objBucket)
	defer destTeardown()
	require.NoError(t, err)

	stream, streamTeardown, err := setupLogStream(t, client, []int{dest.ID})
	defer streamTeardown()
	require.NoError(t, err)

	provisioned, err := waitForLogStreamProvisioned(context.Background(), client, stream.ID, 60, 3600)
	require.NoError(t, err)

	streams, err := client.ListLogStreams(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, streams)

	found := false
	for _, s := range streams {
		if s.ID == provisioned.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created stream not found in list")
}

func TestLogStream_Get(t *testing.T) {
	requireACLPLogsStreamTests(t)

	client, teardown := createTestClient(t, "fixtures/TestLogStream_Get")
	defer teardown()
	requireACLPLogsCapability(t, client)

	objKey, objBucket, objTeardown := setupObjectStorageResources(t, client)
	defer objTeardown()
	dest, destTeardown, err := setupLogDestination(t, client, objKey, objBucket)
	defer destTeardown()
	require.NoError(t, err)

	stream, streamTeardown, err := setupLogStream(t, client, []int{dest.ID})
	defer streamTeardown()
	require.NoError(t, err)

	provisioned, err := waitForLogStreamProvisioned(context.Background(), client, stream.ID, 60, 3600)
	require.NoError(t, err)

	fetched, err := client.GetLogStream(context.Background(), provisioned.ID)
	require.NoError(t, err)
	assert.Equal(t, provisioned.ID, fetched.ID)
	assert.Equal(t, provisioned.Label, fetched.Label)
	assert.Equal(t, provisioned.Status, fetched.Status)
	assert.Len(t, fetched.Destinations, 1)
}

func TestLogStream_Update(t *testing.T) {
	requireACLPLogsStreamTests(t)

	client, teardown := createTestClient(t, "fixtures/TestLogStream_Update")
	defer teardown()
	requireACLPLogsCapability(t, client)

	objKey, objBucket, objTeardown := setupObjectStorageResources(t, client)
	defer objTeardown()
	dest, destTeardown, err := setupLogDestination(t, client, objKey, objBucket)
	defer destTeardown()
	require.NoError(t, err)

	stream, streamTeardown, err := setupLogStream(t, client, []int{dest.ID})
	defer streamTeardown()
	require.NoError(t, err)

	provisioned, err := waitForLogStreamProvisioned(context.Background(), client, stream.ID, 60, 3600)
	require.NoError(t, err)

	// --- update label and verify history ---
	originalLabel := provisioned.Label
	versionBefore := provisioned.Version
	newLabel := originalLabel + "-upd"

	updated, err := client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
		Label: &newLabel,
	})
	require.NoError(t, err)
	assert.Equal(t, newLabel, updated.Label)

	defer func() {
		_, _ = client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
			Label: &originalLabel,
		})
	}()

	history, err := client.ListLogStreamHistory(context.Background(), provisioned.ID, nil)
	require.NoError(t, err)

	var snapOriginal, snapUpdated *linodego.Stream
	for i := range history {
		switch history[i].Version {
		case versionBefore:
			snapOriginal = &history[i]
		case updated.Version:
			snapUpdated = &history[i]
		}
	}
	require.NotNil(t, snapOriginal, "original version not found in history")
	require.NotNil(t, snapUpdated, "updated version not found in history")
	assert.Equal(t, originalLabel, snapOriginal.Label)
	assert.Equal(t, newLabel, snapUpdated.Label)

	// --- toggle status ---
	originalStatus := provisioned.Status
	newStatus := linodego.StreamStatusInactive
	if originalStatus == linodego.StreamStatusInactive {
		newStatus = linodego.StreamStatusActive
	}

	updatedStatus, err := client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
		Status: &newStatus,
	})
	require.NoError(t, err)
	assert.Equal(t, newStatus, updatedStatus.Status)

	defer func() {
		_, _ = client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
			Status: &originalStatus,
		})
	}()
}

func TestLogStream_Update_Destinations(t *testing.T) {
	requireACLPLogsStreamTests(t)

	client, teardown := createTestClient(t, "fixtures/TestLogStream_Update_Destinations")
	defer teardown()
	requireACLPLogsCapability(t, client)

	objKey, objBucket, objTeardown := setupObjectStorageResources(t, client)
	defer objTeardown()
	dest, destTeardown, err := setupLogDestination(t, client, objKey, objBucket)
	defer destTeardown()
	require.NoError(t, err)

	objKey2, objBucket2, objTeardown2 := setupObjectStorageResources(t, client)
	defer objTeardown2()
	secondaryDest, secondaryDestTeardown, err := setupLogDestination(t, client, objKey2, objBucket2)
	defer secondaryDestTeardown()
	require.NoError(t, err)

	stream, streamTeardown, err := setupLogStream(t, client, []int{dest.ID})
	defer streamTeardown()
	require.NoError(t, err)

	provisioned, err := waitForLogStreamProvisioned(context.Background(), client, stream.ID, 60, 3600)
	require.NoError(t, err)

	versionBefore := provisioned.Version

	updated, err := client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
		Destinations: []int{secondaryDest.ID},
	})
	require.NoError(t, err)
	require.Len(t, updated.Destinations, 1)
	assert.Equal(t, secondaryDest.ID, updated.Destinations[0].ID)

	defer func() {
		_, _ = client.UpdateLogStream(context.Background(), provisioned.ID, linodego.StreamUpdateOptions{
			Destinations: []int{dest.ID},
		})
	}()

	history, err := client.ListLogStreamHistory(context.Background(), provisioned.ID, nil)
	require.NoError(t, err)

	var snapOriginal, snapUpdated *linodego.Stream
	for i := range history {
		switch history[i].Version {
		case versionBefore:
			snapOriginal = &history[i]
		case updated.Version:
			snapUpdated = &history[i]
		}
	}
	require.NotNil(t, snapOriginal, "original version not found in history")
	require.NotNil(t, snapUpdated, "updated version not found in history")
	assert.Equal(t, dest.ID, snapOriginal.Destinations[0].ID)
	assert.Equal(t, secondaryDest.ID, snapUpdated.Destinations[0].ID)
}
