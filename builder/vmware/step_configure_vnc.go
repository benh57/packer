package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type stepConfigureVNC struct{}

func (stepConfigureVNC) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	f, err := os.Open(vmxPath)
	if err != nil {
		ui.Error(fmt.Sprintf("Error while reading VMX data: %s", err))
		return multistep.ActionHalt
	}

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		ui.Error(fmt.Sprintf("Error reading VMX data: %s", err))
		return multistep.ActionHalt
	}

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	log.Printf("Looking for available port between %d and %d", config.VNCPortMin, config.VNCPortMax)
	var vncPort uint
	portRange := int(config.VNCPortMax - config.VNCPortMin)
	for {
		vncPort = uint(rand.Intn(portRange) + portRange)
		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	log.Printf("Found available VNC port: %d", vncPort)

	vmxData := ParseVMX(string(vmxBytes))
	vmxData["RemoteDisplay.vnc.enabled"] = "TRUE"
	vmxData["RemoteDisplay.vnc.port"] = fmt.Sprintf("%d", vncPort)

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		ui.Error(fmt.Sprintf("Error writing VMX data: %s", err))
		return multistep.ActionHalt
	}

	state["vnc_port"] = vncPort

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(map[string]interface{}) {
}