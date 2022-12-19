package virt

import (
	"log"

	"github.com/nyanco01/virt-tui/src/operate"
	libvirt "libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/go/libvirtxml"
)
type PCIAddr struct {
    domain      uint
    bus         uint
    slot        uint
    function    uint
}

func GetDomainItems(dom * libvirt.Domain) (items []EditItem, diskNum int, ifaceNum int) {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get xml: %v", err)
    }
    diskNum = 0
    ifaceNum = 0
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    items = append(items, &ItemCPU{
        Number:         domXML.VCPU.Value,
        PlaceMent:      domXML.VCPU.Placement,
        CPUSet:         domXML.VCPU.CPUSet,
        Mode:           domXML.CPU.Mode,
    })
    maxMem := uint(0)
    maxMemSI := ""
    curMem := uint(0)
    curMemSI := ""
    if domXML.MaximumMemory != nil {
        maxMem = domXML.MaximumMemory.Value
        maxMemSI = domXML.MaximumMemory.Unit
    }
    if domXML.CurrentMemory != nil {
        curMem = domXML.CurrentMemory.Value
        curMemSI = domXML.CurrentMemory.Unit
    }
    items = append(items, &ItemMemory{
        Size:               domXML.Memory.Value,
        SizeSI:             domXML.Memory.Unit,
        MaxSize:            maxMem,
        MaxSizeSI:          maxMemSI,
        CurrentMemory:      curMem,
        CurrentMemorySI:    curMemSI,
    })
    for _, disk := range domXML.Devices.Disks {
        p := ""
        if disk.Source != nil {
            p = disk.Source.File.File
        }
        diskNum++
        xml, _ := disk.Marshal()
        items = append(items, &ItemDisk{
            Path:       p,
            Device:     disk.Device,
            ImgType:    disk.Driver.Type,
            Bus:        disk.Target.Bus,
            Target:     disk.Target.Dev,
            ItemXML:    xml,
        })
    }
    for _, cntl := range domXML.Devices.Controllers {
        if *cntl.Index != uint(0) {
            continue
        }
        items = append(items, &ItemController{
            ControllerType:     cntl.Type,
            Model:              cntl.Model,
        })
    }
    for _, iface := range domXML.Devices.Interfaces {
        d := ""
        s := ""
        t := ""
        m := ""
        if iface.Driver != nil {
            d = iface.Driver.Name
            t = "hostdev"
        }
        if iface.Source != nil {
            s = iface.Source.Bridge.Bridge
            t = "bridge"
        }
        if iface.MAC != nil {
            m = iface.MAC.Address
        }
        ifaceNum++
        xml, _ := iface.Marshal()
        items = append(items, &ItemInterface{
            IfType:     t,
            Driver:     d,
            Source:     s,
            Model:      iface.Model.Type,
            MACAddr:    m,
            ItemXML:    xml,
        })
    }
    for _, serial := range domXML.Devices.Serials {
        items = append(items, &ItemSerial{
            TargetType: serial.Target.Type,
        })
    }
    for _, console := range domXML.Devices.Consoles {
        items = append(items, &ItemConsole{
            TargetType: console.Target.Type,
        })
    }
    for _, input := range domXML.Devices.Inputs {
        items = append(items, &ItemInput{
            InputType:  input.Type,
            Bus:        input.Bus,
        })
    }
    for _, graphics := range domXML.Devices.Graphics {
        t := ""
        p := 0
        l := ""
        if graphics.VNC != nil {
            t = "vnc"
            p = graphics.VNC.Port
            l = graphics.VNC.Listen
        }
        if graphics.RDP != nil {
            t = "rdp"
            p = graphics.RDP.Port
            l = graphics.RDP.Listen
        }
        if graphics.Spice != nil {
            t = "spice"
            p = graphics.Spice.Port
            l = ""
        }
        items = append(items, &ItemGraphics{
            GraphicsType:   t,
            Port:           p,
            ListemAddress:  l,
        })
    }
    for _, video := range domXML.Devices.Videos {
        a := ""
        if video.Address.PCI != nil {
            a = "pci"
        }
        items = append(items, &ItemVideo{
            ModelType:      video.Model.Type,
            VRAM:           video.Model.VRam,
            DeviceAddress:  a,
        })
    }
    return
}


