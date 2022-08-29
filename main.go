package main

import (


    "github.com/nyanco01/virt-tui/src/tui"

)

func main() {
    app := tui.CreateApp()

    if err := app.EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}
