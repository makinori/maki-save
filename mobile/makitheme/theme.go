package makitheme

import (
	_ "embed"
	"image/color"
	"sync"

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

	//go:embed NotoColorEmoji.ttf
	fontNotoColorEmojiData []byte
)

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DarkTheme().Color(name, variant)
}

var updateEmojiFontOnce sync.Once

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	updateEmojiFontOnce.Do(func() {
		// default emoji font is old and outdated. replace with custom font.
		// unfortunately can't use `-tags no_emoji`` or below will return nil.
		font, ok := theme.DefaultEmojiFont().(*fyne.StaticResource)
		if ok {
			font.StaticName = "NotoColorEmoji.ttf"
			font.StaticContent = fontNotoColorEmojiData
		}
	})

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
	return theme.DarkTheme().Icon(icon)
}

func (t *Theme) Size(size fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(size)
}
