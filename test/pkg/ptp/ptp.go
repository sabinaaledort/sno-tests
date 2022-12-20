package ptp

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"

	ptpv1 "github.com/redhat-eets/sno-tests/api/v1"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/render"
	corev1 "k8s.io/api/core/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Timeout and Interval settings
	timeout                  = time.Minute * 1
	interval                 = time.Second * 5
	ptpConfigTemplatePath    = "../pkg/config/templates/ptp_config.yaml"
	ptpLeapFileConfigMapName = "ptp-leap-second-configmap"
	ptpLeapFileConfigMapPath = "../pkg/config/ptp_leap_seconds_cm.yaml"
)

// Delete and check the PtpConfig resource is deleted.
func Delete(client *client.ClientSet, ptpConfig *ptpv1.PtpConfig) error {
	err := client.Delete(context.Background(), ptpConfig)
	if errors.IsNotFound(err) { // Ignore err, could be already deleted.
		return nil
	}

	Eventually(func() bool {
		err := client.Get(context.Background(), goclient.ObjectKey{Namespace: ptpConfig.Namespace, Name: ptpConfig.Name}, ptpConfig)
		return errors.IsNotFound(err)
	}, timeout, interval).Should(BeTrue(), "Failed to delete PTPConfig resource")

	return nil
}

// Get PtpConfig resource from file.
func GetFromFile(filePath string) (*ptpv1.PtpConfig, error) {
	ptpConfig := &ptpv1.PtpConfig{}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = decodeYAML(f, ptpConfig)
	if err != nil {
		return nil, err
	}

	return ptpConfig, nil
}

// Get PtpConfig resource from config sample template.
func GetFromTemplate(node string, port string) (*ptpv1.PtpConfig, error) {
	data := render.MakeRenderData()
	data.Data["Node"] = node
	data.Data["Port"] = port

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	obj, err := render.RenderTemplate(ptpConfigTemplatePath, &data)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	ptpv1.AddToScheme(scheme)
	ptpConfig := &ptpv1.PtpConfig{}
	err = scheme.Convert(obj[0], ptpConfig, nil)
	if err != nil {
		return nil, err
	}

	return ptpConfig, nil
}

func ConfigurePTPLeapFileConfigMap(client *client.ClientSet) error {
	cm := &corev1.ConfigMap{}

	f, err := os.Open(ptpLeapFileConfigMapPath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = decodeYAML(f, cm)
	if err != nil {
		return err
	}

	err = client.Create(context.Background(), cm)
	if errors.IsAlreadyExists(err) {
		err = client.Update(context.Background(), cm)
	}
	if err != nil {
		return err
	}

	return nil
}

func decodeYAML(r io.Reader, obj interface{}) error {
	decoder := yaml.NewYAMLToJSONDecoder(r)
	return decoder.Decode(obj)
}
