// 运行: go run ./examples/08_status_and_reset -channel=USBBus1
// 前置: Windows + PCAN-USB
//
// 演示周期检查 Status，发现 BUSOFF / BUSHEAVY 等异常时调用 Reset 恢复。
// 实际工程里 Reset 之后还要重发握手帧、重设过滤器等业务逻辑。

package main

import (
	"context"
	"errors"
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

	bus, err := gocan.Open(ch, gocan.WithBitrate(gocan.Baud500K))
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer bus.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			s, err := bus.Status()
			if err != nil {
				log.Printf("status err: %v", err)
				continue
			}
			log.Printf("status=0x%X (BUSOFF=%v BUSHEAVY=%v BUSPASSIVE=%v)",
				uint32(s),
				gocan.StatusHas(s, gocan.StatusBusOff),
				gocan.StatusHas(s, gocan.StatusBusHeavy),
				gocan.StatusHas(s, gocan.StatusBusPassive),
			)
			if gocan.StatusHas(s, gocan.StatusBusOff) {
				log.Println("BUSOFF detected, calling Reset...")
				if err := bus.Reset(); err != nil {
					// 已关闭场景下也会落到这里
					if errors.Is(err, gocan.ErrBusClosed) {
						return
					}
					log.Printf("reset err: %v", err)
				}
			}
		}
	}
}

func lookupChannel(name string) (gocan.Channel, bool) {
	switch name {
	case "USBBus1":
		return gocan.USBBus1, true
	case "USBBus2":
		return gocan.USBBus2, true
	}
	return 0, false
}