func GetItemPCIAddress(dom *libvirt.Domain) (pciAddrs []PCIAddr, err error){
    var domXML libvirtxml.Domain
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return nil, err
    }
    domXML.Unmarshal(xml)
    Devices := domXML.Devices
    for _, d := range Devices.Disks {
        if d.Address == nil {
            continue
        }
        if d.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *d.Address.PCI.Domain,
            bus:        *d.Address.PCI.Bus,
            slot:       *d.Address.PCI.Slot,
            function:   *d.Address.PCI.Function,
        })
    }
    for _, cntl := range Devices.Controllers {
        if cntl.Address == nil {
            continue
        }
        if cntl.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *cntl.Address.PCI.Domain,
            bus:        *cntl.Address.PCI.Bus,
            slot:       *cntl.Address.PCI.Slot,
            function:   *cntl.Address.PCI.Function,
        })
    }
    for _, iface := range Devices.Interfaces {
        if iface.Address == nil {
            continue
        }
        if iface.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *iface.Address.PCI.Domain,
            bus:        *iface.Address.PCI.Bus,
            slot:       *iface.Address.PCI.Slot,
            function:   *iface.Address.PCI.Function,
        })
    }
    for _, v := range Devices.Videos {
        if v.Address == nil {
            continue
        }
        if v.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *v.Address.PCI.Domain,
            bus:        *v.Address.PCI.Bus,
            slot:       *v.Address.PCI.Slot,
            function:   *v.Address.PCI.Function,
        })
    }
    for _, t := range Devices.TPMs {
        if t.Address == nil {
            continue
        }
        if t.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *t.Address.PCI.Domain,
            bus:        *t.Address.PCI.Bus,
            slot:       *t.Address.PCI.Slot,
            function:   *t.Address.PCI.Function,
        })

    }
    if Devices.MemBalloon != nil {
        if Devices.MemBalloon.Address != nil {
            if Devices.MemBalloon.Address.PCI != nil {
                pciAddrs = append(pciAddrs, PCIAddr{
                    domain:     *Devices.MemBalloon.Address.PCI.Domain,
                    bus:        *Devices.MemBalloon.Address.PCI.Bus,
                    slot:       *Devices.MemBalloon.Address.PCI.Slot,
                    function:   *Devices.MemBalloon.Address.PCI.Function,
                })
            }
        }
    }
    for _, r := range Devices.RNGs {
        if r.Address == nil {
            continue
        }
        if r.Address.PCI == nil {
            continue
        }
        pciAddrs = append(pciAddrs, PCIAddr{
            domain:     *r.Address.PCI.Domain,
            bus:        *r.Address.PCI.Bus,
            slot:       *r.Address.PCI.Slot,
            function:   *r.Address.PCI.Function,
        })
    }
    return pciAddrs, nil
}


func DomainAddNIC(dom *libvirt.Domain, source string) error {
    pcis, err := GetItemPCIAddress(dom)
    if err != nil {
        return err
    }
    var nicXML libvirtxml.DomainInterface
    nicXML.Unmarshal(operate.FileRead("./data/xml/dom_items/br.xml"))
    nicXML.Source.Bridge.Bridge = source
    nicXML.MAC.Address = operate.NewBridgeMAC(source)
    for {
        b := true
        for _, pci := range pcis {
            if *nicXML.Address.PCI.Bus == pci.bus {
                tmp := uint(1) + *nicXML.Address.PCI.Bus
                nicXML.Address.PCI.Bus = &tmp
                b = false
            }
        }
        if b {
            break
        }
    }
    nic, err := nicXML.Marshal()
    if err != nil {
        return err
    }
    err = dom.AttachDeviceFlags(nic, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
    if err != nil {
        return err
    }
    return nil
}


func GetPoolNameList(con *libvirt.Connect) (pNames []string, err error) {
    pools, err := con.ListAllStoragePools(0)
    if err != nil {
        return nil, err
    }
    for _, p := range pools {
        n, err := p.GetName()
        if err != nil {
            return nil, err
        }
        pNames = append(pNames, n)
    }
    return
}


func GetNonAttachDiskByPool(con *libvirt.Connect, name string) []string {
    infos := GetDisksByPool(con, name)
    disks := []string{}
    for _, info := range infos {
        n := GetBelongVM(con, info.Path)
        if n == "none" {
            disks = append(disks, info.Path)
        }
    }
    return disks
}


func GetDiskNameList(dom *libvirt.Domain) (names []string, err error) {
    var domXML libvirtxml.Domain
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return nil, err
    }
    domXML.Unmarshal(xml)
    disks := domXML.Devices.Disks
    for _, disk := range disks {
        if disk.Target != nil {
            names = append(names, disk.Target.Dev)
        }
    }
    return
}


