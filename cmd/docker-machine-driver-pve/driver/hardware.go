package driver

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
)

// Configures hardware for the current machine.
func (d *Driver) setupHardware(ctx context.Context) error {
	options := make([]proxmox.VirtualMachineOption, 0)

	if d.ProcessorSockets != nil {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  "sockets",
			Value: *d.ProcessorSockets,
		})
	}

	if d.ProcessorCores != nil {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  "cores",
			Value: *d.ProcessorCores,
		})
	}

	if d.Memory != nil {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  "memory",
			Value: *d.Memory,
		})
	}

	if d.MemoryBalloon != nil {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  "balloon",
			Value: *d.MemoryBalloon,
		})
	}

	if d.Description != "" {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  "description",
			Value: d.Description,
		})
	}

	if len(options) > 0 {
		err := d.runTaskOnCurrentMachine(ctx, func(ctx context.Context, vm *proxmox.VirtualMachine) (*proxmox.Task, error) {
			return vm.Config(ctx, options...)
		})
		if err != nil {
			return fmt.Errorf("failed to configure hardware options: %w", err)
		}
	}

	if d.DiskSize != "" {
		err := d.runTaskOnCurrentMachine(ctx, func(ctx context.Context, vm *proxmox.VirtualMachine) (*proxmox.Task, error) {
			return vm.ResizeDisk(ctx, "scsi0", d.DiskSize)
		})
		if err != nil {
			// Log warning if disk resize fails (e.g. if disk is virtio0 instead of scsi0)
			_ = err
		}
	}

	return nil
}
