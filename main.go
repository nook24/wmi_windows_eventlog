package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/StackExchange/wmi"
)

// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/eventlogprov/win32-ntlogevent
type Win32_NTLogEvent struct {
	Category         uint16
	CategoryString   string
	ComputerName     string
	Data             []uint8
	EventCode        uint16
	EventIdentifier  uint32
	EventType        uint8
	InsertionStrings []string
	Logfile          string
	Message          string
	RecordNumber     uint32
	SourceName       string
	TimeGenerated    time.Time
	TimeWritten      time.Time
	Type             string
	User             string
}

func main() {
	// This initialization prevents a memory leak on WMF 5+. See
	// https://github.com/martinlindhe/wmi_exporter/issues/77 and linked issues
	// for details.
	// https://github.com/StackExchange/wmi/issues/27#issuecomment-309578576
	// https://github.com/bosun-monitor/bosun/pull/2028/files#diff-a6d7a21df96534b54447f5b1a8936f35e642cacad4a41f911e37a12dc2852e20R17-R23
	//fmt.Println("Initializing SWbemServices")
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		return
	}
	wmi.DefaultClient.SWbemServicesClient = s

	for {
		checkLog()
		PrintMemUsage()
		time.Sleep(250 * time.Millisecond)
	}

}

func checkLog() (map[string][]*Win32_NTLogEvent, error) {

	logfiles := []string{
		"Application",
		"System",
	}

	now := time.Now().UTC()
	now = now.Add((3600 * time.Second) * -1)

	wmidate := now.Format("20060102150405.000000-070")

	var dst []Win32_NTLogEvent
	var sql string
	//var eventBuffer map[string][]*Win32_NTLogEvent
	eventBuffer := make(map[string][]*Win32_NTLogEvent)
	for _, logfile := range logfiles {
		sql = fmt.Sprintf("SELECT * FROM Win32_NTLogEvent WHERE Logfile='%v' AND TimeWritten >= '%v'", logfile, wmidate)
		//fmt.Println(sql)

		err := wmi.Query(sql, &dst)
		if err != nil {
			fmt.Println("Event Log error: ", err)
			continue
		}

		for _, event := range dst {
			// Resolve Memory Leak
			eventBuffer[logfile] = append(eventBuffer[logfile], &Win32_NTLogEvent{
				Category:         event.Category,
				CategoryString:   event.CategoryString,
				ComputerName:     event.ComputerName,
				Data:             event.Data,
				EventCode:        event.EventCode,
				EventIdentifier:  event.EventIdentifier,
				EventType:        event.EventType,
				InsertionStrings: event.InsertionStrings,
				Logfile:          event.Logfile,
				Message:          event.Message,
				RecordNumber:     event.RecordNumber,
				SourceName:       event.SourceName,
				TimeGenerated:    event.TimeGenerated,
				TimeWritten:      event.TimeWritten,
				Type:             event.Type,
				User:             event.User,
			})
		}
	}

	return eventBuffer, nil

}

// Credit to: https://golangcode.com/print-the-current-memory-usage/
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
