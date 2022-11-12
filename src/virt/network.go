package virt

import (
	"log"

	libvirt "libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/libvirt-go-xml"

	//"github.com/nyanco01/virt-tui/src/operate"
)


type NetworkInfo struct {
    Name        string
    Mode        string
    NetType     string
    Source      string
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
