package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/google/go-querystring/query"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	K8SUpgradeDeployment     = "http://localhost:8080/openapi/v1/gateway/action/upgrade_deployment"
	K8SCheckDeploymentStatus = "http://localhost:8080/openapi/v1/gateway/action/get_deployment_status"
)

type K8SConfig struct {
	ApiKey      string   `url:"apikey"`      // 发布密钥
	Namespace   string   `url:"namespace"`   // 所在组
	Deployment  string   `url:"deployment"`  // 要发布的"部署"名称
	Publish     bool     `url:"publish"`     // 是否直接发布上线
	Images      []string `url:"-"`           // 发布的镜像地址
	Clusters    []string `url:"-"`           // 发布的机房
	Description string   `url:"description"` // 发布信息描述
}

type DeployConfig struct {
	Namespace   string       `json:"namespace"`
	ApiKey      string       `json:"-"` // api_key is set in gitlab ci variables
	AppUrl      string       `json:"app_url"`
	Deployments []Deployment `json:"deployments"`
}

type Deployment struct {
	Name     string   `json:"name"`
	Publish  bool     `json:"publish"`
	Clusters []string `json:"clusters"`
	Images   []string `json:"images"`
}

var (
	cfgFile string

	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy to K8S",

		Run: func(cmd *cobra.Command, args []string) {
			config := viper.Get("k8s")
			var deployConfig DeployConfig
			err := mapstructure.Decode(config, &deployConfig)
			if err != nil {
				panic(err)
			}

			description := GetCommitsBetweenTags()
			apiKey := os.Getenv("K8S_API_KEY")
			if apiKey == "" {
				er("failed: K8S_API_KEY os env is needed")
			}

			k8sConfig := K8SConfig{
				ApiKey:      apiKey,
				Namespace:   deployConfig.Namespace,
				Description: description,
			}

			latestTag := GetLatestTag()
			// deploy new images
			for _, deployment := range deployConfig.Deployments {
				k8sConfig.Deployment = deployment.Name
				k8sConfig.Publish = deployment.Publish
				queryStr, _ := query.Values(k8sConfig)
				var images []string
				for _, image := range deployment.Images {
					images = append(images, fmt.Sprintf("%s:%s", image, latestTag))
				}
				clusters := strings.Join(deployment.Clusters, ",")
				api := fmt.Sprintf("%s?%s&images=%s&cluster=%s", K8SUpgradeDeployment, queryStr.Encode(), url.QueryEscape(strings.Join(images, ",")), url.QueryEscape(clusters))
				body, err := httplib.Get(api).Bytes()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"msg": "failed: call /upgrade_deployment api",
					}).Error(err)
				}
				var data struct {
					Code   int      `json:"code"`
					Errors []string `json:"errors,omitempty"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					logrus.WithFields(logrus.Fields{
						"msg": "failed: json decode error",
					}).Error(err)
				}
				if data.Errors != nil {
					logrus.WithFields(logrus.Fields{
						"msg": "failed: call /upgrade_deployment api",
					}).Error(data.Errors)
				}
			}

			// check deployment status
			for _, deployment := range deployConfig.Deployments {
				if !deployment.Publish {
					continue
				}

				for _, cluster := range deployment.Clusters {

					timeout := time.After(5 * time.Minute)
					for {
						success := false
						select {
						case <-timeout:
							fmt.Println("")
							os.Exit(1)
						case <-time.Tick(10 * time.Second):
							// check status every 10s
							type StatusConfig struct {
								ApiKey     string `json:"api_key"`
								Cluster    string `json:"cluster"`
								Namespace  string `json:"namespace"`
								Deployment string `json:"deployment"`
							}
							statusConfig := StatusConfig{
								ApiKey:     apiKey,
								Cluster:    cluster,
								Namespace:  deployConfig.Namespace,
								Deployment: deployment.Name,
							}
							queryStr, _ := query.Values(statusConfig)
							api := fmt.Sprintf("%s?%s", K8SCheckDeploymentStatus, queryStr.Encode())
							response, err := httplib.Get(api).Response()
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"msg": "failed: call /upgrade_deployment api",
								}).Error(err)
							}
							if response.StatusCode == http.StatusOK {
								body, _ := ioutil.ReadAll(response.Body)
								var healthz struct {
									Healthz bool `json:"healthz"` // only focus `healthz` field
								}
								if err := json.Unmarshal(body, &healthz); err != nil {
									logrus.WithFields(logrus.Fields{
										"msg": "failed: json decode error",
									}).Error(err)
								}
								success = healthz.Healthz
							}
						}

						if success {
							break
						}
					}
				}
			}
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	deployCmd.Flags().StringVar(&cfgFile, "config", "", "config file")
	_ = deployCmd.MarkFlagRequired("config")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func GetLatestTag() string {
	tag, err := exec.Command("/bin/sh", "-c", "git describe --tags $(git rev-list --tags --max-count=1 --no-walk)").Output()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(tag))
}

func GetCommitsBetweenTags() string {
	output1, _ := exec.Command("/bin/sh", "-c", "git describe --tags $(git rev-list --tags --max-count=1 --no-walk)").Output()
	tag1 := strings.TrimSpace(string(output1))
	output2, _ := exec.Command("/bin/sh", "-c", "git describe --tags $(git rev-list --tags --skip=1 --max-count=1 --no-walk)").Output()
	tag2 := strings.TrimSpace(string(output2))
	output, _ := exec.Command("/bin/sh", "-c", "git log --pretty=\"format:%h %aN [%s]\" "+tag1+"..."+tag2).Output()
	return strings.TrimSpace(string(output))
}
