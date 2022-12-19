package tui

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
)


type VMEdit struct {
    *tview.Box
    *libvirt.Domain
    name            string
    items           []virt.EditItem
    diskItemCount   int
    ifaceItemCount  int
    addDiskFunc     func()
    addIfaceFunc    func()
    lineOfset       int
    lineOfsetMax    int
    clickItemIndex  int
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
        diskItemCount:  0,
        ifaceItemCount: 0,
        lineOfset:      0,
        lineOfsetMax:   0,
        clickItemIndex: -1,
    }
}


func (e *VMEdit) ClearItems() *VMEdit {
    e.items = []virt.EditItem{}
    return e
}


func (e *VMEdit) SetItemList() *VMEdit {
    e.items, e.diskItemCount, e.ifaceItemCount = virt.GetDomainItems(e.Domain)
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
        if cnt % 5 != 4 {
            if itemIndex == e.clickItemIndex {
                for k := x+1; k <= x+w-2; k++ {
                    screen.SetContent(k, i, ' ', nil, tcell.StyleDefault.Background(tcell.NewRGBColor(40, 40, 40).TrueColor()))
                }
            }
        }
        switch v := e.items[itemIndex].(type) {
        case *virt.ItemCPU:
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
        case *virt.ItemMemory:
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
        case *virt.ItemDisk:
            colorMain := tcell.ColorYellow.TrueColor()
            colorSub := tcell.Color230.TrueColor()
            switch cnt % 5 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Type: [whitesmoke]%s", v.GetItemType()), x+3, i, w, tview.AlignLeft, colorSub)
            case 1:
                tview.Print(screen, fmt.Sprintf("Path: [whitesmoke]%s", v.Path), x+3, i, w, tview.AlignLeft, colorSub)
            case 2:
                tview.Print(screen, fmt.Sprintf("Device Type: [whitesmoke]%s", v.Device), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Target: [whitesmoke]%s", v.Target), x+30, i, w, tview.AlignLeft, colorSub)
            case 3:
                tview.Print(screen, fmt.Sprintf("Bus: [whitesmoke]%s", v.Bus), x+3, i, w, tview.AlignLeft, colorSub)
                tview.Print(screen, fmt.Sprintf("Format: [whitesmoke]%s", v.ImgType), x+30, i, w, tview.AlignLeft, colorSub)
            }
            if cnt % 5 != 4 {
                screen.SetContent(x+1, i, '▐', nil, tcell.StyleDefault.Foreground(colorMain))
            }
        case *virt.ItemController:
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
        case *virt.ItemInterface:
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
        case *virt.ItemSerial:
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
        case *virt.ItemConsole:
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
        case *virt.ItemInput:
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
        case *virt.ItemGraphics:
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
        case *virt.ItemVideo:
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
        var itemSpacers = true
        if (y - 3 + e.lineOfset) % 5 == 0 {
            itemSpacers = true
        } else {
            itemSpacers = false
        }
        px, py, w, _ := e.GetInnerRect()
        switch action {
        case tview.MouseScrollUp:
            if e.lineOfset > 0 {
                e.clickItemIndex = -1
                e.lineOfset--
                consumed = true
            }
        case tview.MouseScrollDown:
            if e.lineOfset < e.lineOfsetMax {
                e.clickItemIndex = -1
                e.lineOfset++
                consumed = true
            }
        case tview.MouseLeftClick:
            if px+(w/2)+1 <= x && py+1 == y {
                if e.addIfaceFunc != nil {
                    e.addIfaceFunc()
                    consumed = true
                }
            } else if x <= px+(w/2) && py+1 == y {
                if e.addDiskFunc != nil {
                    e.addDiskFunc()
                    consumed = true
                }
            } else if py+2 < y && ! itemSpacers {
                e.clickItemIndex = (y - 3 +e.lineOfset) / 5
                if len(e.items)-1 < e.clickItemIndex {
                    return
                }
                if f := e.items[e.clickItemIndex].GetSelectedFunc(); f != nil {
                    f()
                }
            }
        }
        return
    })
}


func MakeItemCPUEditMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("CPU Edit")
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()
    con, err := vm.Domain.DomainGetConnect()
    cpuNum, err := virt.GetCurrentCPUNum(vm.Domain)
    if err != nil {
        if virtErr, ok := err.(libvirt.Error); ok {
            view.SetText(virtErr.Message)
            view.SetTextColor(tcell.ColorRed)
        } else {
            log.Fatalf("failed to get connect: %v", err)
        }
    }
    maxCPUs, _ := virt.GetNodeMax(con)
    optionCPU := []string{}
    currentNum := 0
    for i := 1; i <= maxCPUs; i++ {
        optionCPU = append(optionCPU, strconv.Itoa(i))
        if i == cpuNum {
            currentNum = i-1
        }
    }
    form.AddDropDown("CPU number", optionCPU, currentNum, nil)

    form.AddButton("OK", func() {
        _, cpu := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
        c, _ := strconv.Atoi(cpu)
        err = virt.DomainEditCPU(vm.Domain, uint(c))
        if err != nil {
            if virtErr, ok := err.(libvirt.Error); ok {
                view.SetText(virtErr.Message)
                view.SetTextColor(tcell.ColorRed)
            } else {
                log.Fatalf("failed to edit CPU: %v", err)
            }
        } else {
            edit.ClearItems()
            edit.SetItemList()
            edit.SetItemsFunc(app, vm, page)
            page.SwitchToPage("Edit")
            page.RemovePage("Item CPU")
        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("Item CPU")
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 2, 0, false)

    return pageModal(flex, 40, 10)
}


func SetItemCPUFunc(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) func() {
    return func() {
        modal := MakeItemCPUEditMenu(app, vm, page, edit)
        page.AddPage("Item CPU", modal, true, true)
        page.ShowPage("Item CPU")
        app.SetFocus(modal)
    }
}


func MakeItemMemoryEditMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Memory Edit")
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()
    con, err := vm.Domain.DomainGetConnect()
    memSize, err := virt.GetCurrentMemSize(vm.Domain)
    if err != nil {
        if virtErr, ok := err.(libvirt.Error); ok {
            view.SetText(virtErr.Message)
            view.SetTextColor(tcell.ColorRed)
        } else {
            log.Fatalf("failed to get connect: %v", err)
        }
    }
    _, maxMem := virt.GetNodeMax(con)
    form.AddInputField(fmt.Sprintf("Memory Size [orange](max. %d KB)", maxMem), strconv.Itoa(int(memSize)), 15, InputFieldPositiveInteger, nil)
    form.AddButton("OK", func() {
        mem := form.GetFormItem(0).(*tview.InputField).GetText()
        m, _ := strconv.Atoi(mem)
        if mem == "" {
            view.SetText("A blank character has been entered.")
            view.SetTextColor(tcell.ColorRed)
        } else if uint64(m) > maxMem {
            view.SetText("The maximum memory capacity of the host machine has been exceeded.")
            view.SetTextColor(tcell.ColorRed)
        } else {
            err = virt.DomainEditMemory(vm.Domain, uint(m))
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to edit memory : %v", err)
                }
            } else {
                edit.ClearItems()
                edit.SetItemList()
                edit.SetItemsFunc(app, vm, page)
                page.SwitchToPage("Edit")
                page.RemovePage("Item Memory")
            }
        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("Item Memory")
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 2, 0, false)

    return pageModal(flex, 60, 10)
}


func SetItemMemFunc(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) func() {
    return func() {
        modal := MakeItemMemoryEditMenu(app, vm, page, edit)
        page.AddPage("Item Memory", modal, true, true)
        page.ShowPage("Item Memory")
        app.SetFocus(modal)
    }
}


func MakeItemDiskDeleteMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit, xml string) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Delete Disk")
    viewInfo := tview.NewTextView().SetDynamicColors(true)
    viewInfo.SetText(fmt.Sprintf("[white]Delete Disk by [orange]%s", virt.GetDiskTarget(xml)))
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()
    form.AddButton("OK", func() {
        if edit.diskItemCount == 1 {
            view.SetText("At least one disk is required.")
            view.SetTextColor(tcell.ColorRed)
        } else {
            if err := virt.DomainDeleteDisk(vm.Domain, xml); err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to delete disk: %v", err)
                }
            } else {
                edit.ClearItems()
                edit.SetItemList()
                edit.SetItemsFunc(app, vm, page)
                page.SwitchToPage("Edit")
                page.RemovePage("Item Disk")
            }

        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("Item Disk")
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(viewInfo, 1, 0, false).
        AddItem(form, 0, 1, true).
        AddItem(view, 2, 0, false)
    
    return pageModal(flex, 40, 8)
}


func SetItemDiskFunc(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit, xml string) func() {
    return func() {
        modal := MakeItemDiskDeleteMenu(app, vm, page, edit, xml)
        page.AddPage("Item Disk", modal, true, true)
        page.ShowPage("Item Disk")
        app.SetFocus(modal)
    }
}


