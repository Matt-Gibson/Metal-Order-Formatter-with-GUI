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

func processPanelList(inputText string) *widget.RichText {
	lines := strings.Split(inputText, "\n")
	panelMap := make(map[int]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			return widget.NewRichTextWithText(fmt.Sprintf("⚠️ Invalid format in line: %s\nUse: quantity @ length", line))
		}

		quantityStr := strings.TrimSpace(parts[0])
		lengthStr := strings.TrimSpace(parts[1])

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			return widget.NewRichTextWithText(fmt.Sprintf("⚠️ Invalid quantity '%s'", quantityStr))
		}

		lengthInInches, err := parseLengthInput(lengthStr)
		if err != nil {
			return widget.NewRichTextWithText(fmt.Sprintf("⚠️ Invalid length '%s': %v", lengthStr, err))
		}

		panelMap[lengthInInches] += quantity
	}

	if len(panelMap) == 0 {
		return widget.NewRichTextWithText("No valid panels entered.")
	}

	var panels []Panel
	for length, qty := range panelMap {
		panels = append(panels, Panel{LengthInInches: length, Quantity: qty})
	}

	sort.Slice(panels, func(i, j int) bool {
		return panels[i].LengthInInches > panels[j].LengthInInches
	})

	segments := []widget.RichTextSegment{
		&widget.TextSegment{Text: "🧾 Sorted Panel List (Longest to Shortest):\n\n", Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}}},
	}

	totalInches := 0
	for _, panel := range panels {
		line := fmt.Sprintf("%d @ %s\n", panel.Quantity, formatInchesToFeetAndInches(panel.LengthInInches))
		segments = append(segments, &widget.TextSegment{
			Text:  line,
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
		})
		totalInches += panel.LengthInInches * panel.Quantity
	}

	totalFeet := totalInches / 12
	remainderInches := totalInches % 12
	totalLine := fmt.Sprintf("\n📐 Total Order Length: %d' %d\" (%d inches)", totalFeet, remainderInches, totalInches)

	segments = append(segments, &widget.TextSegment{
		Text:  totalLine,
		Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true, Italic: true}},
	})

	return widget.NewRichText(segments...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Metal Roofing Panel Calculator")

	// Input area
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Example:\n5 @ 12'6\"\n2 @ 150\"\n3 @ 10'\n")
	input.TextStyle = fyne.TextStyle{Monospace: true}

	// Output area
	output := widget.NewRichTextWithText("Results will appear here...")
	output.Wrapping = fyne.TextWrapWord
	output.Scroll = container.ScrollVerticalOnly

	// Process Order button
	processButton := widget.NewButtonWithIcon("Process Order", theme.ConfirmIcon(), func() {
		output.Segments = processPanelList(input.Text).Segments
		output.Refresh()
	})
	processButton.Importance = widget.HighImportance

	// Copy to Clipboard button (only panel lines)
	copyButton := widget.NewButtonWithIcon("Copy to Clipboard", theme.ContentCopyIcon(), func() {
		var builder strings.Builder
		for i, seg := range output.Segments {
			if t, ok := seg.(*widget.TextSegment); ok {
				// Skip the first (header) and last (total) segments
				if i == 0 || i == len(output.Segments)-1 {
					continue
				}
				builder.WriteString(t.Text)
			}
		}
		myWindow.Clipboard().SetContent(builder.String())
	})

	// Buttons side by side
	buttonRow := container.NewHBox(processButton, copyButton)

	// Top area (input)
	topArea := container.NewBorder(
		widget.NewLabelWithStyle("Enter Panel List:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil, input,
	)

	// Bottom area (output)
	bottomArea := container.NewBorder(
		widget.NewLabelWithStyle("Results:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil, output,
	)

	// Split for resizable top/bottom
	split := container.NewVSplit(topArea, bottomArea)
	split.SetOffset(0.45) // 45% input, 55% output

	// Content with buttons
	content := container.NewBorder(nil, buttonRow, nil, nil, split)

	myWindow.SetContent(container.NewPadded(content))
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
