package tui

import (
	//"log"

	//"fmt"

	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
)


type Network struct {
    *tview.Box
    name                string
    networkType         string
    source              string
    ifList              []virt.DomainIF

    lineOfset           int
    lineOfsetMax        int
    oldheight           int

    // For bridge interface, stores master physical interface name
    master              string
    // NAT
    address             string
    dhcpStart           string
    dhcpEnd             string
}


func NewNetwork(con *libvirt.Connect, netInfo virt.NetworkInfo) *Network {
    iflist := virt.GetDomIFListByBridgeName(con, netInfo.Source)

    var ifnames []string
    for _, i := range iflist {
        if i.Name != "" {
            ifnames = append(ifnames, i.Name)
        }
    }
    m := operate.GetBridgeMasterIF(netInfo.Source, ifnames)

    var a, s, l string = "", "", ""
    if netInfo.NetType == "NAT" {
        a, s, l = virt.GetAddressByNATNetwork(con, netInfo.Name)
    }

    return &Network{
        Box:            tview.NewBox(),
        name:           netInfo.Name,
        networkType:    netInfo.NetType,
        source:         netInfo.Source,
        ifList:         iflist,

        lineOfset:      0,
        lineOfsetMax:   0,
        oldheight:      0,

        master:         m,
        address:        a,
        dhcpStart:      s,
        dhcpEnd:        l,
    }
}


