package operate

import (
	//"log"
	"os"
	//"os/exec"
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
}



func Initialize() {
    FolderInit()
    /*
    err := exec.Command("export", "COLORTERM=24bit").Run()
    if err != nil {
        log.Fatalf("failed to run command: %v", err)
    }
    */
}
