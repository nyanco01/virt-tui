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
			//fmt.Printf("%.0f", percent)
			//fmt.Println("%")
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


func GetBridgeMasterIF(brIF string, underIFs []string) string {
    allIF, err := ioutil.ReadDir("/sys/class/net/" + brIF + "/brif/")
    if err != nil {
        //log.Fatalf("failed to get if list by /sys: %v", err)
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


func GetIFNameByMAC(mac string) string {
    listNIC, err := net.Interfaces()
    if err != nil {
        log.Fatalf("failed to get network interface list: %v", err)
    }
    macAddr, err := net.ParseMAC(mac)

    if err != nil {
        log.Fatalf("failed to parse MAC Address: %v", err)
    }
    lastThreeBites := macAddr[3:6]
    firstThreeBites := []byte{254, 84, 00}
    var m net.HardwareAddr = append(firstThreeBites, lastThreeBites...)
    //fmt.Println(m)
    for _, nic := range listNIC {
        if nic.HardwareAddr.String() == m.String() {
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
}


func CreateUUID() string {
    uuidObj, _ := uuid.NewUUID()
    return uuidObj.String()
}
