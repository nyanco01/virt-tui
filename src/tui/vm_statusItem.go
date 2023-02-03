package tui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/constants"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
)

const (
    kilo = constants.Kilo
    mega = constants.Mega
    giga = constants.Giga
    kibi = constants.Kibi
    mebi = constants.Mebi
    gibi = constants.Gibi
)

const (
    leftTriangle    = constants.LeftTriangle
    rightTraiangle  = constants.RightTraiangle
    upTraiangle     = constants.UpTraiangle
    downTraiangle   = constants.DownTraiangle
)

var VirtualMachineStatus    map[string]bool

type CPU struct {
    *tview.Box
    usageGraph      [150][500]string
    usage           [500]float64
    vcpus           uint
}

type Mem struct {
    *tview.Box
    usageGraph      [150][500]string
    usage           [500]float64
    maxMem          uint64
    usedMem         uint64
}

type Disk struct {
    *tview.Box
    infos           []virt.Diskinfo
    index           int
}

type NICMember struct {
    bwUp            [500]int64
    bwDown          [500]int64
    oldUp           int64
    oldDown         int64
    name            string
    MACAddr         string
}

type NIC struct {
    *tview.Box
    bwGraphUp       [150][500]string
    bwGraphDown     [150][500]string
    ifList          []NICMember
    index           int
    /*
    bwUp            [500]int64
    bwDown          [500]int64
    name            string
    MACAddr         string
    */
}


func NotUpVM(name string) *tview.Box {
    box := tview.NewBox().SetBorder(false)
    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, name + " is shutdown", x+1, y + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)

        return x + 1, (y - (height / 2)) + 1, width - 2, height -(y - (height / 2)) + 1 - y
    })

    return box
}

