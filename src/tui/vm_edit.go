package tui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/go/libvirtxml"
)


type VMEdit struct {
    *tview.Box
    *libvirt.Domain
    name            string
    items           []EditItem
    addDiskFunc     func()
    addIfaceFunc    func()
    lineOfset       int
    lineOfsetMax    int
}




func NewVMEdit(dom *libvirt.Domain) *VMEdit {
    n, err := dom.GetName()
    if err != nil {
        log.Fatalf("failed to get domain name: %v", err)
    }
    return &VMEdit{
        Box:            tview.NewBox(),
        Domain:         dom,
        name:           n,
        items:          []EditItem{},
        lineOfset:      0,
        lineOfsetMax:   0,
    }
}


func (e *VMEdit) SetItemList() *VMEdit {
    e.items = GetDomainItems(e.Domain)
    return e
}


func (e *VMEdit) SetAddIfaceFunc(handler func()) *VMEdit {
    e.addIfaceFunc = handler
    return e
}


func (e *VMEdit) SetAddDiskFunc(handler func()) *VMEdit {
    e.addDiskFunc = handler
    return e
}


func (e *VMEdit) Draw(screen tcell.Screen) {
    e.Box.DrawForSubclass(screen, e)
    x, y, w, h := e.GetInnerRect()
    for i := x; i <= x+w; i++ {
        screen.SetContent(i, y, ' ', nil, tcell.StyleDefault.Background(tcell.Color42))
    }
    tview.Print(screen, e.name, x, y, w, tview.AlignCenter, tcell.ColorDarkSlateGray)

    tview.Print(screen,"[+] Disk", x, y+1, w/2, tview.AlignCenter, tcell.Color226)
    tview.Print(screen,"[+] NIC", x+(w/2)+1, y+1, w/2, tview.AlignCenter, tcell.Color45)
    for i := x; i <= x+w; i++ {
        screen.SetContent(i, y+2, ' ', nil, tcell.StyleDefault.Background(tcell.ColorDarkSlateGray))
    }
    d := string(downTraiangle)
    tview.Print(screen, fmt.Sprintf("%s%s Items %s%s", d, d, d, d), x, y+2, w, tview.AlignCenter, tcell.ColorWhite)
    fullHeight := (len(e.items))*5
    e.lineOfsetMax = fullHeight - (h - 2)

    cnt := e.lineOfset
    for i := y+3; i <= y+h; i++ {
        itemIndex := (i - (y+2) + e.lineOfset) / 5
        if itemIndex < 0 {
            itemIndex = 0
        } else if itemIndex >= len(e.items) {
            break
        }
        switch v := e.items[itemIndex].(type) {
        case ItemCPU:
            colorMain := tcell.ColorGreen.TrueColor()
            colorSub := tcell.ColorLightGreen.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("VCPU Count: [whitesmoke]%d", v.Number), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("PlaceMent: [whitesmoke]%s", v.PlaceMent), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("CPU Set: [whitesmoke]%s", v.CPUSet), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Mode: [whitesmoke]%s", v.Mode), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemMemory:
            colorMain := tcell.ColorOrange.TrueColor()
            colorSub := tcell.Color215.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 1:
                tview.Print(screen, fmt.Sprintf("Allocation Size: [whitesmoke]%d %s", v.Size, v.SizeSI), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Max Allocation Size: [whitesmoke]%d %s", v.MaxSize, v.MaxSizeSI), x+3, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Current Size: [whitesmoke]%d %s", v.CurrentMemory, v.CurrentMemorySI), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemDisk:
            colorMain := tcell.ColorYellow.TrueColor()
            colorSub := tcell.Color230.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Device Type: [whitesmoke]%s", v.Device), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Format: [whitesmoke]%s", v.ImgType), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Path: [whitesmoke]%s", v.Path), x+30, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Bus: [whitesmoke]%s", v.Bus), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemController:
            colorMain := tcell.Color46.TrueColor()
            colorSub := tcell.Color155.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Controller Type: [whitesmoke]%s", v.ControllerType), x+3, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Controller Model: [whitesmoke]%s", v.Model), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemInterface:
            colorMain := tcell.Color33.TrueColor()
            colorSub := tcell.Color81.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Interface Type: [whitesmoke]%s", v.IfType), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Driver: [whitesmoke]%s", v.Driver), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                if v.IfType == "bridge" {
                    tview.Print(screen, fmt.Sprintf("Bridge Source: [whitesmoke]%s", v.Source), x+3, i, w, tview.AlignLeft, colorSub)
                } else {
                    tview.Print(screen, fmt.Sprintf("Model: [whitesmoke]%s", v.Model), x+3, i, w, tview.AlignLeft, colorSub)
                }
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemSerial:
            colorMain := tcell.Color240.TrueColor()
            colorSub := tcell.Color244.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Serial Target Type: [whitesmoke]%s", v.TargetType), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemConsole:
            colorMain := tcell.Color240.TrueColor()
            colorSub := tcell.Color244.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Console Target Type: [whitesmoke]%s", v.TargetType), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemInput:
            colorMain := tcell.Color35.TrueColor()
            colorSub := tcell.Color43.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Input Type: [whitesmoke]%s", v.InputType), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Bus: [whitesmoke]%s", v.InputType), x+30, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemGraphics:
            colorMain := tcell.Color133.TrueColor()
            colorSub := tcell.Color140.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("GraphicsType Type: [whitesmoke]%s", v.GraphicsType), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Port: [whitesmoke]%d", v.Port), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Listen Address: [whitesmoke]%s", v.ListemAddress), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case ItemVideo:
            colorMain := tcell.Color133.TrueColor()
            colorSub := tcell.Color140.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Video Model Type: [whitesmoke]%s", v.ModelType), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("VRAM Size: [whitesmoke]%d", v.VRAM), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Device Address: [whitesmoke]%s", v.DeviceAddress), x+3, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        }
        cnt++
    }
}


