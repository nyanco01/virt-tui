package operate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
    "math/rand"
    "strings"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
    "path/filepath"
	"strconv"
	"time"
    "net"

    "github.com/google/uuid"
)


func ShellError(err error) {
    if err != nil {
        log.Fatalf("failed to run shell command: %v", err)
    }
}


func FileCheck(path string) bool {
    if f, err := os.Stat(path); os.IsNotExist(err) || f.IsDir() {
        return false
    } else {
        return true
    }
}


func DirCheck(path string) bool {
    if f, err := os.Stat(path); os.IsNotExist(err) || !f.IsDir() {
        return false
    } else {
        return true
    }
}


func FileCopy(src, dst string) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
	}
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
	}
    defer dstFile.Close()
	io.Copy(dstFile, srcFile)
}


// Not used for large size files.
func FileRead(fileName string) string {
    bytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        panic(err)
    }

    return string(bytes)
}


func FileWrite(path, name, data string) {
    f, err := os.Create(path + "/" + name)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    d := []byte(data)
    _, err = f.Write(d)
    if err != nil {
        panic(err)
    }
}


func FileDelete(path string) {
    if FileCheck(path) {
        if err := os.Remove(path); err != nil {
            log.Fatalf("failed to remove file: %v", err)
        }
    }
}


func GetPWD() string {
    pwd, err := os.Getwd()
    if err != nil {
        panic(err)
    }
    return pwd
}


func PrintDownloadPercent(done chan int64, c chan float64, path string, total int64) {

	var stop bool = false
    file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
    defer file.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percent float64 = float64(size) / float64(total) * 100
            // Adjusted to 70% when download is complete
            percent = percent * 0.7
            c<- percent
		}

		if stop {
			break
		}
		time.Sleep(time.Second)
	}
}


func DownloadFile(url string, dest string, c chan float64) {

	file := path.Base(url)

	var path bytes.Buffer
	path.WriteString(dest)
	path.WriteString("/")
	path.WriteString(file)

	out, err := os.Create(path.String())
	if err != nil {
		fmt.Println(path.String())
		panic(err)
	}
	defer out.Close()

	headResp, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	if err != nil {
		panic(err)
	}

	done := make(chan int64)
	go PrintDownloadPercent(done, c, path.String(), int64(size))

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
	done <- n
}


func ListBridgeIF() []string {
    listNIC, err := net.Interfaces()
    if err != nil {
        log.Fatalf("failed to get network interface list: %v", err)
    }
    var brName []string
    for _, nic := range listNIC {
        if f, err := os.Stat("/sys/class/net/" + nic.Name + "/brif"); !(os.IsNotExist(err) || !f.IsDir()) {
            brName = append(brName, nic.Name)
        }
    }
    return brName
}


func ListPhysicsIF() []string {
    listNIC, err := net.Interfaces()
    if err != nil {
        log.Fatalf("failed to get network interface list: %v", err)
    }
    var brName []string
    for _, nic := range listNIC {
        if f, err := os.Stat("/sys/class/net/" + nic.Name + "/device"); !(os.IsNotExist(err) || !f.IsDir()) {
            brName = append(brName, nic.Name)
        }
    }
    return brName
}


func GetBridgeMasterIF(brIF string, underIFs []string) string {
    allIF, err := ioutil.ReadDir("/sys/class/net/" + brIF + "/brif/")
    if err != nil {
        return ""
    }
    for _, i := range allIF {
        check := false
        for _, u := range underIFs {
            if i.Name() == u {
                check = true
                break
            }
        }
        if !check {
            return i.Name()
        }
    }
    return ""
}


func ConvertMAC(mac string) string {
    macAddr, err := net.ParseMAC(mac)

    if err != nil {
        log.Fatalf("failed to parse MAC Address: %v", err)
    }
    lastThreeBites := macAddr[3:6]
    firstThreeBites := []byte{254, 84, 00}
    var m net.HardwareAddr = append(firstThreeBites, lastThreeBites...)
    return m.String()
}


