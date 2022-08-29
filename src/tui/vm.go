package tui

import (
    "log"
    "fmt"
    "time"

    "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

    "github.com/nyanco01/virt-tui/src/virt"
)


type CPU struct {
    *tview.Box
    usageGraph      [8][500]string
    usage           [500]float64
    vcpus           uint
}

type Mem struct {
    *tview.Box
    usageGraph      [8][500]string
    usage           [500]float64
    maxMem          uint64
    usedMem         uint64
}

type Disk struct {
    *tview.Box
    infos           []virt.Diskinfo
}

type NIC struct {
    *tview.Box
    bwGraphUp       [8][500]string
    bwGraphDown     [8][500]string
    bwUp            [500]int64
    bwDown          [500]int64
    name            string
}


func NotUpVM(name string) *tview.Box {
    box := tview.NewBox().SetBorder(false)
    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, name + " is shutdown", x+1, y + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)

        return x + 1, (y - (height / 2)) + 1, width - 2, height -(y - (height / 2)) + 1 - y
    })

    return box
}

func CreatePages(app *tview.Application) *tview.Pages {
    page := tview.NewPages()
    page.SetBorder(false)

    return page
}

// -------------------------------- Info --------------------------------
func NewVMInfo(dom *libvirt.Domain) *tview.Box {
    box := tview.NewBox().SetBorder(false)
    name, err := dom.GetName()
    if err != nil {
        log.Fatalf("failed to get domain name: %v", err)
    }
    id, err := dom.GetID()
    if err != nil {
        log.Fatalf("failed to get domain id: %v", err)
    }
    uuid, err := dom.GetUUIDString()
    if err != nil {
        log.Fatalf("failed to get domain uuid: %v", err)
    }

    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, fmt.Sprintf("Name : %s", name), x+1, y+1, width, tview.AlignLeft, tcell.ColorWhite)
        tview.Print(screen, fmt.Sprintf("ID   : %d", id), x+1, y+2, width, tview.AlignLeft, tcell.ColorWhite)
        tview.Print(screen, fmt.Sprintf("UUID : %s", uuid), x+1, y+3, width, tview.AlignLeft, tcell.ColorWhite)

        return box.GetInnerRect()
    })

    return box
}

// -------------------------------- CPU --------------------------------
func NewCPU(vcpu uint) *CPU {
    ug := [8][500]string{}
    for i := 0; i < 8; i++ {
        for j := 0; j < 500; j++ {
            ug[i][j] = " "
        }
    }
    u := [500]float64{}
    for i := 0; i < 500; i++ {
        u[i] = 0.0
    }

    return &CPU {
        Box:        tview.NewBox(),
        usageGraph: ug,
        usage:      u,
        vcpus:      vcpu,
    }
}

func (c *CPU) Draw(screen tcell.Screen) {
    c.Box.DrawForSubclass(screen, c)
    x, y, w, _ := c.GetInnerRect()

    // draw graph
    for i := 0; i < w; i++ {
        usage := c.usage[i]
        for j := 0; j < 8; j++ {
            if (usage - 12.5) > 0 {
                c.usageGraph[j][i] = "⣿"
                usage -= 12.5
            } else {
                a := int(usage / 3.125)
                switch {
                case a == 0:
                    c.usageGraph[j][i] = " "
                case a == 1:
                    c.usageGraph[j][i] = "⣀"
                case a == 2:
                    c.usageGraph[j][i] = "⣤"
                case a == 3:
                    c.usageGraph[j][i] = "⣶"
                }
                usage = 0
            }
        }
    }

    line := [8]string{}

    // insert graph into array for drawing
    for i := 1; i <= 8; i++ {
        for j := w; j > 0; j-- {
            line[i-1] += c.usageGraph[8-i][j]
        }
    }

    // draw
    tview.Print(screen, "CPU", x, y, w, tview.AlignCenter, tcell.NewHexColor(558061))
    tview.Print(screen, "╔══", x, y, w, tview.AlignLeft, tcell.NewHexColor(558061))
    tview.Print(screen, "══╗", x, y, w, tview.AlignRight, tcell.NewHexColor(558061))

    tview.Print(screen, fmt.Sprintf("Guest VM CPU utilization is %.2f", c.usage[0]), x, y+1, w, tview.AlignCenter, tcell.ColorSkyblue)
    tview.Print(screen, fmt.Sprintf("%d vCPUs ", c.vcpus), x, y+1, w, tview.AlignRight, tcell.ColorBlue)
    tview.Print(screen, line[0], x, y+2, w, tview.AlignRight, tcell.NewHexColor(16683008))
    tview.Print(screen, line[1], x, y+3, w, tview.AlignRight, tcell.NewHexColor(16741441))
    tview.Print(screen, line[2], x, y+4, w, tview.AlignRight, tcell.NewHexColor(16735084))
    tview.Print(screen, line[3], x, y+5, w, tview.AlignRight, tcell.NewHexColor(16732566))
    tview.Print(screen, line[4], x, y+6, w, tview.AlignRight, tcell.NewHexColor(15817148))
    tview.Print(screen, line[5], x, y+7, w, tview.AlignRight, tcell.NewHexColor(12872154))
    tview.Print(screen, line[6], x, y+8, w, tview.AlignRight, tcell.NewHexColor(8813035))
    tview.Print(screen, line[7], x, y+9, w, tview.AlignRight, tcell.NewHexColor(558061))

    tview.Print(screen, "╚══", x, y+10, w, tview.AlignLeft, tcell.NewHexColor(558061))
    tview.Print(screen, "══╝", x, y+10, w, tview.AlignRight, tcell.NewHexColor(558061))
}

