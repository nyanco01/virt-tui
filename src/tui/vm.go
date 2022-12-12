package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/virt"
)


//Dedicated Modal for placing specific Primitive items inside.
func pageModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		    AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
	    	AddItem(nil, 0, 1, false)
}


func SetButtonDefaultStyle(bt *tview.Button, color tcell.Color) *tview.Button {
    bt.SetBackgroundColor(tcell.ColorBlack)
    bt.SetBackgroundColorActivated(tcell.ColorBlack)
    bt.SetBorder(false)
    bt.SetLabelColor(tcell.ColorWhiteSmoke)
    bt.SetLabelColorActivated(color)
    bt.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        for i := y; i < y+height; i++ {
            screen.SetContent(x, i, '▐', nil, tcell.StyleDefault.Foreground(color))
        }
        /*
        if bt.HasFocus() {
            x, y, _, h := bt.GetInnerRect()
            for i := y; i <= y+h; i++ {
                screen.SetContent(x, i, ' ', nil, tcell.StyleDefault.Background(color))
            }
        }
        */
        return x, y, width, height
    })

    return bt
}


func MakeOnOffModal(app *tview.Application, vm *virt.VM, page *tview.Pages, list *tview.List) tview.Primitive {

    btStart     := tview.NewButton("Start")
    btShutdown  := tview.NewButton("Shutdown")
    btDestroy   := tview.NewButton("Destroy")
    btReboot    := tview.NewButton("Reboot")
    btEdit      := tview.NewButton("Edit")
    btDelete    := tview.NewButton("Delete")
    btQuit      := tview.NewButton("Quit")
    btReboot = SetButtonDefaultStyle(btReboot, tcell.ColorDarkSlateGray)
    btQuit = SetButtonDefaultStyle(btQuit, tcell.Color80)

    btQuit.SetSelectedFunc(func() {
        page.RemovePage("OnOff")
        page.SwitchToPage(vm.Name)
        app.SetFocus(list)
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow)
    flex.SetBorder(true).SetTitle(fmt.Sprintf("%s", vm.Name))
    flex.AddItem(btStart, 3, 0, true)
    flex.AddItem(btShutdown, 3, 0 ,false)
    flex.AddItem(btDestroy, 3, 0, false)
    flex.AddItem(btReboot, 3, 0, false)
    flex.AddItem(btEdit, 3, 0, false)
    flex.AddItem(btDelete, 3, 0, false)
    flex.AddItem(btQuit, 3, 0, false)

    flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        cnt := 0
        for i := 0; i < flex.GetItemCount(); i++ {
            p := flex.GetItem(i)
            if p.HasFocus() {
                cnt = i
            }
        }
        switch event.Key() {
        case tcell.KeyTab, tcell.KeyDown:
            if cnt < (flex.GetItemCount() - 1) {
                app.SetFocus(flex.GetItem(cnt+1))
            }
            return nil
        case tcell.KeyBacktab, tcell.KeyUp:
            if cnt > 0 {
                app.SetFocus(flex.GetItem(cnt-1))
            }
            return nil
        case tcell.KeyEsc:
            page.RemovePage("OnOff")
            page.SwitchToPage(vm.Name)
            app.SetFocus(list)
            return nil
        }
        return event
    })
    if vm.Status {
        // Disable button
        btStart = SetButtonDefaultStyle(btStart, tcell.ColorDarkSlateGray)
        btDelete = SetButtonDefaultStyle(btDelete, tcell.ColorDarkSlateGray)
        btEdit = SetButtonDefaultStyle(btEdit, tcell.ColorDarkSlateGray)
        // Enable button
        btShutdown = SetButtonDefaultStyle(btShutdown, tcell.ColorOrangeRed)
        btDestroy = SetButtonDefaultStyle(btDestroy, tcell.ColorRed)

        btShutdown.SetSelectedFunc(func() {
            _ = vm.Domain.Shutdown()
            time.Sleep(time.Millisecond * 500)
            page.RemovePage("OnOff")
            vm.Status = false
            app.SetFocus(list)
        })
        btDestroy.SetSelectedFunc(func() {
            _ = vm.Domain.Destroy()
            time.Sleep(time.Millisecond * 500)
            page.RemovePage("OnOff")
            vm.Status = false
            app.SetFocus(list)
        })

    } else {
        // Enable button
        btStart = SetButtonDefaultStyle(btStart, tcell.ColorGreen)
        btDelete = SetButtonDefaultStyle(btDelete, tcell.ColorRed)
        btEdit = SetButtonDefaultStyle(btEdit, tcell.Color82)
        // Disable button
        btShutdown = SetButtonDefaultStyle(btShutdown, tcell.ColorDarkSlateGray)
        btDestroy = SetButtonDefaultStyle(btDestroy, tcell.ColorDarkSlateGray)

        btStart.SetSelectedFunc(func() {
            virt.StartDomain(vm.Domain)
            page.RemovePage("OnOff")
            vm.Status = true
            page.RemovePage(vm.Name)
            page.AddAndSwitchToPage(vm.Name, NewVMStatus(app, vm), true)
            list.SetItemText(list.GetCurrentItem(), vm.Name, "")
            app.SetFocus(list)
        })
        btDelete.SetSelectedFunc(func() {
            virt.DeleteDomain(vm.Domain)
            page.RemovePage(vm.Name)
            list.RemoveItem(list.GetCurrentItem())
            app.SetFocus(list)
        })
        btEdit.SetSelectedFunc(func() {
            page.RemovePage("OnOff")
            page.AddAndSwitchToPage("Edit", NewVMEditMenu(app, vm, list, page), true)
            app.SetFocus(page)
        })
    }
    return pageModal(flex, 30, 23)
}


func MakeVMMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))
    list.SetShortcutColor(tcell.Color214)

    VirtualMachineStatus = map[string]bool{}
    virt.VMStatus = map[string]*virt.VM{}

    vms := virt.LookupVMs(con)
    for i, vm := range vms {
        if vm.Status {
            list.AddItem(vm.Name, "", rune(i+'0'), nil)
            page.AddPage(vm.Name, NewVMStatus(app, vm), true, true)
            VirtualMachineStatus[vm.Name] = true
            virt.VMStatus[vm.Name] = vm
        } else {
            list.AddItem(vm.Name, "", rune(i)+'0', nil)
            page.AddPage(vm.Name, NewVMStatus(app, vm), true, true)
            VirtualMachineStatus[vm.Name] = false
            virt.VMStatus[vm.Name] = vm
        }
    }
    app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
        for i := 0; i < list.GetItemCount(); i++ {
            main, second := list.GetItemText(i)
            if _, ok := virt.VMStatus[main]; !ok {
                return true
            }
            if !virt.VMStatus[main].Status {
                if second == "" {
                    list.SetItemText(i, main, "shutdown")
                }
            } else {
                if second == "shutdown" {
                    list.SetItemText(i, main, "")
                }
            }
        }
        return false
    })

    // Displays the page corresponding to the selected item
    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

    list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
        var vmCrnt *virt.VM = virt.VMStatus[s1]

        modal := MakeOnOffModal(app, vmCrnt, page, list)
        if page.HasPage("OnOff") {
            page.RemovePage("OnOff")
        }
        page.AddPage("OnOff", modal, true, true)
        page.ShowPage("OnOff")
        app.SetFocus(modal)
    })
    list.SetDoneFunc(func() {
        a, _ := list.GetItemText(list.GetCurrentItem())
        page.SwitchToPage(a)
    })

    btCreate := tview.NewButton("Create")
    btCreate.SetBackgroundColor(tcell.Color202)

    // If the last item on the list is selected, toggle to move focus to the button
    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if (event.Key() == tcell.KeyTab) || (event.Key() == tcell.KeyDown) {
            if (list.GetItemCount() - 1) == list.GetCurrentItem() {
                app.SetFocus(btCreate)
                return nil
            }
        }
        return event
    })

    // Toggling when the focus is on a button focuses the list
    btCreate.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyTab, tcell.KeyDown:
            list.SetCurrentItem(0)
            app.SetFocus(list)
            return nil
        case tcell.KeyBacktab, tcell.KeyUp:
            list.SetCurrentItem(list.GetItemCount() - 1)
            app.SetFocus(list)
            return nil
        }
        return event
    })

    btCreate.SetSelectedFunc(func() {
        modal := MakeVMCreate(app, con, page, list)
        if page.HasPage("OnOff") {
            page.RemovePage("OnOff")
        }
        page.AddPage("Create", modal, true, true)
        page.ShowPage("Create")
        app.SetFocus(modal)
    })

    // Check if the number of VMs is not zero
    if list.GetItemCount() != 0 {
        main, _ := list.GetItemText(list.GetCurrentItem())
        page.SwitchToPage(main)
    }
    
    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1, true).
        AddItem(btCreate, 5, 0, false)

    return flex
}


func MakePages(app *tview.Application) *tview.Pages {
    page := tview.NewPages()
    page.SetBorder(false)

    return page
}


func MakeVMUI(app *tview.Application, con *libvirt.Connect) *tview.Flex {
    flex := tview.NewFlex()

    page := MakePages(app)
    menu := MakeVMMenu(app, con, page)

    _, _, w, _ := menu.GetInnerRect()
    flex.AddItem(menu, w + 5, 0, true)
    flex.AddItem(page, 0, 1, false)

    return flex
}