func (e *VMEdit)MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return e.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
        if !e.InRect(x, y) {
            return false, nil
        }

        switch action {
        case tview.MouseScrollUp:
            if e.lineOfset > 0 {
                e.lineOfset--
                consumed = true
            }
        case tview.MouseScrollDown:
            if e.lineOfset < e.lineOfsetMax {
                e.lineOfset++
                consumed = true
            }

        }
        return
    })
}


func NewVMEditMenu(app *tview.Application, vm *virt.VM, list *tview.List, page *tview.Pages) *VMEdit {
    edit := NewVMEdit(vm.Domain)
    edit.SetItemList()

    return edit
}



func GetDomainItems(dom * libvirt.Domain) []EditItem {
    xml, err := dom.GetXMLDesc(0 | libvirt.DOMAIN_XML_INACTIVE)
    if err != nil {
        log.Fatalf("failed to get xml: %v", err)
    }
    var domXML libvirtxml.Domain
    domXML.Unmarshal(xml)
    var items []EditItem
    items = append(items, ItemCPU{
        Number:         domXML.VCPU.Value,
        PlaceMent:      domXML.VCPU.Placement,
        CPUSet:         domXML.VCPU.CPUSet,
        Mode:           domXML.CPU.Mode,
    })
    maxMem := uint(0)
    maxMemSI := ""
    curMem := uint(0)
    curMemSI := ""
    if domXML.MaximumMemory != nil {
        maxMem = domXML.MaximumMemory.Value
        maxMemSI = domXML.MaximumMemory.Unit
    }
    if domXML.CurrentMemory != nil {
        curMem = domXML.CurrentMemory.Value
        curMemSI = domXML.CurrentMemory.Unit
    }
    items = append(items, ItemMemory{
        Size:               domXML.Memory.Value,
        SizeSI:             domXML.Memory.Unit,
        MaxSize:            maxMem,
        MaxSizeSI:          maxMemSI,
        CurrentMemory:      curMem,
        CurrentMemorySI:    curMemSI,
    })
    for _, disk := range domXML.Devices.Disks {
        p := ""
        if disk.Source != nil {
            p = disk.Source.File.File
        }
        items = append(items, ItemDisk{
            Path:       p,
            Device:     disk.Device,
            ImgType:    disk.Driver.Type,
            Bus:        disk.Target.Bus,
        })
    }
    for _, cntl := range domXML.Devices.Controllers {
        items = append(items, ItemController{
            ControllerType:     cntl.Type,
            Model:              cntl.Model,
        })
    }
    for _, iface := range domXML.Devices.Interfaces {
        d := ""
        s := ""
        t := ""
        if iface.Driver != nil {
            d = iface.Driver.Name
            t = "hostdev"
        }
        if iface.Source != nil {
            s = iface.Source.Bridge.Bridge
            t = "bridge"
        }
        items = append(items, ItemInterface{
            IfType:     t,
            Driver:     d,
            Source:     s,
            Model:      iface.Model.Type,
        })
    }
    for _, serial := range domXML.Devices.Serials {
        items = append(items, ItemSerial{
            TargetType: serial.Target.Type,
        })
    }
    for _, console := range domXML.Devices.Consoles {
        items = append(items, ItemConsole{
            TargetType: console.Target.Type,
        })
    }
    for _, input := range domXML.Devices.Inputs {
        items = append(items, ItemInput{
            InputType:  input.Type,
            Bus:        input.Bus,
        })
    }
    for _, graphics := range domXML.Devices.Graphics {
        t := ""
        p := 0
        l := ""
        if graphics.VNC != nil {
            t = "vnc"
            p = graphics.VNC.Port
            l = graphics.VNC.Listen
        }
        if graphics.RDP != nil {
            t = "rdp"
            p = graphics.RDP.Port
            l = graphics.RDP.Listen
        }
        if graphics.Spice != nil {
            t = "spice"
            p = graphics.Spice.Port
            l = ""
        }
        items = append(items, ItemGraphics{
            GraphicsType:   t,
            Port:           p,
            ListemAddress:  l,
        })
    }
    for _, video := range domXML.Devices.Videos {
        a := ""
        if video.Address.PCI != nil {
            a = "pci"
        }
        items = append(items, ItemVideo{
            ModelType:      video.Model.Type,
            VRAM:           video.Model.VRam,
            DeviceAddress:  a,
        })
    }
    return items
}

