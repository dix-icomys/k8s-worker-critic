package main

import (
  "fmt"
  "os"
  "net"
  "strconv"
  "strings"
  "flag"
  "time"
  "regexp"
  log "github.com/golang/glog"
  "k8s.io/client-go/kubernetes"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

  kubernetes_rest "k8s.io/client-go/rest"
  // "k8s.io/client-go/tools/clientcmd"
)

var interval = flag.Int("interval", 360, "Timeout in seconds")
var interface_regexp = flag.String("interface_regexp", ".*?", "Interface regexp")
var label_prefix = flag.String("label_prefix", "interface", "Label prefix")

var clientset *kubernetes.Clientset

func main() {
  flag.Parse()

  clientset = getKubernetesClientSet()

  for {
    generateNetworkLabels()

    // Sleep until interval
    fmt.Println("Sleeping for", *interval , "seconds")
    time.Sleep(time.Duration(*interval) * time.Second)
  }
}

func getKubernetesClientSet() *kubernetes.Clientset {
  config, err := kubernetes_rest.InClusterConfig()
  // config, err := clientcmd.BuildConfigFromFlags("", "PATH to config")
  if err != nil {
    log.Fatalf(err.Error())
  }
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    log.Fatalf(err.Error())
  }

  return clientset
}

func generateNetworkLabels() {
  interfaces, _ := net.Interfaces()

  for _, inter := range interfaces {
    match, _ := regexp.MatchString(*interface_regexp, inter.Name)

    if match {
      if addrs, err := inter.Addrs(); err == nil {
        for i, addr := range addrs {
          var interface_name = inter.Name
          var interface_addr, _, _ = net.ParseCIDR(addr.String())

          // Allow only IPv4
          if (strings.Contains(interface_addr.String(), ".")) {
            labels := []string{*label_prefix, ".", interface_name, ".", strconv.Itoa(i)};

            addLabel(strings.Join(labels, ""), interface_addr.String())
          }
        }
      }
    }
  }
}

func addLabel(label_key, label_value string) {
  node, err := clientset.CoreV1().Nodes().Get(os.Getenv("HOSTNAME"), metav1.GetOptions{})
  if err != nil {
    log.Fatalf(err.Error())
  }
  fmt.Println("Add Label:", label_key, "with '", label_value, "' to", node.Name)

  labels := node.GetLabels()
  labels[label_key] = label_value

  _, err = clientset.CoreV1().Nodes().Update(node)
  if err != nil {
    log.Error(err)
  }
}
