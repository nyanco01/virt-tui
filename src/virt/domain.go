package virt

import (
    "log"

	libvirt "libvirt.org/libvirt-go"
)


type VM struct {
    Domain          *libvirt.Domain
    Name            string
    Status          bool
}


func LookupVMs(c *libvirt.Connect) []VM {
    vms := []VM{}
    domActive, err := c.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
    domInactive, err := c.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to domain list: %v", err)
    }

    for _, d := range domActive {
        n, err := d.GetName()
        if err != nil {
            log.Fatalf("failed to get domain name: %v", err)
        }
        tmpDomain := d
        vms = append(vms, VM{Domain: &tmpDomain,Name: n, Status: true})
    }
    for _, d := range domInactive {
        n, err := d.GetName()
        if err != nil {
            log.Fatalf("failed to get domain name: %v", err)
        }
        tmpDomain := d
        vm := VM{Domain: &tmpDomain,Name: n, Status: false}
        vms = append(vms, vm)
    }
    return vms
}
