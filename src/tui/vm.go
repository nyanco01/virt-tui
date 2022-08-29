package tui

import (
    "log"

    "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

    "github.com/nyanco01/virt-tui/src/virt"
)





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

func CreateMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))

    l := virt.LookupVMs(con)
    for i, vm := range l {

        if vm.Status {
            list.AddItem(vm.Name, "", rune(i)+'0', nil)
            //page.AddPage(vm.name, NewVMStatus(app, vm.domain, vm.name), true, true)
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
