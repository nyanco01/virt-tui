package tui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var validInput = [...]rune{'!', 'p', 'l', 'a', 'y', 's', 'l', 'o', 't'}

const laneTextWidth = 6
const laneTextHeight = 5

type laneNumber int

const (
	Zero laneNumber = iota
	One
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
)

func (n laneNumber) text() string {
	switch n {
	case Zero:
		return `
 @@@@
@@  @@
@@  @@
@@  @@
 @@@@`
	case One:
		return `
@@@@
  @@
  @@
  @@
@@@@@@`
	case Two:
		return `
 @@@@
@@  @@
   @@
  @@
@@@@@@`
	case Three:
		return `
 @@@@
@@  @@
   @@@
@@  @@
 @@@@`
	case Four:
		return `
@@  @@
@@  @@
@@@@@@
    @@
    @@`
	case Five:
		return `
@@@@@@
@@
@@@@@
    @@
@@@@@`
	case Six:
		return `
 @@@@
@@
@@@@@
@@  @@
 @@@@`
	case Seven:
		return `
@@@@@@
   @@
  @@
 @@
@@`
	case Eight:
		return `
 @@@@
@@  @@
 @@@@
@@  @@
 @@@@`
	case Nine:
		return `
 @@@@
@@  @@
 @@@@@
    @@
 @@@@`
	default:
		panic("unreachable")
	}
}

type tailedInput struct {
	q []rune // behave as a queue
}

func newTailedInput() tailedInput {
	return tailedInput{make([]rune, 0, len(validInput))}
}

func (t *tailedInput) add(r rune) {
	t.q = append(t.q, r)

	if len(t.q) == (len(validInput) + 1) {
		t.q = t.q[1:]
	}
}

func (t *tailedInput) clear() {
	t.q = make([]rune, 0, len(validInput))
}

func (t *tailedInput) hasCastDone() bool {
	if len(t.q) != len(validInput) {
		return false
	}

	for i, v := range validInput {
		if (t.q[i]) != v {
			return false
		}
	}

	return true
}

type lane struct {
	number  laneNumber
	pressed bool
}

func newLane() lane {
	return lane{Zero, false}
}

func (l *lane) press() {
	l.pressed = true
}

func (l *lane) nextIfNotPressed() {
	if !l.pressed {
		l.number = laneNumber(rand.Int() % 10)
	}
}

type slotResult int

const (
	Success slotResult = iota
	Failure
	Unset
)

func MakeSlotUi(app *tview.Application) *tview.TextView {
	text := tview.NewTextView()

	second := newLane()
	third := newLane()
	fourth := newLane()

	result := Unset

	text.SetDynamicColors(true)
	text.SetTextAlign(tview.AlignCenter)
	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'g' {
			if !second.pressed {
				second.press()
			} else if !third.pressed {
				third.press()
			} else if !fourth.pressed {
				fourth.press()

				if second.number == Zero && third.number == Zero && fourth.number == One {
					result = Success
				} else {
					result = Failure
				}
			} else {
				second = newLane()
				third = newLane()
				fourth = newLane()
			}
		}

		return event
	})

	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			second.nextIfNotPressed()
			third.nextIfNotPressed()
			fourth.nextIfNotPressed()

			const header = "\n\n\n-= < LOCALHOST SLOT > =-\n\n\n\n"
			const footer = "\n\n\nPRESS or RESET: < G >\n\nEXIT: < E >"

			// allow for delays
			resultText := "\n\n\n"
			switch result {
			case Success:
				resultText += "[green]CONGRATULATIONS!"
			case Failure:
				resultText += "[red]FAILURE..."
			default:
				resultText = ""
			}

			t := header + makeSlotText(second.number, third.number, fourth.number) + footer + resultText

			app.QueueUpdateDraw(func() {
				text.SetText(t)
			})
		}
	}()

	return text
}

func makeSlotText(second laneNumber, third laneNumber, fourth laneNumber) string {
	sr := laneToAscii(second)
	tr := laneToAscii(third)
	fr := laneToAscii(fourth)

	const firstWidth = 15

	const gap = 12
	const padding = laneTextWidth + gap
	var r [laneTextHeight][2*firstWidth + 3*padding - gap]rune
	for y := 0; y < len(r); y++ {
		for x := 0; x < len(r[0]); x++ {
			r[y][x] = ' '
		}
	}

	r[laneTextHeight-1][0] = '1'
	r[laneTextHeight-1][1] = '9'
	r[laneTextHeight-1][2] = '2'
	r[laneTextHeight-1][3] = '.'

	for y := 0; y < laneTextHeight; y++ {
		for x := 0; x < laneTextWidth; x++ {
			r[y][firstWidth+x] = sr[y][x]
			r[y][firstWidth+x+padding] = tr[y][x]
			r[y][firstWidth+x+2*padding] = fr[y][x]
		}
	}

	var b []byte
	for y := 0; y < len(r); y++ {
		for x := 0; x < len(r[0]); x++ {
			b = append(b, byte(r[y][x]))
		}
		b = append(b, '\n')
	}

	return string(b)
}

func laneToAscii(n laneNumber) [laneTextHeight][laneTextWidth]rune {
	var r [laneTextHeight][laneTextWidth]rune

	splitted := strings.Split(n.text(), "\n")

	for i, s := range splitted {
		if i == 0 {
			continue // skip blank line
		}

		var cp [laneTextWidth]rune
		copy(cp[:], []rune(s)[:])

		for i := range cp {
			if cp[i] == 0 {
				cp[i] = ' '
			}
		}

		r[i-1] = cp
	}

	return r
}
