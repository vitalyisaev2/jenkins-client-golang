package jenkins

import (
	"context"
	"io"

	"github.com/vitalyisaev2/jenkins-client-golang/result"
)

// Jenkins is an access point to Jenkins API
type Client interface {
	// RootInfo returns basic information about the node that you've connected to
	RootInfo(ctx context.Context) *result.Root
	// JobCreate creates new job for given name
	// and xml configuration dumped into slice of bytes
	JobCreate(ctx context.Context, name string, config io.Reader) *result.Job
	// JobGet requests common job information for a given job name
	JobGet(ctx context.Context, name string, depth int) *result.Job
	// JobDelete deletes the requested job
	JobDelete(ctx context.Context, name string) error
	// JobExists checks wether job with a given name exists or not
	JobExists(ctx context.Context, name string) *result.Bool
	// JobInQueue checks whether job with a given name is in queue at the moment
	JobInQueue(ctx context.Context, name string) *result.Bool
	// JobIsBuilding checks whether job with a given name is building at the moment
	JobIsBuilding(ctx context.Context, name string) *result.Bool
	// BuildInvoke invokes simple (non-paramethrized) build of a given job
	BuildInvoke(ctx context.Context, name string) *result.BuildInvoked
	// BuildGetByNumber returns information about particular jenkins build
	BuildGetByNumber(ctx context.Context, name string, id int) *result.Build
	// BuildGetByNumber returns information about particular jenkins build by given queue id
	BuildGetByQueueID(ctx context.Context, name string, id int) *result.Build
}
