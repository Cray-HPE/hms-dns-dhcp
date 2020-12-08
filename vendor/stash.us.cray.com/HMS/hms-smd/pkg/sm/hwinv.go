// Copyright 2018-2020 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.

package sm

import (
	"encoding/json"
	base "stash.us.cray.com/HMS/hms-base"
	rf "stash.us.cray.com/HMS/hms-smd/pkg/redfish"
)

var ErrHWLocInvalid = base.NewHMSError("sm", "ID is empty or not a valid xname")
var ErrHWFRUIDInvalid = base.NewHMSError("sm", "FRUID is empty or invalid")
var ErrHWInvFmtInvalid = base.NewHMSError("sm", "Invalid HW Inventory format")
var ErrHWInvFmtNI = base.NewHMSError("sm", "HW Inv format not yet implemented")

// Note most of these structures are polymorphic in the sense that they
// are stored generically in the database largely as raw json.  The non
// json parts of the struct are constant across different types.  That said,
// the client only cares about the final payload, so we give each a name
// specific to the type.  In the published spec, the json stored and
// retrieved will have to match the schema for the type in the array.
//
// Also, the hwinv location/FRU schemas the raw Redfish for the type, but
// after sorting the FRU (physical) properties and locatation (specific
// to current location) into separate embedded sub-structs (e.g. SystemFRUInfo
// and SystemLocationInfo) so we can easily snip them out separately.
// these two types of structs determine the actual schema that appears on
// the wire and gets stored as json in the DB.  The structures here should
// be largely static.  Other fields in the Redfish output that do not
// fall into one of these structures is not used for HWInventory, but is
// collected to assist in discovery or whatever other purpose.

// This is an embedded structure for HW inventory.  There should be one
// array for every hms type tracked in the inventory.  This structure
// is also reused to allow individual HWInvByLoc structures to represent
// child components for nested inventory structures.
type hmsTypeArrays struct {
	Nodes          *[]*HWInvByLoc `json:"Nodes,omitempty"`
	Cabinets       *[]*HWInvByLoc `json:"Cabinets,omitempty"`
	Chassis        *[]*HWInvByLoc `json:"Chassis,omitempty"`
	ComputeModules *[]*HWInvByLoc `json:"ComputeModules,omitempty"`
	RouterModules  *[]*HWInvByLoc `json:"RouterModules,omitempty"`
	NodeEnclosures *[]*HWInvByLoc `json:"NodeEnclosures,omitempty"`
	HSNBoards      *[]*HWInvByLoc `json:"HSNBoards,omitempty"`

	Processors *[]*HWInvByLoc `json:"Processors,omitempty"`
	Memory     *[]*HWInvByLoc `json:"Memory,omitempty"`
	Drives     *[]*HWInvByLoc `json:"Drives,omitempty"`

	CabinetPDUs       *[]*HWInvByLoc `json:"CabinetPDUs,omitempty"`
	CabinetPDUOutlets *[]*HWInvByLoc `json:"CabinetPDUOutlets,omitempty"`

	// These don't have hardware inventory location/FRU info yet,
	// either because they aren't known yet or because they are manager
	// types.  Each manager (e.g. BMC) should have some kind of physical
	// enclosure, and for the purposes of HW inventory we might not need
	// both (but probably will).
	CECs           *[]*HWInvByLoc `json:"CECs,omitempty"`
	CDUs           *[]*HWInvByLoc `json:"CDUs,omitempty"`
	CabinetCDUs    *[]*HWInvByLoc `json:"CabinetCDUs,omitempty"`
	CMMRectifiers  *[]*HWInvByLoc `json:"CMMRectifiers,omitempty"`
	CMMFpgas       *[]*HWInvByLoc `json:"CMMFpgas,omitempty"`
	NodeAccels     *[]*HWInvByLoc `json:"NodeAccels,omitempty"`
	NodeFpgas      *[]*HWInvByLoc `json:"NodeFpgas,omitempty"`
	RouterFpgas    *[]*HWInvByLoc `json:"RouterFpgas,omitempty"`
	RouterTORFpgas *[]*HWInvByLoc `json:"RouterTORFpgas,omitempty"`
	HSNAsics       *[]*HWInvByLoc `json:"HSNAsics,omitempty"`

	CabinetBMCs           *[]*HWInvByLoc `json:"CabinetBMCs,omitempty"`
	CabinetPDUControllers *[]*HWInvByLoc `json:"CabinetPDUControllers,omitempty"`
	ChassisBMCs           *[]*HWInvByLoc `json:"ChassisBMCs,omitempty"`
	NodeBMCs              *[]*HWInvByLoc `json:"NodeBMCs,omitempty"`
	RouterBMCs            *[]*HWInvByLoc `json:"RouterBMCs,omitempty"`

	CabinetPDUNics             *[]*HWInvByLoc `json:"CabinetPDUNics,omitempty"`
	NodeEnclosurePowerSupplies *[]*HWInvByLoc `json:"NodeEnclosurePowerSupplies,omitempty"`
	NodePowerConnectors        *[]*HWInvByLoc `json:"NodePowerConnectors,omitempty"`
	NodeBMCNics                *[]*HWInvByLoc `json:"NodeBMCNics,omitempty"`
	NodeNICs                   *[]*HWInvByLoc `json:"NodeNICs,omitempty"`
	NodeHsnNICs                *[]*HWInvByLoc `json:"NodeHsnNICs,omitempty"`
	RouterBMCNics              *[]*HWInvByLoc `json:"RouterBMCNics,omitempty"`

	// Also not implemented yet.  Not clear if these will have any interesting
	// info, so they may never be,
	SMSBoxes             *[]*HWInvByLoc `json:"SMSBoxes,omitempty"`
	HSNLinks             *[]*HWInvByLoc `json:"HSNLinks,omitempty"`
	HSNConnectors        *[]*HWInvByLoc `json:"HSNConnectors,omitempty"`
	HSNConnectorPorts    *[]*HWInvByLoc `json:"HSNConnectorPorts,omitempty"`
	MgmtSwitches         *[]*HWInvByLoc `json:"MgmtSwitches,omitempty"`
	MgmtSwitchConnectors *[]*HWInvByLoc `json:"MgmtSwitchConnectors,omitempty"`
}

