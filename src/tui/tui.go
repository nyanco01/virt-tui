package tui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"
)


func still() *tview.Box {
    box := tview.NewBox().SetBorder(false)
    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, "in progress", x+1, y + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)

        return x + 1, (y - (height / 2)) + 1, width - 2, height -(y - (height / 2)) + 1 - y
    })
    return box
}


func belowMinimumSize() *tview.Box {
    box := tview.NewBox().SetBorder(false)
    box.SetBackgroundColor(tcell.ColorBlack.TrueColor())
    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, "Your terminal size", x+1, y - 1 + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)
        tview.Print(screen, fmt.Sprintf("width: [orange]%d[white], height: [orange]%d", width, height), x+1, y + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)
        tview.Print(screen, "Required terminal size", x+1, y + 1 + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)
        tview.Print(screen, fmt.Sprintf("width: [lightgreen]%d[white], height: [orange]%d", 95, 35), x+1, y + 2 + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)

        return x + 1, (y - (height / 2)) + 1, width - 2, height -(y - (height / 2)) + 1 - y
    })
    return box
}


func configStyles() {
    bgc := tcell.NewRGBColor(0, 0, 0)
    tview.Styles = tview.Theme{
            PrimitiveBackgroundColor:    bgc,
            ContrastBackgroundColor:     tcell.ColorDarkBlue,
            MoreContrastBackgroundColor: tcell.ColorGreen,
            BorderColor:                 tcell.ColorWhite,
            TitleColor:                  tcell.ColorWhite,
            GraphicsColor:               tcell.ColorWhite,
            PrimaryTextColor:            tcell.ColorGhostWhite,
            SecondaryTextColor:          tcell.ColorYellow,
            TertiaryTextColor:           tcell.ColorGreen,
            InverseTextColor:            tcell.ColorDeepSkyBlue,
            ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
    }
}


func MakeApp() *tview.Application {
    configStyles()
    app := tview.NewApplication()

    // connect libvirt
    c, err := libvirt.NewConnect("qemu:///system")
    if err != nil {
        log.Fatalf("failed to qemu connection: %v", err)
    }

    // Flex for layout of buttons and pages as the main UI part
    flex := tview.NewFlex()
    flex.SetDirection(tview.FlexRow)

    page := tview.NewPages()
    page.AddPage("vm", MakeVMUI(app, c), true, true)
    page.AddPage("volume", MakeVolUI(app, c), true, true)
    page.AddPage("network", MakeNetUI(app, c), true, true)
    page.SwitchToPage("vm")

    btVM := tview.NewButton("[#F66640::]VMs").SetSelectedFunc(func() {
        page.SwitchToPage("vm")
    })
    btVM.SetBackgroundColor(tcell.NewRGBColor(120, 120, 120))
    //btVM.SetLabelColor(tcell.NewRGBColor(19, 83, 112))
    btVolume := tview.NewButton("[#FFE15C]Volume").SetSelectedFunc(func() {
        page.SwitchToPage("volume")
    })
    btVolume.SetBackgroundColor(tcell.NewRGBColor(80, 80, 80))
    btNetwork := tview.NewButton("[#1C7AA2]Network").SetSelectedFunc(func() {
        page.SwitchToPage("network")
    })
    btNetwork.SetBackgroundColor(tcell.NewRGBColor(120, 120, 120))
    btQuit := tview.NewButton("[#6DC1B3]Quit").SetSelectedFunc(func() {
        app.Stop()
    })
    btQuit.SetBackgroundColor(tcell.NewRGBColor(80, 80, 80))

    // flexMenu is a Flex that is created to organize buttons for switching between pages.
    flexMenu := tview.NewFlex().
        AddItem(btVM, 0, 1, true).
        AddItem(btVolume, 0, 1, false).
        AddItem(btNetwork, 0, 1, false).
        AddItem(btQuit, 0, 1, false)

    flexMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyLeft:
            if btVolume.HasFocus() {
                app.SetFocus(btVM)
                page.SwitchToPage("vm")
            } else if btNetwork.HasFocus() {
                app.SetFocus(btVolume)
                page.SwitchToPage("volume")
            } else if btQuit.HasFocus() {
                app.SetFocus(btNetwork)
                page.SwitchToPage("network")
            }
        case tcell.KeyRight:
            if btVM.HasFocus() {
                app.SetFocus(btVolume)
                page.SwitchToPage("volume")
            } else if btVolume.HasFocus() {
                app.SetFocus(btNetwork)
                page.SwitchToPage("network")
            } else if btNetwork.HasFocus() {
                app.SetFocus(btQuit)
            }
        case tcell.KeyTab:
            app.SetFocus(page)
        }

        return event
    })

    flex.AddItem(flexMenu, 1, 0, true)
    flex.AddItem(page, 0, 1, false)

    app.SetFocus(flex)
    app.SetRoot(flex, true)

    return app
}
