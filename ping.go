package main

import(
	"fmt"
	"os"
	"log"
	"time"
	"net"
	"flag"
	"golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
)



func receivePing(c *icmp.PacketConn,recv_len *chan int, recv_message *chan [] byte) {
	

	reply := make([]byte, 1500)
	for{
	    n, peer, err := c.ReadFrom(reply)
	   	if err != nil {
	   		log.Printf("Peer %s\n",peer)
	       	fmt.Printf("Reply error")
	       	return
	   	}

	    *recv_len<-n
	    *recv_message<-reply
	}
}

func main(){
	flag.Parse()
	host:=flag.Arg(0)
	
	
	totalPacket:=0
	totalPacketLost:=0
	packetLoss:=0
	icmp_seq:=1
	 c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        return
    }
    defer c.Close()

    // Resolve any DNS (if used) and get the real IP of the target
    dst, err := net.ResolveIPAddr("ip4", host)
    if err != nil {
        return 
    }
    message_len:=make(chan int)
    message:=make(chan []byte)
    timer:=time.NewTimer(4* time.Second)
    go receivePing(c,&message_len,&message)

    for{
    
	m := icmp.Message{
        Type: ipv4.ICMPTypeEcho, Code: 0,
        Body: &icmp.Echo{
            ID: os.Getpid() & 0xffff, Seq: icmp_seq, //<< uint(seq), // TODO
            Data: []byte(""),
        },
    }

    b, err := m.Marshal(nil)
    if err != nil {
        return
    }

    start := time.Now()
    n, err := c.WriteTo(b, dst)
    if err != nil {
        return 
    }else if n != len(b) {
        fmt.Printf("Length problem")
        return 
    }
    
    
    
    	select{
    	case x:=<-message_len:
    		reply:=<-message
    		rtt := time.Since(start).Milliseconds()
    		rm,	_ := icmp.ParseMessage(1, reply[:x])
    		totalPacket=+1
    		
    		packetLoss=(totalPacketLost*100)/totalPacket
    		switch rm.Type {
    			case ipv4.ICMPTypeEchoReply:
        		fmt.Printf("Ping %s, RTT: %d ms, icmp_seq:%d, Loss: %d \n",dst.String(),rtt,icmp_seq,packetLoss) 
    		default:
        		fmt.Printf("got %+v from %v; want echo reply", rm)
        	}
        	icmp_seq++
        	
        	timer.Stop()
        	timer=time.NewTimer(4*time.Second)
    	case t:=<-timer.C:
    		fmt.Printf("%d \n",t)
    		totalPacket++
    		totalPacketLost++
    	}
   
        
    time.Sleep(1*time.Second);
   }
}