// This is a top-level hardware inventory.  We can do a flat mapping
// where every component tracked is its own top-level array, a
// completely hierarchical mapping (since the entry for a component
// can contain it's own set of hmsTypeArrays), or some combination
// (such as node subcomponents being nested, but not higher-level
// components).
type SystemHWInventory struct {
	XName  string
	Format string

	hmsTypeArrays
}

// Valid values for Format field above
const (
	HWInvFormatFullyFlat     = "FullyFlat"
	HWInvFormatHierarchical  = "Hierarchical"  // Not implemented yet.
	HWInvFormatNestNodesOnly = "NestNodesOnly" // Default
)

// Create formatted SystemHWInventory from a random array of HWInvByLoc entries.
// No sorting is done (with components of the same type), so pre/post-sort if
// needed.
// Note: entries in *HWInvByLoc are not copied if modified.  Child entries
// will be appended if format is not FullyFlat, but otherwise no changes will
// be made.
func NewSystemHWInventory(hwlocs []*HWInvByLoc, xName, format string) (*SystemHWInventory, error) {
	hwinv := new(SystemHWInventory)
	hwinv.XName = xName
	if format == HWInvFormatNestNodesOnly ||
		format == HWInvFormatHierarchical ||
		format == HWInvFormatFullyFlat {

		hwinv.Format = format
	} else {
		return nil, ErrHWInvFmtInvalid
	}
	var err error
	for _, hwloc := range hwlocs {
		switch base.ToHMSType(hwloc.Type) {
		// HWInv based on Redfish "Chassis" Type.
		case base.Cabinet:
			if hwinv.Cabinets == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Cabinets = &arr
			}
			*hwinv.Cabinets = append(*hwinv.Cabinets, hwloc)
		case base.Chassis:
			if hwinv.Chassis == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Chassis = &arr
			}
			*hwinv.Chassis = append(*hwinv.Chassis, hwloc)
		case base.ComputeModule:
			if hwinv.ComputeModules == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.ComputeModules = &arr
			}
			*hwinv.ComputeModules = append(*hwinv.ComputeModules, hwloc)
		case base.RouterModule:
			if hwinv.RouterModules == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.RouterModules = &arr
			}
			*hwinv.RouterModules = append(*hwinv.RouterModules, hwloc)
		case base.NodeEnclosure:
			if hwinv.NodeEnclosures == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeEnclosures = &arr
			}
			*hwinv.NodeEnclosures = append(*hwinv.NodeEnclosures, hwloc)
		case base.HSNBoard:
			if hwinv.HSNBoards == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.HSNBoards = &arr
			}
			*hwinv.HSNBoards = append(*hwinv.HSNBoards, hwloc)
		case base.Node:
			if hwinv.Nodes == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Nodes = &arr
			}
			*hwinv.Nodes = append(*hwinv.Nodes, hwloc)
		case base.Processor:
			if hwinv.Processors == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Processors = &arr
			}
			*hwinv.Processors = append(*hwinv.Processors, hwloc)
		case base.Memory:
			if hwinv.Memory == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Memory = &arr
			}
			*hwinv.Memory = append(*hwinv.Memory, hwloc)
		case base.Drive:
			if hwinv.Drives == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Drives = &arr
			}
			*hwinv.Drives = append(*hwinv.Drives, hwloc)
		case base.CabinetPDU:
			if hwinv.CabinetPDUs == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CabinetPDUs = &arr
			}
			*hwinv.CabinetPDUs = append(*hwinv.CabinetPDUs, hwloc)
		case base.CabinetPDUOutlet:
			if hwinv.CabinetPDUOutlets == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CabinetPDUOutlets = &arr
			}
			*hwinv.CabinetPDUOutlets = append(*hwinv.CabinetPDUOutlets, hwloc)
		case base.HMSTypeInvalid:
			err = base.ErrHMSTypeInvalid
		// Not supported for this type.
		default:
			err = base.ErrHMSTypeUnsupported
		}
	}
	// If not completely "FullyFlat", start rolling up subcomponent
	// arrays into their parent components and then dropping them.
	if hwinv.Format == HWInvFormatNestNodesOnly ||
		hwinv.Format == HWInvFormatHierarchical {

		//
		// Nodes
		//

		// Roll up Node subcomponents into their parent nodes.
		// For "NestNodesOnly" this is the only roll-up step needed.

		// Avoid n^2 by first creating map to look up parent nodes in
		// constant-ish time.
		nmap := make(map[string]*HWInvByLoc)
		if hwinv.Nodes != nil {
			for _, n := range *hwinv.Nodes {
				nmap[n.ID] = n
			}
		}

		procArray := hwinv.Processors
		memArray := hwinv.Memory
		driveArray := hwinv.Drives
		// Moving these contents to underneath items in node array.
		// Set these arrays to nil so we won't list them twice.
		hwinv.Processors = nil
		hwinv.Memory = nil
		hwinv.Drives = nil

		// Processors are children of Node
		if procArray != nil {
			for _, p := range *procArray {
				parentID := base.GetHMSCompParent(p.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, p.ID)
					if hwinv.Processors == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Processors = &arr
					}
					// Put orphan components back in their array
					*hwinv.Processors = append(*hwinv.Processors, p)
				} else {
					if parent.Processors == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Processors = &arr
					}
					*parent.Processors = append(*parent.Processors, p)
				}
			}
		}
		// Memory modules are children of Node
		if memArray != nil {
			for _, m := range *memArray {
				parentID := base.GetHMSCompParent(m.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, m.ID)
					if hwinv.Memory == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Memory = &arr
					}
					// Put orphan components back in their array
					*hwinv.Memory = append(*hwinv.Memory, m)
				} else {
					if parent.Memory == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Memory = &arr
					}
					*parent.Memory = append(*parent.Memory, m)
				}
			}
		}

		// Drives are children of Node
		if driveArray != nil {
			for _, d := range *driveArray {
				parentID := d.ID
				for base.GetHMSType(parentID) != base.Node {
					parentID = base.GetHMSCompParent(parentID)
				}
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, d.ID)
					if hwinv.Drives == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Drives = &arr
					}
					// Put orphan components back in their array
					*hwinv.Drives = append(*hwinv.Drives, d)
				} else {
					if parent.Drives == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Drives = &arr
					}
					*parent.Drives = append(*parent.Drives, d)
				}
			}
		}

		//
		// PDUs, nest outlets
		//

		// Avoid n^2 by first creating map to look up parent PDU in
		// constant-ish time.
		pdumap := make(map[string]*HWInvByLoc)
		if hwinv.CabinetPDUs != nil {
			for _, pdu := range *hwinv.CabinetPDUs {
				pdumap[pdu.ID] = pdu
			}
		}

		cabPDUArray := hwinv.CabinetPDUOutlets
		// Moving these contents to underneath items in CabinetPDU array.
		// Set this array to nil so we won't list them twice.
		hwinv.CabinetPDUOutlets = nil

		// CabinetPDUOutlets are children of CabinetPDUs
		if cabPDUArray != nil {
			for _, out := range *cabPDUArray {
				parentID := base.GetHMSCompParent(out.ID)
				parent, ok := pdumap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find pdu key %s for %s",
						parentID, out.ID)
					if hwinv.CabinetPDUOutlets == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.CabinetPDUOutlets = &arr
					}
					// Put orphan components back in their array
					*hwinv.CabinetPDUOutlets = append(*hwinv.CabinetPDUOutlets, out)
				} else {
					if parent.CabinetPDUOutlets == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.CabinetPDUOutlets = &arr
					}
					*parent.CabinetPDUOutlets = append(*parent.CabinetPDUOutlets, out)
				}
			}
		}
		// Moved these contents to underneath items in CabinetPDU array.
		// Set these arrays to nil so we won't list them twice.
		hwinv.CabinetPDUOutlets = nil
	}
	if hwinv.Format == HWInvFormatHierarchical {
		// Continue rolling up components
		// TODO - need to implement controllers first.
		return hwinv, ErrHWInvFmtNI
	}
	return hwinv, err
}

