package gcedns

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
	"golang.org/x/sync/errgroup"
)

// VMInfo represents a single Compute Engine VM's identifying information.
type VMInfo struct {
	Name         string
	ExternalIPv4 string
	ExternalIPv6 string
	ProjectID    string
	Zone         string
}

func Example() error {
	ctx := context.Background()
	c, err := compute.NewAcceleratorTypesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	return nil
}

func GetHostVMInfo(ctx context.Context, name string) (result VMInfo, err error) {
	if !metadata.OnGCE() {
		return VMInfo{}, fmt.Errorf("not running on GCE")
	}

	client := metadata.NewClient(nil)
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() (err error) {
		result.Name, err = client.InstanceName()
		return
	})
	group.Go(func() (err error) {
		result.ExternalIPv4, err = client.ExternalIP()
		return
	})
	group.Go(func() (err error) {
		result.ProjectID, err = client.ProjectID()
		return
	})
	group.Go(func() (err error) {
		result.Zone, err = client.Zone()
		return
	})

	err = group.Wait()
	if err != nil {
		return VMInfo{}, err
	}
	return result, nil
}
