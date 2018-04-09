## Add network interface labels for Kubernetes node

Assignes:

- EC2 Instance tags
- network interface labels

### Usage

```
Usage of critic:
  -interface_regexp string
      Interface regexp (default ".*?")
  -interval int
      Timeout in seconds (default 360)
  -label_prefix string
      Label prefix (default "interface")
```

## Related Works

I have been influenced by the following great works:

- labelgun: https://github.com/Vungle/labelgun
- node-feature-discovery: https://github.com/kubernetes-incubator/node-feature-discovery
- kube-state-metrics: https://github.com/kubernetes/kube-state-metrics

### Build

`make all`

### Install

`helm install --name k8s-worker-critic chart/`

`helm upgrade k8s-worker-critic chart/`

`helm delete k8s-worker-critic`