////////////////////////////////////////////////////////////////////////////
//
// HW Inventory-by-location
//
// This is an individual component in the hardware inventory.  Or more
// accurately a location where the component is, linked to a separate
// object that describes the durable properties of the actual piece of
// physical hardware.  The latter can be tracked independently of
// its current location.
//
////////////////////////////////////////////////////////////////////////////

type HWInvByLoc struct {
	ID      string `json:"ID"`
	Type    string `json:"Type"`
	Ordinal int    `json:"Ordinal"`
	Status  string `json:"Status"`

	// This is used as a descriminator to determine the type of *Info
	// struct that will be included below.
	HWInventoryByLocationType string `json:"HWInventoryByLocationType"`

	// One of:var ErrHMSXnameInvalid = errors.New("got HMSTypeInvalid instead of valid type")
	//    HMSType                  Underlying RF Type          How named in json object
	HMSCabinetLocationInfo       *rf.ChassisLocationInfoRF   `json:"CabinetLocationInfo,omitempty"`
	HMSChassisLocationInfo       *rf.ChassisLocationInfoRF   `json:"ChassisLocationInfo,omitempty"` // Mountain chassis
	HMSComputeModuleLocationInfo *rf.ChassisLocationInfoRF   `json:"ComputeModuleLocationInfo,omitempty"`
	HMSRouterModuleLocationInfo  *rf.ChassisLocationInfoRF   `json:"RouterModuleLocationInfo,omitempty"`
	HMSNodeEnclosureLocationInfo *rf.ChassisLocationInfoRF   `json:"NodeEnclosureLocationInfo,omitempty"`
	HMSHSNBoardLocationInfo      *rf.ChassisLocationInfoRF   `json:"HSNBoardLocationInfo,omitempty"`
	HMSNodeLocationInfo          *rf.SystemLocationInfoRF    `json:"NodeLocationInfo,omitempty"`
	HMSProcessorLocationInfo     *rf.ProcessorLocationInfoRF `json:"ProcessorLocationInfo,omitempty"`
	HMSMemoryLocationInfo        *rf.MemoryLocationInfoRF    `json:"MemoryLocationInfo,omitempty"`
	HMSDriveLocationInfo         *rf.DriveLocationInfoRF     `json:"DriveLocationInfo,omitempty"`

	HMSPDULocationInfo                      *rf.PowerDistributionLocationInfo `json:"PDULocationInfo,omitempty"`
	HMSOutletLocationInfo                   *rf.OutletLocationInfo            `json:"OutletLocationInfo,omitempty"`
	HMSCMMRectifierLocationInfo             *rf.PowerSupplyLocationInfoRF     `json:"CMMRectifierLocationInfo,omitempty"`
	HMSNodeEnclosurePowerSupplyLocationInfo *rf.PowerSupplyLocationInfoRF     `json:"NodeEnclosurePowerSupplyLocationInfo,omitempty"`
	// TODO: Remaining types in hmsTypeArrays

	// If status != empty, up to one of following, matching above *Info.
	PopulatedFRU *HWInvByFRU `json:"PopulatedFRU,omitempty"`

	// These are for nested references for subcomponents.
	hmsTypeArrays
}

