package tui

import (
	"fmt"
    "log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
	"github.com/nyanco01/virt-tui/src/virt"
)


func MakeVolDeleteForm(app *tview.Application, con *libvirt.Connect, page *tview.Pages, view *tview.TextView, p *Pool, volIndex int) *tview.Form {
    form := tview.NewForm()

    delVol := p.volumes[volIndex]

    form.AddButton("Delete", func() {
        if delVol.attachVM == "none" {
            err := virt.DeleteVolume(delVol.info.Path, con)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to delete volume: %v", err)
                }
            } else {
                // Delete the specified volume from the volume slice
                p.volumes = append(p.volumes[:volIndex], p.volumes[volIndex+1:]...)
                page.RemovePage("DeleteVolume")
            }
        } else {
            view.SetText(fmt.Sprintf("The volume is attached to the following VMs: [red::]%s", delVol.attachVM))
        }
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("DeleteVolume")
    })

    return form
}


func MakeVolDelete(app *tview.Application, con *libvirt.Connect, page *tview.Pages, p *Pool, volIndex int) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Delete Volume Menu")
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetText(fmt.Sprintf("Delete of Volume: [orange::]%s", p.volumes[volIndex].info.Path))

    form := MakeVolDeleteForm(app, con, page, view, p, volIndex)

    flex.SetDirection(tview.FlexRow).
        AddItem(view, 2, 0, false).
        AddItem(form, 3, 0, true)

    return pageModal(flex, 40, 7)
}


func MakePoolDeleteForm(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List, view *tview.TextView, name string) *tview.Form {
    form := tview.NewForm()

    form.AddButton("Delete", func() {
        b, vmName := virt.CheckDeletePoolRequest(name, con)
        if b {
            view.SetText(fmt.Sprintf("Delete Pool [orange]%s", name))
            err := virt.DeletePool(name, con)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to Delete pool by %s: %v", name, err)
                }
            } else {
                page.RemovePage(name)
                page.RemovePage("DeletePool")
                list.RemoveItem(list.GetCurrentItem())
                app.SetFocus(list)
            }

        } else {
            view.SetText(fmt.Sprintf("Volumes in the Pool are attached to \nthe following VMs: [red]%s", vmName))
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

    return pageModal(flex, 40, 7)
}
