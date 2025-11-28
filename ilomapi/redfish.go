package ilomapi

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	goredfish "github.com/stmcginnis/gofish/redfish"
)

func (r *RedFishApi) GetInterfaces(onlyActiveIfaces bool) ([]IfaceConfig, error) {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
		return nil, err
	}
	defer c.Logout()

	service := c.Service
	var ifaceConfigs []IfaceConfig

	// Get all chassis
	chassis, err := service.Chassis()
	if err != nil {
		log.Warnf("Failed to get chassis: %v", err)
		return nil, err
	}

	// Iterate through each chassis
	for _, chass := range chassis {

		// Get network adapters for the chassis
		adapters, err := chass.NetworkAdapters()
		if err != nil {
			log.Warningf("Error getting network adapters for chassis %s: %v", chass.ID, err)
			continue
		}

		// Iterate through each network adapter
		for _, adapter := range adapters {
			// Get network ports for the adapter
			ports, err := adapter.NetworkPorts()
			if err != nil {
				log.Warningf("Error getting network ports for adapter %s: %v", adapter.ID, err)
				continue
			}

			// Iterate through each network port
			for _, port := range ports {
				ifaceConfig := IfaceConfig{
					IfaceName:  port.Name,
					IpAddress:  "", // Redfish NetworkPorts don't directly provide IPs
					MacAddress: fmt.Sprintf("%v", port.AssociatedNetworkAddresses),
					Speed:      fmt.Sprintf("%d Mbps", port.CurrentLinkSpeedMbps),
					Status:     fmt.Sprintf("%v", port.LinkStatus),
				}
				ifaceConfigs = append(ifaceConfigs, ifaceConfig)
			}
		}
	}

	// Get all managers
	managers, err := service.Managers()
	if err != nil {
		log.Warnf("Failed to get managers: %v", err)
		return nil, err
	}

	// Iterate through each manager
	for _, manager := range managers {
		// Get Ethernet interfaces for the manager
		interfaces, err := manager.EthernetInterfaces()
		if err != nil {
			log.Warningf("Error getting Ethernet interfaces for manager %s: %v", manager.ID, err)
			continue
		}

		// Iterate through each Ethernet interface
		for _, iface := range interfaces {
			ipAddress := "UNKNOWN"
			if len(iface.IPv4Addresses) > 0 {
				ipAddress = iface.IPv4Addresses[0].Address
			}
			ifaceConfig := IfaceConfig{
				IfaceName:  iface.Name,
				IpAddress:  ipAddress,
				MacAddress: iface.MACAddress,
				Speed:      fmt.Sprintf("%d Mbps", iface.SpeedMbps),
				Status:     fmt.Sprintf("%v", iface.LinkStatus),
			}
			ifaceConfigs = append(ifaceConfigs, ifaceConfig)
		}
	}

	// Get all systems
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return nil, err
	}

	// Iterate through each system
	for _, system := range systems {
		ethIfaces, err := system.EthernetInterfaces()
		if err != nil {
			log.Warningf("Error getting Ethernet interfaces for system %s: %v", system.ID, err)
			continue
		}

		// Iterate through each Ethernet interface
		for _, iface := range ethIfaces {
			ipAddress := "UNKNOWN"
			if len(iface.IPv4Addresses) > 0 {
				ipAddress = iface.IPv4Addresses[0].Address
			}
			ifaceConfig := IfaceConfig{
				IfaceName:  iface.Name,
				IpAddress:  ipAddress,
				MacAddress: iface.MACAddress,
				Speed:      fmt.Sprintf("%d Mbps", iface.SpeedMbps),
				Status:     fmt.Sprintf("%v", iface.LinkStatus),
			}
			ifaceConfigs = append(ifaceConfigs, ifaceConfig)
		}
	}

	if onlyActiveIfaces {
		ifacesUp := make([]IfaceConfig, 0)

		for _, iface := range ifaceConfigs {
			if iface.Status == string(goredfish.LinkUpLinkStatus) {
				ifacesUp = append(ifacesUp, iface)
			}
		}

		return ifacesUp, err
	} else {
		return ifaceConfigs, err
	}

}

type RedFishApi struct {
	config *gofish.ClientConfig
	IlomApi
}

func (r *RedFishApi) GetEndpoint() string {
	return r.config.Endpoint
}

func (r *RedFishApi) GetFlavour() string {
	return "redfish"
}

func NewRedFishApi(iloIpAddr, port, username, password string) *RedFishApi {

	cfg := &gofish.ClientConfig{
		Endpoint:            fmt.Sprintf("https://%s:%s", iloIpAddr, port),
		Username:            username,
		Password:            password,
		Insecure:            true,
		TLSHandshakeTimeout: 3600,
	}
	return &RedFishApi{
		config: cfg,
	}
}
func (r *RedFishApi) GetHostConfig() ([]IfaceConfig, error) {

	ifaces, err := r.GetInterfaces(true)

	if err != nil {
		log.Warnf("Error initializing Redfish API: %v", err)
		return nil, err
	}

	return ifaces, nil

}