// HWInventoryByLocationType
// TODO: Remaining types
const (
	HWInvByLocCabinet                  string = "HWInvByLocCabinet"
	HWInvByLocChassis                  string = "HWInvByLocChassis"
	HWInvByLocComputeModule            string = "HWInvByLocComputeModule"
	HWInvByLocRouterModule             string = "HWInvByLocRouterModule"
	HWInvByLocNodeEnclosure            string = "HWInvByLocNodeEnclosure"
	HWInvByLocHSNBoard                 string = "HWInvByLocHSNBoard"
	HWInvByLocNode                     string = "HWInvByLocNode"
	HWInvByLocProcessor                string = "HWInvByLocProcessor"
	HWInvByLocDrive                    string = "HWInvByLocDrive"
	HWInvByLocMemory                   string = "HWInvByLocMemory"
	HWInvByLocPDU                      string = "HWInvByLocPDU"
	HWInvByLocOutlet                   string = "HWInvByLocOutlet"
	HWInvByLocCMMRectifier             string = "HWInvByLocCMMRectifier"
	HWInvByLocNodeEnclosurePowerSupply string = "HWInvByLocNodeEnclosurePowerSupply"
)

////////////////////////////////////////////////////////////////////////////
// Encoding/decoding: HW Inventory location info
////////////////////////////////////////////////////////////////////////////