func (n *Network)Draw(screen tcell.Screen) {
    n.Box.DrawForSubclass(screen, n)
    x, y, w, h := n.GetInnerRect()

    boxW := 30
    bc := tcell.ColorSkyblue
    netStyle := tcell.StyleDefault
    netStyle = netStyle.Foreground(bc)
    netStyle = netStyle.Background(tview.Styles.PrimitiveBackgroundColor)

    l := len(n.ifList)
    fullHeight := 1 + (l*4)

    // network
    for i := x+1; i <= x+1+boxW; i++ {
        screen.SetContent(i, y+1, tview.Borders.Horizontal, nil, netStyle)
    }
    for i := x+1; i <= x+1+boxW; i++ {
        screen.SetContent(i, y+6, tview.Borders.Horizontal, nil, netStyle)
    }
    for i := y+2; i <= y+5; i++ {
        screen.SetContent(x+1, i, tview.Borders.Vertical, nil, netStyle)
        screen.SetContent(x+1+boxW, i, tview.Borders.Vertical, nil, netStyle)
    }
    // Left corner
    screen.SetContent(x+1, y+1, tview.Borders.TopLeft, nil, netStyle)
    screen.SetContent(x+1, y+6, tview.Borders.BottomLeft, nil, netStyle)
    // Right corner
    screen.SetContent(x+1+boxW, y+1, tview.Borders.TopRight, nil, netStyle)
    screen.SetContent(x+1+boxW, y+6, tview.Borders.BottomRight, nil, netStyle)
    // master name
    tview.Print(screen, "Network", x+2, y+2, len("Network"), tview.AlignCenter, tcell.ColorWhiteSmoke)
    tview.Print(screen, " ------------------- ", x+2, y+3, len(" ------------------- "), tview.AlignCenter, bc)
    tview.Print(screen, fmt.Sprintf("   Name: [blue]%s", n.source), x+2, y+4, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("   Type: [blue]%s", n.networkType), x+2, y+5, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)

    switch n.networkType {
    case "Bridge":
        // master
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, y+9, tview.Borders.Horizontal, nil, netStyle)
        }
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, y+13, tview.Borders.Horizontal, nil, netStyle)
        }
        for i := y+10; i <= y+12; i++ {
            screen.SetContent(x+1, i, tview.Borders.Vertical, nil, netStyle)
            screen.SetContent(x+1+boxW, i, tview.Borders.Vertical, nil, netStyle)
        }
        // Left corner
        screen.SetContent(x+1, y+9, tview.Borders.TopLeft, nil, netStyle)
        screen.SetContent(x+1, y+13, tview.Borders.BottomLeft, nil, netStyle)
        // Right corner
        screen.SetContent(x+1+boxW, y+9, tview.Borders.TopRight, nil, netStyle)
        screen.SetContent(x+1+boxW, y+13, tview.Borders.BottomRight, nil, netStyle)
        // master name
        tview.Print(screen, "Physical Interfaces", x+2, y+10, len("Physical Interfaces"), tview.AlignCenter, tcell.ColorWhiteSmoke)
        tview.Print(screen, " ------------------- ", x+2, y+11, len(" ------------------- "), tview.AlignCenter, bc)
        tview.Print(screen, fmt.Sprintf("   Name: [blue]%s", n.master), x+2, y+12, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)

        screen.SetContent(x+5, y+6,tview.Borders.TopT, nil, netStyle)
        screen.SetContent(x+5, y+7,tview.Borders.Vertical, nil, netStyle)
        screen.SetContent(x+5, y+8,tview.Borders.Vertical, nil, netStyle)
        screen.SetContent(x+5, y+9,tview.Borders.BottomT, nil, netStyle)
    case "NAT":
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, y+6, ' ', nil, netStyle)
        }
        tview.Print(screen, fmt.Sprintf("Address: [blue]%s", n.address), x+2, y+6, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)
        natY := 0
        if n.dhcpStart != "" {
            natY = y+11
            tview.Print(screen, " ------------------- ", x+2, y+7, len(" ------------------- "), tview.AlignCenter, bc)
            tview.Print(screen, "DHCP", x+2, y+8, boxW, tview.AlignLeft, tcell.ColorSkyblue)
            tview.Print(screen, fmt.Sprintf("  Start: [skyblue]%s", n.dhcpStart), x+2, y+9, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("    End: [skyblue]%s", n.dhcpEnd), x+2, y+10, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)
            for i := y+6; i <= y+11; i++ {
                screen.SetContent(x+1, i, tview.Borders.Vertical, nil, netStyle)
                screen.SetContent(x+1+boxW, i, tview.Borders.Vertical, nil, netStyle)
            }

        } else {
            natY = y+7
            screen.SetContent(x+1, y+6, tview.Borders.Vertical, nil, netStyle)
            screen.SetContent(x+1+boxW, y+6, tview.Borders.Vertical, nil, netStyle)
        }
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, natY, tview.Borders.Horizontal, nil, netStyle)
        }
        screen.SetContent(x+1, natY, tview.Borders.BottomLeft, nil, netStyle)
        screen.SetContent(x+1+boxW, natY, tview.Borders.BottomRight, nil, netStyle)

    }

    if len(n.ifList) != 0 {
        if h >= fullHeight {
            // Drawing iface list
            for i, domif := range n.ifList {
                for j := 0; j < 4; j++ {
                    screen.SetContent(x+boxW+6, y+1+(4*i)+j, tview.Borders.Vertical, nil, netStyle)
                }
                for j := 0; j < 3; j++ {
                    screen.SetContent(x+boxW+11, y+1+(4*i)+j, '▌', nil, tcell.StyleDefault.Foreground(tcell.Color87))
                }
                tview.Print(screen,"├───", x+boxW+6, y+1+(4*i), len("├───"), tview.AlignLeft, bc)
                tview.Print(screen, fmt.Sprintf("Name: %s", domif.Name), x+boxW+12, y+1+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
                tview.Print(screen, fmt.Sprintf("AttachVM: %s", domif.AttachVM), x+boxW+12, y+2+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
                tview.Print(screen, fmt.Sprintf("MAC Addr: %s", domif.MacAddr), x+boxW+12, y+3+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
            }
        } else {
            if n.oldheight != 0 && n.oldheight < h && n.lineOfset != 0 {
                n.lineOfset--
            }
            n.lineOfsetMax = fullHeight - h
            cnt := n.lineOfset
            var ifaceNum int
            for i := 0; i < h; i++ {
                ifaceNum = cnt / 4
                if len(n.ifList)-1 < ifaceNum {
                    break
                }
                screen.SetContent(x+boxW+6, y+i+1, tview.Borders.Vertical, nil, netStyle)
                switch cnt % 4 {
                case 0:
                    tview.Print(screen, "├───", x+boxW+6, y+i+1, len("├───"), tview.AlignLeft, bc)
                    tview.Print(screen, fmt.Sprintf("Name: %s", n.ifList[ifaceNum].Name), x+boxW+12, y+i+1, w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
                case 1:
                    tview.Print(screen, fmt.Sprintf("AttachVM: %s", n.ifList[ifaceNum].AttachVM), x+boxW+12, y+i+1, w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
                case 2:
                    tview.Print(screen, fmt.Sprintf("MAC Addr: %s", n.ifList[ifaceNum].MacAddr), x+boxW+12, y+i+1, w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
                }
                if cnt % 4 != 3 {
                    screen.SetContent(x+boxW+11, y+i+1, '▌', nil, tcell.StyleDefault.Foreground(tcell.Color87))
                }
                cnt++
            }
            n.oldheight = h
        }

        screen.SetContent(x+1+boxW, y+3,tview.Borders.LeftT, nil, netStyle)
        screen.SetContent(x+1+boxW+1, y+3,tview.Borders.Horizontal, nil, netStyle)
        screen.SetContent(x+1+boxW+2, y+3,tview.Borders.Horizontal, nil, netStyle)
        screen.SetContent(x+1+boxW+3, y+3,tview.Borders.Horizontal, nil, netStyle)
        screen.SetContent(x+1+boxW+4, y+3,tview.Borders.Horizontal, nil, netStyle)
        screen.SetContent(x+1+boxW+5, y+3,tview.Borders.RightT, nil, netStyle)
    }

}


func (n *Network)Update(con *libvirt.Connect, netInfo virt.NetworkInfo) *Network {
    iflist := virt.GetDomIFListByBridgeName(con, netInfo.Source)

    var ifnames []string
    for _, i := range iflist {
        if i.Name != "" {
            ifnames = append(ifnames, i.Name)
        }
    }

    n.ifList = iflist
    return n
}


func (n *Network)MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return n.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
		if !n.InRect(x, y) {
			return false, nil
		}

        px, _, _, _ := n.GetInnerRect()
        switch action {
        case tview.MouseScrollUp:
            if px + 42 < x {
                if n.lineOfset > 0 {
                    n.lineOfset--
                    consumed = true
                }
            }
        case tview.MouseScrollDown:
            if px + 42 < x {
                if n.lineOfset < n.lineOfsetMax {
                    n.lineOfset++
                    consumed = true
                }
            }
        }
        return
    })
}


func NetworkStatusUpdate(app *tview.Application, n *Network, con *libvirt.Connect, netinfo virt.NetworkInfo) {
    time.Sleep(time.Second * 3)
    for range time.Tick(time.Second * 3) {
        _, err := con.LookupNetworkByName(netinfo.Name)
        if err != nil {
            if virtErr, ok := err.(libvirt.Error); ok {
                text := "no network with matching name"
                if strings.Contains(virtErr.Message, text) {
                    break
                }
            }
        }
        app.QueueUpdateDraw(func() {
            n.Update(con, netinfo)
        })
    }
}


func MakeNetMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40, 40, 40))
    list.SetSecondaryTextColor(tcell.ColorRoyalBlue)
    list.SetShortcutColor(tcell.Color87)

    for i, net := range virt.GetNetworkList(con) {
        list.AddItem(net.Name, net.NetType, rune(i+'0'), nil)
        network := NewNetwork(con, net)
        page.AddPage(net.Name, network, true, true)
        go NetworkStatusUpdate(app, network, con, net)
    }

    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

    list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
        modal := MakeNetDelete(app, con, page, list, s1, s2)
        page.AddPage("Delete", modal, true, true)
    })

    if list.GetItemCount() != 0 {
        main, _ := list.GetItemText(list.GetCurrentItem())
        page.SwitchToPage(main)
    }

    btCreate := tview.NewButton("Create")
    btCreate.SetBackgroundColor(tcell.Color39)
    btCreate.SetLabelColor(tcell.Color232)

    // If the last item on the list is selected, toggle to move focus to the button
    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if (event.Key() == tcell.KeyTab) || (event.Key() == tcell.KeyDown) {
            if (list.GetItemCount() - 1) == list.GetCurrentItem() {
                app.SetFocus(btCreate)
                return nil
            }
        }
        return event
    })
    // Toggling when the focus is on a button focuses the list
    btCreate.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyTab, tcell.KeyDown:
            list.SetCurrentItem(0)
            app.SetFocus(list)
            return nil
        case tcell.KeyBacktab, tcell.KeyUp:
            list.SetCurrentItem(list.GetItemCount() - 1)
            app.SetFocus(list)
            return nil
        }
        return event
    })

    btCreate.SetSelectedFunc(func() {
        modal := MakeNetCreate(app, con, page, list)
        page.AddPage("Create", modal, true, true)
        app.SetFocus(modal)
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1, true).
        AddItem(btCreate, 5, 0, false)

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