// -------------------------------- Info --------------------------------
func NewVMInfo(dom *libvirt.Domain) *tview.Box {
    box := tview.NewBox().SetBorder(false)
    name, err := dom.GetName()
    if err != nil {
        log.Fatalf("failed to get domain name: %v", err)
    }
    var id uint = 0
    if b, _ := dom.IsActive(); b{
    id, err = dom.GetID()
        if err != nil {
            log.Fatalf("failed to get domain id: %v", err)
        }
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
    ug := [150][500]string{}
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
    x, y, w, h := c.GetInnerRect()

    graphHeight := h - 3
    if graphHeight < 0 {
        graphHeight = 0
    }
    brailleGradient := float64(100) / float64(graphHeight * 4)

    // draw graph
    for i := 0; i < w; i++ {
        usage := c.usage[i]
        for j := 0; j < graphHeight; j++ {
            if (usage - (brailleGradient*4)) > 0 {
                c.usageGraph[j][i] = "⣿"
                usage -= (brailleGradient*4)
            } else {
                a := float64(usage / brailleGradient)
                switch {
                case a < 1.0:
                    c.usageGraph[j][i] = " "
                case 1.0 <= a && a < 2.0:
                    c.usageGraph[j][i] = "⣀"
                case 2.0 <= a && a < 3.0:
                    c.usageGraph[j][i] = "⣤"
                case 3.0 <= a && a < 4.0:
                    c.usageGraph[j][i] = "⣶"
                }
                usage = 0
            }
        }
    }

    graph := []string{}

    for i := 0; i <= graphHeight; i++ {
        tmpLine := ""
        for j := w; j > 0; j-- {
            tmpLine += c.usageGraph[graphHeight - i][j]
        }
        graph = append(graph, tmpLine)
    }

    // draw

    tview.Print(screen, "CPU", x, y-1, w, tview.AlignCenter, tcell.NewRGBColor(0, 255, 127))
    tview.Print(screen, constants.LeftTop, x, y-1, w, tview.AlignLeft, tcell.NewRGBColor(0, 255, 127))
    tview.Print(screen, constants.RightTop, x, y-1, w, tview.AlignRight, tcell.NewRGBColor(0, 255, 127))

    tview.Print(screen, fmt.Sprintf("%.2f%%", c.usage[0]), x, y, w, tview.AlignCenter, tcell.ColorForestGreen)
    tview.Print(screen, fmt.Sprintf("%d vCPUs ", c.vcpus), x, y, w, tview.AlignRight, tcell.ColorSpringGreen)

    color := setColorGradation(CPU_COLOR, len(graph))
    for i, line := range graph {
        tview.Print(screen, line, x, y+1+i, w, tview.AlignRight, color[i])
    }

    l := len(graph)
    tview.Print(screen, constants.LeftDown, x, y+1+l, w, tview.AlignLeft, tcell.NewRGBColor(0, 255, 127))
    tview.Print(screen, constants.RightDown, x, y+1+l, w, tview.AlignRight, tcell.NewRGBColor(0, 255, 127))
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

func (c *CPU)GetLastStatus() (float64, int) {
    return c.usage[0], int(c.vcpus)
}


// -------------------------------- Memory --------------------------------
func NewMemory() *Mem {
    ug := [150][500]string{}
    for i := 0; i < 150; i++ {
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
    x, y, w, h := m.GetInnerRect()

    graphHeight := h - 4
    if graphHeight < 0 {
        graphHeight = 0
    }
    brailleGradient := float64(100) / float64(graphHeight * 4)

    // draw graph
    for i := 0; i < w; i++ {
        usage := m.usage[i]
        for j := 0; j < graphHeight; j++ {
            if (usage - (brailleGradient*4)) > 0 {
                m.usageGraph[j][i] = "⣿"
                usage -= (brailleGradient*4)
            } else {
                a := int(usage / brailleGradient)
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
    graph := []string{}

    for i := 0; i <= graphHeight; i++ {
        tmpLine := ""
        for j := w; j > 0; j-- {
            tmpLine += m.usageGraph[graphHeight - i][j]
        }
        graph = append(graph, tmpLine)
    }

    memMax := float64(m.maxMem / kilo)
    memUsed := float64(m.usedMem / kilo)

    tview.Print(screen, "Memory", x, y-1, w, tview.AlignCenter, tcell.NewRGBColor(254, 89, 19))
    tview.Print(screen, constants.LeftTop, x, y-1, w, tview.AlignLeft, tcell.NewRGBColor(254, 89, 19))
    tview.Print(screen, constants.RightTop, x, y-1, w, tview.AlignRight, tcell.NewRGBColor(254, 89, 19))

    tview.Print(screen, fmt.Sprintf("Max %.3f MiB", memMax), x, y, w, tview.AlignRight, tcell.ColorDarkOrange)
    tview.Print(screen, fmt.Sprintf("Used %.3f MiB",memUsed), x, y+1, w, tview.AlignRight, tcell.ColorOrange)

    color := setColorGradation(MEMORY_COLOR, len(graph))
    for i, line := range graph {
        tview.Print(screen, line, x, y+2+i, w, tview.AlignRight, color[i])
    }

    l := len(graph)
    tview.Print(screen, constants.LeftDown, x, y+2+l, w, tview.AlignLeft, tcell.NewRGBColor(254, 89, 19))
    tview.Print(screen, constants.RightDown, x, y+2+l, w, tview.AlignRight, tcell.NewRGBColor(254, 89, 19))

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

    // I can't get memory values for a little while after the VM starts up.
    // So I added it to avoid causing panic.
    if max == 0 {
        m.usage[0] = 0.0
    } else {
        m.usage[0] = float64(used * 100 / max)
    }
}

// -------------------------------- Disk --------------------------------
func NewDisk() *Disk {
    return &Disk {
        Box:        tview.NewBox(),
        infos:      []virt.Diskinfo{},
        index:      0,
    }
}


func (d *Disk)AddInfo(info virt.Diskinfo) *Disk {
    d.infos = append(d.infos, info)
    return d
}


func (d *Disk)Draw(screen tcell.Screen) {
    d.Box.DrawForSubclass(screen, d)
    //x, y, w, h := d.GetInnerRect()
    x, y, w, _ := d.GetInnerRect()

    tview.Print(screen, "Disk", x, y, w, tview.AlignCenter, tcell.ColorDarkOrange)

    usage := float64(d.infos[d.index].Allocation) / float64(d.infos[d.index].Capacity)

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

    tview.Print(screen, fmt.Sprintf("[orange]%s [whitesmoke]%d/%d [orange]%s", string(leftTriangle), d.index+1, len(d.infos), string(rightTraiangle)), x, y+1, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("File : %s",d.infos[d.index].Path), x+9, y + 1, w, tview.AlignLeft, tcell.ColorOrange)
    tview.Print(screen, fmt.Sprintf("Volume size : %.2f", float64(d.infos[d.index].Capacity / gibi)), x, y + 1, w, tview.AlignRight, tcell.ColorGhostWhite)
    tview.Print(screen, fmt.Sprintf("Used        : %.2f", float64(d.infos[d.index].Allocation / gibi)), x, y + 2, w, tview.AlignRight, tcell.ColorOrange)
    // draw Bar
    tview.Print(screen, Bar, x, y + 3, w, tview.AlignLeft, tcell.NewRGBColor(80, 80, 80))

    color := setColorGradation(DISK_COLOR, int(usage * float64(w)))
    for j := 0; j< int(usage * float64(w)); j++ {
        tview.Print(screen, "■", x+j, y + 3, w, tview.AlignLeft,color[j])
    }
}


func (d *Disk)MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return d.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
		if !d.InRect(x, y) {
			return false, nil
		}
        px, py, _, _ := d.GetInnerRect()
        if action == tview.MouseLeftClick {
            if y == py+1 {
                if x == px {
                    if 0 < d.index {
                        d.index--
                        consumed = true
                    }
                } else if x == px+6 {
                    if d.index < len(d.infos)-1 {
                        d.index++
                        consumed = true
                    }
                }
            }
        }
        return
    })
}


// -------------------------- Network interface card ---------------------------
func NewNIC() *NIC {
    bwU := [150][500]string{}
    for i := 0; i < 150; i++ {
        for j := 0; j < 500; j++ {
            bwU[i][j] = " "
        }
    }
    bwD := [150][500]string{}
    for i := 0; i < 150; i++ {
        for j := 0; j < 500; j++ {
            bwD[i][j] = " "
        }
    }

    return &NIC {
        Box:                tview.NewBox(),
        bwGraphUp:          bwU,
        bwGraphDown:        bwD,
        ifList:             []NICMember{},
        index:              0,
    }
}


func (n *NIC)Draw(screen tcell.Screen) {
    n.Box.DrawForSubclass(screen, n)
    x, y, w, h := n.GetInnerRect()

    nicStyle := tcell.StyleDefault
    nicStyle = nicStyle.Foreground(tcell.NewRGBColor(20, 161, 156))
    nicStyle = nicStyle.Background(tview.Styles.PrimitiveBackgroundColor)

    var Uploadjudge int64
    var Downloadjudge int64

    graphHeight := int(h/2) - 1
    if graphHeight < 0 {
        graphHeight = 0
    }
    //brailleGradient := float64(100) / float64(graphHeight * 4)

    // Upload Bandwidth
    Uploadjudge = 0
    for i := 0; i < 5; i++ {
        Uploadjudge += n.ifList[n.index].bwUp[i]
    }
    Uploadjudge = Uploadjudge / 3
    if Uploadjudge <= 0 {
        Uploadjudge = 1
    }


    for i := 0; i < w-30; i++ {
        bandwidth := n.ifList[n.index].bwUp[i]
        for j := 0; j <= graphHeight; j++ {
            if bandwidth > int64(float64(Uploadjudge) / float64(graphHeight)) {
                n.bwGraphUp[j][i] = "⣿"
                bandwidth -= int64(float64(Uploadjudge) / float64(graphHeight))
            } else {
                a := int(float64(bandwidth) / (float64(Uploadjudge) / float64(graphHeight*4)))
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

    // Download Bandwidth
    Downloadjudge = 0
    for i := 0; i < 5; i++ {
        Downloadjudge += n.ifList[n.index].bwDown[i]
    }
    Downloadjudge = Downloadjudge / 3
    if Downloadjudge <= 0 {
        Downloadjudge = 1
    }

    for i := 0; i < w-30; i++ {
        bandwidth := n.ifList[n.index].bwDown[i]
        for j := 0; j <= graphHeight; j++ {
            if bandwidth > int64(float64(Downloadjudge) / float64(graphHeight)) {
                n.bwGraphDown[j][i] = "⣿"
                bandwidth -= int64(float64(Downloadjudge) / float64(graphHeight))
            } else {
                a := int(float64(bandwidth) / (float64(Downloadjudge) / float64(graphHeight*4)))
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

    graphUP := []string{}
    for i := 0; i <= graphHeight; i++ {
        tmpLine := ""
        for j := w; j > 0; j-- {
            tmpLine += n.bwGraphUp[graphHeight - i][j]
        }
        graphUP = append(graphUP, tmpLine)
    }

    graphDOWN := []string{}
    for i := 0; i <= graphHeight; i++ {
        tmpLine := ""
        for j := w; j > 0; j-- {
            tmpLine += n.bwGraphDown[i][j]
        }
        graphDOWN = append(graphDOWN, tmpLine)
    }

    tview.Print(screen, "NIC", x, y-1, w, tview.AlignCenter, tcell.NewRGBColor(20, 161, 156))
    tview.Print(screen, constants.LeftTop, x, y-1, w, tview.AlignLeft, tcell.NewRGBColor(20, 161, 156))
    tview.Print(screen, constants.RightTop, x, y-1, w, tview.AlignRight, tcell.NewRGBColor(20, 161, 156))

    for i := y+1; i < y+h-2; i++ {
        screen.SetContent(x+w-30, i, tview.Borders.Vertical, nil, nicStyle)
    }

    //tview.Print("")
    var rateUPText, rateDOWNText string
    var rateUP, rateDOWN int64
    if n.ifList[n.index].bwUp[0] < int64(mega) {
        rateUPText = "KB/s"
        rateUP = int64(kilo)
    } else if n.ifList[n.index].bwUp[0] < int64(giga) {
        rateUPText = "MB/s"
        rateUP = int64(mega)
    } else {
        rateUPText = "GB/s"
        rateUP = int64(giga)
    }
    if n.ifList[n.index].bwDown[0] < int64(mega) {
        rateDOWNText = "KB/s"
        rateDOWN = int64(kilo)
    } else if n.ifList[n.index].bwDown[0] < int64(giga) {
        rateDOWNText = "MB/s"
        rateDOWN = int64(mega)
    } else {
        rateDOWNText = "GB/s"
        rateDOWN = int64(giga)
    }
    tview.Print(screen, fmt.Sprintf("%s %.2f %s", string(constants.UpTraiangle), float64(n.ifList[n.index].bwUp[0]) / float64(rateUP), rateUPText), x+w-28, y+(h/2)-1, 30, tview.AlignLeft, tcell.NewRGBColor(31, 247, 255))
    tview.Print(screen, fmt.Sprintf("%s %.2f %s", string(constants.DownTraiangle), float64(n.ifList[n.index].bwDown[0]) / float64(rateDOWN), rateDOWNText), x+w-28, y+(h/2), 30, tview.AlignLeft, tcell.NewRGBColor(80, 70, 149))
    tview.Print(screen, fmt.Sprintf("[blue]%s [whitesmoke]%d/%d [blue]%s", string(leftTriangle), n.index+1, len(n.ifList), string(rightTraiangle)), x+w-28, y+1, 30, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("NIC Name: [skyblue]%s", n.ifList[n.index].name), x+w-28, y+2, 30, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("MAC Addr: [skyblue]%s", n.ifList[n.index].MACAddr), x+w-28, y+3, 30, tview.AlignLeft, tcell.ColorWhiteSmoke)

    colorUP := setColorGradation(NIC_UP_COLOR, len(graphUP))
    for i, line := range graphUP {
        tview.Print(screen, line, x, y+i, w-30, tview.AlignRight, colorUP[i])
    }
    l := len(graphUP)

    colorDOWN := setColorGradation(NIC_DOWN_COLOR, len(graphDOWN))
    for i, line := range graphDOWN {
        tview.Print(screen, line, x, y+l+i, w-30, tview.AlignRight, colorDOWN[i])
    }
    l += len(graphDOWN)

    tview.Print(screen, constants.LeftDown, x, y+h-1, w, tview.AlignLeft, tcell.NewRGBColor(20, 161, 156))
    tview.Print(screen, constants.RightDown, x, y+h-1, w, tview.AlignRight, tcell.NewRGBColor(20, 161, 156))

    var upperLimitUp string
    var limitUpDivSI int64
    switch {
    case Uploadjudge < int64(kilo):
        upperLimitUp = "B"
        limitUpDivSI = 1
    case Uploadjudge < int64(mega):
        upperLimitUp = "KB"
        limitUpDivSI = int64(kilo)
    case Uploadjudge < int64(giga):
        upperLimitUp = "MB"
        limitUpDivSI = int64(mega)
    default:
        upperLimitUp = "GB"
        limitUpDivSI = int64(giga)
    }
    var upperLimitDown string
    var limitDownDivSI int64
    switch {
    case Downloadjudge < int64(kilo):
        upperLimitDown = "B"
        limitDownDivSI = 1
    case Downloadjudge < int64(mega):
        upperLimitDown = "KB"
        limitDownDivSI = int64(kilo)
    case Downloadjudge < int64(giga):
        upperLimitDown = "MB"
        limitDownDivSI = int64(mega)
    default:
        upperLimitDown = "GB"
        limitDownDivSI = int64(giga)
    }
    var u float64 = float64(Uploadjudge) / float64(limitUpDivSI)
    var d float64 = float64(Downloadjudge) / float64(limitDownDivSI)
    tview.Print(screen, fmt.Sprintf("%.1f %s", u, upperLimitUp), x, y, w, tview.AlignLeft, tcell.NewRGBColor(20, 161, 156))
    tview.Print(screen, fmt.Sprintf("%.1f %s", d, upperLimitDown), x, y+h-2, w, tview.AlignLeft, tcell.NewRGBColor(141, 232, 237))
}


func (n *NIC)MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return n.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
		if !n.InRect(x, y) {
			return false, nil
		}
        px, py, w, _ := n.GetInnerRect()
        if action == tview.MouseLeftClick {
            if y == py+1 {
                if x == px+w-28 {
                    if 0 < n.index {
                        n.index--
                        consumed = true
                    }
                } else if x == px+w-22 {
                    if n.index < len(n.ifList)-1 {
                        n.index++
                        consumed = true
                    }
                }
            }
        }
        return
    })
}


func (n *NIC)AddIF(mac string) *NIC {
    bwUp := [500]int64{}
    for i := 0; i < 500; i++ {
        bwUp[i] = 0
    }
    bwDown := [500]int64{}
    for i := 0; i < 500; i++ {
        bwDown[i] = 0
    }

    n.ifList = append(n.ifList, NICMember{
        name:       operate.GetIFNameByMAC(mac),
        MACAddr:    mac,
        bwUp:       bwUp,
        bwDown:     bwDown,
        oldUp:      0,
        oldDown:    0,
    })
    return n
}


func (n *NIC)Update(dom *libvirt.Domain) {
    for i := range n.ifList {
        txByte, rxByte := virt.GetTrafficByMAC(dom, n.ifList[i].MACAddr)
        if n.ifList[i].oldUp == 0 || n.ifList[i].oldDown == 0 {
            n.ifList[i].oldUp = txByte
            n.ifList[i].oldDown = rxByte
        }
        // Upload
        l := len(n.ifList[i].bwUp)
        _, _, w, _ := n.GetInnerRect()
        if l < w { w = l }
        for j := w-1; j >= 0; j-- {
            n.ifList[i].bwUp[j+1] = n.ifList[i].bwUp[j]
        }
        n.ifList[i].bwUp[0] = txByte - n.ifList[i].oldUp

        // Download
        l = len(n.ifList[i].bwDown)
        _, _, w, _ = n.GetInnerRect()
        if l < w { w = l }
        for j := w-1; j >= 0; j-- {
            n.ifList[i].bwDown[j+1] = n.ifList[i].bwDown[j]
        }
        n.ifList[i].bwDown[0] = rxByte - n.ifList[i].oldDown

        n.ifList[i].oldUp = txByte
        n.ifList[i].oldDown = rxByte
    }
}


func NewVMStatus(app * tview.Application, vm *virt.VM) tview.Primitive{
    flex := tview.NewFlex()
    flex.SetBorder(false)
    vmstatus := tview.NewFlex().SetDirection(tview.FlexRow)

    domInfo, err := vm.Domain.GetInfo()
    if err != nil {
        log.Fatalf("failed to get domain info: %v", err)
    }
    cpu := NewCPU(domInfo.NrVirtCpu)
    mem := NewMemory()
    disk := NewDisk()
    infos := virt.GetDisks(vm.Domain)
    for _, info := range infos {
        disk.AddInfo(info)
    }
    nic := NewNIC()
    for _, mac := range virt.GetNICListMAC(vm.Domain) {
        nic.AddIF(mac)
    }

    vmstatus.AddItem(NewVMInfo(vm.Domain), 5, 1, false)
    flex.AddItem(cpu, 0, 1, false)
    flex.AddItem(mem, 0, 1, false)
    vmstatus.AddItem(flex, 0, 1, false)
    vmstatus.AddItem(disk, 5, 1, false)
    vmstatus.AddItem(nic, 0, 1, false)

    go func() {
        VMStatusUpdate(app, vmstatus, cpu, mem, nic, vm)
    }()

    return vmstatus
}


func VMStatusUpdate(app *tview.Application, flex *tview.Flex, cpu *CPU, mem *Mem, nic *NIC, vm *virt.VM) {
    sec := time.Second
    if b, _ := vm.Domain.IsActive(); !b {
        if flex.GetItemCount() == 4 {
            flex.Clear()
            flex.AddItem(NotUpVM(vm.Name), 0, 1, false)
        }
    }

    time.Sleep(sec)

    var timeCnt uint64 = 0
    for range time.Tick(sec) {
        if timeCnt >= 3 {
            break
        }
        timeCnt++
    }

    var oldUsage uint64
    if b, _ := vm.Domain.IsActive(); b {
        oldUsage, _, _ = virt.GetCPUUsage(vm.Domain)  // cpu
    }

    for range time.Tick(sec) {
        //timeCnt++
        b, _ := vm.Domain.IsActive()
        if b {
            newUsage, cnt, err := virt.GetCPUUsage(vm.Domain)  // cpu
            if err != nil {
                _, cnt = cpu.GetLastStatus()
                newUsage = oldUsage
            }

            max, used := virt.GetMemUsed(vm.Domain)  // memory
            app.QueueUpdateDraw(func() {
                cpu.Update(float64((newUsage - oldUsage) / (uint64(cnt) * 10000000)))  // cpu
                mem.Update(max, used)
                nic.Update(vm.Domain)
            })

            oldUsage = newUsage  //cpu
        } else {
            vm.Status = false
            if flex.GetItemCount() == 4 {
                flex.Clear()
                flex.AddItem(NotUpVM(vm.Name), 0, 1, false)
            }
        }
    }
}
