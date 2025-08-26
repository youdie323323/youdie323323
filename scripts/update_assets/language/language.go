package language

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"time"
)

var (
	ConfigDir   = flag.String("config-dir", "", "Directory storing config files")
	OutDir      = flag.String("out-dir", "", "Destination of the artifacts")
	TemplateDir = flag.String("template-dir", "", "Directory storing SVG Templates")
)

// ratio is ratio of Courier-new, whose aspect ratio is known.
const ratio = 0.607

func calculateFontWidth(size float64) float64 {
	return ratio * size
}

const ( // Font sizes for larger (title) and smaller (others) texts.
	largeFont       = 24.
	smallFontHeight = 16.
)

var (
	largeFontWidth = calculateFontWidth(largeFont)
	smallFontWidth = calculateFontWidth(smallFontHeight)
)

var (
	largeFontMergin = 1.5 * largeFontWidth
	smallFontMergin = 1.5 * smallFontWidth
)

func convertBytesToSizeFormat(bytes float64) string {
	unit := "B"

	const thres = 1024.

	if bytes > thres {
		bytes /= thres

		unit = "kB"
	}

	if bytes > thres {
		bytes /= thres

		unit = "MB"
	}

	return fmt.Sprintf("%.1f%s", bytes, unit)
}

// set position and text of the title section
func constructTitleComponent(domain *Domain, infos []Information) map[string]any {
	// Move downward a bit, above text
	domain.Height += 1.5 * largeFont

	titleComponent := make(map[string]any)

	titleComponent["x"] = largeFontMergin
	titleComponent["y"] = domain.Height

	{
		var totalBytes float64 = 0

		for _, info := range infos {
			totalBytes += float64(info.Size)
		}

		titleComponent["text"] = fmt.Sprintf("Language Stats (%s)", convertBytesToSizeFormat(totalBytes))
	}

	// Move downward a bit, below text
	domain.Height += 1.5 * largeFont

	return titleComponent
}

func constructLanguageStatisticsComponent(domain *Domain, infos []Information) []map[string]map[string]any {
	// height for each language
	heightOfLine := 2. * smallFontHeight

	// Decide width of the first part by checking the length of the longest name
	// at the same time get the max and sum of the whole values
	labelWidth := 0.

	var vMax float64 = 0
	var vSum float64 = 0

	for _, info := range infos {
		l := smallFontWidth * float64(len(info.Name))

		if labelWidth < l {
			labelWidth = l
		}

		vMax = max(vMax, float64(info.Size))
		vSum += float64(info.Size)
	}

	// Start iterating for each language
	var languageStatisticsComponent []map[string]map[string]any

	for i, info := range infos {
		// Origin
		x := largeFontMergin
		y := domain.Height + float64(i)*heightOfLine

		// Height of rectangles, both label and graph
		rectHeight := 0.7 * heightOfLine

		// A label (language name) and a surrounding rectangle
		labelRectComponent, labelTextComponent := func() (map[string]any, map[string]any) {
			// Rectangle around the label
			rectComponent := make(map[string]any)

			// Top-left corner
			rectComponent["x"] = x
			rectComponent["y"] = y - 0.5*rectHeight

			// Size of the rect, give 0.5 text-width margins for each edge
			rectComponent["width"] = labelWidth + smallFontWidth
			rectComponent["height"] = rectHeight

			rectComponent["round"] = 4

			rectComponent["stroke"] = info.Color

			// Label: language name
			textComponent := make(map[string]any)

			// Give 0.5 text-width margin
			textComponent["x"] = x + 0.5*smallFontWidth
			textComponent["y"] = y

			textComponent["text"] = info.Name

			// Update x edge for the coming elements
			x += labelWidth + smallFontWidth

			return rectComponent, textComponent
		}()

		usagePercentageTextComponent := func() map[string]any {
			maxPercentageTextLen := len("100.00%")

			percentageText := fmt.Sprintf("%.2f%%", 100.*float64(info.Size)/vSum)
			relativePercentageText := maxPercentageTextLen - len(percentageText)

			textComponent := make(map[string]any)

			textComponent["x"] = x + 0.5*smallFontWidth + smallFontWidth*float64(relativePercentageText)
			textComponent["y"] = y

			textComponent["text"] = percentageText

			x += float64(maxPercentageTextLen) * smallFontWidth

			return textComponent
		}()

		usagePercentageGraphComponent := func() map[string]any {
			graphWidth := domain.Width - x - 3.*smallFontWidth

			graphComponent := make(map[string]any)

			graphComponent["x"] = x + 1.*smallFontWidth
			graphComponent["y"] = y - 0.5*rectHeight

			graphComponent["width"] = graphWidth * float64(info.Size) / vMax
			graphComponent["height"] = rectHeight

			graphComponent["round"] = 4

			graphComponent["fill"] = info.Color
			graphComponent["stroke"] = info.Color

			return graphComponent
		}()

		languageStatisticsComponent = append(languageStatisticsComponent, map[string]map[string]any{
			"labelrect": labelRectComponent,
			"labeltext": labelTextComponent,
			"share":     usagePercentageTextComponent,
			"graph":     usagePercentageGraphComponent,
		})
	}

	domain.Height += heightOfLine * float64(len(infos))

	return languageStatisticsComponent
}

