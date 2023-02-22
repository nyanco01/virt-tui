package tui

import (
	"fmt"
    "log"

	"github.com/gdamore/tcell/v2"
	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"
)


func InputFieldIPv4Subnet(text string, ch rune) bool {
    if len(text) > 18 {
        return false
    }
    if ch == '.' || ch == '/' {
        return true
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
            app.SetFocus(page)
        }
    })
    btNAT.SetSelectedFunc(func() {
        if page.HasPage("NAT") {
            page.SwitchToPage("NAT")
            app.SetFocus(page)
        }
    })
    btPrivate.SetSelectedFunc(func() {
        if page.HasPage("Private") {
            page.SwitchToPage("Private")
            app.SetFocus(page)
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
            err := virt.CreateNetworkByBridge(con, name, source)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create bridge network: %v", err)
                }
            } else {
                net := virt.GetNetworkByName(con, name)
                list.AddItem(net.Name, net.NetType, rune(list.GetItemCount()+'0'), nil)
                network := NewNetwork(con, net)
                page.AddPage(net.Name, network, true, true)
                go NetworkStatusUpdate(app, network, con, net)
                page.RemovePage("Create")
                app.SetFocus(list)
            }
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
    descriptionSubnet := tview.NewTextView().SetTextColor(tcell.Color87).SetTextAlign(tview.AlignCenter)
    descriptionSubnet.SetText("Address Range should be put in \nCIDR format. ( e.g. 172.16.10.0/24 )")

    form.AddInputField("Network Name      ", "", 30, nil, nil)
    form.AddInputField("Address Range     ", "", 30, nil, nil)
    form.GetFormItem(1).(*tview.InputField).SetAcceptanceFunc(InputFieldIPv4Subnet)
    form.AddButton("  OK  ", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()
        subnet := form.GetFormItem(1).(*tview.InputField).GetText()

        if name == "" {
            view.SetText("Name is empty.").SetTextColor(tcell.ColorRed)
        } else if !virt.CheckNetworkName(con, name) {
            view.SetText("The same network name exists.").SetTextColor(tcell.ColorRed)
        } else if !operate.CheckNetworkSubnet(subnet) {
            view.SetText("Please enter the correct subnet mask.").SetTextColor(tcell.ColorRed)
        } else if virt.CheckNetworkRange(con, subnet) {
            view.SetText("The address range is overlapped by other NAT networks.").SetTextColor(tcell.ColorRed)
        } else {
            err := virt.CreateNetworkByNAT(con, name, subnet)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create NAT network: %v", err)
                }
            } else {
                net := virt.GetNetworkByName(con, name)
                list.AddItem(net.Name, net.NetType, rune(list.GetItemCount()+'0'), nil)
                network := NewNetwork(con, net)
                page.AddPage(net.Name, network, true, true)
                go NetworkStatusUpdate(app, network, con, net)
                page.RemovePage("Create")
                app.SetFocus(list)
            }
        }
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(menuTitle, 1, 0, false).
        AddItem(form, 0, 1, true).
        AddItem(descriptionSubnet, 2, 0, false).
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
    form.AddButton("  OK  ", func() {
        name := form.GetFormItem(0).(*tview.InputField).GetText()

        if name == "" {
            view.SetText("Name is empty.").SetTextColor(tcell.ColorRed)
        } else if !virt.CheckNetworkName(con, name) {
            view.SetText("The same network name exists.").SetTextColor(tcell.ColorRed)
        } else {
            err :=virt.CreateNetworkByPrivate(con, name)
            if err != nil {
                if virtErr, ok := err.(libvirt.Error); ok {
                    view.SetText(virtErr.Message)
                    view.SetTextColor(tcell.ColorRed)
                } else {
                    log.Fatalf("failed to create Private network: %v", err)
                }
            } else {
                net := virt.GetNetworkByName(con, name)
                list.AddItem(net.Name, net.NetType, rune(list.GetItemCount()+'0'), nil)
                network := NewNetwork(con, net)
                page.AddPage(net.Name, network, true, true)
                go NetworkStatusUpdate(app, network, con, net)
                page.RemovePage("Create")
                app.SetFocus(list)
            }

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
        var left bool = event.Key() == tcell.KeyLeft && event.Modifiers() == tcell.ModCtrl
        var right bool = event.Key() == tcell.KeyRight && event.Modifiers() == tcell.ModCtrl
        pageTitle ,_ := createPage.GetFrontPage()
        switch pageTitle {
        case "Bridge":
            if right {
                createPage.SwitchToPage("NAT")
                return nil
            }
        case "NAT":
            if right {
                createPage.SwitchToPage("Private")
                return nil
            } else if left {
                createPage.SwitchToPage("Bridge")
                return nil
            }
        case "Private":
            if left {
                createPage.SwitchToPage("NAT")
                return nil
            }
        }
        return event
    })

    return pageModal(flex, 62, 15)
}
