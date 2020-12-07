package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	kubeconfig string
	configfile string
	healthIp   string
	healthPort uint

	startTime time.Time
	pods      = map[string]time.Time{}
	mutex     = sync.Mutex{}
)

const (
	TerminatedReasonOOMKilled = "OOMKilled"
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "k8s config path")
	flag.StringVar(&configfile, "configfile", "/root/monitor.yaml", "k8s config path")
	flag.StringVar(&healthIp, "healthIp", "0.0.0.0", "health check ip")
	flag.UintVar(&healthPort, "healthPort", 8082, "health check port")
}

type PodUpdatedGroup struct {
	OldPod *corev1.Pod
	NewPod *corev1.Pod
}
type NodeUpdatedGroup struct {
	OldNode *corev1.Node
	NewNode *corev1.Node
}

// 该脚本watch pod内的container被OOM后，发送通知
// go run . --kubeconfig=/Users/liuxiang/.kube/shbt.kubeconfig.yml
// go run . --kubeconfig=/Users/liuxiang/.kube/minikube.yml
func main() {
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	viper.AutomaticEnv()
	viper.SetConfigFile(configfile)
	if err := viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"errMsg": fmt.Sprintf("failed to read config file: %v", err),
		})
		os.Exit(1)
	}

	//startTime = time.Now()

	//stopChannel := make(chan struct{})
	//InstallSignalHandler(stopChannnel)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	sourceManager, err := sources.NewManager(kubeconfig)
	if err != nil {

	}

	handlerManager, err := handlers.NewManager(configfile)
	if err != nil {

	}

	controller.Start(sourceManager, handlerManager)

	// 方便添加 liveness probe
	go startHealthServer()

	<-sig

	/*for {
		select {
		case nodeAdded := <-nodeAddedCh:
			go handleNodeAdded(nodeAdded)
		case nodeUpdatedGroup := <-nodeUpdatedCh:
			go handleNodeUpdatedGroup(nodeUpdatedGroup)
		case podAdded := <-podAddedCh:
			go handlePodAdded(podAdded)
		case podUpdatedGroup := <-podUpdatedCh:
			go handlePodUpdatedGroup(podUpdatedGroup)
		case <-stopChannel:
			log.WithFields(log.Fields{
				"msg": "stopping...",
			}).Info("[quit]")
			return
		}
	}*/
}

func startHealthServer() {
	log.Error(http.ListenAndServe(net.JoinHostPort(healthIp, strconv.Itoa(int(healthPort))), nil))
}

func handleNodeAdded(node *corev1.Node) {
	log.WithFields(log.Fields{
		"msg": node.Name,
	}).Info("[handleNodeAdded]")
	nodeCollectAndNotifyMessage(node)
}

func handleNodeUpdatedGroup(nodeGroup *NodeUpdatedGroup) {
	if reflect.DeepEqual(nodeGroup.OldNode, nodeGroup.OldNode) {
		return
	}

	node := nodeGroup.NewNode
	log.WithFields(log.Fields{
		"msg": node.Name,
	}).Info("[handleNodeUpdatedGroup]")
	nodeCollectAndNotifyMessage(node)
}

func nodeCollectAndNotifyMessage(node *corev1.Node) {
	conditions := node.Status.Conditions
	bad := false

	for _, condition := range conditions {
		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				bad = true
			}
		case corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodePIDPressure, corev1.NodeNetworkUnavailable:
			if condition.Status == corev1.ConditionTrue || condition.Status == corev1.ConditionUnknown {
				bad = true
			}
		default:
		}
	}

	if bad {
		conditions := node.Status.Conditions
		var message string
		for _, condition := range conditions {
			message += fmt.Sprintf("type:%s, status:%s, reason:%s\t", condition.Type, condition.Status, condition.Reason)
		}

		message = fmt.Sprintf("[%s]:%s", node.Name, message)

		if len(viper.GetString("NOTIFY_OPS_USERS")) != 0 {
			//_, _ = notify360Home(message, viper.GetString("NOTIFY_OPS_USERS"))
		}
	}
}

// 聚合事件，10分钟内只发一次通知
func filterPodByTime(pod *corev1.Pod) bool {
	mutex.Lock()
	defer mutex.Unlock()
	lastTime, ok := pods[pod.Name]
	if !ok {
		pods[pod.Name] = time.Now()
		return true
	}

	now := time.Now()
	if now.Sub(lastTime) > time.Minute*10 {
		pods[pod.Name] = now
		return true
	}

	return false
}

func handlePodAdded(pod *corev1.Pod) {
	if !filterPodByTime(pod) {
		return
	}

	collectAndNotifyMessage(pod)
}

func handlePodUpdatedGroup(group *PodUpdatedGroup) {
	if reflect.DeepEqual(group.OldPod, group.NewPod) {
		return
	}

	if !filterPodByTime(group.NewPod) {
		return
	}

	collectAndNotifyMessage(group.NewPod)
}

func collectAndNotifyMessage(pod *corev1.Pod) {
	var messages []string
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.LastTerminationState.Terminated == nil ||
			containerStatus.LastTerminationState.Terminated.Reason != TerminatedReasonOOMKilled { // only OOM state
			continue
		}
		// only after this app started
		if containerStatus.LastTerminationState.Terminated.FinishedAt.Time.Before(startTime) {
			continue
		}

		messages = append(messages, fmt.Sprintf("container[%s] with image[%s] in pod[%s] was OOMKilled", containerStatus.Name, containerStatus.Image, pod.Name))
	}

	if len(messages) != 0 {
		message := ""
		message += strings.Join(messages, "\n")

		log.WithFields(log.Fields{
			"msg":    message,
			"action": "update",
		}).Info("[oom]")

		if len(viper.GetString("NOTIFY_USERS")) != 0 {
			//_, _ = notify360Home(message, viper.GetString("NOTIFY_USERS"))
		}
	}
}

func InstallSignalHandler(stop chan struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		stop <- struct{}{}
		close(stop)
	}()
}

type Message struct {
	ToUser  []string `json:"touser"`
	Content string   `json:"content"`
}

//func notify360Home(message string, users string) (string, error) {
//	skipProjects := strings.Split(viper.GetString("PREFIX_SKIP_PROJECT"), ",")
//	if len(skipProjects) != 0 {
//		log.WithFields(log.Fields{
//			"PREFIX_SKIP_PROJECT": viper.GetString("PREFIX_SKIP_PROJECT"),
//		}).Info("[PREFIX_SKIP_PROJECT]")
//	}
//	for _, skipProject := range skipProjects {
//		if strings.Contains(message, strings.TrimSpace(skipProject)) {
//			return "", nil
//		}
//	}
//
//	sent := false
//	projects := strings.Split(viper.GetString("PREFIX_PROJECT"), ",")
//	for _, project := range projects {
//		if strings.Contains(message, strings.TrimSpace(project)) {
//			sent = true
//			continue
//		}
//	}
//
//	if !sent {
//		return "", nil
//	}
//
//	url := "http://10.202.5.72:20082/360home/send_custom_message?from=oom-monitor"
//	body := Message{
//		ToUser:  strings.Split(users, ","),
//		Content: viper.GetString("QIHOO_IDC") + ":\n" + message,
//	}
//	request, _ := httplib.Post(url).JSONBody(body)
//	request.SetBasicAuth("360home", "bd57affdeb559fd28592be2560b3bb78")
//	return request.String()
//}
