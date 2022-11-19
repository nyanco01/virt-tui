package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	libvirt "libvirt.org/go/libvirt"
	"github.com/nyanco01/virt-tui/src/virt"
)


type Volume struct {
    info        virt.Diskinfo
    attachVM    string
}


type Pool struct {
    *tview.Box
    name                string
    path                string
    capacity            uint64
    allocation          uint64
    volumes             []Volume

    // Offset for mouse scrolling
    lineOfset           int
    lineOfsetMax        int
    // Height of one previous drawing
    oldheight           int

    // Boolean type of whether the Create button is clicked or not.
    onClickCreate       bool
    clickVolIndex       int

    selectedCreateFunc  func()
    selectedDeleteFunc  func(volIndex int)
}


func NewPool(con *libvirt.Connect, n string) *Pool {
    path, capa, allo := virt.GetPoolInfo(con, n)
    infos := virt.GetDisksByPool(con, n)
    vols := []Volume{}
    for _, info := range infos {
        n := virt.GetBelongVM(con, info.Path)
        vols = append(vols, Volume{
            info:       info,
            attachVM:   n,
        })
    }
    return &Pool{
        Box:                tview.NewBox(),
        name:               n,
        path:               path,
        capacity:           capa,
        allocation:         allo,
        volumes:            vols,
        lineOfset:          0,
        lineOfsetMax:       0,
        oldheight:          0,
        onClickCreate:      false,
        clickVolIndex:      -1,
        selectedCreateFunc: nil,
        selectedDeleteFunc: nil,
    }
}