func (e *VMEdit)SetItemsFunc(app *tview.Application, vm *virt.VM, page *tview.Pages) *VMEdit {
    for _, item := range e.items {
        switch v := item.(type) {
        case *virt.ItemCPU:
            v.SetSelectedFunc(SetItemCPUFunc(app, vm, page, e))
        case *virt.ItemMemory:
            v.SetSelectedFunc(SetItemMemFunc(app, vm, page, e))
        case *virt.ItemDisk:
            v.SetSelectedFunc(SetItemDiskFunc(app, vm, page, e, v.ItemXML))
        }
    }
    return e
}


func MakeAddDiskMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Add Disk Menu")
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()
    con, err := vm.Domain.DomainGetConnect()
    if err != nil {
        if virtErr, ok := err.(libvirt.Error); ok {
            view.SetText(virtErr.Message)
            view.SetTextColor(tcell.ColorRed)
        } else {
            log.Fatalf("failed to get connect for %s: %v", vm.Name, err)
        }
    }
    pools, err := virt.GetPoolNameList(con)
    if err != nil {
        if virtErr, ok := err.(libvirt.Error); ok {
            view.SetText(virtErr.Message)
            view.SetTextColor(tcell.ColorRed)
        } else {
            log.Fatalf("failed to get pool name: %v", err)
        }
    }

    ddVols := tview.NewDropDown().SetLabel("Disk Name")

    ddPools := tview.NewDropDown().SetLabel("Pool Name")
    ddPools.SetOptions(pools, func(text string, index int) {
        disks := virt.GetNonAttachDiskByPool(con, text)
        if len(disks) == 0 {
            disks = append(disks, "No Items")
        }
        ddVols.SetOptions(disks, nil)
    })
    if len(pools) != 0 {
        ddPools.SetCurrentOption(0)
    }
    form.AddFormItem(ddPools)
    form.AddFormItem(ddVols)

    form.AddButton("Add", func() {
        _, diskPath := ddVols.GetCurrentOption()
        if diskPath == "" || diskPath == "No Items" {
            view.SetText("Disk is not selected")
            view.SetTextColor(tcell.ColorOrange.TrueColor())
        } else {
            if operate.CheckIsDisk(diskPath) {
                err = virt.DomainAddDisk(vm.Domain, diskPath)
                if err != nil {
                    if virtErr, ok := err.(libvirt.Error); ok {
                        view.SetText(virtErr.Message)
                        view.SetTextColor(tcell.ColorRed)
                    } else {
                        log.Fatalf("failed to add Disk by %s: %v", vm.Name, err)
                    }
                } else {
                    edit.ClearItems()
                    edit.SetItemList()
                    edit.SetItemsFunc(app, vm, page)
                    page.SwitchToPage("Edit")
                    page.RemovePage("AddDiskMenu")
                }
            } else {
                view.SetText("The selected file is not a disk format.")
                view.SetTextColor(tcell.ColorRed.TrueColor())
            }
        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("AddDiskMenu")
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 2, 0, false)
    
    return pageModal(flex, 60, 11)
}


func MakeAddIfaceMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Add NIC Menu")
    view := tview.NewTextView()
    view.SetTextAlign(tview.AlignCenter)
    form := tview.NewForm()

    form.AddDropDown("Network bridge interface", operate.ListBridgeIF(), 0, nil)
    form.AddButton("Add", func() {
        _, iface := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
        err := virt.DomainAddNIC(vm.Domain, iface)
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
            edit.SetItemsFunc(app, vm, page)
            page.SwitchToPage("Edit")
            page.RemovePage("AddIfaceMenu")
        }
    })
    form.AddButton("Cancel", func() {
        page.SwitchToPage("Edit")
        page.RemovePage("AddIfaceMenu")
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 2, 0, false)

    return pageModal(flex, 40, 10)
}


func SetAddFunc(app *tview.Application, vm *virt.VM, page *tview.Pages, edit *VMEdit) {
    edit.SetAddIfaceFunc(func() {
        modal := MakeAddIfaceMenu(app, vm, page, edit)
        page.AddPage("AddIfaceMenu", modal, true, true)
        page.ShowPage("AddIfaceMenu")
        app.SetFocus(modal)
    })
    edit.SetAddDiskFunc(func() {
        modal := MakeAddDiskMenu(app, vm, page, edit)
        page.AddPage("AddDiskMenu", modal, true, true)
        page.ShowPage("AddDiskMenu")
        app.SetFocus(modal)
    })
}

func MakeVMEditMenu(app *tview.Application, vm *virt.VM, list *tview.List, page *tview.Pages) *VMEdit {
    edit := NewVMEdit(vm.Domain)
    edit.SetItemList()
    SetAddFunc(app, vm, page, edit)
    edit.SetItemsFunc(app, vm, page)

    return edit
}

