package tui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

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

func CreateOnOffModal(app *tview.Application, vm virt.VM, page *tview.Pages, list *tview.List) tview.Primitive {

    btStart     := tview.NewButton("Start")
    btReboot    := tview.NewButton("Reboot")
    btEdit      := tview.NewButton("Edit")
    btShutdown  := tview.NewButton("Shutdown")
    btDestroy   := tview.NewButton("Destroy")
    btQuit      := tview.NewButton("Quit")
    btReboot.SetBackgroundColorActivated(tcell.ColorDarkSlateGray).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)
    btEdit.SetBackgroundColorActivated(tcell.ColorDarkSlateGray).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)
    btQuit.SetBackgroundColorActivated(tcell.ColorLightCyan).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)

    btQuit.SetSelectedFunc(func() {
        page.RemovePage("OnOff")
        page.SwitchToPage(vm.Name)
        app.SetFocus(list)
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow)
    flex.SetBorder(true).SetTitle(fmt.Sprintf("%s %v", vm.Name, vm.Status))
    flex.AddItem(btStart, 3, 0, true)
    flex.AddItem(btReboot, 3, 0, false)
    flex.AddItem(btEdit, 3, 0, false)
    flex.AddItem(btShutdown, 3, 0 ,false)
    flex.AddItem(btDestroy, 3, 0, false)
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
        btStart.SetBackgroundColorActivated(tcell.ColorDarkSlateGray).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)
        btShutdown.SetBackgroundColorActivated(tcell.ColorWhiteSmoke).SetLabelColor(tcell.ColorOrangeRed).SetBackgroundColor(tcell.ColorWhiteSmoke)
        btDestroy.SetBackgroundColorActivated(tcell.ColorWhiteSmoke).SetLabelColor(tcell.ColorRed).SetBackgroundColor(tcell.ColorWhiteSmoke)

        btShutdown.SetSelectedFunc(func() {
            _ = vm.Domain.Shutdown()
            time.Sleep(time.Millisecond * 500)
            page.RemovePage(vm.Name)
            page.AddAndSwitchToPage(vm.Name, NotUpVM(vm.Name), true)
            VirtualMachineStatus[vm.Name] = false
            list.SetItemText(list.GetCurrentItem(), vm.Name, "shutdown")
            app.SetFocus(list)
        })
        btDestroy.SetSelectedFunc(func() {
            _ = vm.Domain.Destroy()
            time.Sleep(time.Millisecond * 500)
            page.RemovePage(vm.Name)
            page.AddAndSwitchToPage(vm.Name, NotUpVM(vm.Name), true)
            VirtualMachineStatus[vm.Name] = false
            list.SetItemText(list.GetCurrentItem(), vm.Name, "shutdown")
            app.SetFocus(list)
        })

    } else {
        btStart.SetBackgroundColorActivated(tcell.ColorWhiteSmoke).SetLabelColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorWhiteSmoke)
        btShutdown.SetBackgroundColorActivated(tcell.ColorDarkSlateGray).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)
        btDestroy.SetBackgroundColorActivated(tcell.ColorDarkSlateGray).SetLabelColor(tcell.ColorWhiteSmoke).SetBackgroundColor(tcell.ColorDarkSlateGray)

        btStart.SetSelectedFunc(func() {
            _ = vm.Domain.Create()
            dur := time.Millisecond * 200
            for range time.Tick(dur) {
                b, _ := vm.Domain.IsActive()
                if b {
                    break
                }
            }
            page.RemovePage("OnOff")
            VirtualMachineStatus[vm.Name] = true
            page.RemovePage(vm.Name)
            page.AddAndSwitchToPage(vm.Name, NewVMStatus(app, vm.Domain, vm.Name), true)
            list.SetItemText(list.GetCurrentItem(), vm.Name, "")
            app.SetFocus(list)
        })
    }

    //return modal(flex, 30, 20)
    return pageModal(flex, 30, 20)
}

func CreateMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))

    VirtualMachineStatus = map[string]bool{}

    vms := virt.LookupVMs(con)
    for i, vm := range vms {
        if vm.Status {
            list.AddItem(vm.Name, "", rune(i)+'0', nil)
            page.AddPage(vm.Name, NewVMStatus(app, vm.Domain, vm.Name), true, true)
            VirtualMachineStatus[vm.Name] = true
        } else {
            list.AddItem(vm.Name, "shutdown", rune(i)+'0', nil)
            page.AddPage(vm.Name, NotUpVM(vm.Name), true, true)
            VirtualMachineStatus[vm.Name] = false
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

    list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
        var vmCrnt virt.VM
        vms = virt.LookupVMs(con)
        for _, vm := range vms {
            if vm.Name == s1 {
                vmCrnt = vm
            }
        }
        modal := CreateOnOffModal(app, vmCrnt, page, list)
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

    main, _ := list.GetItemText(list.GetCurrentItem())
    page.SwitchToPage(main)

    btCreate := tview.NewButton("Create")

    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if (event.Key() == tcell.KeyTab) || (event.Key() == tcell.KeyDown) {
            if (list.GetItemCount() - 1) == list.GetCurrentItem() {
                app.SetFocus(btCreate)
                return nil
            }
        }
        return event
    })
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
        modal := CreateMakeVM(app, con, page, list)
        if page.HasPage("OnOff") {
            page.RemovePage("OnOff")
        }
        page.AddPage("OnOff", modal, true, true)
        page.ShowPage("OnOff")
        app.SetFocus(modal)
    })

    /*
    _, _, w, _ := list.GetInnerRect()
    flex.AddItem(list, w + 5, 1, true)
    flex.AddItem(btCreate, 5, 1, false)
    */

    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1, true).
        AddItem(btCreate, 5, 0, false)

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



