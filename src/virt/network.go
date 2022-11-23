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


func GetNetworkByName(con *libvirt.Connect, name string) NetworkInfo {
    net, err := con.LookupNetworkByName(name)
    if err != nil {
        log.Fatalf("failed to get network: %v", err)
    }
    defer net.Free()

    var netInfo NetworkInfo
    xml, err := net.GetXMLDesc(0 | libvirt.NETWORK_XML_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get xml by network: %v", err)
    }
    var netXML libvirtxml.Network
    err = netXML.Unmarshal(xml)
    if err != nil {
        log.Fatalf("failed to unmarshal xml by network: %v", err)
    }
    
    if netXML.Bridge != nil {
        if netXML.Forward != nil {
            if netXML.Forward.Mode == "nat" {
                netInfo = NetworkInfo{Name: netXML.Name, Mode: netXML.Forward.Mode, NetType: "NAT", Source: netXML.Bridge.Name}
            } else if netXML.Forward.Mode == "bridge" {
                netInfo = NetworkInfo{Name: netXML.Name, Mode: netXML.Forward.Mode, NetType: "Bridge", Source: netXML.Bridge.Name}
            }
        } else {
            netInfo = NetworkInfo{Name: netXML.Name, Mode: "none", NetType: "Private", Source: netXML.Bridge.Name}
        }
    }
    
    return netInfo
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


func GetAddressByNATNetwork(con *libvirt.Connect, name string) (addr, dhcpStart, dhcpEnd string) {
    net, err := con.LookupNetworkByName(name)
    if err != nil {
        log.Fatalf("failed to get network: %v", err)
    }
    defer net.Free()
    xml, err := net.GetXMLDesc(0 | libvirt.NETWORK_XML_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get network xml: %v", err)
    }
    var netXML libvirtxml.Network
    netXML.Unmarshal(xml)
    addr = operate.GetCIDR(netXML.IPs[0].Address, netXML.IPs[0].Netmask)
    if netXML.IPs[0].DHCP != nil {
        dhcpStart = netXML.IPs[0].DHCP.Ranges[0].Start
        dhcpEnd = netXML.IPs[0].DHCP.Ranges[0].End
    } else {
        dhcpStart = ""
        dhcpEnd = ""
    }
    return
}


func CheckNetworkName(con *libvirt.Connect, name string) bool {
    netlist, err := con.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE | libvirt.CONNECT_LIST_NETWORKS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get network list: %v", err)
    }
    for _, net := range netlist {
        n, err := net.GetName()
        if err != nil {
            log.Fatalf("failed to get network name: %v", err)
        }
        if n == name {
            return false
        }
    }
    return true
}


func CreateNetworkByBridge(con *libvirt.Connect, name, source string) {
    operate.CreateShellByBridgeIF(name, source)
    operate.RunShellByCreateBridgeIF(name)

    var netXML libvirtxml.Network
    netXML.Unmarshal(operate.FileRead("./data/xml/network/bridge.xml"))
    netXML.Name = name
    netXML.UUID = operate.CreateUUID()
    netXML.Bridge.Name = name

    xml, err := netXML.Marshal()
    if err != nil {
        log.Fatalf("failed to marshal xml by bridge network: %v", err)
    }
    net, err := con.NetworkDefineXML(xml)
    if err != nil {
        log.Fatalf("failed to create bridge network: %v",err)
    }
    err = net.SetAutostart(true)
    err = net.Create()
    if err != nil {
        log.Fatalf("failed to start bridge network: %v",err)
    }
}


func CreateNetworkByNAT(con *libvirt.Connect, name, network string) {
    var netXML libvirtxml.Network
    netXML.Unmarshal(operate.FileRead("./data/xml/network/nat.xml"))
    netXML.Name = name
    netXML.UUID = operate.CreateUUID()
    netXML.Bridge.Name = name
    IFaddr, dhcpStart, dhcpEnd := operate.CreateIPsBySubnet(network)
    netXML.IPs[0].Address = IFaddr
    netXML.IPs[0].Netmask = operate.ParseMask(network)
    if operate.CheckSubnetLower30(network) {
        netXML.IPs[0].DHCP = nil
    } else {
        netXML.IPs[0].DHCP.Ranges[0].Start = dhcpStart
        netXML.IPs[0].DHCP.Ranges[0].End = dhcpEnd
    }
    xml, err := netXML.Marshal()
    if err != nil {
        log.Fatalf("failed to marshal xml by nat network: %v", err)
    }
    net, err := con.NetworkDefineXML(xml)
    if err != nil {
        log.Fatalf("failed to create nat network: %v",err)
    }
    err = net.SetAutostart(true)
    err = net.Create()
    if err != nil {
        log.Fatalf("failed to start nat network: %v",err)
    }
}


func CreateNetworkByPrivate(con *libvirt.Connect, name string) {
    var netXML libvirtxml.Network
    netXML.Unmarshal(operate.FileRead("./data/xml/network/private.xml"))
    netXML.Name = name
    netXML.UUID = operate.CreateUUID()
    netXML.Bridge.Name = name

    xml, err := netXML.Marshal()
    if err != nil {
        log.Fatalf("failed to marshal xml by private network: %v", err)
    }
    net, err := con.NetworkDefineXML(xml)
    if err != nil {
        log.Fatalf("failed to create private network: %v",err)
    }
    err = net.SetAutostart(true)
    err = net.Create()
    if err != nil {
        log.Fatalf("failed to start private network: %v",err)
    }
}


func CheckNetworkRange(con *libvirt.Connect, subnet string) bool {
    netlist, err := con.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE | libvirt.CONNECT_LIST_NETWORKS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get network list: %v", err)
    }
    for _, net := range netlist {
        defer net.Free()
        xml, err := net.GetXMLDesc(0 | libvirt.NETWORK_XML_INACTIVE)
        if err != nil {
            log.Fatalf("failed to get xml: %v", err)
        }
        var netXML libvirtxml.Network
        netXML.Unmarshal(xml)
        if netXML.Forward != nil {
            if netXML.Forward.Mode == "nat" {
                if netXML.IPs != nil {
                    for _, ip := range netXML.IPs {
                        if operate.CheckOtherNATNetwork(ip.Address, ip.Netmask, subnet) {
                            return true
                        }
                    }
                }
            }
        }
    }
    return false
}
