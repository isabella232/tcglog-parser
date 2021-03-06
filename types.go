// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3 with static-linking exception.
// See LICENCE file for details.

package tcglog

import (
	"crypto"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"fmt"
)

// Spec corresponds to the TCG specification that an event log conforms to.
type Spec uint

// PCRIndex corresponds to the index of a PCR on the TPM.
type PCRIndex uint32

// EventType corresponds to the type of an event in an event log.
type EventType uint32

// AlgorithmId corresponds to the algorithm of digests that appear in the log. The values are in sync with those
// in the TPM Library Specification for the TPM_ALG_ID type.
// See https://trustedcomputinggroup.org/wp-content/uploads/TPM-Rev-2.0-Part-2-Structures-01.38.pdf (Table 9)
type AlgorithmId uint16

func (a AlgorithmId) GetHash() crypto.Hash {
	switch a {
	case AlgorithmSha1:
		return crypto.SHA1
	case AlgorithmSha256:
		return crypto.SHA256
	case AlgorithmSha384:
		return crypto.SHA384
	case AlgorithmSha512:
		return crypto.SHA512
	default:
		return 0
	}
}

func (a AlgorithmId) supported() bool {
	return a.GetHash() != crypto.Hash(0)
}

func (a AlgorithmId) Size() int {
	return a.GetHash().Size()
}

func (a AlgorithmId) hash(data []byte) []byte {
	h := a.GetHash().New()
	h.Write(data)
	return h.Sum(nil)
}

// Digest is the result of hashing some data.
type Digest []byte

// DigestMap is a map of algorithms to digests.
type DigestMap map[AlgorithmId]Digest

func (e EventType) String() string {
	switch e {
	case EventTypePrebootCert:
		return "EV_PREBOOT_CERT"
	case EventTypePostCode:
		return "EV_POST_CODE"
	case EventTypeNoAction:
		return "EV_NO_ACTION"
	case EventTypeSeparator:
		return "EV_SEPARATOR"
	case EventTypeAction:
		return "EV_ACTION"
	case EventTypeEventTag:
		return "EV_EVENT_TAG"
	case EventTypeSCRTMContents:
		return "EV_S_CRTM_CONTENTS"
	case EventTypeSCRTMVersion:
		return "EV_S_CRTM_VERSION"
	case EventTypeCPUMicrocode:
		return "EV_CPU_MICROCODE"
	case EventTypePlatformConfigFlags:
		return "EV_PLATFORM_CONFIG_FLAGS"
	case EventTypeTableOfDevices:
		return "EV_TABLE_OF_DEVICES"
	case EventTypeCompactHash:
		return "EV_COMPACT_HASH"
	case EventTypeIPL:
		return "EV_IPL"
	case EventTypeIPLPartitionData:
		return "EV_IPL_PARTITION_DATA"
	case EventTypeNonhostCode:
		return "EV_NONHOST_CODE"
	case EventTypeNonhostConfig:
		return "EV_NONHOST_CONFIG"
	case EventTypeNonhostInfo:
		return "EV_NONHOST_INFO"
	case EventTypeOmitBootDeviceEvents:
		return "EV_OMIT_BOOT_DEVICE_EVENTS"
	case EventTypeEFIVariableDriverConfig:
		return "EV_EFI_VARIABLE_DRIVER_CONFIG"
	case EventTypeEFIVariableBoot:
		return "EV_EFI_VARIABLE_BOOT"
	case EventTypeEFIBootServicesApplication:
		return "EV_EFI_BOOT_SERVICES_APPLICATION"
	case EventTypeEFIBootServicesDriver:
		return "EV_EFI_BOOT_SERVICES_DRIVER"
	case EventTypeEFIRuntimeServicesDriver:
		return "EV_EFI_RUNTIME_SERVICES_DRIVER"
	case EventTypeEFIGPTEvent:
		return "EF_EFI_GPT_EVENT"
	case EventTypeEFIAction:
		return "EV_EFI_ACTION"
	case EventTypeEFIPlatformFirmwareBlob:
		return "EV_EFI_PLATFORM_FIRMWARE_BLOB"
	case EventTypeEFIHandoffTables:
		return "EV_EFI_HANDOFF_TABLES"
	case EventTypeEFIHCRTMEvent:
		return "EV_EFI_HCRTM_EVENT"
	case EventTypeEFIVariableAuthority:
		return "EV_EFI_VARIABLE_AUTHORITY"
	default:
		return fmt.Sprintf("%08x", uint32(e))
	}
}

func (e EventType) Format(s fmt.State, f rune) {
	switch f {
	case 's', 'v':
		fmt.Fprintf(s, "%s", e.String())
	default:
		fmt.Fprintf(s, makeDefaultFormatter(s, f), uint32(e))
	}
}

func (a AlgorithmId) String() string {
	switch a {
	case AlgorithmSha1:
		return "SHA-1"
	case AlgorithmSha256:
		return "SHA-256"
	case AlgorithmSha384:
		return "SHA-384"
	case AlgorithmSha512:
		return "SHA-512"
	default:
		return fmt.Sprintf("%04x", uint16(a))
	}
}

func (a AlgorithmId) Format(s fmt.State, f rune) {
	switch f {
	case 's', 'v':
		fmt.Fprintf(s, "%s", a.String())
	default:
		fmt.Fprintf(s, makeDefaultFormatter(s, f), uint16(a))
	}
}

// AlgorithmListId is a slice of AlgorithmId values,
type AlgorithmIdList []AlgorithmId

func (l AlgorithmIdList) Contains(a AlgorithmId) bool {
	for _, alg := range l {
		if alg == a {
			return true
		}
	}
	return false
}

// Event corresponds to a single event in an event log.
type Event struct {
	Index     uint      // Sequential index of event in the log
	PCRIndex  PCRIndex  // PCR index to which this event was measured
	EventType EventType // The type of this event
	Digests   DigestMap // The digests corresponding to this event for the supported algorithms
	Data      EventData // The data recorded with this event
}
