package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	bridgeName = "container0"
	vethPrefix = "veth-pair"
	ipAddr     = "10.88.37.1/24"
	ipTmpl     = "10.88.37.%d/24"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		return nil, fmt.Errorf("parse address %s: %v", ipAddr, err)
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

func createVethPair(pid int) (netlink.Link, netlink.Link, error) {
	// get bridge to set as master for one side of veth-pair
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, nil, err
	}
	// generate names for interfaces
	x1, x2 := rand.Intn(10000), rand.Intn(10000)
	parentName := fmt.Sprintf("%s%d", vethPrefix, x1)
	peerName := fmt.Sprintf("%s%d", vethPrefix, x2)

	// create *netlink.Veth
	la := netlink.NewLinkAttrs()
	la.Name = parentName
	la.MasterIndex = br.Attrs().Index

	vp := &netlink.Veth{LinkAttrs: la, PeerName: peerName}
	if err := netlink.LinkAdd(vp); err != nil {
		return nil, nil, fmt.Errorf("veth pair creation %s <-> %s: %v", parentName, peerName, err)
	}

	// get peer by name to put it to namespace
	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		return nil, nil, fmt.Errorf("get peer interface: %v", err)
	}

	// put peer side to network namespace of specified PID
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		return nil, nil, fmt.Errorf("move peer to ns of %d: %v", pid, err)
	}
	if err := netlink.LinkSetUp(vp); err != nil {
		return nil, nil, err
	}
	parentVeth, err := netlink.LinkByName(parentName)
	if err != nil {
		return nil, nil, fmt.Errorf("get parent veth interface err: %v", err)
	}
	return parentVeth, peer, nil
}

func putIface(pid int) error {
	br, err := createBridge()
	if err != nil {
		return fmt.Errorf("create bridge err: %v", err)
	}
	parentVeth, _, err := createVethPair(pid)
	if err != nil {
		return fmt.Errorf("create veth pair err: %v", err)
	}

	if err := netlink.LinkSetMaster(parentVeth, br.(*netlink.Bridge)); err != nil {
		return fmt.Errorf("link set master err: %v", err)
	}

	return nil
}

func setupIface(link netlink.Link, ip string) error {
	// up loopback
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("lo interface: %v", err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("up veth: %v", err)
	}
	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("parse IP: %v", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("addr add err: %v", err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("link set up err: %v", err)
	}

	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: link.Attrs().Index,
		Gw:        net.ParseIP(ipAddr),
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
