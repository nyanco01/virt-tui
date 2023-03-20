package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/virt"
)


func cancelTimer(t int) chan string {
    cancel := make(chan string)
    go func() {
        time.Sleep(time.Duration(t) * time.Second)
        cancel <- fmt.Sprintf("During %d sec", t)
    }()
    return cancel
}


func MakeLoading(app *tview.Application, page *tview.Pages, list *tview.List, text string, done chan string) tview.Primitive {
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetTextColor(tcell.ColorYellow)
    view.SetBorder(true)
    view.SetTextAlign(tview.AlignCenter)

    go func() {
        spin := []rune(`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`)
        cnt := 0
        Loop:
            for {
                select {
                case <-done:
                    app.QueueUpdate(func() {
                        page.RemovePage("Loading")
                        app.Sync()
                    })
                    break Loop
                default:
                    
                    viewtext := string(spin[cnt]) + " " + fmt.Sprintf("[white]%s", text)
                    app.QueueUpdateDraw(func() {
                        /*
                        if !view.HasFocus() {
                            app.SetFocus(view)
                            page.ShowPage("Loading")
                        }
                        */
                        view.SetText(viewtext)
                    })
                    cnt++
                    if cnt == len(spin)-1 {
                        cnt = 0
                    }
                    time.Sleep(200 * time.Millisecond)
                }
            }
    }()

    return pageModal(view, len(text)+10, 3)
}


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
        return x, y, width, height
    })

    return bt
}


func MakeNotification(app *tview.Application, page *tview.Pages, list *tview.List, text string) tview.Primitive {
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetText(text)
    view.SetTextAlign(tview.AlignCenter)

    form := tview.NewForm()
    form.SetButtonBackgroundColor(tcell.NewRGBColor(80, 80, 80))
    form.SetButtonsAlign(tview.AlignCenter)
    form.AddButton(" OK ", func() {
        page.RemovePage("Notification")
        app.SetFocus(list)
    })
    flex := tview.NewFlex().SetDirection(tview.FlexRow)
    flex.SetBorder(true)

    flex.AddItem(view, 2, 0, false)
    flex.AddItem(form, 3, 0, true)

    return pageModal(flex, 40, 7)
}


func MakeDeleteVMMenu(app *tview.Application, vm *virt.VM, page *tview.Pages, list *tview.List) tview.Primitive {
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetText(fmt.Sprintf("Virtual machine name to be \n deleted : [red]%s", vm.Name))

    form := tview.NewForm()
    form.SetButtonBackgroundColor(tcell.NewRGBColor(80, 80, 80))
    form.SetButtonsAlign(tview.AlignCenter)
    form.AddButton(" OK ", func() {
        virt.DeleteDomain(vm.Domain)
        page.RemovePage(vm.Name)
        list.RemoveItem(list.GetCurrentItem())
        page.RemovePage("Delete")
        app.SetFocus(list)
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("Delete")
        app.SetFocus(list)
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow)
    flex.SetBorder(true)

    flex.AddItem(view, 2, 0, false)
    flex.AddItem(form, 3, 0, true)

    return pageModal(flex, 40, 7)
}


func MakeOnOffModal(app *tview.Application, vm *virt.VM, page *tview.Pages, list *tview.List) tview.Primitive {

    btStart     := tview.NewButton("Start")
    btShutdown  := tview.NewButton("Shutdown")
    btDestroy   := tview.NewButton("Destroy")
    btReboot    := tview.NewButton("Reboot")
    btEdit      := tview.NewButton("Edit")
    btDelete    := tview.NewButton("Delete")
    btQuit      := tview.NewButton("Quit")
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
        btReboot = SetButtonDefaultStyle(btReboot, tcell.Color56.TrueColor())

        btShutdown.SetSelectedFunc(func() {

            cancel := cancelTimer(10)

            done, e := virt.ShutdownDomain(vm.Domain, cancel)

            page.RemovePage("OnOff")

            if page.HasPage("Loading") {
                page.RemovePage("Loading")
            }
            
            loading := MakeLoading(app, page, list, "VM is shutting down", done)
            page.AddPage("Loading", loading, true, true)
            page.ShowPage("Loading")
            
            //page.AddAndSwitchToPage("Loading", MakeLoading(app, page, list, "VM is shutting down", done), true)

            go func() {
                select {
                case <-done:
                    vm.Status = false
                    app.SetFocus(list)
                case v := <-e:
                    done <- "done"
                    notice := MakeNotification(app, page, list, v)
                    page.AddPage("Notification", notice, true, true)
                    page.ShowPage("Notification")
                    app.SetFocus(notice)
                }
            }()

        })
        btDestroy.SetSelectedFunc(func() {
            _ = vm.Domain.Destroy()
            time.Sleep(time.Millisecond * 500)
            page.RemovePage("OnOff")
            vm.Status = false
            app.SetFocus(list)
        })
        btReboot.SetSelectedFunc(func() {
            _ = vm.Domain.Reboot(libvirt.DOMAIN_REBOOT_ACPI_POWER_BTN)
            time.Sleep(time.Second)
            page.RemovePage("OnOff")
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
        btReboot = SetButtonDefaultStyle(btReboot, tcell.ColorDarkSlateGray)

        btStart.SetSelectedFunc(func() {
            cancel := cancelTimer(20) 

            done, e := virt.StartDomain(vm.Domain, cancel)

            page.RemovePage("OnOff")

            if page.HasPage("Loading") {
                page.RemovePage("Loading")
            }
            
            loading := MakeLoading(app, page, list, "VM is starting", done)
            page.AddPage("Loading", loading, true, true)
            page.ShowPage("Loading")
            
            //page.AddAndSwitchToPage("Loading", MakeLoading(app, page, list, "VM is starting", done), true)

            go func(){
                select {
                case <-done:
                    vm.Status = true
                    page.RemovePage(vm.Name)
                    page.AddAndSwitchToPage(vm.Name, NewVMStatus(app, vm), true)
                    list.SetItemText(list.GetCurrentItem(), vm.Name, "")
                    app.SetFocus(list)
                case v := <-e:
                    done <- "done"
                    notice := MakeNotification(app, page, list, v)
                    page.AddPage("Notification", notice, true, true)
                    page.ShowPage("Notification")
                    app.SetFocus(notice)
                }
            }()
        })
        btDelete.SetSelectedFunc(func() {
            delPage := MakeDeleteVMMenu(app, vm, page, list)
            page.AddPage("Delete", delPage, true, true)
            page.RemovePage("OnOff")
            page.ShowPage("Delete")
            app.SetFocus(delPage)
        })
        btEdit.SetSelectedFunc(func() {
            page.RemovePage("OnOff")
            if page.HasPage("Edit") {
                page.RemovePage("Edit")
            }
            page.AddAndSwitchToPage("Edit", MakeVMEditMenu(app, vm, list, page), true)
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