func (r *RedFishApi) SetVLANID(vlanID int) error {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
		return err
	}
	defer c.Logout()

	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return err
	}

	for _, system := range systems {

		bios, err := system.Bios()
		// Check if the BIOS settings are accessible
		if err != nil {
			log.Warnf("Failed to get BIOS from system: %v", err)
			return err
		}

		attributes := bios.Attributes

		//Print current BIOS attributes
		log.Debugf("Current BIOS VLAN attributes: %v %v", attributes["VlanId"], attributes["VlanControl"])

		attributes["VlanId"] = vlanID
		attributes["VlanControl"] = "Enabled"

		err = bios.UpdateBiosAttributes(attributes)
		if err != nil {
			log.Warnf("Failed to set BIOS attributes: %v", err)
			return err
		}

	}

	// // Define the BIOS settings payload
	// payload := map[string]interface{}{
	// 	"Attributes": map[string]interface{}{
	// 		"VlanId":            fmt.Sprintf("%d", vlanID),
	// 		"VlanControl":       "Enabled",
	// 		"NetworkBootRetry":  "Enabled",
	// 		"PreBootNetworkEnv": "Auto",
	// 	},
	// }

	// // Send a PATCH request to the BIOS endpoint
	// biosEndpoint := "/redfish/v1/Systems/1/Bios/"
	// _, err = c.Patch(biosEndpoint, payload)

	log.Debugf("Successfully set VLAN ID %d and other BIOS attributes", vlanID)
	return nil
}

func (r *RedFishApi) SetOneTimeHTTPBoot() error {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
		return err
	}
	defer c.Logout()
	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return err
	}

	for _, system := range systems {
		// Set the boot source to HTTP
		// Check the system's current state
		if system.PowerState != goredfish.OnPowerState && system.PowerState != goredfish.OffPowerState {
			log.Warnf("System %s is in POST. Cannot modify BootSourceOverrideTarget.", system.ID)
			return fmt.Errorf("system %s is in State", system.ID)
		}
		if system.Status.State == common.StartingState {
			log.Warnf("System %s is in POST. Cannot modify BootSourceOverrideTarget.", system.ID)
			return fmt.Errorf("system %s is in State %s", system.ID, system.Status.State)
		}
		// Set the boot source to HTTP
		err = system.SetBoot(goredfish.Boot{
			BootSourceOverrideTarget:  goredfish.UefiHTTPBootSourceOverrideTarget,
			BootSourceOverrideEnabled: goredfish.OnceBootSourceOverrideEnabled,
		})
		if err != nil {
			log.Warnf("Failed to set boot source to HTTP: %v", err)
			return err
		}
		log.Infof("Successfully set one-time boot to HTTP for system %s", system.ID)
	}

	return err
}

func (r *RedFishApi) GetServerPowerState() (PowerState, error) {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
		return PowerStateUnknown, err
	}
	defer c.Logout()

	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return PowerStateUnknown, err
	}

	for _, system := range systems {
		switch system.PowerState {
		case goredfish.OnPowerState:
			return PowerStateOn, nil
		case goredfish.OffPowerState:
			return PowerStateOff, nil
		case goredfish.PoweringOnPowerState:
			return PowerStateStarting, nil
		case goredfish.PoweringOffPowerState:
			return PowerStateStopping, nil
		default:
			return PowerStateUnknown, nil
		}
	}

	return PowerStateUnknown, nil
}

func (r *RedFishApi) RebootServer() error {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
	}
	defer c.Logout()

	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return err
	}

	for _, system := range systems {
		// Reboot the system
		err = system.Reset(goredfish.ForceRestartResetType)
		if err != nil {
			log.Warningf("Failed to reboot system %s: %v", system.ID, err)
			continue
		}
		log.Infof("Successfully triggered reboot for system %s", system.ID)
	}
	return err
}

func (r *RedFishApi) StartServer() error {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
	}
	defer c.Logout()

	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return err
	}

	for _, system := range systems {
		// Start the system
		err = system.Reset(goredfish.OnResetType)
		if err != nil {
			log.Warningf("Failed to start system %s: %v", system.ID, err)
			continue
		}
		log.Infof("Successfully triggered start for system %s", system.ID)
	}
	return err
}

func (r *RedFishApi) StopServer() error {
	// Connect to the Redfish API
	c, err := gofish.Connect(*r.config)
	if err != nil {
		log.Warnf("Failed to connect to Redfish API: %v", err)
	}
	defer c.Logout()

	// Get all systems
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		log.Warnf("Failed to get systems: %v", err)
		return err
	}

	for _, system := range systems {
		// Start the system
		err = system.Reset(goredfish.ForceOffResetType)
		if err != nil {
			log.Warningf("Failed to stop system %s: %v", system.ID, err)
			continue
		}
		log.Infof("Successfully triggered for system %s to shut down", system.ID)
	}
	return err
}
