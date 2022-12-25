package devices

import (
	"fmt"
	"strings"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func GetDevInfo(api *client.ClientSet, intf string, ptpPod *corev1.Pod) (string, string, string, error) {
	var commands []string
	var outstring string

	commands = []string{
		"readlink", "/sys/class/net/" + intf + "/device",
	}
	buf, err := pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf.String()
	if err != nil {
		return "", "", "", fmt.Errorf("Error to get device info due to %s", outstring)
	}
	s := strings.Split(strings.TrimSpace(outstring), "/")
	busid := s[len(s)-1]
	logrus.Infof("busid: %s", busid)

	parts := strings.Split(busid, ":")
	ttyGNSS := parts[1] + strings.Split(parts[2], ".")[0]
	ttyGNSS = "/dev/ttyGNSS_" + ttyGNSS + "_0"
	commands = []string{
		"cat", "/sys/class/net/" + intf + "/device/device",
	}
	buf, err = pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf.String()
	if err != nil {
		return "", "", "", fmt.Errorf("Error to get device info due to %s", outstring)
	}
	device := strings.TrimSpace(outstring)
	commands = []string{
		"cat", "/sys/class/net/" + intf + "/device/vendor",
	}

	buf, err = pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf.String()
	if err != nil {
		return "", "", "", fmt.Errorf("Error to get device info due to %s", outstring)
	}
	vendor := strings.TrimSpace(outstring)
	logrus.Infof("vendor: %s, device ID: %s, ttyGNSS: %s", vendor, device, ttyGNSS)

	return vendor, device, ttyGNSS, nil
}

func GetBusID(api *client.ClientSet, intf string, ptpPod *corev1.Pod) (string, error) {
	command := []string{
		"readlink", "/sys/class/net/" + intf + "/device",
	}

	buf, err := pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, command)
	outstring := buf.String()
	if err != nil {
		return "", fmt.Errorf("Error to get device info due to %s", outstring)
	}

	s := strings.Split(strings.TrimSpace(outstring), "/")
	busid := s[len(s)-1]

	return busid, nil
}
