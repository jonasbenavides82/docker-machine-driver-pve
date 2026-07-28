package driver

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/machine/libmachine/drivers"
	"github.com/rancher/machine/libmachine/mcnflag"
)

// Available flags.
const (
	flagURL              = "pve-url"
	flagInsecureTLS      = "pve-insecure-tls"
	flagTokenID          = "pve-token-id" //nolint:gosec // False-positive
	flagTokenSecret      = "pve-token-secret"
	flagUsername         = "pve-username"
	flagPassword         = "pve-password"
	flagRealm            = "pve-realm"
	flagNode             = "pve-node"
	flagResourcePool     = "pve-resource-pool"
	flagTemplateID       = "pve-template"
	flagISODevice        = "pve-iso-device"
	flagNetworkInterface = "pve-network-interface"
	flagSSHUser          = "pve-ssh-user"
	flagSSHPort          = "pve-ssh-port"
	flagProcessorSockets = "pve-processor-sockets"
	flagProcessorCores   = "pve-processor-cores"
	flagMemory           = "pve-memory"
	flagMemoryBalloon    = "pve-memory-balloon"
	flagFullClone        = "pve-full-clone"
	flagDiskSize         = "pve-disk-size"
	flagBridge           = "pve-bridge"
	flagVLAN             = "pve-vlan"
	flagCIPassword       = "pve-cipassword"
	flagNameserver       = "pve-nameserver"
	flagSearchdomain     = "pve-searchdomain"
	flagDescription      = "pve-description"
	flagTags             = "pve-tags"
)

// Default values for flags.
const (
	defaultSSHUser = "service"
	defaultSSHPort = 22
)

// Driver's configuration.
type config struct {
	// Proxmox VE URL (e.g. 'https://<PROXMOX VE ADDRESS>:8006').
	URL string

	// Disables Proxmox VE TLS certificate verification.
	InsecureTLS bool

	// Proxmox VE API Token ID (including username and realm, e.g. 'root@pam!rancher').
	TokenID string

	// Proxmox VE API Token secret.
	TokenSecret string

	// Fallback credentials
	Username string
	Password string
	Realm    string

	// Target Proxmox node name
	NodeName string

	// Proxmox VE Resource Pool name.
	ResourcePoolName string

	// ID of the Proxmox VE template.
	TemplateID int

	// Bus/Device of the CD/DVD Drive to mount cloud-init ISO to (e.g. 'scsi1').
	ISODeviceName string

	// Bus/Device of the network interface to read machine's IP address from (e.g. 'net0').
	NetworkInterfaceName string

	// If set, number of processor sockets to configure for the machine.
	ProcessorSockets *int

	// If set, number of processor cores to configure for the machine.
	ProcessorCores *int

	// If set, amount of memory in MiB to configure for the machine.
	Memory *int

	// If set, minimum amount of memory in MiB to configure for the machine.
	// If set to 0, disables memory ballooning.
	MemoryBalloon *int

	// Forces full copy of all disks, even if underlying storage supports linked clones.
	FullClone bool

	// Disk size expansion or target size (e.g. 20G or +10G)
	DiskSize string

	// Network bridge (e.g. vmbr0)
	Bridge string

	// VLAN Tag
	VLAN int

	// Cloud-init user password
	CIPassword string

	// Custom DNS nameserver for cloud-init
	Nameserver string

	// Custom DNS search domain for cloud-init
	Searchdomain string

	// VM Description
	Description string

	// Tags to apply to the machine.
	Tags []string
}

