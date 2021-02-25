package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/comail/colog"
	"github.com/zanghao2/k8s-controller/pkg"
	"github.com/zanghao2/k8s-controller/pkg/utils"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const RecommendedKubeConfigPathEnv = "KUBECONFIG"

var (
	clientset    *kubernetes.Clientset
	resyncPeriod = 30 * time.Second
	clientConfig clientcmd.ClientConfig
)

func initKubeClient() {
	kubeConfig := ""
	if len(os.Getenv(RecommendedKubeConfigPathEnv)) > 0 {
		// use the current context in kubeconfig
		// This is very useful for running locally.
		kubeConfig = os.Getenv(RecommendedKubeConfigPathEnv)
	}

	// Get kubernetes config.
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("fatal: Failed to init rest config due to %v", err)
	}
}

func main() {

	flag.CommandLine.Parse([]string{})

	colog.SetDefaultLevel(colog.LInfo)
	colog.SetMinLevel(colog.LInfo)
	colog.SetFormatter(&colog.StdFormatter{
		Colors: true,
		Flag:   log.Ldate | log.Ltime | log.Lshortfile,
	})

	colog.Register()

	level := StringToLevel(os.Getenv("LOG_LEVEL"))
	log.Print("Log level was set to ", strings.ToUpper(level.String()))
	colog.SetMinLevel(level)
	threadness := StringToInt(os.Getenv("THREADNESS"))

	//
	initKubeClient()
	stopCh := utils.SetupSignalHandler()

	informerFactory := kubeinformers.NewSharedInformerFactory(clientset, resyncPeriod)

	controller, err := pkg.NewController(clientset, informerFactory, stopCh)
	if err != nil {
		log.Fatal("Failed to start due to %v", err)
	}

	err = controller.Run(threadness, stopCh)
	if err != nil {
		log.Fatal("controller run error %v", err)
	}

}

// StringToLevel set level
func StringToLevel(levelStr string) colog.Level {
	switch level := strings.ToUpper(levelStr); level {
	case "TRACE":
		return colog.LTrace
	case "DEBUG":
		return colog.LDebug
	case "INFO":
		return colog.LInfo
	case "WARNING":
		return colog.LWarning
	case "ERROR":
		return colog.LError
	case "ALERT":
		return colog.LAlert
	default:
		log.Printf("warning: LOG_LEVEL=\"%s\" is empty or invalid, fallling back to \"INFO\".\n", level)
		return colog.LInfo
	}
}

func StringToInt(sThread string) int {
	thread := 1

	return thread
}
