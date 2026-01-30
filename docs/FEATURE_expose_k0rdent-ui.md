# Exposing K0rdent UI Feature

## Feature Description

`K0rdentd` should have a `expose-ui` command that does the following:
- check if the Kubernetes deployment `k0rdent-k0rdent-ui` in the `kcm-system` namespace is ready. If it is not ready, wait and repeat the check. Put a configurable timeout which defaults to 5min. 
- check if the Kubernetes Service `k0rdent-k0rdent-ui` in the `kcm-system` namespace is present.
- create an ingress object to expose that service on the endpoint `/k0rdent-ui` 
- look for the VM's IP addresses. Also, the VM on which `k0rdentd` might run could be a cloud instance, might have an external IP address that is not easily visible. The program should check if the VM is running on a cloud (maybe check AWS, GCP and Azure), and depending on the result, extract the `external IP` from a metadata service endpoint or some other way.
- In the final IP list, ignore all Calico interfaces (`caliXXX` or `vxlan.calico`) as well as potential docker interfaces. You can keep adresses created by VPNs, such as tailscale or others.
- test if accessing the k0rdent UI using `/k0rdent-ui` path on the external IP gives a valid HTML document.
- output the possible links to k0rdent UI using all the IPs found. That's because the end user might be communicating with the VM on any of these interfaces.

The `expose-ui` process (implemented as func `ExposeUI()` in `./pkg/ui/ui.go`) should also be implemented after installing `k0rdent` in the `install` command.

## Verification

The following details just some ways to find out information:
- Internal IP can be found using `ip a` and ignoring all interfaces such as `caliXXX` or `vxlan.calico`.
- Cloud provider might recognized using content from `dmidecode -s bios-vendor`. 
- Use additional verifications that you might know about. For instance, that Amazon EC2 instances, would have a metadata endpoint available on `http://169.254.169.254/latest/meta-data/`.
- Go ahead and implement ways to find the external IP, based on what the cloud provider might be.

## Error handling
- If anything goes wrong during the search for an external IP, ignore the error and continue
- Only if absolutely no IP address was extracted, including using `ip a`, then throw a warning, followed by the command the end user should use to do port-forwarding to the `k0rdent-ui`.

