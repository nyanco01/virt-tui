package tui

import (
	"log"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
)


func MakeNetDeleteForm(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List, view *tview.TextView, name, netType string) *tview.Form {
    form := tview.NewForm()

    form.AddButton("Delete", func() {
        if virt.CheckContainDomIF(con, name) {
            view.SetText("Cannot delete because some VMs belong to the network.").SetTextColor(tcell.ColorRed).SetTextAlign(tview.AlignCenter)
        } else {
            if netType == "Bridge" {
                operate.FileDelete("./tmp/shell/" + name + ".sh")
                operate.DeleteBridgeIF(name)
            }
            err := virt.DeleteNetwork(con, name)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create bridge network: %v", err)
                }
            } else {
                page.RemovePage(name)
                page.RemovePage("Delete")
                list.RemoveItem(list.GetCurrentItem())
                app.SetFocus(list)
            }
        }
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("Delete")
        app.SetFocus(list)
    })

    return form
}


func MakeNetDelete(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List, name, netType string) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Delete Network Menu")
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetText(fmt.Sprintf("Delete of Network: [orange]%s", name))

    form := MakeNetDeleteForm(app, con, page, list, view, name, netType)

    flex.SetDirection(tview.FlexRow).
        AddItem(view, 2, 0, false).
        AddItem(form, 3, 0, true)

    return pageModal(flex, 40, 7)
}
