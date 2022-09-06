package tui

import (
	"fmt"
    "strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

	"github.com/nyanco01/virt-tui/src/virt"
)

func butItemCheck(item string) string {
    switch item {
    case "VMName":
        return "VM name field is wrong."
    case "Memory":
        return "Memory field is wrong."
    case "Disk":
        return "Disk field is wrong."
    case "VNC":
        return "VNC Port field is wrong."
    case "HostName":
        return "No host name"
    case "UserName":
        return "No user name"
    case "UserPass":
        return "No user password"
    }
    return ""
}

func makeVMForm(app *tview.Application, con *libvirt.Connect, view *tview.TextView, list *tview.List) *tview.Form {
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
    view := tview.NewTextView()

    form := makeVMForm(app, con, view, list)

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)


    return pageModal(flex, 65, 30)
}
