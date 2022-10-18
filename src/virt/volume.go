package virt

import (
	"log"
	"time"

	libvirt "libvirt.org/libvirt-go"
	libvirtxml "libvirt.org/libvirt-go-xml"

	"github.com/nyanco01/virt-tui/src/operate"
)


type Diskinfo struct {
    Path            string
    Capacity        uint64
    Allocation      uint64
}


func GetDisksByPool(con *libvirt.Connect, name string) []Diskinfo {
    pool, err := con.LookupStoragePoolByName(name)
    defer pool.Free()
    if err != nil {
        log.Fatalf("failed to connect pool %s:%v",name ,err)
    }
    disks, err := pool.ListAllStorageVolumes(0)
    if err != nil {
        log.Fatalf("failed to get volume list by %s:%v",name ,err)
    }
    infos := []Diskinfo{}
    for _, disk := range disks {
        path, _ := disk.GetPath()
        volInfo, _ := disk.GetInfo()
        infos = append(infos, Diskinfo{
            Path:       path,
            Capacity:   volInfo.Capacity,
            Allocation: volInfo.Allocation,
        })
    }
    return infos
}

func GetPoolInfo(con *libvirt.Connect, name string) (path string, capacity, allocation uint64) {
    pool, err := con.LookupStoragePoolByName(name)
    defer pool.Free()
    if err != nil {
        log.Fatalf("failed to connect pool %s:%v",name ,err)
    }
    xml, err := pool.GetXMLDesc(0)
    if err != nil {
        log.Fatalf("failed to get xml by %s:%v",name ,err)
    }
    var poolxml libvirtxml.StoragePool
    poolxml.Unmarshal(xml)
    path = poolxml.Target.Path
    info ,_ := pool.GetInfo()
    capacity = info.Capacity
    allocation = info.Allocation

    return
}

func GetBelongVM(con *libvirt.Connect ,name string) string {
    doms, err := con.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
    if err != nil {
        log.Fatalf("failed to connect domain:%v",err)
    }
    for _, dom := range doms {
        defer dom.Free()
        _, err := dom.GetBlockInfo(name, 0)
        if err == nil {
            n, _ := dom.GetName()
            return n
        }
    }

    return "none"
}


func GetVolumeInfo(poolPath string, con *libvirt.Connect) Diskinfo {
    vol, err := con.LookupStorageVolByPath(poolPath)
    if err != nil {
        log.Fatalf("failed to get volume: %v", err)
    }
    defer vol.Free()
    path, _ := vol.GetPath()
    volinfo, _ := vol.GetInfo()
    return Diskinfo{
        Path:       path,
        Capacity:   volinfo.Capacity,
        Allocation: volinfo.Allocation,
    }
}


func CheckCreateVolumeRequest(name string, size int, available uint64) (OK bool, ErrInfo string) {
    OK = true
    ErrInfo = ""
    if name == "" {
        OK = false
        ErrInfo = "Volume name is empty."
        return
    }
    if size == 0 {
        OK = false
        ErrInfo = "Volume size is empty."
        return
    }

    if available <= uint64(size * 1024 * 1024 * 1024) {
        OK = false
        ErrInfo = "The maximum size of Pool is exceeded"
        return
    }

    return
}

func CreateVolume(name, poolPath string, size int, con *libvirt.Connect) {
    pool, err := con.LookupStoragePoolByTargetPath(poolPath)
    defer pool.Free()
    if err != nil {
        log.Fatalf("failed to get pool: %v", err)
    }
    xml := CreateVolumeXML(name, poolPath, size)
    vol, _ := pool.StorageVolCreateXML(xml, libvirt.STORAGE_VOL_CREATE_PREALLOC_METADATA)
    pool.Refresh(0)
    vol.Free()
}

func CreateVolumeXML(name, poolPath string, size int) string {
    tmpXML := operate.FileRead("./data/xml/volume/qcow2.xml")
    var volXML libvirtxml.StorageVolume
    volXML.Unmarshal(tmpXML)
    volXML.Name = name
    volXML.Key = poolPath + "/" + name
    volXML.Target.Path = poolPath + "/" + name
    volXML.Capacity.Value = uint64(size*1024*1024*1024)
    volXML.Allocation.Value = 0

    xmlData, _ := volXML.Marshal()
    return xmlData
}


func CheckCreatePoolRequest(name, path string, con *libvirt.Connect) (OK bool, ErrInfo string) {
    OK = true
    ErrInfo = ""
    if name == "" {
        OK = false
        ErrInfo = "Pool name is empty."
        return
    }
    if path == "" {
        OK = false
        ErrInfo = "Pool path is empty."
        return
    }

    pools, err := con.ListAllStoragePools(0)
    if err != nil {
        log.Fatalf("failed to get pool list: %v", err)
    }
    var listName []string
    var listPath []string
    for _, pool := range pools {
        poolxml, _ := pool.GetXMLDesc(0)
        var xml libvirtxml.StoragePool
        xml.Unmarshal(poolxml)
        listName = append(listName, xml.Name)
        listPath = append(listPath, xml.Target.Path)
        pool.Free()
    }

    for _, n := range listName {
        if n == name {
            OK = false
            ErrInfo = "The same Pool name is already defined."
            return
        }
    }

    for _, p := range listPath {
        if p == path {
            OK = false
            ErrInfo = "The same Pool path is already defined."
            return
        }
    }

    if !operate.DirCheck(path) {
        OK = false
        ErrInfo = "Pool path does not exist"
        return
    }

    return
}


func CreatePool(name, path string, con *libvirt.Connect) {
    poolxml := CreatePoolXML(name, path)
    pool, err := con.StoragePoolDefineXML(poolxml, 0)
    if err != nil {
        log.Fatalf("failed to create pool: %v", err)
    }
    time.Sleep(time.Second)
    pool.SetAutostart(true)
    pool.Create(0)
    pool.Free()
}


func CreatePoolXML(name, path string) string {
    tmpXML := operate.FileRead("./data/xml/pool/directory.xml")
    var poolXML libvirtxml.StoragePool
    poolXML.Unmarshal(tmpXML)
    poolXML.Name = name
    poolXML.Target.Path = path
    poolXML.UUID = operate.CreateUUID()

    xmlData, _ := poolXML.Marshal()
    return xmlData
}