// GetCreateFlags implements drivers.Driver.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   flagURL,
			EnvVar: flagEnvVarFromFlagName(flagURL),
			Usage:  "Proxmox VE URL (e.g. 'https://<PROXMOX VE ADDRESS>:8006')",
		},
		mcnflag.BoolFlag{
			Name:   flagInsecureTLS,
			EnvVar: flagEnvVarFromFlagName(flagInsecureTLS),
			Usage:  "Disables Proxmox VE TLS certificate verification",
		},
		mcnflag.StringFlag{
			Name:   flagTokenID,
			EnvVar: flagEnvVarFromFlagName(flagTokenID),
			Usage:  "Proxmox VE API Token ID (e.g. 'root@pam!rancher')",
		},
		mcnflag.StringFlag{
			Name:   flagTokenSecret,
			EnvVar: flagEnvVarFromFlagName(flagTokenSecret),
			Usage:  "Proxmox VE API Token secret",
		},
		mcnflag.StringFlag{
			Name:   flagUsername,
			EnvVar: flagEnvVarFromFlagName(flagUsername),
			Usage:  "Proxmox VE Username (legacy auth fallback)",
		},
		mcnflag.StringFlag{
			Name:   flagPassword,
			EnvVar: flagEnvVarFromFlagName(flagPassword),
			Usage:  "Proxmox VE Password (legacy auth fallback)",
		},
		mcnflag.StringFlag{
			Name:   flagRealm,
			EnvVar: flagEnvVarFromFlagName(flagRealm),
			Usage:  "Proxmox VE Auth Realm (defaults to 'pam')",
		},
		mcnflag.StringFlag{
			Name:   flagNode,
			EnvVar: flagEnvVarFromFlagName(flagNode),
			Usage:  "Proxmox VE Node name to host the VM (optional)",
		},
		mcnflag.StringFlag{
			Name:   flagResourcePool,
			EnvVar: flagEnvVarFromFlagName(flagResourcePool),
			Usage:  "Proxmox VE Resource Pool name",
		},
		mcnflag.IntFlag{
			Name:   flagTemplateID,
			EnvVar: flagEnvVarFromFlagName(flagTemplateID),
			Usage:  "ID of the Proxmox VE template",
			Value:  0,
		},
		mcnflag.IntFlag{
			Name:   "pve-template-id",
			EnvVar: "PVE_TEMPLATE_ID",
			Usage:  "ID of the Proxmox VE template (alias)",
			Value:  0,
		},
		mcnflag.StringFlag{
			Name:   flagISODevice,
			EnvVar: flagEnvVarFromFlagName(flagISODevice),
			Usage:  "Bus/Device of CD/DVD Drive for cloud-init ISO (e.g. 'scsi1')",
		},
		mcnflag.StringFlag{
			Name:   "pve-isodevice",
			EnvVar: "PVE_ISODEVICE",
			Usage:  "Bus/Device of CD/DVD Drive for cloud-init ISO (alias)",
		},
		mcnflag.StringFlag{
			Name:   flagNetworkInterface,
			EnvVar: flagEnvVarFromFlagName(flagNetworkInterface),
			Usage:  "Bus/Device of network interface (e.g. 'net0')",
		},
		mcnflag.StringFlag{
			Name:   "pve-networkinterface",
			EnvVar: "PVE_NETWORKINTERFACE",
			Usage:  "Bus/Device of network interface (alias)",
		},
		mcnflag.StringFlag{
			Name:   flagSSHUser,
			EnvVar: flagEnvVarFromFlagName(flagSSHUser),
			Usage:  fmt.Sprintf("SSH user created via cloud-init, defaults to '%s'", defaultSSHUser),
		},
		mcnflag.IntFlag{
			Name:   flagSSHPort,
			EnvVar: flagEnvVarFromFlagName(flagSSHPort),
			Usage:  fmt.Sprintf("SSH port, defaults to '%d'", defaultSSHPort),
			Value:  defaultSSHPort,
		},
		mcnflag.IntFlag{
			Name:   flagProcessorSockets,
			EnvVar: flagEnvVarFromFlagName(flagProcessorSockets),
			Usage:  "Processor sockets count for the VM",
			Value:  1,
		},
		mcnflag.IntFlag{
			Name:   flagProcessorCores,
			EnvVar: flagEnvVarFromFlagName(flagProcessorCores),
			Usage:  "Processor cores count for the VM",
			Value:  1,
		},
		mcnflag.IntFlag{
			Name:   flagMemory,
			EnvVar: flagEnvVarFromFlagName(flagMemory),
			Usage:  "Memory size in MiB for the VM",
			Value:  2048,
		},
		mcnflag.IntFlag{
			Name:   flagMemoryBalloon,
			EnvVar: flagEnvVarFromFlagName(flagMemoryBalloon),
			Usage:  "Minimum memory in MiB for memory ballooning (0 to disable)",
			Value:  0,
		},
		mcnflag.BoolFlag{
			Name:   flagFullClone,
			EnvVar: flagEnvVarFromFlagName(flagFullClone),
			Usage:  "Forces full copy of all disks during template clone",
		},
		mcnflag.StringFlag{
			Name:   flagDiskSize,
			EnvVar: flagEnvVarFromFlagName(flagDiskSize),
			Usage:  "Target disk size (e.g. '20G')",
		},
		mcnflag.StringFlag{
			Name:   "pve-disksize",
			EnvVar: "PVE_DISKSIZE",
			Usage:  "Target disk size (alias)",
		},
		mcnflag.StringFlag{
			Name:   flagBridge,
			EnvVar: flagEnvVarFromFlagName(flagBridge),
			Usage:  "Network bridge interface (e.g. 'vmbr0')",
		},
		mcnflag.IntFlag{
			Name:   flagVLAN,
			EnvVar: flagEnvVarFromFlagName(flagVLAN),
			Usage:  "Network VLAN tag number",
			Value:  0,
		},
		mcnflag.StringFlag{
			Name:   flagCIPassword,
			EnvVar: flagEnvVarFromFlagName(flagCIPassword),
			Usage:  "Cloud-init user password",
		},
		mcnflag.StringFlag{
			Name:   "pve-ci-password",
			EnvVar: "PVE_CI_PASSWORD",
			Usage:  "Cloud-init user password (alias)",
		},
		mcnflag.StringFlag{
			Name:   flagNameserver,
			EnvVar: flagEnvVarFromFlagName(flagNameserver),
			Usage:  "Custom DNS nameserver for cloud-init",
		},
		mcnflag.StringFlag{
			Name:   flagSearchdomain,
			EnvVar: flagEnvVarFromFlagName(flagSearchdomain),
			Usage:  "Custom DNS search domain for cloud-init",
		},
		mcnflag.StringFlag{
			Name:   flagDescription,
			EnvVar: flagEnvVarFromFlagName(flagDescription),
			Usage:  "VM description text in Proxmox",
		},
		mcnflag.StringFlag{
			Name:   flagTags,
			EnvVar: flagEnvVarFromFlagName(flagTags),
			Usage:  "Comma-separated list of tags to assign to the VM",
		},
	}
}

