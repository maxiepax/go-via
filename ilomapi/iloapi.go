package ilomapi

type IlomApi interface {
	// GetEndpoint
	GetEndpoint() string
	// GetFlavour of ilom (e.g. HPE, Dell, etc.)
	GetFlavour() string

	// GetServerPowerState
	GetServerPowerState() (PowerState, error)

	// GetHostConfig returns the host configuration for the given iloIpAddr and port
	GetHostConfig() ([]IfaceConfig, error)

	// SetVLANID sets the VLAN ID in the BIOS
	SetVLANID(vlanID int) error

	// SetOneTimeHTTPBoot sets the boot source to HTTP for one-time boot
	SetOneTimeHTTPBoot() error

	// RebootServer reboots the server
	RebootServer() error

	// StartServer starts the server
	StartServer() error

	// StopServer stops the server
	StopServer() error
}

type IfaceConfig struct {
	IfaceName  string `json:"ifaceName"`
	IpAddress  string `json:"ipAddress"`
	MacAddress string `json:"macAddress"`
	Speed      string `json:"speed"`  // Added Speed field
	Status     string `json:"status"` // Added Status field
}

type PowerState int

const (
	PowerStateUnknown PowerState = iota
	PowerStateOff
	PowerStateOn
	PowerStateStarting
	PowerStateStopping
)

// String returns the string representation of PowerState
func (ps PowerState) String() string {
	switch ps {
	case PowerStateOn:
		return "on"
	case PowerStateOff:
		return "off"
	case PowerStateUnknown:
		return "unknown"
	case PowerStateStarting:
		return "starting"
	case PowerStateStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// ParsePowerState parses a string into a PowerState
func ParsePowerState(s string) PowerState {
	switch s {
	case "on":
		return PowerStateOn
	case "off":
		return PowerStateOff
	case "starting":
		return PowerStateStarting
	case "stopping":
		return PowerStateStopping
	default:
		return PowerStateUnknown
	}
}
