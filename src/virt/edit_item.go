package virt

type EditItem interface {
    GetItemType() string
    //SetSelectedFunc(handler func()) EditItem
    GetSelectedFunc() func()
}


type ItemCPU struct {
    EditItem
    Number          uint
    PlaceMent       string
    CPUSet          string
    Mode            string
    Selectedfunc    func()
}


func (c ItemCPU) GetItemType() string {
    return "CPU"
}


func (c *ItemCPU) SetSelectedFunc(handler func()) *ItemCPU {
    c.Selectedfunc = handler
    return c
}


func (c *ItemCPU) GetSelectedFunc() func() {
    return c.Selectedfunc
}


type ItemMemory struct {
    EditItem
    Size            uint
    SizeSI          string
    MaxSize         uint
    MaxSizeSI       string
    CurrentMemory   uint
    CurrentMemorySI string
    Selectedfunc    func()
}


func (m ItemMemory) GetItemType() string {
    return "Memory"
}


func (m *ItemMemory) SetSelectedFunc(handler func()) *ItemMemory {
    m.Selectedfunc = handler
    return m
}


func (m *ItemMemory) GetSelectedFunc() func() {
    return m.Selectedfunc
}


type ItemDisk struct {
    EditItem
    Path            string
    Device          string
    ImgType         string
    Bus             string
    Selectedfunc    func()
}

func (d ItemDisk) GetItemType() string {
    return "Disk"
}


func (d *ItemDisk) SetSelectedFunc(handler func()) *ItemDisk {
    d.Selectedfunc = handler
    return d
}


func (d *ItemDisk) GetSelectedFunc() func() {
    return d.Selectedfunc
}


type ItemController struct {
    EditItem
    ControllerType  string
    Model           string
    Selectedfunc    func()
}


func (c ItemController) GetItemType() string {
    return "Controller"
}


func (c *ItemController) SetSelectedFunc(handler func()) *ItemController {
    c.Selectedfunc = handler
    return c
} 


func (c *ItemController) GetSelectedFunc() func() {
    return c.Selectedfunc
}


type ItemInterface struct {
    EditItem
    IfType          string
    Source          string
    Model           string
    Driver          string
    Selectedfunc    func()
}


func (i ItemInterface) GetItemType() string {
    return "Interface"
}


func (i *ItemInterface) SetSelectedFunc(handler func()) *ItemInterface {
    i.Selectedfunc = handler
    return i
}


func (i *ItemInterface) GetSelectedFunc() func() {
    return i.Selectedfunc
}


type ItemSerial struct {
    EditItem
    TargetType      string
    Selectedfunc    func()
}


func (s ItemSerial) GetItemType() string {
    return "Serial"
}


func (s *ItemSerial) SetSelectedFunc(handler func()) *ItemSerial {
    s.Selectedfunc = handler
    return s
}


func (s *ItemSerial) GetSelectedFunc() func() {
    return s.Selectedfunc
}


type ItemConsole struct {
    EditItem
    TargetType      string
    Selectedfunc    func()
}


func (c ItemConsole) GetItemType() string {
    return "Console"
}


func (c *ItemConsole) SetSelectedFunc(handler func()) *ItemConsole {
    c.Selectedfunc = handler
    return c
}


func (c *ItemConsole) GetSelectedFunc() func() {
    return c.Selectedfunc
}


type ItemInput struct {
    EditItem
    InputType       string
    Bus             string
    Selectedfunc    func()
}


func (i ItemInput) GetItemType() string {
    return "Input"
}


func (i *ItemInput) SetSelectedFunc(handler func()) *ItemInput {
    i.Selectedfunc = handler
    return i
}


func (i *ItemInput) GetSelectedFunc() func() {
    return i.Selectedfunc
}


type ItemGraphics struct {
    EditItem
    GraphicsType    string
    Port            int
    ListemAddress   string
    Selectedfunc    func()
}


func (g ItemGraphics) GetItemType() string {
    return "Graphics"
}


func (g *ItemGraphics) SetSelectedFunc(handler func()) *ItemGraphics {
    g.Selectedfunc = handler
    return g
}


func (g *ItemGraphics) GetSelectedFunc() func() {
    return g.Selectedfunc
}


type ItemVideo struct {
    EditItem
    ModelType       string
    VRAM            uint
    DeviceAddress   string
    Selectedfunc    func()
}


func (v ItemVideo) GetItemType() string {
    return "Video"
}


func (v *ItemVideo) SetSelectedFunc(handler func()) *ItemVideo {
    v.Selectedfunc = handler
    return v
}


func (v *ItemVideo) GetSelectedFunc() func() {
    return v.Selectedfunc
}


