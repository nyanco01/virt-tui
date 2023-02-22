package tui

import (
	"fmt"
    "log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
	"github.com/nyanco01/virt-tui/src/virt"
)


func MakeVolumeCreateForm(app * tview.Application, con *libvirt.Connect, view *tview.TextView, page *tview.Pages, pool *Pool) *tview.Form {
    available := pool.capacity-pool.allocation
    form := tview.NewForm()

    // Volume name      item index 0
    form.AddInputField("Volume name", "", 20, nil, nil)

    // Volume size      item index 1
    form.AddInputField(fmt.Sprintf("Volume size [orange]GB (max %.1f GB)", float64(available)/float64(gibi)), "", 6, nil, nil)
    form.GetFormItem(1).(*tview.InputField).SetAcceptanceFunc(tview.InputFieldInteger)

    form.AddButton("Create", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()
        size, _ := strconv.Atoi(form.GetFormItem(1).(*tview.InputField).GetText())
        b, ErrInfo := virt.CheckCreateVolumeRequest(name, size, available)

        if b {
            //view.SetText("OK").SetTextColor(tcell.ColorSkyblue)
            err := virt.CreateVolume(name, pool.path, size, con)

            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create volume: %v", err)
                }
            } else {
                vol := Volume {
                    info:       virt.GetVolumeInfo(pool.path + "/" + name, con),
                    attachVM:   "none",
                }
                pool.volumes = append(pool.volumes, vol)
                pool.onClickCreate = false
                page.RemovePage("CreateVolume")
            }
        } else {
            view.SetText(ErrInfo).SetTextColor(tcell.ColorRed)
        }
    })

    form.AddButton("Cancel", func() {
        page.RemovePage("CreateVolume")
    })

    return form
}


func MakeVolumeCreate(app *tview.Application, con *libvirt.Connect, pool *Pool, page *tview.Pages) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create Volume Menu")
    view := tview.NewTextView()

    form := MakeVolumeCreateForm(app, con, view, page, pool)

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return pageModal(flex, 70, 10)
}


func MakePoolCreateForm(app *tview.Application, con *libvirt.Connect, view *tview.TextView, list *tview.List, page *tview.Pages) *tview.Form {
    form := tview.NewForm()

    // Pool name            item index 0
    form.AddInputField("Pool name", "", 20, nil, nil)
    
    // Pool path            item index 1
    form.AddInputField("Path (absolute path)", "", 40, nil, nil)

    form.AddButton("Create", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()
        path := form.GetFormItem(1).(*tview.InputField).GetText()

        b, ErrInfo := virt.CheckCreatePoolRequest(name, path, con)

        if b {
            // Start creating a Pool
            view.SetText("OK").SetTextColor(tcell.ColorSkyblue)
            err := virt.CreatePool(name, path, con)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create pool: %v", err)
                }
            } else {
                list.AddItem(name, "", rune(list.GetItemCount())+'0', nil)
                list.SetCurrentItem(list.GetItemCount())
                p := NewPool(con, name)
                go PoolStatusUpdate(app, p, con, name)
                SetModal(app, con, p, page)
                page.AddPage(name, p, true, true)
                app.SetFocus(list)
                page.RemovePage("Create")
            }
        } else {
            view.SetText(ErrInfo).SetTextColor(tcell.ColorRed)
        }
        
    })

    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    return form
}


func MakePoolCreate(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create Pool Menu")
    view := tview.NewTextView()

    form := MakePoolCreateForm(app, con, view, list, page)

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return pageModal(flex, 70, 10)
}