func constructLastUpdateComponent(domain *Domain) map[string]any {
	domain.Height += 0.5 * smallFontHeight

	now := time.Now()

	lastUpdateComponent := make(map[string]any)

	lastUpdateComponent["x"] = largeFontMergin
	lastUpdateComponent["y"] = domain.Height

	lastUpdateComponent["text"] = fmt.Sprintf("%d %s %d", now.Day(), now.Month(), now.Year())

	domain.Height += largeFont

	return lastUpdateComponent
}

const borderStrokeWidth = 2.

func constructBorderComponent(domain *Domain) map[string]float64 {
	borderComponent := make(map[string]float64)

	borderComponent["x"] = 0.5 * borderStrokeWidth
	borderComponent["y"] = 0.5 * borderStrokeWidth

	borderComponent["width"] = domain.Width - borderStrokeWidth
	borderComponent["height"] = domain.Height - borderStrokeWidth

	borderComponent["round"] = 10.

	borderComponent["stroke_width"] = borderStrokeWidth

	return borderComponent
}

func executeTemplateThenOut(input string, output string, data any) {
	inputFile, err := os.Open(input)
	if err != nil {
		panic(err)
	}

	defer inputFile.Close()

	templateContents, err := io.ReadAll(inputFile)
	if err != nil {
		panic(err)
	}

	t, err := template.New("template").Parse(string(templateContents))
	if err != nil {
		panic(err)
	}

	outputFile, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	defer outputFile.Close()

	err = t.Execute(outputFile, data)
	if err != nil {
		panic(err)
	}
}

// Domain is a general data type to store 2d sizes.
type Domain struct {
	Width  float64
	Height float64
}

func ConstructLanguageInformationSVG() {
	infos := CreateRealtimeInformations()

	// Overall svg size
	// Width is fixed, while height is subject to change
	domain := Domain{500, 0}

	// Collect information to construct svg
	titleComponent := constructTitleComponent(&domain, infos)
	languageStatisticsComponent := constructLanguageStatisticsComponent(&domain, infos)
	lastUpdateComponent := constructLastUpdateComponent(&domain)
	borderComponent := constructBorderComponent(&domain)

	result := struct {
		Domain     Domain
		Title      map[string]any
		Border     map[string]float64
		Languages  []map[string]map[string]any
		LastUpdate map[string]any
	}{
		Domain:     domain,
		Title:      titleComponent,
		Border:     borderComponent,
		Languages:  languageStatisticsComponent,
		LastUpdate: lastUpdateComponent,
	}

	executeTemplateThenOut(
		path.Join(*TemplateDir, "language.svg"),
		path.Join(*OutDir, "language.svg"),
		result,
	)
}
