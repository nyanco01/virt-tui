package tui

import (
	//"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/virt"
)


func MakeNetMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
    list.SetSecondaryTextColor(tcell.Color33)
    list.SetShortcutColor(tcell.Color87)

    for i, net := range virt.GetNetworkList(con) {
        list.AddItem(net.Name, net.NetType, rune(i+'0'), nil)
    }


    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1, true)

    return flex
}


func MakeNetUI(app *tview.Application, con *libvirt.Connect) *tview.Flex {
    flex := tview.NewFlex()

    page := MakePages(app)
    menu := MakeNetMenu(app, con, page)

    _, _, w, _ := menu.GetInnerRect()
    flex.AddItem(menu, w + 1, 0, true)
    flex.AddItem(page, 0, 1, false)

    return flex
}
