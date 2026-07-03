// 运行: go run ./examples/02_receive_polling -channel=USBBus1
// 前置: Windows + PCAN-USB + 总线上有其他节点在发帧
//
// 演示用 Polling 模式接收 Classical CAN 帧。
// 适合对 CPU 占用敏感、可接受 1ms 级抖动的场景。

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/zhuzx17/gocan"
)

func main() {
	chName := flag.String("channel", "USBBus1", "channel name")
	flag.Parse()

	ch, ok := lookupChannel(*chName)
	if !ok {
		log.Fatalf("unknown channel: %s", *chName)
	}

	bus, err := gocan.Open(ch,
		gocan.WithBitrate(gocan.Baud500K),
		gocan.WithReceiveMode(gocan.ModePolling),
		gocan.WithPollInterval(time.Millisecond),
	)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer bus.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("polling mode receive started, Ctrl-C to stop")
	for {
		fr, err := bus.ReadOne(ctx)
		if err != nil {
			log.Printf("stop: %v", err)
			return
		}
		log.Printf("rx id=0x%X ext=%v rtr=%v data=%X ts=%dµs",
			fr.ID,
			fr.Has(gocan.FlagExtended),
			fr.Has(gocan.FlagRemote),
			fr.Data, fr.TimestampMicros)
	}
}

func lookupChannel(name string) (gocan.Channel, bool) {
	switch name {
	case "USBBus1":
		return gocan.USBBus1, true
	case "USBBus2":
		return gocan.USBBus2, true
	case "USBBus3":
		return gocan.USBBus3, true
	case "USBBus4":
		return gocan.USBBus4, true
	}
	return 0, false
}