// This routine takes raw location info captured as free-form JSON (e.g.
// from a schema-free database field) and unmarshals it into the correct struct
// for the type with the proper type-specific name.
//
// NOTEs: The location info should be that produced by EncodeLocationInfo.
//        MODIFIES caller.
//
// Return: If err != nil hw is unmodified,
//         Else, the type's *LocationInfo pointer is set to the expected struct.
func (hw *HWInvByLoc) DecodeLocationInfo(locInfoJSON []byte) error {
	var err error
	var rfChassisLocationInfo *rf.ChassisLocationInfoRF
	var rfSystemLocationInfo *rf.SystemLocationInfoRF
	var rfProcessorLocationInfo *rf.ProcessorLocationInfoRF
	var rfDriveLocationInfo *rf.DriveLocationInfoRF
	var rfMemoryLocationInfo *rf.MemoryLocationInfoRF
	var rfPDULocationInfo *rf.PowerDistributionLocationInfo
	var rfOutletLocationInfo *rf.OutletLocationInfo
	var rfCMMRectifierLocationInfo *rf.PowerSupplyLocationInfoRF
	var rfNodeEnclosurePowerSupplyLocationInfo *rf.PowerSupplyLocationInfoRF

	switch base.ToHMSType(hw.Type) {
	// HWInv based on Redfish "Chassis" Type.  Identical structs (for now).
	case base.Cabinet:
		fallthrough
	case base.Chassis:
		fallthrough
	case base.ComputeModule:
		fallthrough
	case base.RouterModule:
		fallthrough
	case base.NodeEnclosure:
		fallthrough
	case base.HSNBoard:
		rfChassisLocationInfo = new(rf.ChassisLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfChassisLocationInfo)
		if err == nil {
			// Assign struct to appropriate name for type.
			switch base.ToHMSType(hw.Type) {
			case base.Cabinet:
				hw.HMSCabinetLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocCabinet
			case base.Chassis:
				hw.HMSChassisLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocChassis
			case base.ComputeModule:
				hw.HMSComputeModuleLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocComputeModule
			case base.RouterModule:
				hw.HMSRouterModuleLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocRouterModule
			case base.NodeEnclosure:
				hw.HMSNodeEnclosureLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocNodeEnclosure
			case base.HSNBoard:
				hw.HMSHSNBoardLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocHSNBoard
			}
		}
	// HWInv based on Redfish "System" Type.
	case base.Node:
		rfSystemLocationInfo = new(rf.SystemLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfSystemLocationInfo)
		if err == nil {
			hw.HMSNodeLocationInfo = rfSystemLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNode
		}
	// HWInv based on Redfish "Processor" Type.
	case base.Processor:
		rfProcessorLocationInfo = new(rf.ProcessorLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfProcessorLocationInfo)
		if err == nil {
			hw.HMSProcessorLocationInfo = rfProcessorLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocProcessor
		}
	// HWInv based on Redfish "Memory" Type.
	case base.Memory:
		rfMemoryLocationInfo = new(rf.MemoryLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfMemoryLocationInfo)
		if err == nil {
			hw.HMSMemoryLocationInfo = rfMemoryLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocMemory
		}
	// HWInv based on Redfish "Processor" Type.
	case base.Drive:
		rfDriveLocationInfo = new(rf.DriveLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfDriveLocationInfo)
		if err == nil {
			hw.HMSDriveLocationInfo = rfDriveLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocDrive
		}
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case base.CabinetPDU:
		rfPDULocationInfo = new(rf.PowerDistributionLocationInfo)
		err = json.Unmarshal(locInfoJSON, rfPDULocationInfo)
		if err == nil {
			hw.HMSPDULocationInfo = rfPDULocationInfo
			hw.HWInventoryByLocationType = HWInvByLocPDU
		}
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case base.CabinetPDUOutlet:
		rfOutletLocationInfo = new(rf.OutletLocationInfo)
		err = json.Unmarshal(locInfoJSON, rfOutletLocationInfo)
		if err == nil {
			hw.HMSOutletLocationInfo = rfOutletLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocOutlet
		}
	case base.CMMRectifier:
		rfCMMRectifierLocationInfo = new(rf.PowerSupplyLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfCMMRectifierLocationInfo)
		if err == nil {
			hw.HMSCMMRectifierLocationInfo = rfCMMRectifierLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocCMMRectifier
		}
	case base.NodeEnclosurePowerSupply:
		rfNodeEnclosurePowerSupplyLocationInfo = new(rf.PowerSupplyLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfNodeEnclosurePowerSupplyLocationInfo)
		if err == nil {
			hw.HMSNodeEnclosurePowerSupplyLocationInfo = rfNodeEnclosurePowerSupplyLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNodeEnclosurePowerSupply
		}
	// No match - not a valid HMSType, always an error
	case base.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return err
}

//
// This function encode's the hwinv's type-specific LocationInfo struct
// into a free-form JSON byte array that can be stored schema-less in the
// database.
//
// NOTE: This function is the counterpart to DecodeLocationInfo().
//
// Returns: type's location info as JSON []byte representation, err = nil
//          Else, err != nil if encoding failed (and location_info is empty)
func (hw *HWInvByLoc) EncodeLocationInfo() ([]byte, error) {
	var err error
	var locInfoJSON []byte

	switch base.ToHMSType(hw.Type) {
	// HWInv based on Redfish "Chassis" Type.
	case base.Cabinet:
		locInfoJSON, err = json.Marshal(hw.HMSCabinetLocationInfo)
	case base.Chassis:
		locInfoJSON, err = json.Marshal(hw.HMSChassisLocationInfo)
	case base.ComputeModule:
		locInfoJSON, err = json.Marshal(hw.HMSComputeModuleLocationInfo)
	case base.RouterModule:
		locInfoJSON, err = json.Marshal(hw.HMSRouterModuleLocationInfo)
	case base.NodeEnclosure:
		locInfoJSON, err = json.Marshal(hw.HMSNodeEnclosureLocationInfo)
	case base.HSNBoard:
		locInfoJSON, err = json.Marshal(hw.HMSHSNBoardLocationInfo)
	// HWInv based on Redfish "System" Type.
	case base.Node:
		locInfoJSON, err = json.Marshal(hw.HMSNodeLocationInfo)
	// HWInv based on Redfish "Processor" Type.
	case base.Processor:
		locInfoJSON, err = json.Marshal(hw.HMSProcessorLocationInfo)
	// HWInv based on Redfish "Memory" Type.
	case base.Memory:
		locInfoJSON, err = json.Marshal(hw.HMSMemoryLocationInfo)
	// HWInv based on Redfish "Drive" Type.
	case base.Drive:
		locInfoJSON, err = json.Marshal(hw.HMSDriveLocationInfo)
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case base.CabinetPDU:
		locInfoJSON, err = json.Marshal(hw.HMSPDULocationInfo)
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case base.CabinetPDUOutlet:
		locInfoJSON, err = json.Marshal(hw.HMSOutletLocationInfo)
	case base.CMMRectifier:
		locInfoJSON, err = json.Marshal(hw.HMSCMMRectifierLocationInfo)
	case base.NodeEnclosurePowerSupply:
		locInfoJSON, err = json.Marshal(hw.HMSNodeEnclosurePowerSupplyLocationInfo)
	// No match - not a valid HMS Type, always an error
	case base.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	// Not supported for this type.
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return locInfoJSON, err
}

////////////////////////////////////////////////////////////////////////////
//
// Hardware Inventory - Field Replaceable Unit data
//
//   These are the properties of components that move with the physical
//   unit and may or may not have a matching location at the moment.  These
//   will eventually have their location histories tracked.
//
////////////////////////////////////////////////////////////////////////////

type HWInvByFRU struct {
	FRUID   string `json:"FRUID"`
	Type    string `json:"Type"`
	Subtype string `json:"Subtype"`

	// This is used as a descriminator to specify the type of *Info
	// struct that will be included below.
	HWInventoryByFRUType string `json:"HWInventoryByFRUType"`

	// One of (based on HWFRUInfoType):
	//   HMSType             Underlying RF Type      How named in json object
	HMSCabinetFRUInfo       *rf.ChassisFRUInfoRF   `json:"CabinetFRUInfo,omitempty"`
	HMSChassisFRUInfo       *rf.ChassisFRUInfoRF   `json:"ChassisFRUInfo,omitempty"` // Mountain chassis
	HMSComputeModuleFRUInfo *rf.ChassisFRUInfoRF   `json:"ComputeModuleFRUInfo,omitempty"`
	HMSRouterModuleFRUInfo  *rf.ChassisFRUInfoRF   `json:"RouterModuleFRUInfo,omitempty"`
	HMSNodeEnclosureFRUInfo *rf.ChassisFRUInfoRF   `json:"NodeEnclosureFRUInfo,omitempty"`
	HMSHSNBoardFRUInfo      *rf.ChassisFRUInfoRF   `json:"HSNBoardFRUInfo,omitempty"`
	HMSNodeFRUInfo          *rf.SystemFRUInfoRF    `json:"NodeFRUInfo,omitempty"`
	HMSProcessorFRUInfo     *rf.ProcessorFRUInfoRF `json:"ProcessorFRUInfo,omitempty"`
	HMSMemoryFRUInfo        *rf.MemoryFRUInfoRF    `json:"MemoryFRUInfo,omitempty"`
	HMSDriveFRUInfo         *rf.DriveFRUInfoRF     `json:"DriveFRUInfo,omitempty"`

	HMSPDUFRUInfo                      *rf.PowerDistributionFRUInfo `json:"PDUFRUInfo,omitempty"`
	HMSOutletFRUInfo                   *rf.OutletFRUInfo            `json:"OutletFRUInfo,omitempty"`
	HMSCMMRectifierFRUInfo             *rf.PowerSupplyFRUInfoRF     `json:"CMMRectifierFRUInfo,omitempty"`
	HMSNodeEnclosurePowerSupplyFRUInfo *rf.PowerSupplyFRUInfoRF     `json:"NodeEnclosurePowerSupplyFRUInfo,omitempty"`

	// TODO: Remaining types in hmsTypeArrays
}

// HWInventoryByFRUType properties.  Used to select proper subtype in
// api schema.
// TODO: Remaining types
const (
	HWInvByFRUCabinet                  string = "HWInvByFRUCabinet"
	HWInvByFRUChassis                  string = "HWInvByFRUChassis"
	HWInvByFRUComputeModule            string = "HWInvByFRUComputeModule"
	HWInvByFRURouterModule             string = "HWInvByFRURouterModule"
	HWInvByFRUNodeEnclosure            string = "HWInvByFRUNodeEnclosure"
	HWInvByFRUHSNBoard                 string = "HWInvByFRUHSNBoard"
	HWInvByFRUNode                     string = "HWInvByFRUNode"
	HWInvByFRUProcessor                string = "HWInvByFRUProcessor"
	HWInvByFRUMemory                   string = "HWInvByFRUMemory"
	HWInvByFRUDrive                    string = "HWInvByFRUDrive"
	HWInvByFRUPDU                      string = "HWInvByFRUPDU"
	HWInvByFRUOutlet                   string = "HWInvByFRUOutlet"
	HWInvByFRUCMMRectifier             string = "HWInvByFRUCMMRectifier"
	HWInvByFRUNodeEnclosurePowerSupply string = "HWInvByFRU"
)

////////////////////////////////////////////////////////////////////////////
// Encoding/decoding: HW Inventory FRU info
///////////////////////////////////////////////////////////////////////////

// This routine takes raw FRU info captured as free-form JSON (e.g.
// from a schema-free database field) and unmarshals it into the correct struct
// for the type with the proper type-specific name.
//
// NOTEs: The fruInfoJSON array should be that produced by EncodeFRUInfo.
//        MODIFIES caller.
//
// Return: If err != nil hf is unmodified and operation failed.
//         Else, the type's *FRUInfo pointer is set to the expected struct.
func (hf *HWInvByFRU) DecodeFRUInfo(fruInfoJSON []byte) error {
	var err error = nil
	var rfChassisFRUInfo *rf.ChassisFRUInfoRF
	var rfSystemFRUInfo *rf.SystemFRUInfoRF
	var rfProcessorFRUInfo *rf.ProcessorFRUInfoRF
	var rfMemoryFRUInfo *rf.MemoryFRUInfoRF
	var rfDriveFRUInfo *rf.DriveFRUInfoRF
	var rfPDUFRUInfo *rf.PowerDistributionFRUInfo
	var rfOutletFRUInfo *rf.OutletFRUInfo
	var rfCMMRectifierFRUInfo *rf.PowerSupplyFRUInfoRF
	var rfNodeEnclosurePowerSupplyFRUInfo *rf.PowerSupplyFRUInfoRF

	switch base.ToHMSType(hf.Type) {
	// HWInv based on Redfish "Chassis" Type.  Identical structs (for now).
	case base.Cabinet:
		fallthrough
	case base.Chassis:
		fallthrough
	case base.ComputeModule:
		fallthrough
	case base.RouterModule:
		fallthrough
	case base.NodeEnclosure:
		fallthrough
	case base.HSNBoard:
		rfChassisFRUInfo = new(rf.ChassisFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfChassisFRUInfo)
		if err == nil {
			// Assign struct to appropriate name for type.
			switch base.ToHMSType(hf.Type) {
			case base.Cabinet:
				hf.HMSCabinetFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUCabinet
			case base.Chassis:
				hf.HMSChassisFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUChassis
			case base.ComputeModule:
				hf.HMSComputeModuleFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUComputeModule
			case base.RouterModule:
				hf.HMSRouterModuleFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRURouterModule
			case base.NodeEnclosure:
				hf.HMSNodeEnclosureFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUNodeEnclosure
			case base.HSNBoard:
				hf.HMSHSNBoardFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUHSNBoard
			}
		}
	// HWInv based on Redfish "System" Type.
	case base.Node:
		rfSystemFRUInfo = new(rf.SystemFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfSystemFRUInfo)
		if err == nil {
			hf.HMSNodeFRUInfo = rfSystemFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNode
		}
	// HWInv based on Redfish "Processor" Type.
	case base.Processor:
		rfProcessorFRUInfo = new(rf.ProcessorFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfProcessorFRUInfo)
		if err == nil {
			hf.HMSProcessorFRUInfo = rfProcessorFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUProcessor
		}
	// HWInv based on Redfish "Memory" Type.
	case base.Memory:
		rfMemoryFRUInfo = new(rf.MemoryFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfMemoryFRUInfo)
		if err == nil {
			hf.HMSMemoryFRUInfo = rfMemoryFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUMemory
		}
	// HWInv based on Redfish "Drive" Type.
	case base.Drive:
		rfDriveFRUInfo = new(rf.DriveFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfDriveFRUInfo)
		if err == nil {
			hf.HMSDriveFRUInfo = rfDriveFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUDrive
		}
	// HWInv based on Redfish "PowerDistribution" Type.
	case base.CabinetPDU:
		rfPDUFRUInfo = new(rf.PowerDistributionFRUInfo)
		err = json.Unmarshal(fruInfoJSON, rfPDUFRUInfo)
		if err == nil {
			hf.HMSPDUFRUInfo = rfPDUFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUPDU
		}
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case base.CabinetPDUOutlet:
		rfOutletFRUInfo = new(rf.OutletFRUInfo)
		err = json.Unmarshal(fruInfoJSON, rfOutletFRUInfo)
		if err == nil {
			hf.HMSOutletFRUInfo = rfOutletFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUOutlet
		}
	// HWInv based on Redfish "PowerSupply" Type.
	case base.CMMRectifier:
		rfCMMRectifierFRUInfo = new(rf.PowerSupplyFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfCMMRectifierFRUInfo)
		if err == nil {
			hf.HMSCMMRectifierFRUInfo = rfCMMRectifierFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUCMMRectifier
		}
	// HWInv based on Redfish "PowerSupply" Type.
	case base.NodeEnclosurePowerSupply:
		rfNodeEnclosurePowerSupplyFRUInfo = new(rf.PowerSupplyFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfNodeEnclosurePowerSupplyFRUInfo)
		if err == nil {
			hf.HMSNodeEnclosurePowerSupplyFRUInfo = rfNodeEnclosurePowerSupplyFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNodeEnclosurePowerSupply
		}
	// No match - not a valid HMSType, always an error
	case base.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return err
}

//
// This function encode's the hwinv's type-specific FRU info struct
// into a free-form JSON byte array that can be stored schema-less in the
// database.
//
// NOTE: This function is the counterpart to DecodeFRUInfo().
//
// Returns: FRU's info as JSON []byte representation, err = nil
//          Else, err != nil if encoding failed (plus, []byte value is empty)
func (hf *HWInvByFRU) EncodeFRUInfo() ([]byte, error) {
	var err error
	var fruInfoJSON []byte

	switch base.ToHMSType(hf.Type) {
	// HWInv based on Redfish "Chassis" Type.
	case base.Cabinet:
		fruInfoJSON, err = json.Marshal(hf.HMSCabinetFRUInfo)
	case base.Chassis:
		fruInfoJSON, err = json.Marshal(hf.HMSChassisFRUInfo)
	case base.ComputeModule:
		fruInfoJSON, err = json.Marshal(hf.HMSComputeModuleFRUInfo)
	case base.RouterModule:
		fruInfoJSON, err = json.Marshal(hf.HMSRouterModuleFRUInfo)
	case base.NodeEnclosure:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeEnclosureFRUInfo)
	case base.HSNBoard:
		fruInfoJSON, err = json.Marshal(hf.HMSHSNBoardFRUInfo)
	// HWInv based on Redfish "System" Type.
	case base.Node:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeFRUInfo)
	// HWInv based on Redfish "Processor" Type.
	case base.Processor:
		fruInfoJSON, err = json.Marshal(hf.HMSProcessorFRUInfo)
	// HWInv based on Redfish "Memory" Type.
	case base.Memory:
		fruInfoJSON, err = json.Marshal(hf.HMSMemoryFRUInfo)
	// HWInv based on Redfish "Processor" Type.
	case base.Drive:
		fruInfoJSON, err = json.Marshal(hf.HMSDriveFRUInfo)
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case base.CabinetPDU:
		fruInfoJSON, err = json.Marshal(hf.HMSPDUFRUInfo)
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case base.CabinetPDUOutlet:
		fruInfoJSON, err = json.Marshal(hf.HMSOutletFRUInfo)
	// HWInv based on Redfish "PowerSupply" Type.
	case base.CMMRectifier:
		fruInfoJSON, err = json.Marshal(hf.HMSCMMRectifierFRUInfo)
	// HWInv based on Redfish "PowerSupply" Type.
	case base.NodeEnclosurePowerSupply:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeEnclosurePowerSupplyFRUInfo)
	// No match - not a valid HMS Type, always an error
	case base.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	// Not supported for this type.
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return fruInfoJSON, err
}
