package gcedns

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
)

func Example() error {
	ctx := context.Background()
	c, err := compute.NewAcceleratorTypesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	return nil
}
