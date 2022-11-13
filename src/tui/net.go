package tui

import (
	//"log"

	//"fmt"

	"fmt"

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
    // For bridge interface, stores master physical interface name
    master              string
    source              string
    ifList              []virt.DomainIF
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

    return &Network{
        Box:            tview.NewBox(),
        name:           netInfo.Name,
        networkType:    netInfo.NetType,
        master:         m,
        source:         netInfo.Source,
        ifList:         iflist,
    }
}


func (n *Network)Draw(screen tcell.Screen) {
    n.Box.DrawForSubclass(screen, n)
    x, y, w, _ := n.GetInnerRect()

    boxW := 30


    bc := tcell.ColorSkyblue

    // network
    for i := x+1; i <= x+1+boxW; i++ {
        screen.SetContent(i, y+1, tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
    }
    for i := x+1; i <= x+1+boxW; i++ {
        screen.SetContent(i, y+6, tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
    }
    for i := y+2; i <= y+5; i++ {
        screen.SetContent(x+1, i, tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW, i, tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
    }
    // Left
    screen.SetContent(x+1, y+1, tview.Borders.TopLeft, nil, tcell.StyleDefault.Foreground(bc))
    screen.SetContent(x+1, y+6, tview.Borders.BottomLeft, nil, tcell.StyleDefault.Foreground(bc))
    // Right
    screen.SetContent(x+1+boxW, y+1, tview.Borders.TopRight, nil, tcell.StyleDefault.Foreground(bc))
    screen.SetContent(x+1+boxW, y+6, tview.Borders.BottomRight, nil, tcell.StyleDefault.Foreground(bc))
    // master name
    tview.Print(screen, "Network", x+2, y+2, len("Network"), tview.AlignCenter, tcell.ColorWhiteSmoke)
    tview.Print(screen, " ------------------- ", x+2, y+3, len(" ------------------- "), tview.AlignCenter, bc)
    tview.Print(screen, fmt.Sprintf("Name: [blue]%s", n.source), x+2, y+4, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Type: [blue]%s", n.networkType), x+2, y+5, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)

    if n.networkType != "Private" {

        // master
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, y+9, tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        }
        for i := x+1; i <= x+1+boxW; i++ {
            screen.SetContent(i, y+13, tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        }
        for i := y+10; i <= y+12; i++ {
            screen.SetContent(x+1, i, tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
            screen.SetContent(x+1+boxW, i, tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
        }
        // Left
        screen.SetContent(x+1, y+9, tview.Borders.TopLeft, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1, y+13, tview.Borders.BottomLeft, nil, tcell.StyleDefault.Foreground(bc))
        // Right
        screen.SetContent(x+1+boxW, y+9, tview.Borders.TopRight, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW, y+13, tview.Borders.BottomRight, nil, tcell.StyleDefault.Foreground(bc))
        // master name
        tview.Print(screen, "Physical Interfaces", x+2, y+10, len("Physical Interfaces"), tview.AlignCenter, tcell.ColorWhiteSmoke)
        tview.Print(screen, " ------------------- ", x+2, y+11, len(" ------------------- "), tview.AlignCenter, bc)
        tview.Print(screen, fmt.Sprintf("Name: [blue]%s", n.master), x+2, y+12, boxW, tview.AlignLeft, tcell.ColorWhiteSmoke)

        screen.SetContent(x+5, y+6,tview.Borders.TopT, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+5, y+7,tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+5, y+8,tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+5, y+9,tview.Borders.BottomT, nil, tcell.StyleDefault.Foreground(bc))
    }

    if len(n.ifList) != 0 {
        for i, domif := range n.ifList {
            for j := 0; j < 4; j++ {
                screen.SetContent(x+boxW+6, y+1+(4*i)+j, tview.Borders.Vertical, nil, tcell.StyleDefault.Foreground(bc))
            }
            for j := 0; j < 3; j++ {
                screen.SetContent(x+boxW+11, y+1+(4*i)+j, '▌', nil, tcell.StyleDefault.Foreground(tcell.Color87))
            }
            tview.Print(screen,"├───", x+boxW+6, y+1+(4*i), len("├───"), tview.AlignLeft, bc)
            tview.Print(screen, fmt.Sprintf("Name: %s", domif.Name), x+boxW+12, y+1+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("AttachVM: %s", domif.AttachVM), x+boxW+12, y+2+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("MAC Addr: %s", domif.MacAddr), x+boxW+12, y+3+(4*i), w-(boxW+11), tview.AlignLeft, tcell.ColorWhiteSmoke)

        }
        screen.SetContent(x+1+boxW, y+3,tview.Borders.LeftT, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW+1, y+3,tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW+2, y+3,tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW+3, y+3,tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW+4, y+3,tview.Borders.Horizontal, nil, tcell.StyleDefault.Foreground(bc))
        screen.SetContent(x+1+boxW+5, y+3,tview.Borders.RightT, nil, tcell.StyleDefault.Foreground(bc))
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
        page.AddPage(net.Name, NewNetwork(con, net), true, true)
    }

    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

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
