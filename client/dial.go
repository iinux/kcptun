package main

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/pkg/errors"
	kcp "github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/tcpraw"
	"log"
	"net"
	"time"
	_ "unsafe"
)

func dial(config *Config, block kcp.BlockCrypt) (*kcp.UDPSession, error) {
	if config.TCP {
		conn, err := tcpraw.Dial("tcp", config.RemoteAddr)
		if err != nil {
			return nil, errors.Wrap(err, "tcpraw.Dial()")
		}
		return kcp.NewConn(config.RemoteAddr, block, config.DataShard, config.ParityShard, conn)
	}
	return DialWithOptions(config.RemoteAddr, block, config.DataShard, config.ParityShard)
}

//go:linkname newUDPSession github.com/xtaci/kcp-go/v5.newUDPSession
func newUDPSession(conv uint32, dataShards, parityShards int, l *kcp.Listener, conn net.PacketConn, ownConn bool, remote net.Addr, block kcp.BlockCrypt) *kcp.UDPSession

func DialWithOptions(raddr string, block kcp.BlockCrypt, dataShards, parityShards int) (*kcp.UDPSession, error) {
	// network type detection
	udpaddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	network := "udp4"
	if udpaddr.IP.To4() == nil {
		network = "udp"
	}

	laddr := &net.UDPAddr{Port: 8021}
	conn, err := net.ListenUDP(network, laddr)
	log.Println("new conn", conn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	go func() {
		for true {
			log.Println("heartbeat", conn)
			_, err := conn.WriteTo([]byte(""), udpaddr)
			if err != nil {
				log.Printf("%+v\n", err)
				return
			}
			time.Sleep(20 * time.Second)
		}
	}()

	var convid uint32
	binary.Read(rand.Reader, binary.LittleEndian, &convid)
	return newUDPSession(convid, dataShards, parityShards, nil, conn, true, udpaddr, block), nil
}
