package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

	"github.com/nyanco01/virt-tui/src/virt"
)

type Volume struct {
    info        virt.Diskinfo
    attachVM    string
}

type Pool struct {
    *tview.Box
    name        string
    path        string
    capacity    uint64
    allocation  uint64
    volumes     []Volume
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
        Box:            tview.NewBox(),
        name:           n,
        path:           path,
        capacity:       capa,
        allocation:     allo,
        volumes:        vols,
    }
}

func(p *Pool)Draw(screen tcell.Screen) {
    p.Box.DrawForSubclass(screen, p)
    x, y, w, _ := p.GetInnerRect()

    // fill background color test
    for fillY := y; fillY < y+5; fillY++ {
        /*
        for fillX := x; fillX < x+w; fillX++ {
            screen.SetContent(fillX, fillY, ' ', nil, tcell.StyleDefault.Background(tcell.ColorLightSeaGreen))
        }
        */
        tview.Print(screen, "▐", x+1, fillY, w, tview.AlignLeft, tcell.ColorYellow)
    }

    tview.Print(screen, fmt.Sprintf("Name       : %s", p.name), x+3, y, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Path       : %s", p.path), x+3, y+1, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("Capacity   : %.2f GB", float64(p.capacity)/1024/1024/1024), x+3, y+2, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("allocation : %.2f GB", float64(p.allocation)/1024/1024/1024), x+3, y+3, w, tview.AlignLeft, tcell.ColorWhiteSmoke)
    tview.Print(screen, fmt.Sprintf("available  : %.2f GB", float64(p.capacity-p.allocation)/1024/1024/1024), x+3, y+4, w, tview.AlignLeft, tcell.ColorWhiteSmoke)

    l := len(p.volumes)-1
    volY := y+6
    tview.Print(screen, "│", x+1,volY-1, w, tview.AlignLeft, tcell.ColorLightYellow)
    for i, vol := range p.volumes {
        tview.Print(screen, fmt.Sprintf("Attached VM : %s", vol.attachVM), x+5, volY+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
        tview.Print(screen, fmt.Sprintf("Path        : %s", vol.info.Path), x+5, volY+1+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
        tview.Print(screen, fmt.Sprintf("Capacity    : %.2f", float64(vol.info.Capacity)/1024/1024/1024), x+5, volY+2+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
        tview.Print(screen, fmt.Sprintf("Allocation  : %.2f", float64(vol.info.Allocation)/1024/1024/1024), x+5, volY+3+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)
        tview.Print(screen, fmt.Sprintf("Available   : %.2f", float64(vol.info.Capacity-vol.info.Allocation)/1024/1024/1024), x+5, volY+4+(i*6), w, tview.AlignLeft, tcell.ColorWhiteSmoke)

        for yy := volY+(i*6); yy <= volY+4+(i*6); yy++ {
            tview.Print(screen, "▐", x+3, yy, w, tview.AlignLeft, tcell.ColorLightYellow)
        }
        if i == l {
            tview.Print(screen, "┕", x+1, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
        } else {
            for j := volY+(i*6); j <= volY+5+(i*6); j++ {
                tview.Print(screen, "│", x+1, j, w, tview.AlignLeft, tcell.ColorLightYellow)
            }
            tview.Print(screen, "├", x+1, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
        }
        tview.Print(screen, "─", x+2, volY+(i*6), w, tview.AlignLeft, tcell.ColorLightYellow)
    }
}

func NewPoolStatus(app *tview.Application, con *libvirt.Connect, name string) *tview.Flex {
    flex := tview.NewFlex()
    flex.AddItem(NewPool(con, name), 0, 1, true)
    return flex
}


func CreateVolMenu(app *tview.Application, con *libvirt.Connect, page *tview.Pages) *tview.Flex {
    flex := tview.NewFlex()
    list := tview.NewList()
    list.SetBorder(false).SetBackgroundColor(tcell.NewRGBColor(40,40,40))

    poolInfo := virt.GetPoolList(con)
    for i, name := range poolInfo.Name {
        list.AddItem(name, "", rune(i)+'0', nil)
        page.AddPage(name, NewPoolStatus(app, con, name), true, true)
    }

    list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if page.HasPage(mainText) {
            page.SwitchToPage(mainText)
        }
    })

    flex.SetDirection(tview.FlexRow).
        AddItem(list, 0, 1,true)

    return flex
}


func CreateVolUI(app *tview.Application, con *libvirt.Connect) *tview.Flex {
    flex := tview.NewFlex()

    page := CreatePages(app)
    menu := CreateVolMenu(app, con, page)

    _, _, w, _ := menu.GetInnerRect()
    flex.AddItem(menu, w + 1, 0, true)
    flex.AddItem(page, 0, 1, false)

    return flex
}
