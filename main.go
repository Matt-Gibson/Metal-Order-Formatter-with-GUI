package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Panel struct {
	LengthInInches int
	Quantity       int
}

func formatInchesToFeetAndInches(totalInches int) string {
	feet := totalInches / 12
	inches := totalInches % 12
	return fmt.Sprintf("%d' %d\"", feet, inches)
}

func parseLengthInput(input string) (int, error) {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	var feet, inches int
	var err error

	if strings.Contains(input, "'") {
		parts := strings.Split(input, "'")
		feetStr := strings.TrimSpace(parts[0])
		if feetStr != "" {
			feet, err = strconv.Atoi(feetStr)
			if err != nil {
				return 0, fmt.Errorf("invalid feet: %v", err)
			}
		}
		if len(parts) > 1 {
			inchPart := strings.TrimSpace(parts[1])
			inchPart = strings.TrimSuffix(inchPart, "\"")
			if inchPart != "" {
				inches, err = strconv.Atoi(inchPart)
				if err != nil {
					return 0, fmt.Errorf("invalid inches: %v", err)
				}
			}
		}
	} else if strings.Contains(input, "\"") {
		inchStr := strings.TrimSuffix(input, "\"")
		inches, err = strconv.Atoi(strings.TrimSpace(inchStr))
		if err != nil {
			return 0, fmt.Errorf("invalid inches: %v", err)
		}
	} else {
		inches, err = strconv.Atoi(input)
		if err != nil {
			return 0, fmt.Errorf("invalid numeric input: %v", err)
		}
	}

	return feet*12 + inches, nil
}

func processPanelList(inputText string) string {
	lines := strings.Split(inputText, "\n")
	panelMap := make(map[int]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			return fmt.Sprintf("⚠️ Invalid format in line: %s\nUse: quantity @ length", line)
		}

		quantityStr := strings.TrimSpace(parts[0])
		lengthStr := strings.TrimSpace(parts[1])

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			return fmt.Sprintf("⚠️ Invalid quantity '%s'", quantityStr)
		}

		lengthInInches, err := parseLengthInput(lengthStr)
		if err != nil {
			return fmt.Sprintf("⚠️ Invalid length '%s': %v", lengthStr, err)
		}

		panelMap[lengthInInches] += quantity
	}

	if len(panelMap) == 0 {
		return "No valid panels entered."
	}

	var panels []Panel
	for length, qty := range panelMap {
		panels = append(panels, Panel{LengthInInches: length, Quantity: qty})
	}

	sort.Slice(panels, func(i, j int) bool {
		return panels[i].LengthInInches > panels[j].LengthInInches
	})

	result := "🧾 Sorted Panel List (Longest to Shortest):\n"
	totalInches := 0
	for _, panel := range panels {
		lengthStr := formatInchesToFeetAndInches(panel.LengthInInches)
		result += fmt.Sprintf("%d @ %s\n", panel.Quantity, lengthStr)
		totalInches += panel.LengthInInches * panel.Quantity
	}

	totalFeet := totalInches / 12
	remainderInches := totalInches % 12
	result += fmt.Sprintf("\n📐 Total Order Length: %d' %d\" (%d inches)\n", totalFeet, remainderInches, totalInches)

	return result
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Metal Roofing Panel Calculator")

	// Input area
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Example:\n5 @ 12'6\"\n2 @ 150\"\n3 @ 10'\n")
	input.TextStyle = fyne.TextStyle{Monospace: true}

	// Output area
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Results will appear here...")
	output.Disable()
	output.TextStyle = fyne.TextStyle{Monospace: true}

	// Full-width process button with padding
	processButton := widget.NewButtonWithIcon("Process Order", theme.ConfirmIcon(), func() {
		output.SetText(processPanelList(input.Text))
	})
	processButton.Importance = widget.HighImportance
	buttonArea := container.NewVBox(
		widget.NewLabel(""), // small spacer
		processButton,
	)

	// Split view for proportional height
	topArea := container.NewVBox(
		widget.NewLabelWithStyle("Enter Panel List:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		input,
	)
	bottomArea := container.NewVBox(
		widget.NewLabelWithStyle("Results:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		output,
	)
	split := container.NewVSplit(topArea, bottomArea)
	split.SetOffset(0.45) // ~45% top, 55% bottom

	// Final layout with padded main content
	mainContent := container.NewBorder(nil, buttonArea, nil, nil, split)

	myWindow.SetContent(container.NewPadded(mainContent))
	myWindow.Resize(fyne.NewSize(750, 600))
	myWindow.ShowAndRun()
}