func DomainAddDisk(dom *libvirt.Domain, source string) error {
    pcis, err := GetItemPCIAddress(dom)
    if err != nil {
        return err
    }
    var diskXML libvirtxml.DomainDisk
    diskXML.Unmarshal(operate.FileRead("./data/xml/dom_items/qcow2.xml"))
    diskXML.Source.File.File = source
    for {
        b := true
        for _, pci := range pcis {
            if *diskXML.Address.PCI.Bus == pci.bus {
                tmp := uint(1) + *diskXML.Address.PCI.Bus
                diskXML.Address.PCI.Bus = &tmp
                b = false
            }
        }
        if b {
            break
        }
    }
    names, err := GetDiskNameList(dom)
    cnt := 0
    var n string = "vd" + string(rune('a' + cnt))
    for {
        b := true
        n = "vd" + string(rune('a' + cnt))
        for _, name := range names {
            if n == name {
                b = false
                break
            }
        }
        cnt++
        if b {
            break
        }
    }
    diskXML.Target.Dev = n
    disk, err := diskXML.Marshal()
    if err != nil {
        return err
    }
    err = dom.AttachDeviceFlags(disk, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
    if err != nil {
        return err
    }

    return nil
}


func GetCurrentCPUNum(dom *libvirt.Domain) (int, error) {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return -1, err
    }
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    return int(domXML.VCPU.Value), nil
}


func DomainEditCPU(dom *libvirt.Domain, cpuNum uint) error {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return err
    }
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    domXML.VCPU.Value = cpuNum
    xml, err = domXML.Marshal()
    if err != nil {
        return err
    }
    
    con, err := dom.DomainGetConnect()
    if err != nil {
        return err
    }
    _, err = con.DomainDefineXML(xml)
    if err != nil {
        return err
    }

    return nil
}


func GetCurrentMemSize(dom *libvirt.Domain) (uint, error) {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return 0, err
    }
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    return domXML.Memory.Value, nil
}


func DomainEditMemory(dom *libvirt.Domain, memSize uint) error {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        return err
    }
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    domXML.Memory.Value = memSize
    if domXML.CurrentMemory != nil {
        domXML.CurrentMemory.Value = memSize
    }
    xml, err = domXML.Marshal()
    if err != nil {
        return err
    }
    
    con, err := dom.DomainGetConnect()
    if err != nil {
        return err
    }
    _, err = con.DomainDefineXML(xml)
    if err != nil {
        return err
    }

    return nil
}


func GetDiskTarget(xml string) string {
    var diskXML libvirtxml.DomainDisk
    diskXML.Unmarshal(xml)
    return diskXML.Target.Dev
}


func DomainDeleteDisk(dom *libvirt.Domain, diskXML string) error {
    err := dom.DetachDeviceFlags(diskXML, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
    return err
}


func GetIfaceMAC(xml string) string {
    var ifaceXML libvirtxml.DomainInterface
    ifaceXML.Unmarshal(xml)
    return ifaceXML.MAC.Address
}


func DomainDeleteIface(dom *libvirt.Domain, ifaceXML string) error {
    err := dom.DetachDeviceFlags(ifaceXML, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG)
    return err
}
