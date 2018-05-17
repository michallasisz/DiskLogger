package main

//This is first experiment in Go after Hello World:)
//usefull only for me
import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"syscall"
)

const appname = "DiskLogger"

var onlystdout string

func main() {
	textPtr := flag.String("print-only", "NO", "YES/NO")
	flag.Parse()
	if *textPtr == "YES" {
		onlystdout = "YES"
	} else {
		onlystdout = "NO"
	}

	file, err := os.Open("/root/DiskLogger.mountpoints")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		CheckWriteLog(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

//DiskUsagePerc -this function return current mountpoint free space[%] or 99999999  in case of error
func DiskUsagePerc(path string) uint64 {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return 99999999
	}
	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	freeperc := (free * 100) / total
	return freeperc

}

//CheckWriteLog -this function writes messages to stdout or/and syslog
func CheckWriteLog(mountpoint string) {
	freeperc := DiskUsagePerc(mountpoint)
	message, status := "", ""
	l3, err := syslog.New(syslog.LOG_LOCAL7, appname)
	// l3, err := syslog.Dial("udp", "syslog_ip:514", sev, appname) // connection to a log daemon
	defer l3.Close()
	if err != nil {
		log.Fatal(err)
	}
	if freeperc == 99999999 {
		status = "ERROR:"
		message = fmt.Sprintf("%s Error reading freeperc of: %s mountpoint. Check config or system...", status, mountpoint)
	}
	//Critical threshold is 10 %
	if freeperc < 11 {
		status = "ALERT:"
		message = fmt.Sprintf("%s There is %d %% free space on: %s", status, freeperc, mountpoint)
	}
	//Critical threshold is 30 %
	if freeperc > 10 && freeperc < 31 {
		status = "WARNING:"
		message = fmt.Sprintf("%s There is %d %% free space on: %s", status, freeperc, mountpoint)
	}
	//Everything else is normal
	if freeperc > 30 && freeperc < 101 {
		status = "NORMAL:"
		message = fmt.Sprintf("%s There is %d %% free space on: %s", status, freeperc, mountpoint)
	}
	if onlystdout == "YES" {
		fmt.Println(message)
	} else {
		switch status {
		case "ERROR:":
			l3.Alert(message)
		case "ALERT:":
			l3.Alert(message)
		case "WARNING:":
			l3.Warning(message)
		case "NORMAL:":
			l3.Notice(message)
		}
	}
}
