package virt

import (
	"log"

	libvirt "libvirt.org/libvirt-go"
	libvirtxml "libvirt.org/libvirt-go-xml"
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
