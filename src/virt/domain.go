package virt

import (
	"log"
	"sort"
	"time"

	"github.com/nyanco01/virt-tui/src/operate"
	libvirt "libvirt.org/go/libvirt"
    libvirtxml "libvirt.org/go/libvirtxml"
)


type PoolInfos struct {
    Name        []string
    Avalable    []uint64
    Path        []string
}

type VM struct {
    Domain          *libvirt.Domain
    Name            string
    Status          bool
}

type TrafficByMAC struct {
    MACAddr         string
    TxBytes         int64
    RxBytes         int64
}

type CreateVMRequest struct {
    DomainName      string
    CPUNum          int
    MemNum          int
    DiskPath        string
    DiskSize        int
    VNCPort         int
    NICBridgeIF     string
    HostName        string
    UserName        string
    UserPassword    string
}


func butVMItemCheck(item string) string {
    switch item {
    case "VMName":
        return "VM name field is wrong."
    case "Memory":
        return "Memory field is wrong."
    case "Disk":
        return "Disk field is wrong."
    case "VNC":
        return "VNC Port field is wrong."
    case "HostName":
        return "No host name"
    case "UserName":
        return "No user name"
    case "UserPass":
        return "No user password"
    }
    return ""
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


func GetTrafficByMAC(d *libvirt.Domain, mac string) (txByte, rxByte int64) {
    ifState, err := d.InterfaceStats(mac)
    if err != nil {
        log.Fatalf("failed to get iface state: %v", err)
    }

    return ifState.TxBytes, ifState.RxBytes
}


func GetNICListMAC(d *libvirt.Domain) (mac []string) {
    xml, err := d.GetXMLDesc(0)
    if err != nil {
        log.Fatalf("failed to open xml: %v", err)
    }
    var xmlDomain libvirtxml.Domain
    xmlDomain.Unmarshal(xml)
    for _, iface := range xmlDomain.Devices.Interfaces {
        if iface.MAC != nil {
            mac = append(mac, iface.MAC.Address)
        }
    }
    return
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
            Path:           name,
            Allocation:     info.Allocation,
            Capacity:       info.Capacity,
        })
    }

    return infos
}


func GetNICInfo(d *libvirt.Domain) (name, MAC string) {
    xml, err := d.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get xml: %v", err)
    }
    var domXML libvirtxml.Domain
    err = domXML.Unmarshal(xml)
    if err != nil {
        log.Fatalf("failed to unmarshal xml: %v", err)
    }
    for _, iface := range domXML.Devices.Interfaces {
        MAC = iface.MAC.Address
    }
    name = operate.GetIFNameByMAC(MAC)
    MAC = operate.ConvertMAC(MAC)
    return
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


func GetNodeMax(c *libvirt.Connect) (maxCPU int, maxMem uint64) {
    nodeInfo, err := c.GetNodeInfo()
    if err != nil {
        log.Fatalf("failed to get node info: %v", err)
    }
    maxCPU = int(nodeInfo.Cpus) - 2
    maxMem = nodeInfo.Memory - uint64(2 * 1024 * 1024)
    return
}

// Make a list of the VNC ports you are using from the list of VMs
func GetUsedResources(vms []VM) (name []string, vnc []int) {
    var domainXml libvirtxml.Domain
    for _, vm := range vms {
        xml, _ := vm.Domain.GetXMLDesc(0)
        domainXml.Unmarshal(xml)
        name = append(name, domainXml.Name)
        for _, graphics := range domainXml.Devices.Graphics {
            vnc = append(vnc, graphics.VNC.Port)
        }
    }
    sort.Slice(vnc, func(i, j int) bool { return vnc[i] < vnc[j] })
    return
}


func GetPoolList(c *libvirt.Connect) PoolInfos {
    pools, err := c.ListAllStoragePools(0)
    if err != nil {
        log.Fatalf("failed to get pools: %v", err)
    }
    var xmlPool libvirtxml.StoragePool
    var Infos PoolInfos
    for _, pool := range pools {
        xml, err := pool.GetXMLDesc(1)
        if err != nil {
            log.Fatalf("failed to get pool xml: %s", err)
        }
        xmlPool.Unmarshal(xml)
        Infos.Name = append(Infos.Name, xmlPool.Name)
        Infos.Avalable = append(Infos.Avalable, xmlPool.Available.Value)
        Infos.Path = append(Infos.Path, xmlPool.Target.Path)
    }
    return Infos
}


