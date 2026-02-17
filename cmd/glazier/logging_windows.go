//go:build windows

package main

import (
	"fmt"
	"os"

	"github.com/google/deck"
	"github.com/google/deck/backends/eventlog"
	"golang.org/x/sys/windows/registry"
)

// isWinPE checks for the MiniNT registry key which indicates WinPE.
func isWinPE() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `System\CurrentControlSet\Control\MiniNT`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

func init() {
	if isWinPE() {
		fmt.Println("WinPE detected — skipping eventlog backend")
		return
	}

	evt, err := eventlog.Init("Glazier")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to initialize eventlog: %v\n", err)
		os.Exit(1)
	}
	deck.Add(evt)
}
