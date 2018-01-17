package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

var (
	bridgeName = "container0"
	vethPrefix = "veth-pair"
	ipAddr     = "10.88.37.1/24"
	ipTmpl     = "10.88.37.%d/24"
	globalVpi  = new(vethPair)
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

type vethPair struct {
	Veth         netlink.Link
	VethAddr     string
	VethName     string
	VethPeer     netlink.Link
	VethPeerAddr string
	VethPeerName string
}

func createVethPair(pid int) error {
	// get bridge to set as master for one side of veth-pair
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// generate names for interfaces
	x1, x2 := rand.Intn(10000), rand.Intn(10000)
	globalVpi.VethName = fmt.Sprintf("%s%d", vethPrefix, x1)
	globalVpi.VethPeerName = fmt.Sprintf("%s%d", vethPrefix, x2)
	globalVpi.VethAddr = fmt.Sprintf(ipTmpl, rand.Intn(253)+2)
	globalVpi.VethPeerAddr = fmt.Sprintf(ipTmpl, rand.Intn(253)+2)

	// create *netlink.Veth
	la := netlink.NewLinkAttrs()
	la.Name = globalVpi.VethName
	la.MasterIndex = br.Attrs().Index

	vp := &netlink.Veth{LinkAttrs: la, PeerName: globalVpi.VethPeerName}
	if err := netlink.LinkAdd(vp); err != nil {
		return fmt.Errorf("veth pair creation %s <-> %s: %v", globalVpi.VethName, globalVpi.VethPeerName, err)
	}
	globalVpi.Veth = vp

	// get peer by name to put it to namespace
	peer, err := netlink.LinkByName(globalVpi.VethPeerName)
	if err != nil {
		return fmt.Errorf("get peer interface: %v", err)
	}
	globalVpi.VethPeer = peer

	// put peer side to network namespace of specified PID
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		return fmt.Errorf("move peer to ns of %d: %v", pid, err)
	}

	addr, err := netlink.ParseAddr(globalVpi.VethAddr)
	if err != nil {
		return fmt.Errorf("veth addr parse IP: %v", err)
	}

	if err := netlink.AddrAdd(vp, addr); err != nil {
		return fmt.Errorf("veth addr add err: %v", err)
	}

	if err := netlink.LinkSetUp(vp); err != nil {
		return fmt.Errorf("veth set up err: %v", err)
	}

	return nil
}

func putIface(pid int) error {
	br, err := createBridge()
	if err != nil {
		return fmt.Errorf("create bridge err: %v", err)
	}
	if err := createVethPair(pid); err != nil {
		return fmt.Errorf("create veth pair err: %v", err)
	}

	if err := netlink.LinkSetMaster(globalVpi.Veth, br.(*netlink.Bridge)); err != nil {
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
	addr, err := netlink.ParseAddr(globalVpi.VethPeerAddr)
	if err != nil {
		return fmt.Errorf("parse IP: %v", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("addr add err: %v", err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("link set up err: %v", err)
	}

	vethIP := net.ParseIP(globalVpi.VethAddr)
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
