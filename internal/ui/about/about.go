package about

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/ui/icons"
	"github.com/samber/lo"

	"image/color"
)

func NewAboutTab() *container.TabItem {
	padded := func(obj fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewPaddedLayout(), obj)
	}

	en, cn := config.Name()
	title := canvas.NewText(fmt.Sprintf("%s (%s)", en, cn), color.NRGBA{R: 0, G: 120, B: 215, A: 255})
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}

	aboutImage := canvas.NewImageFromResource(icons.LogoTransparent)
	aboutImage.FillMode = canvas.ImageFillContain
	aboutImage.SetMinSize(fyne.NewSize(24, 24))

	header := container.NewHBox(
		container.NewVBox(
			layout.NewSpacer(),
			title,
			layout.NewSpacer(),
		),
		container.NewVBox(
			layout.NewSpacer(),
			aboutImage,
			layout.NewSpacer(),
		),
	)

	u, _ := url.Parse("https://github.com/owu/share-sniffer/issues")
	link := widget.NewHyperlinkWithStyle("如果您有任何问题，请在 GitHub 上提交议题", u, fyne.TextAlignTrailing, fyne.TextStyle{Underline: false})

	version := widget.NewLabelWithStyle(
		"版本信息："+config.Version(),
		fyne.TextAlignLeading, fyne.TextStyle{Italic: true},
	)

	headerSpacer := container.NewVBox(
		padded(header),

		version,
		widget.NewSeparator(),
	)

	note := widget.NewLabelWithStyle(
		"功能说明：可批量检测多种网盘分享链接是否过期，同时也提供了cli工具供服务端调用",
		fyne.TextAlignLeading, fyne.TextStyle{Italic: true},
	)

	featuresTitle := widget.NewLabelWithStyle("支持列表：", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	quarkImage := canvas.NewImageFromResource(icons.LogoQuark)
	quarkImage.FillMode = canvas.ImageFillContain
	quarkImage.SetMinSize(fyne.NewSize(40, 40))

	baiduImage := canvas.NewImageFromResource(icons.LogoBaidu)
	baiduImage.FillMode = canvas.ImageFillContain
	baiduImage.SetMinSize(fyne.NewSize(40, 40))

	thunderImage := canvas.NewImageFromResource(icons.LogoThunder)
	thunderImage.FillMode = canvas.ImageFillContain
	thunderImage.SetMinSize(fyne.NewSize(40, 40))

	ecloudImage := canvas.NewImageFromResource(icons.LogoEcloud)
	ecloudImage.FillMode = canvas.ImageFillContain
	ecloudImage.SetMinSize(fyne.NewSize(40, 40))

	alipanImage := canvas.NewImageFromResource(icons.LogoAliPan)
	alipanImage.FillMode = canvas.ImageFillContain
	alipanImage.SetMinSize(fyne.NewSize(40, 40))

	yywImage := canvas.NewImageFromResource(icons.LogoYyw)
	yywImage.FillMode = canvas.ImageFillContain
	yywImage.SetMinSize(fyne.NewSize(40, 40))

	yesImage := canvas.NewImageFromResource(icons.LogoYes)
	yesImage.FillMode = canvas.ImageFillContain
	yesImage.SetMinSize(fyne.NewSize(40, 40))

	ucImage := canvas.NewImageFromResource(icons.LogoUc)
	ucImage.FillMode = canvas.ImageFillContain
	ucImage.SetMinSize(fyne.NewSize(40, 40))

	ydImage := canvas.NewImageFromResource(icons.LogoYd)
	ydImage.FillMode = canvas.ImageFillContain
	ydImage.SetMinSize(fyne.NewSize(40, 40))

	features := []struct {
		icon  fyne.Resource
		title string
		desc  string
	}{
		{quarkImage.Resource, "夸克网盘", fmt.Sprintf("%s*", config.GetSupportedQuark())},
		{ecloudImage.Resource, "天翼云盘", fmt.Sprintf("%s*", config.GetSupportedTelecom())},
		{baiduImage.Resource, "百度网盘", fmt.Sprintf("%s*", config.GetSupportedBaidu())},
		{alipanImage.Resource, "阿里云盘", fmt.Sprintf("%s*", config.GetSupportedAliPan())},
		{yywImage.Resource, "115网盘", fmt.Sprintf("%s*", config.GetSupportedYyw())},
		{yesImage.Resource, "123网盘", fmt.Sprintf("%s*", config.GetSupportedYes())},
		{ucImage.Resource, "UC网盘", fmt.Sprintf("%s*", config.GetSupportedUc())},
		{thunderImage.Resource, "迅雷云盘", "TODO"},
		{ydImage.Resource, "移动云盘", "TODO"},
		{theme.QuestionIcon(), "TODO", "TODO"},
	}

	// 将features切片按每2个元素一组进行分组
	groupedFeatures := lo.Chunk(features, 2)

	featureItems := container.NewVBox()

	// 遍历分组后的features，每行显示两个feature，确保左对齐并兼容奇数/偶数个元素
	for _, pair := range groupedFeatures {
		// 使用GridLayoutWithColumns确保两个元素平分空间
		rowContainer := container.New(layout.NewGridLayoutWithColumns(2))

		// 添加第一个feature
		if len(pair) > 0 {
			rowContainer.Add(createFeatureItem(pair[0]))
		}

		// 如果有第二个feature，添加第二个feature
		if len(pair) > 1 {
			rowContainer.Add(createFeatureItem(pair[1]))
		}

		// 添加到featureItems容器
		featureItems.Add(padded(rowContainer))
		featureItems.Add(widget.NewSeparator())
	}

	content := container.NewVBox(
		padded(headerSpacer),
		padded(note),
		padded(widget.NewSeparator()),
		padded(featuresTitle),
		padded(widget.NewSeparator()),
		padded(featureItems),
		padded(link),
	)

	return container.NewTabItemWithIcon(
		"关于",
		theme.InfoIcon(),
		container.NewVScroll(
			container.New(layout.NewPaddedLayout(), content),
		),
	)
}

// 辅助函数：创建单个feature的UI元素
func createFeatureItem(f struct {
	icon  fyne.Resource
	title string
	desc  string
}) *fyne.Container {
	// 创建图标
	iconWidget := widget.NewIcon(f.icon)

	// 创建标题标签
	titleLabel := widget.NewLabelWithStyle(f.title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 创建描述标签
	descLabel := widget.NewLabelWithStyle(f.desc, fyne.TextAlignLeading, fyne.TextStyle{})

	// 创建紧凑的水平布局
	return container.NewHBox(
		// 添加一个小的空白间距，保持图标与文字的适当距离
		container.NewPadded(iconWidget),
		// 使用垂直压缩布局
		container.NewVBox(
			// 移除标题上方的空间
			container.NewWithoutLayout(titleLabel),
			// 移除描述上方的空间
			container.NewWithoutLayout(descLabel),
		),
	)
}
