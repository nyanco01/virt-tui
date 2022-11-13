package virt

import (
	"log"

	libvirt "libvirt.org/go/libvirt"
    libvirtxml "libvirt.org/go/libvirtxml"

	"github.com/nyanco01/virt-tui/src/operate"
)


type DomainIF struct {
    AttachVM        string
    MacAddr         string
    // If the VM is running, the interface is given a name
    Name            string
}



type NetworkInfo struct {
    Name        string
    Mode        string
    NetType     string
    Source      string
}





func GetDomIFListByBridgeName(con *libvirt.Connect, source string) []DomainIF {
    doms, err := con.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get domain list: %v", err)
    }

    var ifList []DomainIF

    for _, dom := range doms {
        defer dom.Free()
        xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
        if err != nil {
            log.Fatalf("failed to get domain xml: %v", err)
        }
        var domXML libvirtxml.Domain
        err = domXML.Unmarshal(xml)
        if err != nil {
            log.Fatalf("failed to unmarshal xml by domain: %v", err)
        }
        for _, iface := range domXML.Devices.Interfaces {
            if iface.Source.Bridge.Bridge == source {
                n := operate.GetIFNameByMAC(iface.MAC.Address)
                ifList = append(ifList, DomainIF{AttachVM: domXML.Name, MacAddr: iface.MAC.Address, Name: n})
            }
        }
    }
    return ifList
}


func GetNetworkList(con *libvirt.Connect) []NetworkInfo {
    netList, err := con.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE | libvirt.CONNECT_LIST_NETWORKS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get network list: %v", err)
    }

    var libvirtNetList []NetworkInfo

    for _, net := range netList {
        defer net.Free()
        xml, err := net.GetXMLDesc(0 | libvirt.NETWORK_XML_INACTIVE)
        if err != nil {
            log.Fatalf("failed to get network xml: %v", err)
        }
        var netXML libvirtxml.Network
        err = netXML.Unmarshal(xml)
        if err != nil {
            log.Fatalf("failed to unmarshal xml: %v", err)
        }
        // Processing when a structure has no members
        if netXML.Bridge != nil {
            if netXML.Forward != nil {
                if netXML.Forward.Mode == "nat" {
                    libvirtNetList = append(libvirtNetList, NetworkInfo{Name: netXML.Name, Mode: netXML.Forward.Mode, NetType: "NAT", Source: netXML.Bridge.Name})
                } else if netXML.Forward.Mode == "bridge" {
                    libvirtNetList = append(libvirtNetList, NetworkInfo{Name: netXML.Name, Mode: netXML.Forward.Mode, NetType: "Bridge", Source: netXML.Bridge.Name})
                }
            } else {
                // NOTE: The 'Private' defined here is different from libvirt's Private.
                libvirtNetList = append(libvirtNetList, NetworkInfo{Name: netXML.Name, Mode: "none", NetType: "Private", Source: netXML.Bridge.Name})
            }
        }
    }

    return libvirtNetList
}
