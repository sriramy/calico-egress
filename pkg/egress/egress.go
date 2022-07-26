package egress

import "github.com/vishvananda/netlink"

type Network struct{}

func (n *Network) EnsureDummyDevice(deviceName string) (netlink.Link, error) {
	link, err := netlink.LinkByName(deviceName)
	if err == nil {
		return link, nil
	}
	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{Name: deviceName},
	}
	if err = netlink.LinkAdd(dummy); err != nil {
		return nil, err
	}
	return dummy, nil
}
