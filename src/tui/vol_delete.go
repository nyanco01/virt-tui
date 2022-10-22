package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"
	//"github.com/nyanco01/virt-tui/src/virt"
)


func MakePoolDeleteForm(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List, view *tview.TextView, name string) *tview.Form {
    form := tview.NewForm()

    form.AddButton("Delete", func() {
        b, vmName := virt.CheckDeletePoolRequest(name, con)
        if b {
            view.SetText(fmt.Sprintf("Delete Pool [orange]%s", name))
            virt.DeletePool(name, con)
            page.RemovePage(name)
            page.RemovePage("DeletePool")
            list.RemoveItem(list.GetCurrentItem())
            app.SetFocus(list)
        } else {
            view.SetText(fmt.Sprintf("Volumes in the Pool are attached to the following VMs: [red]%s", vmName))
        }
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("DeletePool")
        app.SetFocus(list)
    })

    return form
}


func MakePoolDelete(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List, name string) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Delete Pool Menu")
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetText(fmt.Sprintf("Delete of Pool: [orange::]%s", name))
    view.SetTextColor(tcell.ColorWhiteSmoke)

    form := MakePoolDeleteForm(app, con, page, list, view, name)

    flex.SetDirection(tview.FlexRow).
        AddItem(view, 2, 0, false).
        AddItem(form, 3, 0, true)

    return pageModal(flex, 60, 7)
}
