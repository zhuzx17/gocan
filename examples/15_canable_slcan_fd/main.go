// 示例 15：Windows 上通过 CANable 2.0 SLCAN-FD 固件发送一帧 CAN FD。
package main

import (
	"context"
	"flag"
	"log"

	"github.com/zhuzx17/gocan"
)

func main() {
	port := flag.String("port", "COM5", "CANable 2.0 serial port")
	flag.Parse()

	bus, err := gocan.OpenSLCANFD(
		*port,
		gocan.SLCANBitrate500K,
		gocan.SLCANDataBitrate2M,
	)
	if err != nil {
		log.Fatalf("open SLCAN-FD: %v", err)
	}
	defer bus.Close()

	frame, err := gocan.NewFDFrame(
		0x456,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		true,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := bus.Send(context.Background(), frame); err != nil {
		log.Fatalf("send: %v", err)
	}
	log.Printf("sent CAN FD frame on %s: id=0x%X data=%X", *port, frame.ID, frame.Data)
}