func GetIFNameByMAC(mac string) string {
    listNIC, err := net.Interfaces()
    if err != nil {
        log.Fatalf("failed to get network interface list: %v", err)
    }

    for _, nic := range listNIC {
        if nic.HardwareAddr.String() == ConvertMAC(mac) {
            return nic.Name
        }
    }
    return ""
}


func NewMacAddrNotOverlap(listAddr []net.HardwareAddr) string {
    var macAddr net.HardwareAddr
    for {
        rand.Seed(time.Now().UnixNano())
        underBit := make([]byte, 3)
        _, err :=rand.Read(underBit)
        if err != nil {
            log.Fatalf("failed to generate random numbers: %v", err)
        }
        // virtio upper 24bit 52:54:00
        upperBit := []byte{82, 84, 00}
        m := append(upperBit, underBit...)
        macAddr = m
        cnt := 0
        for _, mac := range listAddr {
            if mac.String() == macAddr.String() {
                cnt++
            }
        }
        if cnt == 0 {
            break
        }
    }
    return macAddr.String()
}


func NewBridgeMAC(bridgeIF string) string {
    pathIF := "/sys/class/net/" + bridgeIF + "/brif/"

    listIF := []string{}

	err := filepath.Walk(pathIF, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
        tmp := strings.Replace(path, pathIF, "", 1)
        if tmp != "" {
            listIF = append(listIF, tmp)
        }
		return nil
	})

	if err != nil {
		panic(err)
	}
    var macAddrs []net.HardwareAddr
    for _, iFace := range listIF {
        nic, err := net.InterfaceByName(iFace)
        if err != nil {
            log.Fatalf("failed to get network interface: %v", err)
        }
        macAddrs = append(macAddrs, nic.HardwareAddr)
        //fmt.Printf("%s\t%s\n", iFace, nic.HardwareAddr.String())
    }
    return NewMacAddrNotOverlap(macAddrs)
}


func MakeISO(user, pass, host, domain string) {
    fUser, err := os.Create("./tmp/init/user-data")
    if err!=nil {
        panic(err)
    }
    defer fUser.Close()
    str := "#cloud-config\n"
    str += "user: " + user + "\n"
    str += "password: " + pass + "\n"
    str += "chpasswd: {expire: False}\n"
    str += "ssh_pwauth: True\n"
    data := []byte(str)
    _, err = fUser.Write(data)
    if err != nil {
        panic(err)
    }
    str = ""

    fMeta, err := os.Create("./tmp/init/meta-data")
    if err!=nil {
        panic(err)
    }
    defer fMeta.Close()
    str += "instance-id: " + host + "\n"
    str += "local-hostname: " + host + ".local" + "\n"
    data = []byte(str)
    _, err = fMeta.Write(data)
    if err != nil {
        panic(err)
    }
    path := "./tmp/iso/" + domain + ".iso"
    u := "./tmp/init/user-data"
    m := "./tmp/init/meta-data"
    err = exec.Command("genisoimage", "-output", path, "-volid", "cidata", "-joliet", "-rock", u, m).Run()
    ShellError(err)
}


func CreateUUID() string {
    uuidObj, _ := uuid.NewUUID()
    return uuidObj.String()
}


func CheckBridgeSource(name string) bool {
    brList := ListBridgeIF()
    for _, br := range brList {
        allIFs, _ := ioutil.ReadDir("/sys/class/net/" + br + "/brif/")
        for _, iface := range allIFs {
            if iface.Name() == name {
                return false
            }
        }
    }

    return true
}


func CreateShellByBridgeIF(name, source string) {
    path := "./tmp/shell/" + name + ".sh"
    err := exec.Command("touch", path).Run()
    ShellError(err)

    file, err := os.OpenFile(path, os.O_WRONLY | os.O_APPEND, 0644)
    if err != nil {
        log.Fatalf("failed to open file: %v",err)
    }
    defer file.Close()

    writeText := "#!/bin/bash"
    fmt.Fprintln(file, writeText)
    writeText = "sudo brctl addbr " + name
    fmt.Fprintln(file, writeText)
    writeText = "sudo brctl addif " + name + " " + source
    fmt.Fprintln(file, writeText)
}


