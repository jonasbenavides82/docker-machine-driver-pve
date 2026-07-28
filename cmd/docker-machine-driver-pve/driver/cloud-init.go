package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/rancher/machine/libmachine/log"
	yaml "gopkg.in/yaml.v3"
)

// Configures cloud-init for the current machine.
func (d *Driver) setupCloudinit(ctx context.Context) error {
	machine, err := d.getCurrentMachine(ctx)
	if err != nil {
		return err
	}

	cloudinitMetadata, err := d.generateCloudinitMetadata()
	if err != nil {
		return fmt.Errorf("failed to generate cloud-init metadata: %w", err)
	}

	cloudinitUserdata, err := d.generateCloudinitUserdata()
	if err != nil {
		return fmt.Errorf("failed to generate cloud-init userdata: %w", err)
	}

	if machine.VirtualMachineConfig != nil && !strings.HasPrefix(machine.VirtualMachineConfig.Boot, "order=") {
		bootVal := fmt.Sprintf("order=scsi0;%s", d.ISODeviceName)
		_ = machine.Config(ctx, proxmox.VirtualMachineConfigOption{Name: "boot", Value: bootVal})
	}

	if err := machine.CloudInit(ctx, d.ISODeviceName, cloudinitUserdata, cloudinitMetadata, "", ""); err != nil {
		return fmt.Errorf("failed to configure cloud-init for Proxmox VE virtual machine ID='%d': %w", machine.VMID, err)
	}

	return nil
}

// Blocks until QEMU Guest Agent is up and running on the machine.
func (d *Driver) waitForQemuAgent(ctx context.Context) error {
	log.Infof("Waiting for QEMU guest agent to respond on VM '%s'...", d.MachineName)
	pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for {
		machine, err := d.getCurrentMachine(pollCtx)
		if err == nil {
			_, err = machine.AgentOsInfo(pollCtx)
			if err == nil {
				log.Infof("QEMU guest agent is active on VM '%s'", d.MachineName)
				return nil
			}
		}

		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timed out waiting for QEMU guest agent on VM '%s': %w", d.MachineName, pollCtx.Err())
		case <-time.After(3 * time.Second):
			continue
		}
	}
}

// Blocks until cloud-init finishes setup on the current machine.
func (d *Driver) waitForCloudinit() error {
	log.Infof("Waiting for cloud-init to finish on VM '%s'...", d.MachineName)
	ctx, cancel := context.WithTimeout(context.TODO(), pveTaskPollingTimeout)
	defer cancel()

	for {
		err := d.runCommandOnCurrentMachine("sudo cloud-init status --wait")
		if err == nil {
			log.Infof("Cloud-init completed successfully on VM '%s'", d.MachineName)
			return nil
		}

		if errors.Is(err, ErrNonZeroExitCode) {
			return fmt.Errorf("cloud-init finished with non-zero exit code on VM '%s': %w", d.MachineName, err)
		}

		log.Debugf("waiting for SSH execution of 'sudo cloud-init status --wait' on VM '%s': %v", d.MachineName, err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for cloud-init to finish on VM '%s': %w", d.MachineName, context.DeadlineExceeded)
		case <-time.After(pveTaskPollingInterval):
			continue
		}
	}
}

// Removes cloud-init configuration from the current machine.
func (d *Driver) cleanupCloudinit(ctx context.Context) error {
	machine, err := d.getCurrentMachine(ctx)
	if err != nil {
		return err
	}

	if err := machine.UnmountCloudInitISO(ctx, d.ISODeviceName); err != nil {
		log.Warnf("Failed to unmount cloud-init ISO: %v", err)
	}

	err = d.runTaskOnCurrentMachine(ctx, func(ctx context.Context, vm *proxmox.VirtualMachine) (*proxmox.Task, error) {
		return vm.RemoveTag(ctx, proxmox.MakeTag(proxmox.TagCloudInit))
	})
	if err != nil {
		log.Warnf("Failed to remove cloud-init tag: %v", err)
	}

	return nil
}

// Generates cloud-init metadatadata for the current machine.
func (d *Driver) generateCloudinitMetadata() (string, error) {
	metadata := map[string]interface{}{
		"instance-id": d.MachineName,
		"hostname":    d.MachineName,
	}

	metadataYAML, err := yaml.Marshal(&metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cloud-init metadata: %w", err)
	}

	return string(metadataYAML), nil
}

// Generates cloud-init userdata for the current machine.
func (d *Driver) generateCloudinitUserdata() (string, error) {
	sshPublicKey, err := os.ReadFile(d.GetSSHPublicKeyPath())
	if err != nil {
		return "", fmt.Errorf("failed to read machine's SSH public key: %w", err)
	}

	userMap := map[string]interface{}{
		"name":        d.SSHUser,
		"lock_passwd": true,
		"sudo":        "ALL=(ALL) NOPASSWD:ALL",
		"ssh_authorized_keys": []string{
			string(sshPublicKey),
		},
	}

	if d.CIPassword != "" {
		userMap["passwd"] = d.CIPassword
		userMap["lock_passwd"] = false
	}

	userdata := map[string]interface{}{
		"hostname":             d.MachineName,
		"preserve_hostname":    false,
		"create_hostname_file": true,
		"users": []map[string]interface{}{
			userMap,
		},
		"runcmd": []interface{}{
			"systemctl enable --now qemu-guest-agent || true",
		},
	}

	if d.Nameserver != "" {
		userdata["nameservers"] = map[string]interface{}{
			"addresses": []string{d.Nameserver},
		}
	}

	if d.Searchdomain != "" {
		userdata["searchdomains"] = []string{d.Searchdomain}
	}

	userdataYAML, err := yaml.Marshal(&userdata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cloud-init userdata: %w", err)
	}

	return fmt.Sprintf("#cloud-config\n%s", userdataYAML), nil
}
