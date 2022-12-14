package operate

import (
	"fmt"
	"log"
	"os"
    "os/exec"
	"strings"
)


func FolderInit() {
    if !DirCheck("./data/image") {
        os.Mkdir("./data/image", os.ModePerm)
    }
    if !DirCheck("./tmp/init") {
        os.MkdirAll("./tmp/init", os.ModePerm)
    }
    if !DirCheck("./tmp/iso") {
        os.MkdirAll("./tmp/iso", os.ModePerm)
    }
    if !DirCheck("./tmp/shell") {
        os.MkdirAll("./tmp/shell", os.ModePerm)
    }
}


func ShellInit() {
    if !FileCheck("/etc/rc.local") {
        fp, err := os.OpenFile("/etc/rc.local", os.O_RDWR | os.O_CREATE, 0755)
        if err != nil {
            log.Fatalf("failed to create file: %v", err)
        }
        err = fp.Close()
        if err != nil {
            log.Fatalf("failed to close file: %v", err)
        }
        FileWrite("/etc", "rc.local", "#!/bin/bash\n")
    }

    fileText := FileRead("/etc/rc.local")
    if !strings.Contains(fileText, "BridgeNetwork.sh") {
        pwd := GetPWD()
        cmd := "bash " + pwd + "/BridgeNetwork.sh"
        file, err := os.OpenFile("/etc/rc.local", os.O_WRONLY | os.O_APPEND, 0755)
        if err != nil {
            log.Fatalf("failed to open file: %v", err)
        }
        fmt.Fprintf(file, cmd)
    }
}


func OSTypeListInit() {
    ps, err := exec.Command("osinfo-query", "os").CombinedOutput()
    if err != nil {
            log.Fatalf("failed to get os list: %v", err)
    }
    s := string(ps)
    osList = []string{}
    if strings.Contains(s, "ubuntu20.04") {
        osList = append(osList, "Ubuntu20.04")
    }
    if strings.Contains(s, "ubuntu22.04") {
        osList = append(osList, "Ubuntu22.04")
    }
    if strings.Contains(s, "centos-stream8") {
        osList = append(osList, "CentOS8")
    }
    if strings.Contains(s, "centos-stream9") {
        osList = append(osList, "CentOS9")
    }

}


func Initialize() {
    FolderInit()
    ShellInit()
    OSTypeListInit()
    /*
    err := exec.Command("export", "COLORTERM=24bit").Run()
    if err != nil {
        log.Fatalf("failed to run command: %v", err)
    }
    */
}