func(p *Pool)Draw(screen tcell.Screen) {
    p.Box.DrawForSubclass(screen, p)
    x, y, w, h := p.GetInnerRect()

    usagePool := float64(p.allocation) / float64(p.capacity)

    PoolBar := ""
    for k := 0; k < w-5; k++ {
        PoolBar += "■"
    }

    tview.Print(screen, fmt.Sprintf("Name       : %s", p.name), x+3, y, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Path       : %s", p.path), x+3, y+1, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Capacity   : %.2f GB", float64(p.capacity)/1024/1024/1024), x+50, y, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("allocation : %.2f GB", float64(p.allocation)/1024/1024/1024), x+50, y+1, w, tview.AlignLeft, tcell.ColorWhiteSmoke)

    var avaPool float64
    if p.capacity <= p.allocation {
        avaPool = 0
    } else {
        avaPool = float64(p.capacity - p.allocation)/1024/1024/1024
    }
    tview.Print(screen, fmt.Sprintf("available  : %.2f GB", avaPool), x+50, y+2, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, PoolBar, x+3, y+4, w, tview.AlignLeft, tcell.NewRGBColor(80, 80, 80))
    poolColor := setColorGradation(DISK_COLOR, int(usagePool * float64(w-5)))
    for k := 0; k < int(usagePool * float64(w-5)); k++ {
        tview.Print(screen, "■", x+3 + k, y+4, w, tview.AlignLeft, poolColor[k])
    }
    tview.Print(screen, fmt.Sprintf("used : %.2f%%", usagePool*100), x+3, y+3, w, tview.AlignLeft,tcell.ColorWhiteSmoke)

    for fillY := y; fillY < y+5; fillY++ {
        tview.Print(screen, "▐", x+1, fillY, w, tview.AlignLeft, tcell.ColorYellow)
    }

    l := len(p.volumes)-1
    volY := y+8
    tview.Print(screen, "│", x+1,volY-3, w, tview.AlignLeft, tcell.ColorLightYellow)
    tview.Print(screen, "├─", x+1,volY-2, w, tview.AlignLeft, tcell.ColorLightYellow)
    tview.Print(screen, "│", x+1,volY-1, w, tview.AlignLeft, tcell.ColorLightYellow)

    // Height required for the entire list of volumes to be drawn.
    fullHeight := 6*(l+1)

    def := tcell.StyleDefault
    background := def.Background(tcell.NewRGBColor(40, 40, 40))
    var createButtonColor tcell.Color
    if p.onClickCreate {
        for i := 3+x; i < 24+x; i++ {
            screen.SetContent(i, 7, ' ', nil, background)
        }
        createButtonColor = tcell.ColorGray
    } else {
        createButtonColor = tcell.ColorYellow
    }
    tview.Print(screen, "[+] New volume create", x+3, volY-2, w, tview.AlignLeft, createButtonColor)

    VolBar := ""
    for k := 0; k < w-10; k++ {
        VolBar += "■"
    }

    var avaVol float64

    if h - 7 >= fullHeight {
        // Drawing a Volume
        for i, vol := range p.volumes {
            for n := volY+(i*6); n <= volY+4+(i*6); n++ {
                var volColor tcell.Color
                if i == p.clickVolIndex {
                    volColor = tcell.ColorOrange
                    for k := x+3; k <= x+w-7; k++ {
                        screen.SetContent(k, n, ' ', nil, background)
                    }
                } else {
                    volColor = tcell.ColorLightYellow
                }
                tview.Print(screen, "▐", x+3, n, w, tview.AlignLeft, volColor)
            }

            usageVol := float64(vol.info.Allocation) / float64(vol.info.Capacity)
            tview.Print(screen, fmt.Sprintf("Attached VM : %s", vol.attachVM), x+5, volY+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("Path        : %s", vol.info.Path), x+5, volY+1+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("Capacity    : %.2f GB", float64(vol.info.Capacity)/1024/1024/1024), x+55, volY+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("Allocation  : %.2f GB", float64(vol.info.Allocation)/1024/1024/1024), x+55, volY+1+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            if vol.info.Capacity <= vol.info.Allocation {
                avaVol = 0
            } else {
                avaVol = float64(vol.info.Capacity - vol.info.Allocation)/1024/1024/1024
            }
            tview.Print(screen, fmt.Sprintf("Available   : %.2f GB", avaVol), x+55, volY+2+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, fmt.Sprintf("Used : %.2f%%",usageVol*100), x+5, volY+3+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            tview.Print(screen, VolBar, x+5, volY+4+(i*6), w, tview.AlignLeft, tcell.NewRGBColor(80, 80, 80))
            volColor := setColorGradation(DISK_COLOR, int(usageVol * float64(w-10)))
            for k := 0; k < int(usageVol * float64(w-10)); k++ {
                tview.Print(screen, "■", x+5 + k, volY+4+(i*6), w, tview.AlignLeft, volColor[k])
            }


            if i == l {
                tview.Print(screen, "└", x+1, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
            } else {
                for j := volY+(i*6); j <= volY+5+(i*6); j++ {
                    tview.Print(screen, "│", x+1, j, w, tview.AlignLeft, tcell.ColorLightYellow)
                }
                tview.Print(screen, "├", x+1, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
            }
            tview.Print(screen, "─", x+2, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
        }

        p.lineOfset = 0
    } else {
        // Volume is displayed by mouse scroll.
        if p.oldheight != 0 && p.oldheight < h && p.lineOfset != 0 {
            p.lineOfset--
        }
        p.lineOfsetMax = fullHeight - (h - 7)
        cnt := p.lineOfset
        var vols int = 0
        for i := volY; i <= y+h; i++ {
            vols = (i + p.lineOfset - volY)/6
            // If the terminal is vigorously resized vertically, the calculation may not be completed in time.
            if vols < 0 {
                vols = 0
            }

            // Prevent out-of-array references when volume is deleted
            if len(p.volumes) - 1 < vols {
                break
            }

            if cnt % 6 != 5 {
                var volColor tcell.Color
                // Change the display depending on whether it is selected or not
                if vols == p.clickVolIndex {
                    volColor = tcell.ColorOrange
                    for k := x+3; k <= x+w-7; k++ {
                        screen.SetContent(k, i, ' ', nil, background)
                    }
                } else {
                    volColor = tcell.ColorLightYellow
                }
                tview.Print(screen, "▐", x+3, i, w, tview.AlignLeft, volColor)
            }

            usageVol := float64(p.volumes[vols].info.Allocation) / float64(p.volumes[vols].info.Capacity)
            switch cnt % 6 {
            case 0:
                tview.Print(screen, fmt.Sprintf("Attached VM : %s", p.volumes[vols].attachVM), x+5, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
                tview.Print(screen, fmt.Sprintf("Capacity    : %.2f GB", float64(p.volumes[vols].info.Capacity)/1024/1024/1024), x+55, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
                if vols == l {
                    tview.Print(screen, "└─", x+1, i, w, tview.AlignLeft, tcell.ColorLightYellow)
                } else {
                    tview.Print(screen, "├─", x+1, i, w, tview.AlignLeft, tcell.ColorLightYellow)
                }
            case 1:
                tview.Print(screen, fmt.Sprintf("Path        : %s", p.volumes[vols].info.Path), x+5, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
                tview.Print(screen, fmt.Sprintf("Allocation  : %.2f GB", float64(p.volumes[vols].info.Allocation)/1024/1024/1024), x+55, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            case 2:
                if p.volumes[vols].info.Capacity <= p.volumes[vols].info.Allocation {
                    avaVol = 0
                } else {
                    avaVol = float64(p.volumes[vols].info.Capacity - p.volumes[vols].info.Allocation)/1024/1024/1024
                }
                tview.Print(screen, fmt.Sprintf("Available   : %.2f GB", avaVol), x+55, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            case 3:
                tview.Print(screen, fmt.Sprintf("Used : %.2f%%", usageVol*100), x+5, i, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
            case 4:
                tview.Print(screen, VolBar, x+5, i, w, tview.AlignLeft, tcell.NewRGBColor(80, 80, 80))
                volColor := setColorGradation(DISK_COLOR, int(usageVol * float64(w-10)))
                for k := 0; k < int(usageVol * float64(w-10)); k++ {
                    tview.Print(screen, "■", x+5 + k, i, w, tview.AlignLeft, volColor[k])
                }
            }

            if vols != l && cnt % 6 != 0 {
                tview.Print(screen, "│", x+1, i, w, tview.AlignLeft, tcell.ColorLightYellow)
            }
            cnt++
        }

        p.oldheight = h
    }
}


func (p *Pool)SetCreateVolFunc(handler func()) *Pool {
    p.selectedCreateFunc = handler
    return p
}


func (p *Pool)SetDeleteVolFunc(handler func(volIndex int)) *Pool {
    p.selectedDeleteFunc = handler
    return p
}


func (p *Pool)MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return p.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
		if !p.InRect(x, y) {
			return false, nil
		}

        var volSpacers bool
        if (y - 8 + p.lineOfset) % 6 == 0 {
            volSpacers = true
        } else {
            volSpacers = false
        }

        px, py, _, _ := p.GetInnerRect()
        p.clickVolIndex = -1
        switch action {
        case tview.MouseScrollUp:
            if p.lineOfset > 0 {
                p.lineOfset--
                consumed = true
            }
        case tview.MouseScrollDown:
            if p.lineOfset < p.lineOfsetMax {
                p.lineOfset++
                consumed = true
            }
        case tview.MouseLeftClick:
            if 3+px <= x && x <= 23+px && y == 6+py {
                p.onClickCreate = true
                if p.selectedCreateFunc != nil {
                    p.selectedCreateFunc()
                }
            }else {
                p.onClickCreate = false
                if 3+px <= x && 8+py <= y && !volSpacers {
                    p.clickVolIndex = (y - 8 + p.lineOfset) / 6
                    // Exception handling when an item is clicked on a location after it has been erased
                    if len(p.volumes)-1 < p.clickVolIndex {
                        break
                    }
                    if p.selectedDeleteFunc != nil {
                        p.selectedDeleteFunc(p.clickVolIndex)
                    }
                }
            }
            consumed = true
        }

        return
    })
}

// Added a function to display Modal to Primitive that displays Pool information.
func SetModal(app *tview.Application, con *libvirt.Connect, pool *Pool, page *tview.Pages) {
    pool.SetCreateVolFunc(func() {
        CreateVolModal := MakeVolumeCreate(app, con, pool, page)
        page.AddPage("CreateVolume", CreateVolModal, true, true)
        app.SetFocus(CreateVolModal)
    })
    pool.SetDeleteVolFunc(func(volIndex int) {
        DeleteVolModal := MakeVolDelete(app, con, page, pool, volIndex)
        page.AddPage("DeleteVolume", DeleteVolModal, true, true)
        app.SetFocus(DeleteVolModal)
    })
}


func MakeVolMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))

    poolInfo := virt.GetPoolList(con)
    for i, name := range poolInfo.Name {
        list.AddItem(name, "", rune(i+'0'), nil)
        pool := NewPool(con, name)
        SetModal(app, con, pool, page)
        page.AddPage(name, pool, true, true)
    }

    // Displays the page corresponding to the selected item
    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

    list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
        modal := MakePoolDelete(app, con, page, list, s1)
        page.AddPage("DeletePool", modal, true, true)
    })

    btCreate := tview.NewButton("Create")
    btCreate.SetBackgroundColor(tcell.Color220)
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
        modal := MakePoolCreate(app, con, page, list)
        if page.HasPage("DeletePool"){
            page.RemovePage("DeletePool")
        }
        page.AddPage("Create", modal, true, true)
        app.SetFocus(modal)
    })

    if list.GetItemCount() != 0 {
        main, _ := list.GetItemText(list.GetCurrentItem())
        page.SwitchToPage(main)
    }

    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1,true).
        AddItem(btCreate, 5, 0, false)

    return flex
}


func MakeVolUI(app *tview.Application, con *libvirt.Connect) *tview.Flex {
    flex := tview.NewFlex()

    page := MakePages(app)
    menu := MakeVolMenu(app, con, page)

    _, _, w, _ := menu.GetInnerRect()
    flex.AddItem(menu, w + 1, 0, true)
    flex.AddItem(page, 0, 1, false)

    return flex
}
