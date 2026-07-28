package main

import (
	"github.com/rancher/machine/libmachine/drivers/plugin"
	"github.com/stellatarum/docker-machine-driver-pve/cmd/docker-machine-driver-pve/driver"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}
