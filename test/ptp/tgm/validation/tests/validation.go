package tests

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/devices"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
)

var _ = Describe("PTP T-GM Validation", func() {
	var testPort string
	testPort, ok := os.LookupEnv("TGM_TESTING_PORT")
	if !ok {
		testPort = consts.DefaultTGMTestingPort
	}

	client := client.New("")

	Context("Configuration Check", func() {
		var ptpRunningPods []*corev1.Pod

		BeforeEach(func() {
			var err error
			ptpRunningPods, err = pods.GetPTPDaemonPods(client)
			Expect(err).NotTo(HaveOccurred())

			vendor, device, _, err := devices.GetDevInfo(client, testPort, ptpRunningPods[0])
			Expect(err).NotTo(HaveOccurred())
			if vendor != "0x8086" && device != "0x1593" {
				Skip("NIC is not a WPC")
			}
		})

		It("Should have the desired firmware version", func() {
			commands := []string{
				"/bin/sh", "-c", "ethtool -i " + testPort,
			}

			buf, err := pods.ExecCommand(client, ptpRunningPods[0], pkg.PtpContainerName, commands)
			outstring := buf.String()
			Expect(err).NotTo(HaveOccurred(), "Error to find device info due to %s", outstring)
			Expect(outstring).To(Not(BeEmpty()))

			By(fmt.Sprintf("checking the firmware version equals or greater than %.2f", consts.ICEDriverFirmwareVerMinVersion))
			scanner := bufio.NewScanner(strings.NewReader(outstring))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "firmware-version:") {
					firmwareVerArr := strings.Split(line, ":")[1]
					firmwareVer := strings.Split(firmwareVerArr, " ")[1]
					Expect(strconv.ParseFloat(firmwareVer, 64)).To(BeNumerically(">=", consts.ICEDriverFirmwareVerMinVersion), "linuxptp-daemon is not deployed on cluster")
				}
			}
		})
	})
})
