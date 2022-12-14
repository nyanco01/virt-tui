package tui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
	//libvirtxml "libvirt.org/go/libvirtxml"
)


type VMEdit struct {
    *tview.Box
    *libvirt.Domain
    name            string
    items           []virt.EditItem
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
        items:          []virt.EditItem{},
        lineOfset:      0,
        lineOfsetMax:   0,
    }
}


func (e *VMEdit) ClearItems() *VMEdit {
    e.items = []virt.EditItem{}
    return e
}


func (e *VMEdit) SetItemList() *VMEdit {
    e.items = virt.GetDomainItems(e.Domain)
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
        case virt.ItemCPU:
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
        case virt.ItemMemory:
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
        case virt.ItemDisk:
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
        case virt.ItemController:
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
        case virt.ItemInterface:
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
        case virt.ItemSerial:
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
        case virt.ItemConsole:
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
        case virt.ItemInput:
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
        case virt.ItemGraphics:
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
        case virt.ItemVideo:
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
        px, py, w, _ := e.GetInnerRect()
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
        case tview.MouseLeftClick:
            if px+(w/2)+1 <= x && py+1 == y {
                if e.addIfaceFunc != nil {
                    e.addIfaceFunc()
                    consumed = true
                }
            }
        }
        return
    })
}

func MakeAddIfaceMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Add NIC Menu")
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()

    form.AddDropDown("Network bridge interface", operate.ListBridgeIF(), 0, nil)
    form.AddButton("Create", func() {
        _, iface := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
        err := virt.CreateAddNIC(vm.Domain, iface)
        if err != nil {
            if virtErr, ok := err.(libvirt.Error); ok {
                view.SetText(virtErr.Message)
                view.SetTextColor(tcell.ColorRed)
            } else {
                log.Fatalf("failed to add interface by %s: %v", vm.Name, err)
            }
        } else {
            edit.ClearItems()
            edit.SetItemList()
            page.SwitchToPage("Edit")
            page.RemovePage("AddIfaceMenu")
        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("AddIfaceMenu")
    })

    flex.
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return pageModal(flex, 40, 10)
}


func SetAddFunc(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) {
    edit.SetAddIfaceFunc(func() {
        modal := MakeAddIfaceMenu(app, vm, page, edit)
        page.AddPage("AddIfaceMenu", modal, true, true)
        page.ShowPage("AddIfaceMenu")
        app.SetFocus(modal)
    })
}

func MakeVMEditMenu(app *tview.Application, vm *virt.VM, list *tview.List, page *tview.Pages) *VMEdit {
    edit := NewVMEdit(vm.Domain)
    edit.SetItemList()
    SetAddFunc(app, vm, page, edit)

    return edit
}

