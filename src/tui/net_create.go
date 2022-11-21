package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"
)


func InputFieldIPv4Subnet(text string, ch rune) bool {
    if ch == '.' || ch == '/' {
        return true
    }
    if len(text) > 18 {
        return false
    }
    for i := 0; i <= 9; i++ {
        if ch == rune(i + '0') {
            return true
        }
    }
    return false
}


func SelectPageButton(app *tview.Application, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()

    btBridge := tview.NewButton("Bridge")
    btNAT := tview.NewButton("NAT")
    btPrivate := tview.NewButton("Private")

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

    app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
        pageTitle, _ := page.GetFrontPage()
        switch pageTitle {
        case "Bridge":
            btBridge.SetBackgroundColor(tcell.Color81)
            btBridge.SetLabelColor(tcell.Color232)
            btNAT.SetBackgroundColor(tcell.NewRGBColor(80, 80, 80))
            btNAT.SetLabelColor(tcell.ColorWhiteSmoke)
            btPrivate.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
            btPrivate.SetLabelColor(tcell.ColorWhiteSmoke)
        case "NAT":
            btBridge.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
            btBridge.SetLabelColor(tcell.ColorWhiteSmoke)
            btNAT.SetBackgroundColor(tcell.Color87)
            btNAT.SetLabelColor(tcell.Color232)
            btPrivate.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
            btPrivate.SetLabelColor(tcell.ColorWhiteSmoke)
        case "Private":
            btBridge.SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
            btBridge.SetLabelColor(tcell.ColorWhiteSmoke)
            btNAT.SetBackgroundColor(tcell.NewRGBColor(80, 80, 80))
            btNAT.SetLabelColor(tcell.ColorWhiteSmoke)
            btPrivate.SetBackgroundColor(tcell.Color33)
            btPrivate.SetLabelColor(tcell.Color232)
        }
        return false
    })

    flex.
        AddItem(btBridge, 0, 1, false).
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

    form.AddInputField("Network Name     ", "", 20, nil, nil)
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

    form.AddInputField("Network Name      ", "", 20, nil, nil)
    form.AddInputField("Address Range     ", "", 20, nil, nil)
    form.GetFormItem(1).(*tview.InputField).SetAcceptanceFunc(InputFieldIPv4Subnet)
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

    form.AddInputField("Network Name      ", "", 20, nil, nil)
    form.AddButton("  OK  ", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()

        if name == "" {
            view.SetText("Name is empty.").SetTextColor(tcell.ColorRed)
        } else if !virt.CheckNetworkName(con, name) {
            view.SetText("The same network name exists.").SetTextColor(tcell.ColorRed)
        } else {
            virt.CreateNetworkByPrivate(con, name)
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


func MakeNetCreate(app *tview.Application, con *libvirt.Connect, mainPage *tview.Pages, list *tview.List) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create Network Menu")
    createPage := tview.NewPages()

    createPage.AddPage("Bridge", MakeNetCreatePageByBridge(app, con, mainPage, list), true, true)
    createPage.AddPage("NAT", MakeNetCreatePageByNAT(app, con, mainPage, list), true, true)
    createPage.AddPage("Private", MakeNetCreatePageByPrivate(app, con, mainPage, list), true, true)

    createPage.SwitchToPage("Bridge")

    buttons := SelectPageButton(app, createPage)
    flex.SetDirection(tview.FlexRow).
        AddItem(buttons, 1, 0, false).
        AddItem(createPage, 0, 1, true)

    flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        pageTitle ,_ := createPage.GetFrontPage()
        switch pageTitle {
        case "Bridge":
            if event.Key() == tcell.KeyRight {
                createPage.SwitchToPage("NAT")
                return nil
            }
        case "NAT":
            if event.Key() == tcell.KeyRight {
                createPage.SwitchToPage("Private")
                return nil
            } else if event.Key() == tcell.KeyLeft {
                createPage.SwitchToPage("Bridge")
                return nil
            }
        case "Private":
            if event.Key() == tcell.KeyLeft {
                createPage.SwitchToPage("NAT")
                return nil
            }
        default:
            return event
        }
        return event
    })

    return pageModal(flex, 62, 15)
}
