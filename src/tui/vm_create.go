package tui

import (
	"fmt"
    "strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	libvirt "libvirt.org/libvirt-go"

	"github.com/nyanco01/virt-tui/src/virt"
)

func butItem(item string) string {
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

func makeVMForm(app *tview.Application, con *libvirt.Connect, view *tview.TextView) *tview.Form {
    // get libvirt status
    vms := virt.LookupVMs(con)
    maxCPUs, maxMem := virt.GetNodeMax(con)
    listVMName, listVNCPort := virt.GetUsedResources(vms)
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
            form.GetFormItem(4).(*tview.InputField).SetLabel(fmt.Sprintf("Disk Size [orange]GB (max %.1f GB)", float64((listPool.Avalable[index] - uint64(1024*1024*1024)) / uint64(1024*1024*1024)) ))
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
        check := map[string]bool{}
        // VM Name
        nameVM := form.GetFormItem(0).(*tview.InputField).GetText()
        check["VMName"] = true
        for _, n := range listVMName {
            if n == nameVM {
                check["VMName"] = false
            }
        }
        if nameVM == "" { check["VMName"] = false }
        // Memory
        memSize, _ := strconv.Atoi(form.GetFormItem(2).(*tview.InputField).GetText())
        if memSize > int(maxMem / 1024) {
            check["Memory"] = false
        }
        if memSize == 0 { check["Memory"] = false }
        // Disk
        diskSize, _ := strconv.Atoi(form.GetFormItem(4).(*tview.InputField).GetText())
        if diskSize > 50 {
            check["Disk"] = false
        }
        if diskSize == 0 { check["Disk"] = false }
        // VNC Port
        port, _ := strconv.Atoi(form.GetFormItem(5).(*tview.InputField).GetText())
        for _, p := range listVNCPort {
            if p == port {
                check["VNC"] = false
            }
        }
        if port == 0 { check["VNC"] = false }
        // Host name
        hostName := form.GetFormItem(6).(*tview.InputField).GetText()
        if hostName == "" { check["HostName"] = false }
        // User name
        userName := form.GetFormItem(7).(*tview.InputField).GetText()
        if userName == "" { check["UserName"] = false }
        // User password
        userPass := form.GetFormItem(8).(*tview.InputField).GetText()
        if userPass == "" { check["UserPass"] = false }

        b := true
        out := ""
        for key, value := range check {
            if !value {
                out = key
                b = false
                break
            }
        }
        if b {

        } else {
            notice := butItem(out)
            view.SetText(notice).SetTextColor(tcell.ColorRed)
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

    form := makeVMForm(app, con, view)

    flex.SetDirection(tview.FlexRow).
        AddItem(form, 0, 1, true).
        AddItem(view, 1, 0, false)


    return pageModal(flex, 65, 30)
}
