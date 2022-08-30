package tui

import "github.com/gdamore/tcell/v2"

const (
    CPU_COLOR = iota
    MEMORY_COLOR
    DISK_COLOR
    NIC_UP_COLOR
    NIC_DOWN_COLOR
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

        red         := float64(diffRed) / float64(num)
        green       := float64(diffGreen) / float64(num)
        blue        := float64(diffBlue) / float64(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(206 - int32(red*float64(i)), 56 - int32(green*float64(i)), 64 - int32(blue*float64(i))))
        }
    case MEMORY_COLOR:
        // first color rgb 254 78 19
        // second color rgb 126 38 9
        diffRed     := 126 - 254
        diffGreen   := 38 - 78
        diffBlue    := 9 - 19

        red         := float64(diffRed) / float64(num)
        green       := float64(diffGreen) / float64(num)
        blue        := float64(diffBlue) / float64(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(126 - int32(red*float64(i)), 38 - int32(green*float64(i)), 9 - int32(blue*float64(i))))
        }
    case DISK_COLOR:
        // first color rgb 244 202 44
        // second color rgb 248 215 128
        diffRed     := 244 - 248
        diffGreen   := 202 - 215
        diffBlue    := 44 - 128

        red         := float64(diffRed) / float64(num)
        green       := float64(diffGreen) / float64(num)
        blue        := float64(diffBlue) / float64(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(244 - int32(red*float64(i)), 202 - int32(green*float64(i)), 44 - int32(blue*float64(i))))
        }
    case NIC_UP_COLOR:
        // first color rgb 20 161 156
        // second color rgb 31 247 255
        diffRed     := 31 - 20
        diffGreen   := 247 - 161
        diffBlue    := 255 - 156

        red         := float64(diffRed) / float64(num)
        green       := float64(diffGreen) / float64(num)
        blue        := float64(diffBlue) / float64(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(31 - int32(red*float64(i)), 247 - int32(green*float64(i)), 255 - int32(blue*float64(i))))
        }
    case NIC_DOWN_COLOR:
        // first color rgb 80 70 149
        // second color rgb 141 232 237
        diffRed     := 80 - 141
        diffGreen   := 70 - 232
        diffBlue    := 149 - 237

        red         := float64(diffRed) / float64(num)
        green       := float64(diffGreen) / float64(num)
        blue        := float64(diffBlue) / float64(num)

        for i := 0; i < num; i++ {
            colors = append(colors, tcell.NewRGBColor(80 - int32(red*float64(i)), 70 - int32(green*float64(i)), 149 - int32(blue*float64(i))))
        }
    default:
        for i := 0; i < num; i++ {
            colors = append(colors, tcell.ColorWhite)
        }
    }

    return colors
}