func DeleteBridgeIF(name string) {
    err := exec.Command("brctl", "delbr", name).Run()
    if err != nil {
        log.Fatalf("failed to delete bridge interface: %v", err)
    }
}


func RunShellByCreateBridgeIF(name string) {
    path := "./tmp/shell/" + name + ".sh"
    err := exec.Command("bash", path).Run()
    ShellError(err)
}


func CheckNetworkSubnet(subnet string) bool {
    ip, ipnet, err := net.ParseCIDR(subnet)
    if err != nil {
        return false
    }
    return ip.String() == ipnet.IP.String()
}


func CheckOtherNATNetwork(addr, mask, subnet string) bool {
    m := net.ParseIP(mask).To4()
    CIDR1 := net.IPNet{IP: net.ParseIP(addr), Mask: net.IPv4Mask(m[0], m[1], m[2], m[3])}
    _, CIDR2, err := net.ParseCIDR(subnet)
    if err != nil {
        return true
    }
    return (CIDR1.Contains(CIDR2.IP.To4()) || CIDR2.Contains(CIDR1.IP.To4()))
}


func CheckSubnetLower30(subnet string) bool {
    _, CIDR, _ := net.ParseCIDR(subnet)
    a, b := CIDR.Mask.Size()
    return b-a <= 2
}


func ParseMask(subnet string) string {
    _, CIDR, _ := net.ParseCIDR(subnet)
    return net.IPv4(CIDR.Mask[0], CIDR.Mask[1], CIDR.Mask[2], CIDR.Mask[3]).String()
}


func GetCIDR(addr, mask string) string {
    tmp := net.ParseIP(mask).To4()
    m := net.IPv4Mask(tmp[0], tmp[1], tmp[2], tmp[3])
    cidr := net.IPNet{IP: net.ParseIP(addr), Mask: m}
    return cidr.String()
}

// Generate the starting address and the address allocated by DHCP from the subnet in CIDR format
func CreateIPsBySubnet(subnet string) (firstIP, secondIP, lastIP string) {
    _, CIDR, _ := net.ParseCIDR(subnet)
    firstIP = net.IPv4(CIDR.IP[0], CIDR.IP[1], CIDR.IP[2], CIDR.IP[3]+1).String()
    secondIP = net.IPv4(CIDR.IP[0], CIDR.IP[1], CIDR.IP[2], CIDR.IP[3]+2).String()
    f, l := CIDR.Mask.Size()
    s := l - f
    if s != 0 {
        switch {
        case s <= 8:
            lastIP = net.IPv4(CIDR.IP[0], CIDR.IP[1], CIDR.IP[2], (CIDR.IP[3] + 1<<s)-2).String()
        case 8 < s && s <= 16:
            tmp := s - 8
            lastIP = net.IPv4(CIDR.IP[0], CIDR.IP[1], (CIDR.IP[2]+ 1<<tmp)-1, 254).String()
        // Commented out because the address range that can be distributed 
        // by DHCP of libvirt is up to /16.
        /*
        case 16 < s && s <= 24:
            tmp := s - 16
            lastIP = net.IPv4(CIDR.IP[0], (CIDR.IP[1]+ 1<<tmp)-1, 255, 254).String()
        case 24 < s:
            tmp := s - 24
            lastIP = net.IPv4((CIDR.IP[0]+ 1<<tmp)-1, 255, 255, 254).String()
        */
        default:
            tmp := s - 8
            lastIP = net.IPv4(CIDR.IP[0], CIDR.IP[1], (CIDR.IP[2]+ 1<<tmp)-1, 254).String()
        }
    }
    return
}
