package tui

import (

	//"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

	"github.com/nyanco01/virt-tui/src/virt"
)


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
            virt.CreatePool(name, path, con)

            list.AddItem(name, "", rune(list.GetItemCount())+'0', nil)
            list.SetCurrentItem(list.GetItemCount())
            page.AddPage(name, NewPool(con, name), true, true)
            app.SetFocus(list)
            page.RemovePage("Create")
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
