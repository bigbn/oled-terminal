package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	memory "github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
	serial "go.bug.st/serial"
)

const handshake = "handshake"
const scale = 12

func main() {
	cpu.Percent(100, true)

	ports, err := serial.GetPortsList()
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		log.Printf("Checking port: %v\n", port)
		openedPort, err := serial.Open(port, mode)
		if err != nil {
			log.Fatal(err)
		}

		openedPort.Write([]byte(handshake))
		chunk := make([]byte, 1)
		buffer := ""
		for i := 0; i < len(handshake); i++ {
			_, err := openedPort.Read(chunk)
			if err != nil {
				log.Fatal(err)
				break
			}
			buffer = buffer + string(chunk)
		}

		if buffer == handshake {
			log.Printf("Handshake success with %v\n", port)
			portLoop(openedPort)
		}
	}
}

func portLoop(port serial.Port) {
	gbDivider := uint64(math.Pow(1024, 3))
	lastUsed := 0
	lastCpuUsed := 0

	for {
		total := float64(memory.TotalMemory()) / float64(gbDivider)
		free := float64(memory.FreeMemory()) / float64(gbDivider)
		used := float64(total - free)

		scaledUsed := int(scale / total * used)

		cpuUsage, _ := cpu.Percent(0, false)
		zeroCore := cpuUsage[0]

		scaledCpuUsed := int(zeroCore / 100 * scale)

		if scaledUsed == lastUsed && scaledCpuUsed == lastCpuUsed {
			time.Sleep(time.Second)
			continue
		}

		scaledFree := int(scale / total * free)
		scaledCpuUnused := int((100 - zeroCore) / 100 * scale)

		memory := fmt.Sprintf("%s%s|%s%s", strings.Repeat("~", scaledUsed), strings.Repeat("-", scaledFree), strings.Repeat("~", scaledCpuUsed), strings.Repeat("-", scaledCpuUnused))
		lastUsed = scaledUsed
		lastCpuUsed = scaledCpuUsed

		port.Write([]byte(memory))
		time.Sleep(time.Millisecond * 50)
	}

}
