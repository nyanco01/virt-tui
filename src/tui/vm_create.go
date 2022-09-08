package tui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

	"github.com/nyanco01/virt-tui/src/virt"
)

type ProgressBar struct {
    *tview.Box
    rate        float64
}

func NewProgressBar() *ProgressBar {
    return &ProgressBar {
        Box:    tview.NewBox(),
        rate:   0.0,
    }
}

func (p *ProgressBar)Draw(screen tcell.Screen) {
    p.Box.DrawForSubclass(screen, p)
    x, y, w, _ := p.GetInnerRect()

    gradient := float64(100) / float64(w * 8)

    bar := ""
    r := p.rate
    for i := 0; i < w; i++ {
        if r > (gradient * 8) {
            bar += "█"
            r -= gradient * 8
        } else {
            switch int(r / gradient) {
            case 0:
                break
            case 1:
                bar += "▏"
            case 2:
                bar += "▎"
            case 3:
                bar += "▍"
            case 4:
                bar += "▌"
            case 5:
                bar += "▋"
            case 6:
                bar += "▊"
            case 7:
                bar += "▉"
            }
            r = 0
        }
    }
    tview.Print(screen, bar, x, y, w, tview.AlignLeft, tcell.ColorSkyblue)
}

func (p *ProgressBar)Update(newrate float64) {
    p.rate = newrate
}

func UpdateBar(c chan float64, status chan string, p *ProgressBar, view *tview.TextView, app *tview.Application) {
    for {
        select {
        case par := <-c:
            app.QueueUpdateDraw(func() {
                p.Update(par)
            })
        case st := <- status:
            app.QueueUpdateDraw(func() {
                view.SetText(st)
            })
        default:
            time.Sleep(500 * time.Millisecond)
        }
    }
    /*
    for i := range c {
        app.QueueUpdateDraw(func() {
            p.Update(i)
        })
    }
    */
}

func makeVMForm(app *tview.Application, con *libvirt.Connect, view *tview.TextView, list *tview.List, bar *ProgressBar) *tview.Form {
    // get libvirt status
    vms := virt.LookupVMs(con)
    maxCPUs, maxMem := virt.GetNodeMax(con)
    _, listVNCPort := virt.GetUsedResources(vms)
    listPool := virt.GetPoolList(con)

    form := tview.NewForm()

    // domain name              item index 0
    form.AddInputField("VM name", "", 30, nil, nil)
    
    // CPU count                item index 1
    optionCPU := []string{}
    for i := 1; i <= maxCPUs; i++ {
        optionCPU = append(optionCPU, strconv.Itoa(i))
    }
    form.AddDropDown("CPU number", optionCPU, 0, nil)
    
    // Memory size              item index 2
    form.AddInputField(fmt.Sprintf("Memory Size [orange]MB (max. %d MB) ", int(maxMem/1024)), "", 10, nil, nil)
    form.GetFormItem(2).(*tview.InputField).SetAcceptanceFunc(tview.InputFieldInteger)
    
    // Disk pool path           item index 3
    form.AddDropDown("Strage pool", listPool.Name, 0, nil)
    // Disk Size                item index 4
    form.AddInputField(fmt.Sprintf("Disk Size [orange]GB (max %.1f GB)", float64((listPool.Avalable[0] - uint64(1024*1024*1024)) / uint64(1024*1024*1024))), "", 6, nil, nil)
    form.GetFormItem(4).(*tview.InputField).SetAcceptanceFunc(tview.InputFieldInteger)

    // VNC port number          item index 5
    vncPort := 5901
    for i := range listVNCPort {
        if listVNCPort[i] == vncPort {
            vncPort++
        }
    }
    form.AddInputField("VNC Port", strconv.Itoa(vncPort), 6, nil, nil)
    form.GetFormItem(5).(*tview.InputField).SetAcceptanceFunc(tview.InputFieldInteger)

    // Changing the maximum disk size
    form.GetFormItem(3).(*tview.DropDown).SetDoneFunc(func(key tcell.Key) {
        if (key == tcell.KeyTab) || (key == tcell.KeyBacktab) {
            index, _ := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
            form.GetFormItem(4).(*tview.InputField).SetLabel(fmt.Sprintf("Disk Size [orange]GB (max %.1f GB)", float64((listPool.Avalable[index] - uint64(2*1024*1024*1024)) / uint64(1024*1024*1024)) ))
        }
    })

    //cloud-init
    // Host name                item index 6
    form.AddInputField("host name", "", 30, nil, nil)
    // guest vm user name       item index 7
    form.AddInputField("user name", "", 30, nil, nil)
    // guest vm user password   item index 8
    form.AddPasswordField("user password", "", 30, '*', nil)

    form.AddButton("Create", func() {
        _, cpunum := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
        cpu, _ := strconv.Atoi(cpunum)
        memSize, _ := strconv.Atoi(form.GetFormItem(2).(*tview.InputField).GetText())
        poolIndex, _ := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
        dSize, _ := strconv.Atoi(form.GetFormItem(4).(*tview.InputField).GetText())
        VNCp, _ := strconv.Atoi(form.GetFormItem(5).(*tview.InputField).GetText())
        request := virt.CreateRequest{
            DomainName:     form.GetFormItem(0).(*tview.InputField).GetText(),
            CPUNum:         cpu,
            MemNum:         memSize,
            DiskPath:       listPool.Path[poolIndex],
            DiskSize:       dSize,
            VNCPort:        VNCp,
            HostName:       form.GetFormItem(6).(*tview.InputField).GetText(),
            UserName:       form.GetFormItem(7).(*tview.InputField).GetText(),
            UserPassword:   form.GetFormItem(8).(*tview.InputField).GetText(),
        }

        b, ErrInfo := virt.CheckCreateRequest(request, con)

        if b {
            view.SetText("OK!").SetTextColor(tcell.ColorSkyblue)
            c := make(chan float64)
            statusTxt := make(chan string)
            go virt.CreateDomain(request, con, c, statusTxt)
            go UpdateBar(c, statusTxt, bar, view, app)

        } else {
            view.SetText(ErrInfo).SetTextColor(tcell.ColorRed)
        }
    })
    form.AddButton("Cancel", func() {
        app.Stop()
    })

    return form
}

func CreateMakeVM(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create Menu")
    bar := NewProgressBar()
    view := tview.NewTextView()

    form := makeVMForm(app, con, view, list, bar)

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(bar, 1, 0, false).
        AddItem(view, 1, 0, false)


    return pageModal(flex, 65, 30)
}
