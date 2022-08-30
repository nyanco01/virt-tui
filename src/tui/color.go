package tui

import "github.com/gdamore/tcell/v2"

const (
    CPU_COLOR = iota
    MEMORY_COLOR
    DISK_COLOR
    NIC_COLOR
)

func setColorGradation(colorType, num int) []tcell.Color{

    colors := []tcell.Color{}

    switch colorType {
    case CPU_COLOR:
        // first color rgb 0 255 127
        // second color rgb 206 56 64
        diffRed     := 206 - 0
        diffGreen   := 56 - 255
        diffBlue    := 64 - 127

        red         := float32(diffRed) / float32(num)
        green       := float32(diffGreen) / float32(num)
        blue        := float32(diffBlue) / float32(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(206 - int32(red*float32(i)), 56 - int32(green*float32(i)), 64 - int32(blue*float32(i))))
        }
    case MEMORY_COLOR:
        // first color rgb 254 78 19
        // second color rgb 126 38 9
        diffRed     := 126 - 254
        diffGreen   := 38 - 78
        diffBlue    := 9 - 19

        red         := float32(diffRed) / float32(num)
        green       := float32(diffGreen) / float32(num)
        blue        := float32(diffBlue) / float32(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(126 - int32(red*float32(i)), 38 - int32(green*float32(i)), 9 - int32(blue*float32(i))))
        }
    default:
        for i := 0; i < num; i++ {
            colors = append(colors, tcell.ColorWhite)
        }
    }

    return colors
}
