package driver

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/luthermonson/go-proxmox"
)

const (
	// Polling interval for Proxmox task status.
	pveTaskPollingInterval = 3 * time.Second

	// Polling timeout for Proxmox task status.
	pveTaskPollingTimeout = 10 * time.Minute
)

// Creates a new Proxmox VE virtual machine from the current template.
func (d *Driver) createPVEVirtualMachine(ctx context.Context) (int, error) {
	template, err := d.getPVETemplate(ctx)
	if err != nil {
		return -1, err
	}

	vmid, task, err := template.Clone(ctx, &proxmox.VirtualMachineCloneOptions{
		Name: d.MachineName,
		Pool: d.ResourcePoolName,
		Full: map[bool]uint8{false: 0, true: 1}[d.FullClone],
	})
	if err != nil {
		return vmid, fmt.Errorf("failed to clone template ID='%d': %w", d.TemplateID, err)
	}

	if err := d.waitForPVETaskToSucceed(ctx, task); err != nil {
		return vmid, fmt.Errorf("failed to clone template ID='%d': %w", d.TemplateID, err)
	}

	return vmid, nil
}

// Returns the current Proxmox VE template.
func (d *Driver) getPVETemplate(ctx context.Context) (*proxmox.VirtualMachine, error) {
	template, err := d.getPVEVirtualMachine(ctx, d.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proxmox VE template: %w", err)
	}

	return template, nil
}

// Returns a Proxmox VE virtual machine from the current resource pool.
func (d *Driver) getPVEVirtualMachine(ctx context.Context, vmid int) (*proxmox.VirtualMachine, error) {
	if d.ResourcePoolName == "" {
		if d.NodeName != "" {
			return d.getPVEVirtualMachineOnNode(ctx, vmid, d.NodeName)
		}
		nodes, err := d.getPVEClient().Nodes(ctx)
		if err == nil {
			for _, n := range nodes {
				vm, err := d.getPVEVirtualMachineOnNode(ctx, vmid, n.Name)
				if err == nil && vm != nil {
					return vm, nil
				}
			}
		}
		return nil, fmt.Errorf("failed to retrieve Proxmox VE virtual machine ID='%d': not found on any node", vmid)
	}

	resourcePool, err := d.getCurrentPVEResourcePool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proxmox VE virtual machine ID='%d': %w", vmid, err)
	}

	for _, member := range resourcePool.Members {
		if member.VMID > math.MaxInt {
			continue
		}

		if member.Type != "qemu" || int(member.VMID) != vmid {
			continue
		}

		return d.getPVEVirtualMachineOnNode(ctx, vmid, member.Node)
	}

	return nil, fmt.Errorf("failed to retrieve Proxmox VE virtual machine ID='%d' in resource pool name='%s': not found", vmid, d.ResourcePoolName)
}

// Returns Proxmox VE virtual machine from a given node.
func (d *Driver) getPVEVirtualMachineOnNode(ctx context.Context, vmid int, nodeName string) (*proxmox.VirtualMachine, error) {
	node, err := d.getPVEClient().Node(ctx, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proxmox VE node name='%s': %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proxmox VE virtual machine ID='%d' on node name='%s': %w", vmid, nodeName, err)
	}

	return vm, nil
}

// Returns the current Proxmox VE resource pool.
func (d *Driver) getCurrentPVEResourcePool(ctx context.Context) (*proxmox.Pool, error) {
	if d.ResourcePoolName == "" {
		return nil, nil
	}
	resourcePool, err := d.getPVEClient().Pool(ctx, d.ResourcePoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proxmox VE resource pool name='%s': %w", d.ResourcePoolName, err)
	}

	return resourcePool, nil
}

// Blocks until a Proxmox VE task finishes successfully.
func (d *Driver) waitForPVETaskToSucceed(ctx context.Context, task *proxmox.Task) error {
	if err := task.Wait(ctx, pveTaskPollingInterval, pveTaskPollingTimeout); err != nil {
		return fmt.Errorf("failed waiting for task ID='%s' to complete: %w", task.ID, err)
	}

	if !task.IsSuccessful {
		return fmt.Errorf("task ID='%s' failed", task.ID)
	}

	return nil
}

// Returns a client for Proxmox VE.
func (d *Driver) getPVEClient() *proxmox.Client {
	if d.pveClient != nil {
		return d.pveClient
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec
				InsecureSkipVerify: true,
			},
		},
	}

	pveURL, err := url.Parse(d.URL)
	if err != nil {
		// Note that parsing is already checked in SetConfigFromFlags()
		panic(fmt.Errorf("failed to parse Proxmox VE URL: %w", err).Error())
	}

	options := []proxmox.Option{
		proxmox.WithHTTPClient(&httpClient),
	}

	if d.TokenID != "" && d.TokenSecret != "" {
		options = append(options, proxmox.WithAPIToken(d.TokenID, d.TokenSecret))
	} else if d.Username != "" && d.Password != "" {
		realm := d.Realm
		if realm == "" {
			realm = "pam"
		}
		options = append(options, proxmox.WithCredentials(&proxmox.Credentials{
			Username: d.Username,
			Password: d.Password,
			Realm:    realm,
		}))
	}

	d.pveClient = proxmox.NewClient(
		pveURL.JoinPath("/api2/json").String(),
		options...,
	)

	return d.pveClient
}