func CheckCreateVMRequest(request CreateVMRequest, con *libvirt.Connect) (OK bool, ErrInfo string) {
    vms := LookupVMs(con)
    _, maxMem := GetNodeMax(con)
    listVMName, listVNCPort := GetUsedResources(vms)
    //listPool := GetPoolList(con)

    check := map[string]bool{}

    // domain name check
    check["VMName"] = true
    for _, n := range listVMName {
        if n == request.DomainName {
            check["VMName"] = false
        }
    }
    if request.DomainName == "" { check["VMName"] = false }
    // memory size check
    if request.MemNum > int(maxMem / 1024) {
        check["Memory"] = false
    }
    if request.MemNum == 0 { check["Memory"] = false }
    // Disk
    pool, _ := con.LookupStoragePoolByTargetPath(request.DiskPath)
    poolInfo, _ := pool.GetInfo()
    ava := int(poolInfo.Available / uint64(1024*1024*1024)) - 2
    if request.DiskSize > ava {
        check["Disk"] = false
    }
    if request.DiskSize == 0 { check["Disk"] = false }
    // VNC port check
    for _, p := range listVNCPort {
        if p == request.VNCPort {
            check["VNC"] = false
        }
    }
    if request.VNCPort == 0 { check["VNC"] = false }
    // host name check
    if request.HostName == "" { check["HostName"] = false }
    // user name check
    if request.UserName == "" { check["UserName"] = false }
    // user password check
    if request.UserPassword == "" { check["UserPass"] = false }

    OK = true
    out := ""
    for key, value := range check {
        if !value {
            out = key
            OK = false
            break
        }
    }
    if OK {
        ErrInfo = ""
    } else {
        ErrInfo = butVMItemCheck(out)
    }
    return
}


func CreateDomain(request CreateVMRequest, con *libvirt.Connect, c chan float64, status chan string, done chan int) {
    if !operate.FileCheck("./data/image/ubuntu-20.04-server-cloudimg-amd64.img") {
        status <- "Download image file"
        operate.DownloadFile("https://cloud-images.ubuntu.com/releases/focal/release-20220824/ubuntu-20.04-server-cloudimg-amd64.img","./data/image", c)
    } else {
        c <- 70.0
    }
    status <- "Create volume"
    CreateVol("ubuntu-20.04-server-cloudimg-amd64.img", request.DiskPath, request.DomainName, request.DiskSize, con)
    c <- 80.0
    // cloud-init make iso file
    status <- "cloud-init"
    operate.MakeISO(request.UserName, request.UserPassword, request.HostName, request.DomainName)
    c <- 85.0
    // create xml file
    status <- "Create xml file"
    xml := CreateDomainXML(request.DomainName, request.DiskPath, request.CPUNum, request.MemNum, request.VNCPort)
    c <- 90.0
    // create domain
    //dom, err := con.DomainDefineXML(xml)
    dom, err :=con.DomainDefineXML(xml)
    if err!=nil {
        log.Fatalf("failed to create domain: %v", err)
    }
    c <- 95.0
    status <- "Attach network interface"
    AttachBridgeNIC(dom, request.NICBridgeIF)
    c <- 100.0
    status <- "Complete !"
    time.Sleep(time.Second)
    dom.Free()
    done <- 1
}


func CreateDomainXML(domain, diskPath string, vcpu, mem, vnc int) string {
    tmpXML := operate.FileRead("./data/xml/domain/ubuntu-20.04-server.xml")
    var domXML libvirtxml.Domain
    domXML.Unmarshal(tmpXML)
    domXML.Name = domain
    domXML.UUID = operate.CreateUUID()
    domXML.VCPU.Value = uint(vcpu)
    domXML.Memory.Value = uint(mem*1024)
    domXML.CurrentMemory.Value = uint(mem*1024)
    for _, disk := range domXML.Devices.Disks {
        if disk.Device == "disk" {
            disk.Source.File.File = diskPath + "/" + domain
        } else if disk.Device == "cdrom" {
            disk.Source.File.File = operate.GetPWD() + "/tmp/iso/" + domain + ".iso"
        }
    }
    for _, graphics := range domXML.Devices.Graphics {
        graphics.VNC.Port = vnc
    }
    xmlData, _ := domXML.Marshal()
    return xmlData
}


func CreateVol(item, path, name string, resize int, con *libvirt.Connect) {
    // connect pool
    pool, err := con.LookupStoragePoolByTargetPath(path)
    if err != nil {
        log.Fatalf("failed to get pool: %v", err)
    }
    // create xml file
    xml := CreateVolXML(path, name, resize)
    vol, _ := pool.StorageVolCreateXML(xml, 1)

    //Move image files to pool
    src := "./data/image/" + item
    dst := path + "/" + name
    operate.FileCopy(src, dst)

    //get now capacity
    info, _ := vol.GetInfo()
    size := uint64(resize*1024*1024*1024) - info.Capacity
    // resize
    vol.Resize(size, 2)
}


func CreateVolXML(path, name string, resize int) string {
    tmpXML := operate.FileRead("./data/xml/volume/qcow2.xml")
    var volXML libvirtxml.StorageVolume
    volXML.Unmarshal(tmpXML)
    volXML.Name = name
    volXML.Key = path + "/" + name
    volXML.Target.Path = path + "/" + name
    volXML.Capacity.Value = uint64(resize*1024*1024*1024)

    xmlData, _ := volXML.Marshal()
    //operate.FileWrite("./tmp/xml/volume", name, xmlData)
    return xmlData
}


func AttachBridgeNIC(d *libvirt.Domain, ifName string) {
    var nicXML libvirtxml.DomainInterface
    nicXML.Unmarshal(operate.FileRead("./data/xml/domif/br.xml"))
    nicXML.Source.Bridge.Bridge = ifName
    nicXML.MAC.Address = operate.NewBridgeMAC(ifName)
    xml, err := nicXML.Marshal()
    if err != nil {
        log.Fatalf("failed to marshal xml: %v", err)
    }
    err = d.AttachDeviceFlags(xml, 2)
    if err != nil {
        log.Fatalf("failed to attach network interface: %v\n", err)
    }
}
