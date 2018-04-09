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

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/ec2metadata"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/ec2"
)

var interval = flag.Int("interval", 360, "Timeout in seconds")
var interface_regexp = flag.String("interface_regexp", ".*?", "Interface regexp")
var label_prefix = flag.String("label_prefix", "node", "Label prefix")

var clientset *kubernetes.Clientset

var ec2_session = session.New()
var ec2_svc = ec2.New(ec2_session, &aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))})
var ec2_metadata = ec2metadata.New(ec2_session)

func main() {
  flag.Parse()

  clientset = getKubernetesClientSet()

  for {
    generateNetworkLabels()
    generateEC2instanceTagsLabels()

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
            labels := []string{*label_prefix, ".", "network.interface", ".", interface_name, ".", strconv.Itoa(i)};

            addLabel(strings.Join(labels, ""), interface_addr.String())
          }
        }
      }
    }
  }
}

func generateEC2instanceTagsLabels() {
  meta_document, err := ec2_metadata.GetInstanceIdentityDocument()

  if err != nil {
    log.Fatalf(err.Error())
  }

  instance_id := meta_document.InstanceID
  fmt.Println("Instance ID:", string(instance_id))

  params := &ec2.DescribeTagsInput{
    Filters: []*ec2.Filter{
      &ec2.Filter{
        Name: aws.String("resource-id"),
        Values: []*string{
          aws.String(string(instance_id)),
        },
      },
      &ec2.Filter{
        Name: aws.String("resource-type"),
        Values: []*string{
          aws.String("instance"),
        },
      },
    },
  }
  response, err := ec2_svc.DescribeTags(params)

  if err != nil {
    log.Fatalf(err.Error())
  }

  for _, tag := range response.Tags {
    tag_key := strings.ToLower(*tag.Key)
    tag_key = strings.Replace(tag_key, ":", ".", -1)
    tag_key = strings.Replace(tag_key, "-", "", -1)

    labels := []string{*label_prefix, ".", "ec2.tag", ".", tag_key};

    addLabel(strings.Join(labels, ""), *tag.Value)
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
