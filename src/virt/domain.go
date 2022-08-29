package virt

import (
    "log"

	libvirt "libvirt.org/libvirt-go"
	libvirtxml "libvirt.org/libvirt-go-xml"
)


type VM struct {
    Domain          *libvirt.Domain
    Name            string
    Status          bool
}

type Diskinfo struct {
    Name            string
    Capacity        uint64
    Allocation      uint64
}

func GetCPUUsage(d *libvirt.Domain) (uint64, int) {
    cpuGuest, err := d.GetVcpus()
    if err != nil {
        log.Fatalf("failed to get cpu status: %v", err)
    }
    var all uint64
    all = 0
    cnt := 0
    for _, c := range cpuGuest {
        all += c.CpuTime
        cnt++
    }

    return all, cnt
}

func GetMemUsed(d *libvirt.Domain) (max, used uint64) {
    domMemStatus, err := d.MemoryStats(13, 0)
    if err != nil {
        log.Fatalf("failed to get memory: %v", err)
    }

    memStatus := make(map[int]uint64)
    for i := 0; i < 13; i++ {
        memStatus[i] = 0
    }

    for _, status := range domMemStatus {
        memStatus[int(status.Tag)] = status.Val
    }

    max, _ = memStatus[5]
    u, _ := memStatus[8]
    used = max - u

    return
}

func GetNICStatus(d *libvirt.Domain) (txByte, rxByte int64) {
    xml, err := d.GetXMLDesc(0)
    if err != nil {
        log.Fatalf("failed to open xml: %v", err)
    }
    var xmlDomain libvirtxml.Domain
    xmlDomain.Unmarshal(xml)

    /*
    I'm still trying to figure out how to display it, so right now it's in VM and 
    there are multiple Returns the last state of the NIC.
    (This is a very bad implementation and will be fixed as soon as possible.)
    */
    var mac string
    for _, iface := range xmlDomain.Devices.Interfaces {
        mac = iface.MAC.Address
    }
    ifState, err := d.InterfaceStats(mac)
    if err != nil {
        log.Fatalf("failed to get iface state: %v", err)
    }

    return ifState.TxBytes, ifState.RxBytes
}

func GetDisks(d *libvirt.Domain) []Diskinfo {
    xml, err := d.GetXMLDesc(0)
    if err != nil {
        log.Fatalf("failed to open xml: %v", err)
    }
    var xmlDomain libvirtxml.Domain
    xmlDomain.Unmarshal(xml)

    names := []string{}
    for _, disk := range xmlDomain.Devices.Disks {
        if disk.Device == "disk" {
            names = append(names, disk.Source.File.File)
        }
    }
    infos := []Diskinfo{}
    for _, name := range names {
        info, err := d.GetBlockInfo(name, 0)
        if err != nil {
            log.Fatalf("failed to get disk status: %v", err)
        }
        infos = append(infos, Diskinfo{
            Name:           name,
            Allocation:     info.Allocation,
            Capacity:       info.Capacity,
        })
    }

    return infos
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
