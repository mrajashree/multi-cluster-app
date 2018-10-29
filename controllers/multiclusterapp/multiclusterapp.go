package multiclusterapp

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	log "github.com/sirupsen/logrus"

	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	multiClusterAppController = "mgmt-multicluster-app-controller"
	helmName                  = "rancherhelm"
	tillerName                = "ranchertiller"
	base                      = 32768
	end                       = 61000
)

var httpTimeout = time.Second * 30

var httpClient = &http.Client{
	Timeout: httpTimeout,
}

type MCAppController struct {
	mcappClient v3.MultiClusterAppInterface
	mcappLister v3.MultiClusterAppLister
}

func newMultiClusterApp(mgmt *config.ManagementContext) *MCAppController {
	m := MCAppController{
		mcappClient: mgmt.Management.MultiClusterApps(""),
		mcappLister: mgmt.Management.MultiClusterApps("").Controller().Lister(),
	}

	return &m
}

func Register(ctx context.Context, management *config.ManagementContext) {
	m := newMultiClusterApp(management)
	management.Management.MultiClusterApps("").AddHandler(multiClusterAppController, m.sync)
}

func (m MCAppController) sync(key string, app *v3.MultiClusterApp) error {
	if app == nil || app.DeletionTimestamp != nil {
		return nil
	}

	log.Infof("\nApp 1: %v\n", app)

	// get the Helm Chart contents into a temp dir
	if app.Spec.ChartRepositoryURL == "" {
		return nil
	}

	log.Infof("\nApp: %#v\n", app)

	for _, target := range app.Spec.Targets {
		log.Infof("\nTarget: %#v\n", target)
		config := target.Spec.ClusterConfig
		kubeconfigPath, err := m.writeKubeConfig(config, app.Namespace)
		if err != nil {
			return err
		}

		cont, cancel := context.WithCancel(context.Background())
		defer cancel()
		addr := generateRandomPort()
		probeAddr := generateRandomPort()
		if app.Spec.ReleaseNamespace != "" {

		}
		go m.StartTiller(cont, addr, probeAddr, app.Spec.ReleaseNamespace, kubeconfigPath)
		err = m.InstallChart(app, target)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MCAppController) writeKubeConfig(config v3.ClusterConfig, namespace string) (string, error) {
	kubeConfig := &clientcmdapi.Config{
		CurrentContext: "default",
		APIVersion:     "v1",
		Kind:           "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			"default": {
				Server:                   config.Server,
				CertificateAuthorityData: config.CertificateAuthorityData,
				//CertificateAuthority: config.CertificateAuthorityPath,
				//InsecureSkipTLSVerify: true,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"default": {
				AuthInfo: "user",
				Cluster:  "default",
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"user": {
				ClientCertificateData: config.ClientCertificateData,
				ClientKeyData:         config.ClientKeyData,
				//Token:             config.TokenFile,
				//ClientCertificate: config.ClientCertificatePath,
				//ClientKey:         config.ClientKeyPath,
			},
		},
	}

	log.Infof("Kubeconfig: %v", kubeConfig)
	tempDir, err := ioutil.TempDir("", "kubeconfig-")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Join(tempDir, namespace), 0755); err != nil {
		return "", err
	}
	kubeConfigPath := filepath.Join(tempDir, namespace, ".kubeconfig")
	if err := clientcmd.WriteToFile(*kubeConfig, kubeConfigPath); err != nil {
		return "", err
	}
	return kubeConfigPath, nil
}

func (m MCAppController) InstallChart(app *v3.MultiClusterApp, target v3.Target) error {
	var installCommand []string
	installCommand = append([]string{"install", "--repo", app.Spec.ChartRepositoryURL, app.Spec.ChartReference})

	if app.Spec.ChartVersion != "" {
		installCommand = append(installCommand, []string{"--version", app.Spec.ChartVersion}...)
	}

	setValues := []string{}
	if target.Spec.Answers != nil {
		answers := target.Spec.Answers
		result := []string{}
		for k, v := range answers {
			result = append(result, fmt.Sprintf("%s=%s", k, v))
		}
		setValues = append([]string{"--set"}, strings.Join(result, ","))
	}
	commands := make([]string, 0)
	commands = append(installCommand, setValues...)
	cmd := exec.Command(helmName, commands...)
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderrBuf
	if err := cmd.Start(); err != nil {
		return err
		//return fmt.Errorf("failed to install app %s. %s", obj.Name, stderrBuf.String())
	}
	if err := cmd.Wait(); err != nil {
		// if the first install failed, the second install would have error message like `has no deployed releases`, then the
		// original error is masked. We need to filter out the message and always return the last one if error matches this pattern
		if strings.Contains(stderrBuf.String(), "has no deployed releases") {
			//return errors.New(v3.AppConditionInstalled.GetMessage(obj))
		}
		return fmt.Errorf("failed to install app: %v", err)
		//return errors.Errorf("failed to install app %s. %s", obj.Name, stderrBuf.String())
	}
	return nil
}

func (m *MCAppController) StartTiller(context context.Context, port, probePort, namespace, kubeConfigPath string) error {
	cmd := exec.Command(tillerName, "--listen", ":"+port, "--probe", ":"+probePort)
	cmd.Env = []string{fmt.Sprintf("%s=%s", "KUBECONFIG", kubeConfigPath)}
	if namespace != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "TILLER_NAMESPACE", namespace))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	select {
	case <-context.Done():
		return cmd.Process.Kill()
	}
}

func generateRandomPort() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for {
		port := base + r1.Intn(end-base+1)
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			continue
		}
		ln.Close()
		return strconv.Itoa(port)
	}
}
