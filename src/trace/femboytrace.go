package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// -target ip address you want to point at
// -maxTTL set amount of TTL
// -icmpHop number of ICMP packets each hop
// -times Time it takes to respond

func icmpPacket(id, seq int) []byte {
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: []byte("GO-Route"),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		log.Fatal("Failed to marshal ICMP message:", err)
	}
	return b
}

func waitResponses(c *icmp.PacketConn, timeout time.Duration, _ time.Time, targetHost string, maxTTL, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	timer := time.NewTimer(timeout)
	results := make(map[int]string)

	go func() {
		for {
			reply := make([]byte, 1500)
			n, peer, err := c.ReadFrom(reply)
			if err != nil {
				break
			}

			rm, err := icmp.ParseMessage(1, reply[:n])
			if err != nil {
				continue
			}

			switch rm.Type {
			case ipv4.ICMPTypeEchoReply:
				if echo, ok := rm.Body.(*icmp.Echo); ok && peer.String() == targetHost {
					results[echo.Seq] = peer.String()
				}

			case ipv4.ICMPTypeTimeExceeded:

				if exceeded, ok := rm.Body.(*icmp.TimeExceeded); ok {
					if len(exceeded.Data) >= 28 {
						seq := int(exceeded.Data[26])<<8 | int(exceeded.Data[27])
						results[seq] = peer.String()
					}
				}

			}
		}
	}()

	<-timer.C
	for i := 1; i <= maxTTL; i++ {
		if ip, exists := results[i]; exists {
			fmt.Printf("%d\t%s\n", i, ip)
			if ip == targetHost {
				break
			}
		} else {
			fmt.Printf("%d\t*\n", i)
		}
	}
}

func instruction() {
	fmt.Println("Welcome to FemboyTace!\n", "(✿^‿^)")
	fmt.Println("Here's a basic tutorial!\n", "./femboytrace -target=`8.8.8.8` -maxTTL=20 -icmpHop=2 -times=2")
	fmt.Println("Compiler flags!\n", "-target = ip address you want to point at\n", "-highTTL set amount of TTL\n", "-icmpHop number of ICMP packets each hop\n", "-times Time it takes to respond")

}

func main() {

	instruction()

	target := flag.String("target", "1.1.1.1", "Target host for traceroute")
	highTTL := flag.Int("highTTL", 30, "Maximum number of hops")
	icmpHop := flag.Int("icmpHop", 3, "Number of probes per hop")
	times := flag.Int("times", 3, "Response timeout in seconds")

	flag.Parse()

	c, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		log.Fatal("Make sure you have Sudo privilages!")
	}
	defer c.Close()

	id := os.Getpid() & 0xffff
	startTime := time.Now()
	timeoutDuration := time.Duration(*times) * time.Second

	remoteHost, err := net.ResolveIPAddr("ip4", *target)
	if err != nil {
		log.Fatal("Failed to resolve target host:", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go waitResponses(c, timeoutDuration, startTime, remoteHost.IP.String(), *highTTL, id, &wg)

	for ttl := 1; ttl <= *highTTL; ttl++ {
		if err := c.IPv4PacketConn().SetTTL(ttl); err != nil {
			log.Printf("Failed to set TTL (%d): %v", ttl, err)
			continue
		}

		for i := 0; i < *icmpHop; i++ {
			b := icmpPacket(id, ttl)
			if _, err := c.WriteTo(b, remoteHost); err != nil {
				log.Println("Failed to send ICMP packet:", err)
			}
		}
	}

	wg.Wait()
}
