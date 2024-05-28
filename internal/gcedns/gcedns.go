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
	ExternalIPv6 netip.Addr
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
		if !addr.Is4() {
			return fmt.Errorf("expected IPv4 address, got %s", addr)
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
func getHostIPv6Addr() (netip.Addr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return netip.IPv6Unspecified(), err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagRunning == 0 {
			continue
		}

		ips, err := iface.Addrs()
		if err != nil {
			return netip.IPv6Unspecified(), err
		}

		for _, ip := range ips {
			ipnet, ok := ip.(*net.IPNet)
			if !ok {
				return netip.IPv6Unspecified(),
					fmt.Errorf("cannot parse net.Addr as IPNet: %v", ip)
			}
			addr, ok := netip.AddrFromSlice(ipnet.IP)
			if !ok {
				return netip.IPv6Unspecified(),
					fmt.Errorf("cannot convert slice to Addr: %v", ipnet.IP)
			}
			if !addr.Is6() || addr.Is4In6() {
				continue
			}
			if !addr.IsGlobalUnicast() {
				continue
			}
			return addr, nil
		}
	}

	return netip.IPv6Unspecified(), errors.New("no IPv6 addresses found")
}
