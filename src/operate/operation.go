package operate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

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
