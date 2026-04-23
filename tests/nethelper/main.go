package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type iface struct {
	Name string `json:"name"`
	Up   bool   `json:"up"`
}

type probe struct {
	Interfaces        []iface `json:"interfaces"`
	LoopbackReachable bool    `json:"loopbackReachable"`
}

func main() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list interfaces: %v\n", err)
		os.Exit(1)
	}

	result := probe{
		Interfaces:        make([]iface, 0, len(ifaces)),
		LoopbackReachable: loopbackReachable(),
	}
	for _, netIface := range ifaces {
		result.Interfaces = append(result.Interfaces, iface{
			Name: netIface.Name,
			Up:   netIface.Flags&net.FlagUp != 0,
		})
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "encode probe: %v\n", err)
		os.Exit(1)
	}
}

func loopbackReachable() bool {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return false
	}
	defer ln.Close()

	accepted := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			accepted <- err
			return
		}
		_ = conn.Close()
		accepted <- nil
	}()

	conn, err := net.DialTimeout("tcp4", ln.Addr().String(), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()

	select {
	case err := <-accepted:
		return err == nil
	case <-time.After(time.Second):
		return false
	}
}
