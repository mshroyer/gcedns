package gcedns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
	"golang.org/x/sync/errgroup"
)

// VMInfo represents a single Compute Engine VM's identifying information.
type VMInfo struct {
	Name         string
	ExternalIPv4 netip.Addr
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

// GetHostVMInfo returns information about the host VM.
func GetHostVMInfo(ctx context.Context, name string) (result VMInfo, err error) {
	if !metadata.OnGCE() {
		return VMInfo{}, errors.New("not running on GCE")
	}

	// Using the default client
	client := metadata.NewClient(nil)

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() (err error) {
		result.Name, err = client.InstanceName()
		return
	})
	group.Go(func() error {
		ipstr, err := client.ExternalIP()
		if err != nil {
			return err
		}

		addr, err := netip.ParseAddr(ipstr)
		if err != nil {
			return err
		}
		result.ExternalIPv4 = addr
		return nil
	})
	group.Go(func() (err error) {
		result.ExternalIPv6, err = getHostIPv6Addr()
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

// getHostIPv6Addr returns the public, non-temporary IPv6 address of the host.
//
// Returns the empty string if the host has no such IPv6 addresses.  If the
// host has multiple eligible IPv6 addresses, one of them will be returned
// arbitrarily.
func getHostIPv6Addr() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagRunning == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			fmt.Printf("Found address: %s\n", addr)
		}
	}

	return "", nil
}
