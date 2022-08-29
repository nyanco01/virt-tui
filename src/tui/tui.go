package tui

import (

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func still() *tview.Box {
    box := tview.NewBox().SetBorder(false)
    box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
        tview.Print(screen, "in progress", x+1, y + (height / 2), width - 2, tview.AlignCenter, tcell.ColorWhite)

        return x + 1, (y - (height / 2)) + 1, width - 2, height -(y - (height / 2)) + 1 - y
    })

    return box
}

func CreateApp() *tview.Application {
    app := tview.NewApplication()

    // Flex for layout of buttons and pages as the main UI part
    flex := tview.NewFlex()
    flex.SetDirection(tview.FlexRow)

    page := tview.NewPages()
    page.AddPage("vm", still(), true, true)
    page.AddPage("volume", still(), true, true)
    page.AddPage("network", still(), true, true)

    btVM := tview.NewButton("VMs").SetSelectedFunc(func() {
        page.SwitchToPage("vm")
    })
    btVolume := tview.NewButton("Volume").SetSelectedFunc(func() {
        page.SwitchToPage("volume")
    })
    btNetwork := tview.NewButton("Network").SetSelectedFunc(func() {
        page.SwitchToPage("network")
    })
    btQuit := tview.NewButton("Quit").SetSelectedFunc(func() {
        app.Stop()
    })




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
            } else if btNetwork.HasFocus() {
                app.SetFocus(btVolume)
            } else if btQuit.HasFocus() {
                app.SetFocus(btNetwork)
            }
        case tcell.KeyRight:
            if btVM.HasFocus() {
                app.SetFocus(btVolume)
            } else if btVolume.HasFocus() {
                app.SetFocus(btNetwork)
            } else if btNetwork.HasFocus() {
                app.SetFocus(btQuit)
            }
        }

        return event
    })

    flex.AddItem(flexMenu, 1, 0, true)
    flex.AddItem(page, 0, 1, false)

    app.SetFocus(flex)
    app.SetRoot(flex, true)

    return app
}
