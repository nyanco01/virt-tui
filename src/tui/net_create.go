package tui

import (
	//"log"

	//"fmt"

	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"
)


func SelectPageButton(app *tview.Application, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()

    btBridge := tview.NewButton("Bridge")
    btBridge.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
    btBridge.SetBackgroundColorActivated(tcell.Color81)
    btBridge.SetLabelColorActivated(tcell.Color232)

    btNAT := tview.NewButton("NAT")
    btNAT.SetBackgroundColor(tcell.NewRGBColor(80, 80, 80))
    btNAT.SetBackgroundColorActivated(tcell.Color87)
    btNAT.SetLabelColorActivated(tcell.Color232)

    btPrivate := tview.NewButton("Private")
    btPrivate.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
    btPrivate.SetBackgroundColorActivated(tcell.Color33)
    btPrivate.SetLabelColorActivated(tcell.Color232)

    btBridge.SetSelectedFunc(func() {
        if page.HasPage("Bridge") {
            page.SwitchToPage("Bridge")
        }
    })
    btNAT.SetSelectedFunc(func() {
        if page.HasPage("NAT") {
            page.SwitchToPage("NAT")
        }
    })
    btPrivate.SetSelectedFunc(func() {
        if page.HasPage("Private") {
            page.SwitchToPage("Private")
        }
    })

    flex.
        AddItem(btBridge, 0, 1, true).
        AddItem(btNAT, 0, 1, false).
        AddItem(btPrivate, 0, 1, false)

    return flex
}


func MakeNetCreatePageByBridge(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) *tview.Flex {
    flex := tview.NewFlex()
    menuTitle := tview.NewTextView()
    menuTitle.SetText("Create Bridge Network").SetTextAlign(tview.AlignCenter).SetTextColor(tcell.Color81)
    form := tview.NewForm()
    form.SetLabelColor(tcell.Color81)
    view := tview.NewTextView()
    view.SetDynamicColors(true)
    view.SetTextAlign(tview.AlignCenter)

    form.AddInputField("Network Name     ", "", 30, nil, nil)
    listPhysicsIF := operate.ListPhysicsIF()
    form.AddDropDown("Physical Interface", listPhysicsIF, 0, nil)
    form.AddButton("  OK  ", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()
        _, source := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()

        if name == "" {
            view.SetText("Name is empty.").SetTextColor(tcell.ColorRed)
        } else if !virt.CheckNetworkName(con, name) {
            view.SetText("The same network name exists.").SetTextColor(tcell.ColorRed)
        } else if !operate.CheckBridgeSource(source) {
            view.SetText(fmt.Sprintf("[skyblue]%s [red]is already a member of the bridge interface.", source))
        } else {
            virt.CreateNetworkByBridge(con, name, source)
            net := virt.GetNetworkByName(con, name)
            list.AddItem(net.Name, net.NetType, rune(list.GetItemCount()+'0'), nil)
            page.AddPage(net.Name, NewNetwork(con, net), true, true)

            page.RemovePage("Create")
            app.SetFocus(list)
        }

    })
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(menuTitle, 1, 0, false).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return flex
}


func MakeNetCreatePageByNAT(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) *tview.Flex {
    flex := tview.NewFlex()
    menuTitle := tview.NewTextView()
    menuTitle.SetText("Create NAT Network").SetTextAlign(tview.AlignCenter).SetTextColor(tcell.Color87)
    form := tview.NewForm()
    form.SetLabelColor(tcell.Color81)
    view := tview.NewTextView()

    form.AddInputField("Network Name      ", "", 30, nil, nil)
    form.AddButton("  OK  ", nil)
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(menuTitle, 1, 0, false).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return flex
}


func MakeNetCreatePageByPrivate(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) *tview.Flex {
    flex := tview.NewFlex()
    menuTitle := tview.NewTextView()
    menuTitle.SetText("Create Private Network").SetTextAlign(tview.AlignCenter).SetTextColor(tcell.Color33)
    form := tview.NewForm()
    form.SetLabelColor(tcell.Color81)
    view := tview.NewTextView()

    form.AddInputField("Network Name      ", "", 30, nil, nil)
    form.AddButton("  OK  ", nil)
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(menuTitle, 1, 0, false).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)

    return flex
}


func MakeNetCreate(app *tview.Application, con *libvirt.Connect, mainPage *tview.Pages, list *tview.List) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create Network Menu")
    createPage := tview.NewPages()
    buttons := SelectPageButton(app, createPage)

    createPage.AddPage("Bridge", MakeNetCreatePageByBridge(app, con, mainPage, list), true, true)
    createPage.AddPage("NAT", MakeNetCreatePageByNAT(app, con, mainPage, list), true, true)
    createPage.AddPage("Private", MakeNetCreatePageByPrivate(app, con, mainPage, list), true, true)

    createPage.SwitchToPage("Bridge")
    app.SetFocus(createPage)

    flex.SetDirection(tview.FlexRow).
        AddItem(buttons, 1, 0, false).
        AddItem(createPage, 0, 1, true)

    return pageModal(flex, 62, 15)
}
