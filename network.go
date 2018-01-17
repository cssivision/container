package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

var (
	bridgeName   = "container0"
	bridgeIP     = "10.88.37.1/24"
	vethName     = "veth"
	vethPeerName = "veth-peer"
	vethAddr     = "10.88.37.11/24"
	vethPeerAddr = "10.88.37.22/24"
)

func createBridge() (netlink.Link, error) {
	if br, err := netlink.LinkByName(bridgeName); err == nil {
		return br, nil
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return nil, fmt.Errorf("bridge creation: %v", err)
	}

	addr, err := netlink.ParseAddr(bridgeIP)
	if err != nil {
		return nil, fmt.Errorf("parse address %s: %v", bridgeIP, err)
	}

	if err := netlink.AddrAdd(br, addr); err != nil {
		return nil, fmt.Errorf("br add addr err: %v", err)
	}

	// sets up bridge ( ip link set dev container0 up )
	if err := netlink.LinkSetUp(br); err != nil {
		return nil, err
	}
	return br, nil
}

type vethPair struct {
	Veth         netlink.Link
	VethAddr     string
	VethName     string
	VethPeer     netlink.Link
	VethPeerAddr string
	VethPeerName string
}

func createVethPair(pid int) (netlink.Link, error) {
	// get bridge to set as master for one side of veth-pair
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, fmt.Errorf("find bridge err: %v", err)
	}

	// create *netlink.Veth
	la := netlink.NewLinkAttrs()
	la.Name = vethName
	la.MasterIndex = br.Attrs().Index

	vp := &netlink.Veth{LinkAttrs: la, PeerName: vethPeerName}
	if err := netlink.LinkAdd(vp); err != nil {
		return nil, fmt.Errorf("veth pair creation %s <-> %s: %v", vethName, vethPeerName, err)
	}

	// get peer by name to put it to namespace
	peer, err := netlink.LinkByName(vethPeerName)
	if err != nil {
		return nil, fmt.Errorf("get peer interface: %v", err)
	}

	// put peer side to network namespace of specified PID
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		return nil, fmt.Errorf("move peer to ns of %d: %v", pid, err)
	}

	addr, err := netlink.ParseAddr(vethAddr)
	if err != nil {
		return nil, fmt.Errorf("veth addr parse IP: %v", err)
	}

	if err := netlink.AddrAdd(vp, addr); err != nil {
		return nil, fmt.Errorf("veth addr add err: %v", err)
	}

	if err := netlink.LinkSetUp(vp); err != nil {
		return nil, fmt.Errorf("veth set up err: %v", err)
	}

	return vp, nil
}

func putIface(pid int) error {
	br, err := createBridge()
	if err != nil {
		return fmt.Errorf("create bridge err: %v", err)
	}
	veth, err := createVethPair(pid)
	if err != nil {
		return fmt.Errorf("create veth pair err: %v", err)
	}

	if err := netlink.LinkSetMaster(veth, br.(*netlink.Bridge)); err != nil {
		return fmt.Errorf("link set master err: %v", err)
	}

	if err := setIptables(); err != nil {
		return fmt.Errorf("set iptables err: %v", err)
	}

	return nil
}

func setIptables() error {
	return nil
}

func setupIface(link netlink.Link) error {
	// up loopback
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("lo interface: %v", err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("up veth: %v", err)
	}
	addr, err := netlink.ParseAddr(vethPeerAddr)
	if err != nil {
		return fmt.Errorf("parse IP: %v", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("addr add err: %v", err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("link set up err: %v", err)
	}

	vethIP := net.ParseIP(vethAddr)
	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: link.Attrs().Index,
		Gw:        vethIP,
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("route add err: %v", err)
	}
	return nil
}

func waitForIface() (netlink.Link, error) {
	log.Println("Starting to wait for network interface")
	start := time.Now()
	for {
		fmt.Printf(".")
		if time.Since(start) > 5*time.Second {
			fmt.Printf("\n")
			return nil, fmt.Errorf("failed to find veth interface in 5 seconds")
		}
		// get list of all interfaces
		lst, err := netlink.LinkList()
		if err != nil {
			fmt.Printf("\n")
			return nil, err
		}
		for _, l := range lst {
			// if we found "veth" interface - it's time to continue setup
			if l.Type() == "veth" {
				fmt.Printf("\n")
				return l, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
