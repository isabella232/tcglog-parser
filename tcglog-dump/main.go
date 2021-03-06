// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3 with static-linking exception.
// See LICENCE file for details.

package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/canonical/tcglog-parser"
	"github.com/canonical/tcglog-parser/internal"
)

var (
	alg                  string
	verbose              bool
	hexDump              bool
	varDataHexDump       bool
	extractDataPrefix    string
	extractVarDataPrefix string
	withGrub             bool
	withSdEfiStub        bool
	sdEfiStubPcr         int
	pcrs                 internal.PCRArgList
)

func init() {
	flag.StringVar(&alg, "alg", "sha1", "Name of the hash algorithm to display")
	flag.BoolVar(&verbose, "verbose", false, "Display details of event data")
	flag.BoolVar(&verbose, "v", false, "Display details of event data (shorthand)")
	flag.BoolVar(&hexDump, "hexdump", false, "Display hexdump of event data")
	flag.BoolVar(&hexDump, "x", false, "Display hexdump of event data (shorthand)")
	flag.BoolVar(&varDataHexDump, "vardatahexdump", false, "Display hexdump of EFI variable data")
	flag.StringVar(&extractDataPrefix, "extract-data", "", "Extract event data to individual files named with the specified prefix (format: <prefix>-<pcr>-<num>")
	flag.StringVar(&extractVarDataPrefix, "extract-vardata", "", "Extract EFI variable data to individual files named with the specified prefix (format: <prefix>-<pcr>-<num>")
	flag.BoolVar(&withGrub, "with-grub", false, "Interpret measurements made by GRUB to PCR's 8 and 9")
	flag.BoolVar(&withSdEfiStub, "with-systemd-efi-stub", false, "Interpret measurements made by systemd's EFI stub Linux loader")
	flag.IntVar(&sdEfiStubPcr, "systemd-efi-stub-pcr", 8, "Specify the PCR that systemd's EFI stub Linux loader measures to")
	flag.Var(&pcrs, "pcrs", "Display events associated with the specified PCRs. Can be specified multiple times")
}

func shouldDisplayEvent(event *tcglog.Event) bool {
	if len(pcrs) == 0 {
		return true
	}
	return pcrs.Contains(event.PCRIndex)
}

func main() {
	flag.Parse()

	algorithmId, err := internal.ParseAlgorithm(alg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Too many arguments\n")
		os.Exit(1)
	}

	var path string
	if len(args) == 1 {
		path = args[0]
	} else {
		path = "/sys/kernel/security/tpm0/binary_bios_measurements"
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	log, err := tcglog.ParseLog(file, &tcglog.LogOptions{EnableGrub: withGrub, EnableSystemdEFIStub: withSdEfiStub, SystemdEFIStubPCR: tcglog.PCRIndex(sdEfiStubPcr)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse log file: %v\n", err)
		os.Exit(1)
	}

	if !log.Algorithms.Contains(algorithmId) {
		fmt.Fprintf(os.Stderr,
			"The log doesn't contain entries for the %s digest algorithm\n", algorithmId)
		os.Exit(1)
	}

	for _, event := range log.Events {
		if !shouldDisplayEvent(event) {
			continue
		}

		var builder bytes.Buffer
		fmt.Fprintf(&builder, "%2d %x %s", event.PCRIndex, event.Digests[algorithmId], event.EventType)
		if verbose || hexDump {
			data := event.Data.String()
			if data != "" {
				fmt.Fprintf(&builder, " [ %s ]", data)
			}
		}

		if hexDump {
			fmt.Fprintf(&builder, "\n  Event data:\n  %s", strings.Replace(hex.Dump(event.Data.Bytes()), "\n", "\n  ", -1))
		}

		if varDataHexDump {
			varData, ok := event.Data.(*tcglog.EFIVariableData)
			if ok {
				fmt.Fprintf(&builder, "\n  EFI variable data:\n  %s", strings.Replace(hex.Dump(varData.VariableData), "\n", "\n  ", -1))
			}
		}

		fmt.Println(builder.String())

		if extractDataPrefix != "" {
			ioutil.WriteFile(fmt.Sprintf("%s-%d-%d", extractDataPrefix, event.PCRIndex, event.Index), event.Data.Bytes(), 0644)
		}

		if extractVarDataPrefix != "" {
			varData, ok := event.Data.(*tcglog.EFIVariableData)
			if ok {
				ioutil.WriteFile(fmt.Sprintf("%s-%d-%d", extractVarDataPrefix, event.PCRIndex, event.Index), varData.VariableData, 0644)
			}
		}
	}
}
