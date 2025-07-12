package makitheme

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Theme struct{}

var (
	//go:embed MakiLibertinusMono-Regular.ttf
	fontRegularData []byte
	fontRegular     = fyne.NewStaticResource("MakiLibertinusMono-Regular.ttf", fontRegularData)

	//go:embed MakiLibertinusMono-RegularItalic.ttf
	fontRegularItalicData []byte
	fontRegularItalic     = fyne.NewStaticResource("MakiLibertinusMono-RegularItalic.ttf", fontRegularItalicData)

	//go:embed MakiLibertinusMono-Bold.ttf
	fontBoldData []byte
	fontBold     = fyne.NewStaticResource("MakiLibertinusMono-Bold.ttf", fontBoldData)

	//go:embed MakiLibertinusMono-BoldItalic.ttf
	fontBoldItalicData []byte
	fontBoldItalic     = fyne.NewStaticResource("MakiLibertinusMono-BoldItalic.ttf", fontBoldItalicData)
)

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	// our font is monospace
	// if style.Monospace {
	// 	return theme.DefaultTextMonospaceFont()
	// }
	if style.Bold {
		if style.Italic {
			return fontBoldItalic
		}
		return fontBold
	}
	if style.Italic {
		return fontRegularItalic
	}
	if style.Symbol {
		return theme.DefaultSymbolFont()
	}
	return fontRegular
}

func (t *Theme) Icon(icon fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(icon)
}

func (t *Theme) Size(size fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(size)
}
