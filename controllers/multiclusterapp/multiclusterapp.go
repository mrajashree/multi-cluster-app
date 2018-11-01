package multiclusterapp

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	//log "github.com/sirupsen/logrus"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	multiClusterAppController = "mgmt-multicluster-app-controller"
	helmName                  = "helm"
	tillerName                = "tiller"
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

	// get the Helm Chart contents into a temp dir
	if app.Spec.ChartRepositoryURL == "" {
		return nil
	}

	for _, target := range app.Spec.Targets {
		config := target.Spec.ClusterConfig
		kubeconfigPath, err := m.writeKubeConfig(config, app.Namespace)
		if err != nil {
			return err
		}

		cont, cancel := context.WithCancel(context.Background())
		defer cancel()
		addr := generateRandomPort()
		probeAddr := generateRandomPort()
		fmt.Printf("\nkubeconfigPath: %v\n", kubeconfigPath)
		go m.StartTiller(cont, addr, probeAddr, app.Spec.ReleaseNamespace, kubeconfigPath)
		err = m.InstallChart(app, target, addr)
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

	tempDir, err := ioutil.TempDir("", "kubeconfig-")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Join(tempDir, ""), 0755); err != nil {
		return "", err
	}
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")
	if err := clientcmd.WriteToFile(*kubeConfig, kubeConfigPath); err != nil {
		return "", err
	}
	return kubeConfigPath, nil
}

func (m MCAppController) InstallChart(app *v3.MultiClusterApp, target v3.Target, port string) error {
	var installCommand []string
	var initCommand []string
	initCommand = []string{"init", "--client-only"}
	cm1 := exec.Command(helmName, initCommand...)
	cm1.Env = []string{fmt.Sprintf("%s=%s", "HELM_HOME", ".helm")}
	stderrBuf := &bytes.Buffer{}
	cm1.Stdout = os.Stdout
	cm1.Stderr = stderrBuf
	err := cm1.Run()
	if err != nil {
		return fmt.Errorf("Error: %v, %v", err, stderrBuf.String())
	}
	installCommand = append([]string{"install", "--debug", "--repo", app.Spec.ChartRepositoryURL, app.Spec.ChartReference})

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
	cmd.Env = []string{fmt.Sprintf("%s=%s", "HELM_HOST", "127.0.0.1:"+port)}
	stderrBuf = &bytes.Buffer{}
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderrBuf
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to install app%s", stderrBuf.String())
	}
	if err := cmd.Wait(); err != nil {
		// if the first install failed, the second install would have error message like `has no deployed releases`, then the
		// original error is masked. We need to filter out the message and always return the last one if error matches this pattern
		if strings.Contains(stderrBuf.String(), "has no deployed releases") {
			//return errors.New(v3.AppConditionInstalled.GetMessage(obj))
		}
		return fmt.Errorf("failed to install app: %v, %v", err, stderrBuf.String())
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