func (c *CPU)Update(u float64) {
    l := len(c.usage)
    _, _, w, _ := c.GetInnerRect()

    if l < w { w = l }
    for i := w-1; i >= 0; i-- {
        c.usage[i+1] = c.usage[i]
    }

    c.usage[0] = u
}

// -------------------------------- Memory --------------------------------
func NewMemory() *Mem {
    ug := [8][500]string{}
    for i := 0; i < 8; i++ {
        for j := 0; j < 500; j++ {
            ug[i][j] = " "
        }
    }
    u := [500]float64{}
    for i := 0; i < 500; i++ {
        u[i] = 0.0
    }

    return &Mem {
        Box:        tview.NewBox(),
        usageGraph: ug,
        usage:      u,
        maxMem:     0,
        usedMem:    0,
    }
}

func (m *Mem)Draw(screen tcell.Screen) {
    m.Box.DrawForSubclass(screen, m)
    x, y, w, _ := m.GetInnerRect()

    // draw graph
    for i := 0; i < w; i++ {
        usage := m.usage[i]
        for j := 0; j < 8; j++ {
            if (usage - 12.5) > 0 {
                m.usageGraph[j][i] = "⣿"
                usage -= 12.5
            } else {
                a := int(usage / 3.125)
                switch {
                case a == 0:
                    m.usageGraph[j][i] = " "
                case a == 1:
                    m.usageGraph[j][i] = "⣀"
                case a == 2:
                    m.usageGraph[j][i] = "⣤"
                case a == 3:
                    m.usageGraph[j][i] = "⣶"
                }
                usage = 0
            }
        }
    }

    line := [8]string{}
    // insert graph into array for drawing
    for i := 1; i <= 8; i++ {
        for j := w; j > 0; j-- {
            line[i-1] += m.usageGraph[8-i][j]
        }
    }

    memMax := float64(m.maxMem / 1000)
    memUsed := float64(m.usedMem / 1000)

    tview.Print(screen, "Memory", x, y, w, tview.AlignCenter, tcell.NewHexColor(9225790))
    tview.Print(screen, "╔══", x, y, w, tview.AlignLeft, tcell.NewHexColor(9225790))
    tview.Print(screen, "══╗", x, y, w, tview.AlignRight, tcell.NewHexColor(9225790))

    tview.Print(screen, fmt.Sprintf("Max %.3f MiB", memMax), x, y+1, w, tview.AlignRight, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Used %.3f MiB",memUsed), x, y+2, w, tview.AlignRight, tcell.NewHexColor(10209336))
    tview.Print(screen, line[0], x, y+3, w, tview.AlignRight, tcell.NewHexColor(16043302))
    tview.Print(screen, line[1], x, y+4, w, tview.AlignRight, tcell.NewHexColor(15060261))
    tview.Print(screen, line[2], x, y+5, w, tview.AlignRight, tcell.NewHexColor(14077222))
    tview.Print(screen, line[3], x, y+6, w, tview.AlignRight, tcell.NewHexColor(13093929))
    tview.Print(screen, line[4], x, y+7, w, tview.AlignRight, tcell.NewHexColor(12176173))
    tview.Print(screen, line[5], x, y+8, w, tview.AlignRight, tcell.NewHexColor(11192626))
    tview.Print(screen, line[6], x, y+9, w, tview.AlignRight, tcell.NewHexColor(10209336))
    tview.Print(screen, line[7], x, y+10, w, tview.AlignRight, tcell.NewHexColor(9225790))

    tview.Print(screen, "╚══", x,y+11, w, tview.AlignLeft, tcell.NewHexColor(9225790))
    tview.Print(screen, "══╝", x,y+11, w, tview.AlignRight, tcell.NewHexColor(9225790))

}

func (m *Mem)Update(max, used uint64){
    m.maxMem = max
    m.usedMem = used

    l := len(m.usage)
    _, _, w, _ := m.GetInnerRect()

    if l < w { w = l }
    for i := w-1; i >= 0; i-- {
        m.usage[i+1] = m.usage[i]
    }

    m.usage[0] = float64(used * 100 / max)
}

// -------------------------------- Disk --------------------------------
func NewDisk() *Disk {
    return &Disk {
        Box:        tview.NewBox(),
        infos:      []virt.Diskinfo{},
    }
}

func (d *Disk)AddInfo(info virt.Diskinfo) *Disk {
    d.infos = append(d.infos, info)
    return d
}

func (d *Disk)GetInfoSize() int {
    return len(d.infos)
}

func (d *Disk)Draw(screen tcell.Screen) {
    d.Box.DrawForSubclass(screen, d)
    x, y, w, h := d.GetInnerRect()

    tview.Print(screen, "Disk", x, y, w, tview.AlignCenter, tcell.ColorDarkOrange)
    for i, info := range d.infos {
        if h >= 4 {
            usage := float64(info.Allocation) / float64(info.Capacity)

            // create usage bar
            usageBar := ""
            for i := 0; i < int(usage * float64(w)); i++ {
                usageBar += "■"
            }
            // create bar
            Bar := ""
            for i := 0; i < w; i++{
                Bar += "■"
            }

            tview.Print(screen, fmt.Sprintf("File : %s",info.Name), x, y + (i*4) + 1, w, tview.AlignLeft, tcell.ColorOrange)
            tview.Print(screen, fmt.Sprintf("Volume size : %.2f", float64(info.Capacity / (1024 * 1024 * 1024))), x, y + (i*4) + 1, w, tview.AlignRight, tcell.ColorGhostWhite)
            tview.Print(screen, fmt.Sprintf("Used        : %.2f", float64(info.Allocation / (1024 * 1024 * 1024))), x, y + (i*4) + 2, w, tview.AlignRight, tcell.ColorOrange)
            // draw Bar
            tview.Print(screen, Bar, x, y + (i*4) + 3, w, tview.AlignLeft, tcell.ColorGhostWhite)
            tview.Print(screen, usageBar, x, y + (i*4) + 3, w, tview.AlignLeft, tcell.ColorOrange)
        }
        h -= 4
    }
}

// -------------------------- Network interface card ---------------------------
func NewNIC() *NIC {
    bwU := [8][500]string{}
    for i := 0; i < 8; i++ {
        for j := 0; j < 500; j++ {
            bwU[i][j] = " "
        }
    }
    bwD := [8][500]string{}
    for i := 0; i < 8; i++ {
        for j := 0; j < 500; j++ {
            bwD[i][j] = " "
        }
    }

    bwUp := [500]int64{}
    for i := 0; i < 500; i++ {
        bwUp[i] = 0
    }

    bwDown := [500]int64{}
    for i := 0; i < 500; i++ {
        bwDown[i] = 0
    }

    return &NIC {
        Box:                tview.NewBox(),
        bwGraphUp:          bwU,
        bwGraphDown:        bwD,
        bwUp:               bwUp,
        bwDown:             bwDown,
    }
}

func (n *NIC)Draw(screen tcell.Screen) {
    n.Box.DrawForSubclass(screen, n)
    x, y, w, _ := n.GetInnerRect()

    var Uploadjudge int64
    var Downloadjudge int64

    // Upload Bandwidth
    Uploadjudge = 0
    for i := 0; i < 5; i++ {
        Uploadjudge += n.bwUp[i]
    }
    if (Uploadjudge / 5) > (1000 * 1000) {
        for i := 0; i < w; i++ {
            bandwidth := n.bwUp[i]
            for j := 0; j < 8; j++ {
                if bandwidth > (1000 * 1000 * 500 / 8) {
                    n.bwGraphUp[j][i] = "⣿"
                    bandwidth -= (1000 * 1000 * 500 / 8)
                } else {
                    a := int(bandwidth / int64(15625000))
                    switch {
                    case a == 0:
                        n.bwGraphUp[j][i] = " "
                    case a == 1:
                        n.bwGraphUp[j][i] = "⣀"
                    case a == 2:
                        n.bwGraphUp[j][i] = "⣤"
                    case a == 3:
                        n.bwGraphUp[j][i] = "⣶"
                    }
                    bandwidth = 0
                }
            }
        }
    } else {
        for i := 0; i < w; i++ {
            bandwidth := n.bwUp[i]
            for j := 0; j < 8; j++ {
                if bandwidth > (1000 * 1000 / 8) {
                    n.bwGraphUp[j][i] = "⣿"
                    bandwidth -= 125000
                } else {
                    a := int(bandwidth / int64(31250))
                    switch {
                    case a == 0:
                        n.bwGraphUp[j][i] = " "
                    case a == 1:
                        n.bwGraphUp[j][i] = "⣀"
                    case a == 2:
                        n.bwGraphUp[j][i] = "⣤"
                    case a == 3:
                        n.bwGraphUp[j][i] = "⣶"
                    }
                    bandwidth = 0
                }
            }
        }
    }

    // Download Bandwidth
    Downloadjudge = 0
    for i := 0; i < 5; i++ {
        Downloadjudge += n.bwDown[i]
    }

    var a int

    if (Downloadjudge / 5) > (1000 * 1000) {
        for i := 0; i < w; i++ {
            bandwidth := n.bwDown[i]
            for j := 0; j < 8; j++ {
                if bandwidth > (1000 * 1000 * 500 / 8) {
                    n.bwGraphDown[j][i] = "⣿"
                    bandwidth -= (1000 * 1000 * 500 / 8)
                } else {
                    a = int(bandwidth / int64(15625000))
                    switch {
                    case a == 0:
                        n.bwGraphDown[j][i] = " "
                    case a == 1:
                        n.bwGraphDown[j][i] = "⠉"
                    case a == 2:
                        n.bwGraphDown[j][i] = "⠛"
                    case a == 3:
                        n.bwGraphDown[j][i] = "⠿"
                    }
                    bandwidth = 0
                }
            }
        }
    } else {
        for i := 0; i < w; i++ {
            bandwidth := n.bwDown[i]
            for j := 0; j < 8; j++ {
                if bandwidth > (1000 * 1000 / 8) {
                    n.bwGraphDown[j][i] = "⣿"
                    bandwidth -= 125000
                } else {
                    a = int(bandwidth / int64(31250))
                    switch {
                    case a == 0:
                        n.bwGraphDown[j][i] = " "
                    case a == 1:
                        n.bwGraphDown[j][i] = "⠉"
                    case a == 2:
                        n.bwGraphDown[j][i] = "⠛"
                    case a == 3:
                        n.bwGraphDown[j][i] = "⠿"
                    }
                    bandwidth = 0
                }
            }
        }
    }
    //log.Print(a)

    lineUp := [8]string{}
    // insert graph into array for drawing
    for i := 1; i <= 8; i++ {
        for j := w; j > 0; j-- {
            lineUp[i-1] += n.bwGraphUp[8-i][j]
        }
    }

    lineDown := [8]string{}
    // insert graph into array for drawing
    for i := 1; i <= 8; i++ {
        for j := w; j > 0; j-- {
            lineDown[i-1] += n.bwGraphDown[i-1][j]
        }
    }

    tview.Print(screen, "NIC", x, y, w, tview.AlignCenter, tcell.NewHexColor(558061))
    tview.Print(screen, "╔══", x, y, w, tview.AlignLeft, tcell.NewHexColor(558061))
    tview.Print(screen, "══╗", x, y, w, tview.AlignRight, tcell.NewHexColor(558061))

    tview.Print(screen, fmt.Sprintf("Upload : %.2f KiB", float64(n.bwUp[0] / 1000)), x-30, y+1, w, tview.AlignRight, tcell.ColorSkyblue)
    tview.Print(screen, lineUp[0], x, y+2, w, tview.AlignRight, tcell.NewHexColor(16683008))
    tview.Print(screen, lineUp[1], x, y+3, w, tview.AlignRight, tcell.NewHexColor(16741441))
    tview.Print(screen, lineUp[2], x, y+4, w, tview.AlignRight, tcell.NewHexColor(16735084))
    tview.Print(screen, lineUp[3], x, y+5, w, tview.AlignRight, tcell.NewHexColor(16732566))
    tview.Print(screen, lineUp[4], x, y+6, w, tview.AlignRight, tcell.NewHexColor(15817148))
    tview.Print(screen, lineUp[5], x, y+7, w, tview.AlignRight, tcell.NewHexColor(12872154))
    tview.Print(screen, lineUp[6], x, y+8, w, tview.AlignRight, tcell.NewHexColor(8813035))
    tview.Print(screen, lineUp[7], x, y+9, w, tview.AlignRight, tcell.NewHexColor(558061))

    tview.Print(screen, fmt.Sprintf("Download : %.2f KiB", float64(n.bwDown[0] / 1000)), x, y+1, w, tview.AlignRight, tcell.ColorSkyblue)
    tview.Print(screen, lineDown[0], x, y+10, w, tview.AlignRight, tcell.NewHexColor(16043302))
    tview.Print(screen, lineDown[1], x, y+11, w, tview.AlignRight, tcell.NewHexColor(15060261))
    tview.Print(screen, lineDown[2], x, y+12, w, tview.AlignRight, tcell.NewHexColor(14077222))
    tview.Print(screen, lineDown[3], x, y+13, w, tview.AlignRight, tcell.NewHexColor(13093929))
    tview.Print(screen, lineDown[4], x, y+14, w, tview.AlignRight, tcell.NewHexColor(12176173))
    tview.Print(screen, lineDown[5], x, y+15, w, tview.AlignRight, tcell.NewHexColor(11192626))
    tview.Print(screen, lineDown[6], x, y+16, w, tview.AlignRight, tcell.NewHexColor(10209336))
    tview.Print(screen, lineDown[7], x, y+17, w, tview.AlignRight, tcell.NewHexColor(9225790))

    tview.Print(screen, "╚══", x, y+18, w, tview.AlignLeft, tcell.NewHexColor(558061))
    tview.Print(screen, "══╝", x, y+19, w, tview.AlignRight, tcell.NewHexColor(558061))

    if (Uploadjudge / 5) > (1000 * 1000) {
        tview.Print(screen, "500 MiB", x, y+2, w, tview.AlignLeft, tcell.ColorPurple)
        tview.Print(screen, "1 MiB", x, y+9, w, tview.AlignLeft, tcell.ColorPurple)
    } else {
        tview.Print(screen, "1 MiB", x, y+2, w, tview.AlignLeft, tcell.ColorRebeccaPurple)
        tview.Print(screen, "1 KiB", x, y+9, w, tview.AlignLeft, tcell.ColorRebeccaPurple)
    }
    if (Downloadjudge / 5) > (1000 * 1000) {
        tview.Print(screen, "500 MiB", x, y+17, w, tview.AlignLeft, tcell.ColorPurple)
        tview.Print(screen, "1 MiB", x, y+10, w, tview.AlignLeft, tcell.ColorPurple)
    } else {
        tview.Print(screen, "1 MiB", x, y+17, w, tview.AlignLeft, tcell.ColorRebeccaPurple)
        tview.Print(screen, "1 KiB", x, y+10, w, tview.AlignLeft, tcell.ColorRebeccaPurple)
    }

}

func (n *NIC)Update(upload, download int64) {
    // Upload
    l := len(n.bwUp)
    _, _, w, _ := n.GetInnerRect()
    if l < w { w = l }
    for i := w-1; i >= 0; i-- {
        n.bwUp[i+1] = n.bwUp[i]
    }
    n.bwUp[0] = upload

    // Download
    l = len(n.bwDown)
    _, _, w, _ = n.GetInnerRect()
    if l < w { w = l }
    for i := w-1; i >= 0; i-- {
        n.bwDown[i+1] = n.bwDown[i]
    }
    n.bwDown[0] = download
}


func NewVMStatus(app * tview.Application, dom *libvirt.Domain, name string) tview.Primitive{
    vmstatus := tview.NewFlex().SetDirection(tview.FlexRow)
    //vmstatus.SetTitle(name)
    //vmstatus.SetBorder(true).SetBorderColor(tcell.NewHexColor(16683008))

    domInfo, err := dom.GetInfo()
    if err != nil {
        log.Fatalf("failed to get domain info: %v", err)
    }
    cpu := NewCPU(domInfo.NrVirtCpu)
    mem := NewMemory()
    disk := NewDisk()
    infos := virt.GetDisks(dom)
    for _, info := range infos {
        disk.AddInfo(info)
    }
    nic := NewNIC()

    vmstatus.AddItem(NewVMInfo(dom), 5, 1, false)
    vmstatus.AddItem(cpu, 12, 1, false)
    vmstatus.AddItem(mem, 12, 1, false)
    vmstatus.AddItem(disk, 2 + (4 * disk.GetInfoSize()), 1, false)
    vmstatus.AddItem(nic, 19, 1, false)

    go func() {
        VMStatusUpdate(app, dom, cpu, mem, nic, name)
    }()

    return vmstatus
}

func VMStatusUpdate(app *tview.Application, d *libvirt.Domain, cpu *CPU, mem *Mem, nic *NIC, name string) {
    sec := time.Second

    oldUsage, _ := virt.GetCPUUsage(d)  // cpu
    oldTX, oldRX := virt.GetNICStatus(d)  // nic

    for range time.Tick(sec) {
        newUsage, cnt := virt.GetCPUUsage(d)  // cpu
        newTX, newRX := virt.GetNICStatus(d)  // nic

        max, used := virt.GetMemUsed(d)  // memory
        app.QueueUpdateDraw(func() {
            cpu.Update(float64((newUsage - oldUsage) / (uint64(cnt) * 10000000)))  // cpu
            mem.Update(max, used)
            nic.Update(newTX - oldTX, newRX - oldRX)
        })

        oldUsage = newUsage  //cpu
        oldTX = newTX
        oldRX = newRX
    }
}

func CreateMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))

    l := virt.LookupVMs(con)
    for i, vm := range l {

        if vm.Status {
            list.AddItem(vm.Name, "", rune(i)+'0', nil)
            page.AddPage(vm.Name, NewVMStatus(app, vm.Domain, vm.Name), true, true)
        } else {
            list.AddItem(vm.Name, "shutdown", rune(i)+'0', nil)
            page.AddPage(vm.Name, NotUpVM(vm.Name), true, true)
        }
    }

    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyDown {
            list.SetCurrentItem(list.GetCurrentItem() + 1)
            return nil
        } else if event.Key() == tcell.KeyUp {
            list.SetCurrentItem(list.GetCurrentItem() - 1)
            return nil
        }
        return event
    })

    main, _ := list.GetItemText(list.GetCurrentItem())
    page.SwitchToPage(main)

    _, _, w, _ := list.GetInnerRect()
    flex.AddItem(list, w + 5, 1, true)


    return flex
}

func CreateVMUI(app *tview.Application) *tview.Flex {
    flex := tview.NewFlex()

    c, err := libvirt.NewConnect("qemu:///system")
    if err != nil {
        log.Fatalf("failed to qemu connection: %v", err)
    }

    Pages := CreatePages(app)
    Menu := CreateMenu(app, c, Pages)

    _, _, w, _ := Menu.GetInnerRect()
    flex.AddItem(Menu, w + 5, 0, true)
    flex.AddItem(Pages, 0, 1, false)

    return flex
}
