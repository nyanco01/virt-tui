package tui

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/go/libvirt"

	"github.com/nyanco01/virt-tui/src/operate"
	"github.com/nyanco01/virt-tui/src/virt"
)


type ProgressBar struct {
    *tview.Box
    rate        float64
}


func InputFieldPositiveInteger(text string, ch rune) bool {
		if text == "-" {
			return false
		}
		_, err := strconv.Atoi(text)
		return err == nil
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
}

func errorForm(app *tview.Application, list *tview.List, page *tview.Pages) *tview.Form {
    form := tview.NewForm()
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })
    return form
}


func MakeVMCreateForm(app *tview.Application, con *libvirt.Connect, view *tview.TextView, list *tview.List, page *tview.Pages, bar *ProgressBar) (*tview.Form, error) {
    // get libvirt status
    vms := virt.LookupVMs(con)
    maxCPUs, maxMem := virt.GetNodeMax(con)
    _, listVNCPort := virt.GetUsedResources(vms)
    listPool := virt.GetPoolList(con)

    form := tview.NewForm()

    var err error = nil
    if maxCPUs < 0 {
        err = errors.New("The number of CPUs must be higher than 2 to create a VM")
        return errorForm(app, list, page), err
    }
    if len(listPool.Name) == 0 {
        err = errors.New("No storage pool has been created")
        return errorForm(app, list, page), err
    }

    // domain name              item index 0
    form.AddInputField("VM name", "", 30, nil, nil)
    
    // CPU count                item index 1
    optionCPU := []string{}
    for i := 1; i <= maxCPUs; i++ {
        optionCPU = append(optionCPU, strconv.Itoa(i))
    }
    form.AddDropDown("CPU number", optionCPU, 0, nil)
    
    // Memory size              item index 2
    form.AddInputField(fmt.Sprintf("Memory Size [orange]MB (max. %d MB) ", int(maxMem/kibi)), "", 10, nil, nil)
    form.GetFormItem(2).(*tview.InputField).SetAcceptanceFunc(InputFieldPositiveInteger)
    
    // Disk pool path           item index 3
    form.AddDropDown("Storage pool", listPool.Name, 0, nil)
    // Disk Size                item index 4
    form.AddInputField(fmt.Sprintf("Disk Size [orange]GB (max %.1f GB)", float64((listPool.Avalable[0] - gibi) / gibi)), "", 6, nil, nil)
    form.GetFormItem(4).(*tview.InputField).SetAcceptanceFunc(InputFieldPositiveInteger)

    // VNC port number          item index 5
    vncPort := 5901
    for i := range listVNCPort {
        if listVNCPort[i] == vncPort {
            vncPort++
        }
    }
    form.AddInputField("VNC Port", strconv.Itoa(vncPort), 6, nil, nil)
    form.GetFormItem(5).(*tview.InputField).SetAcceptanceFunc(InputFieldPositiveInteger)

    // Changing the maximum disk size
    form.GetFormItem(3).(*tview.DropDown).SetDoneFunc(func(key tcell.Key) {
        if (key == tcell.KeyTab) || (key == tcell.KeyBacktab) {
            index, _ := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
            form.GetFormItem(4).(*tview.InputField).SetLabel(fmt.Sprintf("Disk Size [orange]GB (max %.1f GB)", float64((listPool.Avalable[index] - (2*gibi)) / gibi) ))
        }
    })

    // Network Interface        item index 6
    form.AddDropDown("Network bridge interface", operate.ListBridgeIF(), 0, nil)

    // OS Type List             item index 7
    form.AddDropDown("OS Type", operate.GetOSTypeList(), 0, nil)

    //cloud-init
    // Host name                item index 8
    form.AddInputField("host name", "", 30, nil, nil)
    // guest vm user name       item index 9
    form.AddInputField("user name", "", 30, nil, nil)
    // guest vm user password   item index 10
    form.AddPasswordField("user password", "", 30, '*', nil)

    c := make(chan float64)
    statusTxt := make(chan string)
    done := make(chan int)

    form.AddButton("Create", func() {
        _, cpunum := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
        cpu, _ := strconv.Atoi(cpunum)
        memSize, _ := strconv.Atoi(form.GetFormItem(2).(*tview.InputField).GetText())
        poolIndex, _ := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
        dSize, _ := strconv.Atoi(form.GetFormItem(4).(*tview.InputField).GetText())
        VNCp, _ := strconv.Atoi(form.GetFormItem(5).(*tview.InputField).GetText())
        _, brName := form.GetFormItem(6).(*tview.DropDown).GetCurrentOption()
        _, ostype := form.GetFormItem(7).(*tview.DropDown).GetCurrentOption()
        request := virt.CreateVMRequest{
            DomainName:     form.GetFormItem(0).(*tview.InputField).GetText(),
            CPUNum:         cpu,
            MemNum:         memSize,
            DiskPath:       listPool.Path[poolIndex],
            DiskSize:       dSize,
            VNCPort:        VNCp,
            NICBridgeIF:    brName,
            OSType:         ostype,
            HostName:       form.GetFormItem(8).(*tview.InputField).GetText(),
            UserName:       form.GetFormItem(9).(*tview.InputField).GetText(),
            UserPassword:   form.GetFormItem(10).(*tview.InputField).GetText(),
        }

        b, ErrInfo := virt.CheckCreateVMRequest(request, con)

        if b {
            // Start creating a VM
            view.SetText("OK!").SetTextColor(tcell.ColorSkyblue)
            go UpdateBar(c, statusTxt, bar, view, app)
            go virt.CreateDomain(request, con, c, statusTxt, done)
            go doneCreate(request.DomainName,con, list, page, app, done)
        } else {
            view.SetText(ErrInfo).SetTextColor(tcell.ColorRed)
        }
    })
    form.AddButton("Cancel", func() {
        page.RemovePage("Create")
        app.SetFocus(list)
    })

    return form, nil
}


func MakeVMCreate(app *tview.Application, con *libvirt.Connect, page *tview.Pages, list *tview.List) tview.Primitive {
    flex := tview.NewFlex()
    flex.SetBorder(true).SetTitle("Create VM Menu")
    bar := NewProgressBar()
    view := tview.NewTextView()

    form, err := MakeVMCreateForm(app, con, view, list, page, bar)

    if err != nil {
        view.SetText(err.Error())
        flex.SetDirection(tview.FlexRow).
            AddItem(view, 1, 0, false).
            AddItem(form, 3, 0, true)
        return pageModal(flex, 65, 6)
    }

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(bar, 1, 0, false).
        AddItem(view, 1, 0, false)

    return pageModal(flex, 65, 29)
}


func doneCreate(name string, con *libvirt.Connect, list *tview.List, page *tview.Pages, app *tview.Application, done chan int) {
    <-done

    list.AddItem(name, "shutdown", rune(list.GetItemCount())+'0', nil)
    list.SetCurrentItem(list.GetItemCount())
    page.AddPage(name, NewVMStatus(app, virt.VMStatus[name]), true, true)
    app.SetFocus(list)
    page.RemovePage("Create")
    app.Draw()
}