// SetConfigFromFlags implements drivers.Driver.
//
//nolint:cyclop,gocyclo
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.URL = opts.String(flagURL)
	if d.URL == "" {
		return fmt.Errorf("flag '--%s' is required", flagURL)
	}

	if _, err := url.Parse(d.URL); err != nil {
		return fmt.Errorf("failed to parse Proxmox VE URL (flag '--%s'): %w", flagURL, err)
	}

	d.InsecureTLS = opts.Bool(flagInsecureTLS)

	d.TokenID = opts.String(flagTokenID)
	d.TokenSecret = opts.String(flagTokenSecret)
	d.Username = opts.String(flagUsername)
	d.Password = opts.String(flagPassword)
	d.Realm = opts.String(flagRealm)

	if d.TokenID == "" && (d.Username == "" || d.Password == "") {
		return fmt.Errorf("either API token ('--%s' & '--%s') or credentials ('--%s' & '--%s') are required", flagTokenID, flagTokenSecret, flagUsername, flagPassword)
	}

	d.NodeName = opts.String(flagNode)
	d.ResourcePoolName = opts.String(flagResourcePool)

	var err error

	if tStr := opts.String(flagTemplateID); tStr != "" {
		d.TemplateID, _ = strconv.Atoi(tStr)
	} else if tStr := opts.String("pve-template-id"); tStr != "" {
		d.TemplateID, _ = strconv.Atoi(tStr)
	} else {
		d.TemplateID = opts.Int(flagTemplateID)
		if d.TemplateID <= 0 {
			d.TemplateID = opts.Int("pve-template-id")
		}
	}

	if d.TemplateID <= 0 {
		return fmt.Errorf("flag '--%s' is required and must be > 0", flagTemplateID)
	}

	d.ISODeviceName = strings.ToLower(opts.String(flagISODevice))
	if d.ISODeviceName == "" {
		d.ISODeviceName = strings.ToLower(opts.String("pve-isodevice"))
	}
	if d.ISODeviceName == "" {
		d.ISODeviceName = "scsi1"
	}

	d.NetworkInterfaceName = opts.String(flagNetworkInterface)
	if d.NetworkInterfaceName == "" {
		d.NetworkInterfaceName = opts.String("pve-networkinterface")
	}
	if d.NetworkInterfaceName == "" {
		d.NetworkInterfaceName = "net0"
	}

	d.SSHUser = opts.String(flagSSHUser)
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	if pStr := opts.String(flagSSHPort); pStr != "" {
		d.SSHPort, _ = strconv.Atoi(pStr)
	} else {
		d.SSHPort = opts.Int(flagSSHPort)
	}

	if d.SSHPort == 0 {
		d.SSHPort = defaultSSHPort
	} else if d.SSHPort < 0 {
		return fmt.Errorf("flag '--%s' must be > 0", flagSSHPort)
	}

	if d.ProcessorSockets, err = parseStringFlagToInt(opts.String(flagProcessorSockets)); err != nil {
		return fmt.Errorf("failed to parse '--%s': %w", flagProcessorSockets, err)
	} else if d.ProcessorSockets != nil && *d.ProcessorSockets < 1 {
		return fmt.Errorf("flag '--%s' must be >= 1", flagProcessorSockets)
	}

	if d.ProcessorCores, err = parseStringFlagToInt(opts.String(flagProcessorCores)); err != nil {
		return fmt.Errorf("failed to parse '--%s': %w", flagProcessorCores, err)
	} else if d.ProcessorCores != nil && *d.ProcessorCores < 1 {
		return fmt.Errorf("flag '--%s' must be >= 1", flagProcessorCores)
	}

	if d.Memory, err = parseStringFlagToInt(opts.String(flagMemory)); err != nil {
		return fmt.Errorf("failed to parse '--%s': %w", flagMemory, err)
	} else if d.Memory != nil && *d.Memory < 1 {
		return fmt.Errorf("flag '--%s' must be >= 1", flagMemory)
	}

	if d.MemoryBalloon, err = parseStringFlagToInt(opts.String(flagMemoryBalloon)); err != nil {
		return fmt.Errorf("failed to parse '--%s': %w", flagMemoryBalloon, err)
	} else if d.MemoryBalloon != nil && *d.MemoryBalloon < 0 {
		return fmt.Errorf("flag '--%s' must be >= 1; set to 0 to disable", flagMemoryBalloon)
	}

	// Default memory/memory balloon to the other one if it's set
	if d.Memory != nil && d.MemoryBalloon == nil {
		d.MemoryBalloon = d.Memory
	}

	if d.MemoryBalloon != nil && *d.MemoryBalloon != 0 && d.Memory == nil {
		d.Memory = d.MemoryBalloon
	}

	if d.Memory != nil && d.MemoryBalloon != nil && *d.MemoryBalloon > *d.Memory {
		return fmt.Errorf("flag '--%s' must be <= than flag '--%s'", flagMemoryBalloon, flagMemory)
	}

	d.FullClone = opts.Bool(flagFullClone)
	d.DiskSize = opts.String(flagDiskSize)
	if d.DiskSize == "" {
		d.DiskSize = opts.String("pve-disksize")
	}

	d.Bridge = opts.String(flagBridge)

	if vStr := opts.String(flagVLAN); vStr != "" {
		d.VLAN, _ = strconv.Atoi(vStr)
	} else {
		d.VLAN = opts.Int(flagVLAN)
	}
	d.CIPassword = opts.String(flagCIPassword)
	if d.CIPassword == "" {
		d.CIPassword = opts.String("pve-ci-password")
	}
	d.Nameserver = opts.String(flagNameserver)
	d.Searchdomain = opts.String(flagSearchdomain)
	d.Description = opts.String(flagDescription)

	if opts.String(flagTags) != "" {
		d.Tags = strings.Split(opts.String(flagTags), ",")
	}

	return nil
}

// Creates flag's EnvVar from it's name.
func flagEnvVarFromFlagName(name string) string {
	return strings.ToUpper(
		strings.ReplaceAll(
			name,
			"-",
			"_",
		),
	)
}

// Parses string flag to integer. Returns nil if the flag was unset/empty.
func parseStringFlagToInt(value string) (*int, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		//nolint:nilnil
		return nil, nil
	}

	numberValue, err := strconv.Atoi(trimmedValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to int: %w", err)
	}

	return &numberValue, nil
}
