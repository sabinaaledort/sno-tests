package ptp_test

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/redhat-eets/sno-tests/api/v1"
	client "github.com/redhat-eets/sno-tests/test/pkg/client"
	ptputil "github.com/redhat-eets/sno-tests/test/pkg/ptp"
	"gopkg.in/yaml.v3"
)

const (
	defaultCfgDir = "/testconfig"
	cfgFile       = "topology.yaml"
)

var (
	cfgdir    string
	topo      Topology
	clients   map[string]*client.ClientSet
	nodenames map[string]string
	roles     = [...]string{"GM", "Tester"}
)

type Topology struct {
	PTP *struct {
		GM *struct {
			Node       string  `yaml:"node"`
			PortTester *string `yaml:"toTester,omitempty"`
		} `yaml:"gm,omitempty"`
		Tester *struct {
			Node   string `yaml:"node"`
			PortGM string `yaml:"toGM"`
		} `yaml:"tester,omitempty"`
	} `yaml:"ptp,omitempty"`
}

func TestPtp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ptp Suite")
}

var ptpConfigGM *v1.PtpConfig
var ptpConfigTester *v1.PtpConfig

var _ = BeforeSuite(func() {
	// Get the config file location from enviroment
	val, ok := os.LookupEnv("CFG_DIR")
	if !ok {
		cfgdir = defaultCfgDir
	} else {
		cfgdir = val
	}

	cfg := cfgdir + "/" + cfgFile
	yfile, err := os.ReadFile(cfg)
	Expect(err).NotTo(HaveOccurred())

	Expect(yaml.Unmarshal(yfile, &topo)).To(Succeed())

	// Skip this test suite if no ptp topology is specified
	if topo.PTP == nil {
		Skip("PTP test suite not requested.")
	}

	// Setup k8s api client for each cluster under PTP section
	clients = make(map[string]*client.ClientSet)
	// clusters can be shared by multiple nodes, where clients is per node
	clusters := make(map[string]*client.ClientSet)
	nodenames = make(map[string]string)
	info := make(map[string][]string)
	if topo.PTP.GM != nil && topo.PTP.GM.Node != "" {
		info["GM"] = strings.Split(topo.PTP.GM.Node, "/")
	}

	if topo.PTP.Tester != nil && topo.PTP.Tester.Node != "" {
		info["Tester"] = strings.Split(topo.PTP.Tester.Node, "/")
	}

	for _, role := range roles {
		if _, ok := info[role]; !ok {
			clients[role] = nil
			nodenames[role] = ""
			continue
		}

		if len(info[role]) > 1 {
			nodenames[role] = info[role][1]
		} else {
			nodenames[role] = ""
		}
		kubecfg := cfgdir + "/" + info[role][0]
		if _, ok := clusters[kubecfg]; !ok {
			clusters[kubecfg] = client.New(kubecfg)
		}
		clients[role] = clusters[kubecfg]
	}

	ptpConfigGM, err = ptputil.GetFromTemplate(topo.PTP.GM.Node, *topo.PTP.GM.PortTester)
	Expect(err).NotTo(HaveOccurred())

	ptpConfigTester = ptpConfigGM

	if topo.PTP.GM != nil && topo.PTP.GM.Node != "" {
		err := configurePTP(clients["GM"], ptpConfigGM, "PTP_GM_CONFIG_FILE")
		Expect(err).NotTo(HaveOccurred())
	}

	if topo.PTP.Tester != nil && topo.PTP.Tester.Node != "" {
		err := configurePTP(clients["Tester"], ptpConfigTester, "PTP_TESTER_CONFIG_FILE")
		Expect(err).NotTo(HaveOccurred())
	}

	mapFirstElement := reflect.ValueOf(clients).MapKeys()[0]
	client := clients[mapFirstElement.String()]
	err = ptputil.ConfigurePTPLeapFileConfigMap(client)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if topo.PTP.GM != nil && topo.PTP.GM.Node != "" {
		err := deletePTPConfig(clients["GM"], ptpConfigGM)
		Expect(err).NotTo(HaveOccurred())
	}

	if topo.PTP.Tester != nil && topo.PTP.Tester.Node != "" {
		err := deletePTPConfig(clients["Tester"], ptpConfigTester)
		Expect(err).NotTo(HaveOccurred())
	}
})

func configurePTP(client *client.ClientSet, ptpConfig *v1.PtpConfig, configFile string) error {
	var err error
	if ptpConfigFile := os.Getenv(configFile); len(ptpConfigFile) != 0 {
		ptpConfig, err = ptputil.GetFromFile(ptpConfigFile)
		if err != nil {
			return err
		}
	}

	err = client.Create(context.Background(), ptpConfig)
	if err != nil {
		return err
	}

	return nil
}

func deletePTPConfig(client *client.ClientSet, ptpConfig *v1.PtpConfig) error {
	err := ptputil.Delete(client, ptpConfig)
	if err != nil {
		return err
	}

	return nil
}
