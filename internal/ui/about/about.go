package about

import (
	"fmt"
	"net/url"

	"share-sniffer/internal/config"
	"share-sniffer/internal/ui/icons"
	"share-sniffer/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"

	"image/color"
)

func NewAboutTab(window fyne.Window) *container.TabItem {
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

	u, _ := url.Parse(fmt.Sprintf("%s/issues", config.HomePage()))
	link := widget.NewHyperlinkWithStyle("如果您有任何问题，请在项目主页上提交议题", u, fyne.TextAlignTrailing, fyne.TextStyle{Underline: false})

	// 创建可点击的版本号标签
	versionText := "版本信息：v" + config.Version()
	versionLink := widget.NewHyperlinkWithStyle(versionText, nil, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	// 设置点击事件处理函数
	versionLink.OnTapped = func() {
		// 在协程中执行版本检查，避免阻塞UI
		go func() {
			CheckUpdate(window, true)
		}()
	}

	headerSpacer := container.NewVBox(
		padded(header),

		versionLink,
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

	chromeImage := canvas.NewImageFromResource(icons.LogoChrome)
	chromeImage.FillMode = canvas.ImageFillContain
	chromeImage.SetMinSize(fyne.NewSize(40, 40))

	features := []struct {
		icon  fyne.Resource
		title string
		desc  string
		ico   fyne.Resource
	}{
		{quarkImage.Resource, "夸克网盘", fmt.Sprintf("%s*", config.GetSupportedQuark()), nil},
		{ecloudImage.Resource, "天翼云盘", fmt.Sprintf("%s*", config.GetSupportedTelecom()), nil},
		{baiduImage.Resource, "百度网盘", fmt.Sprintf("%s*", config.GetSupportedBaidu()), nil},
		{alipanImage.Resource, "阿里云盘", fmt.Sprintf("%s*", config.GetSupportedAliPan()), nil},
		{yywImage.Resource, "115网盘", fmt.Sprintf("%s*", config.GetSupportedYyw()), nil},
		{yesImage.Resource, "123云盘", fmt.Sprintf("%s*", config.GetSupportedYes()), nil},
		{ucImage.Resource, "UC网盘", fmt.Sprintf("%s*", config.GetSupportedUc()), nil},
		//{theme.QuestionIcon(), "TODO", "TODO", nil},
	}

	// 只有在非安卓平台才添加的功能
	if utils.IsDesktop() {
		features = append(features, struct {
			icon  fyne.Resource
			title string
			desc  string
			ico   fyne.Resource
		}{thunderImage.Resource, "迅雷云盘", fmt.Sprintf("%s*", config.GetSupportedXunlei()), icons.LogoChrome})
		features = append(features, struct {
			icon  fyne.Resource
			title string
			desc  string
			ico   fyne.Resource
		}{ydImage.Resource, "移动云盘", fmt.Sprintf("%s*", config.GetSupportedYd()), icons.LogoChrome})
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
	ico   fyne.Resource
}) *fyne.Container {
	// 创建图标
	iconWidget := widget.NewIcon(f.icon)

	// 创建标题标签
	titleLabel := widget.NewLabelWithStyle(f.title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 创建描述标签
	descLabel := widget.NewLabelWithStyle(f.desc, fyne.TextAlignLeading, fyne.TextStyle{})

	// 创建标题行，包含标题和Chrome图标（如果有）
	var titleRow fyne.CanvasObject
	if f.ico != nil {
		icoWidget := widget.NewIcon(f.ico)
		// 设置图标更小的尺寸
		iconSize := fyne.NewSize(18, 18) // 图标尺寸
		icoWidget.Resize(iconSize)

		// 使用自定义HBox布局，设置负的水平间距以实现更紧密的贴合
		titleRow = container.New(
			&customHBoxLayout{spacing: -2}, // 设置负间距，让图标与文字更加紧密贴合
			titleLabel,
			icoWidget,
		)
	} else {
		titleRow = titleLabel
	}

	// 创建右侧内容区域，使用自定义布局来控制垂直间距
	// 直接使用HBox来布局整个项，然后手动排列标题和描述
	contentArea := container.New(
		// 使用自定义布局函数来精确控制间距
		&customVBoxLayout{spacing: 1}, // 设置最小的垂直间距
		titleRow,
		descLabel,
	)

	// 创建整体布局，保持适当的水平间距
	return container.NewHBox(
		iconWidget,
		contentArea,
	)
}

// 自定义VBox布局，允许设置更小的垂直间距
type customVBoxLayout struct {
	spacing float32
}

func (l *customVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var minSize fyne.Size
	for _, obj := range objects {
		objSize := obj.MinSize()
		if objSize.Width > minSize.Width {
			minSize.Width = objSize.Width
		}
		minSize.Height += objSize.Height
	}
	minSize.Height += l.spacing * float32(len(objects)-1)
	return minSize
}

func (l *customVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, obj := range objects {
		objSize := obj.MinSize()
		obj.Resize(fyne.NewSize(size.Width, objSize.Height))
		obj.Move(fyne.NewPos(0, y))
		y += objSize.Height + l.spacing
	}
}

// 自定义HBox布局，允许设置更小的水平间距
type customHBoxLayout struct {
	spacing float32
}

func (l *customHBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var minSize fyne.Size
	for _, obj := range objects {
		objSize := obj.MinSize()
		if objSize.Height > minSize.Height {
			minSize.Height = objSize.Height
		}
		minSize.Width += objSize.Width
	}
	minSize.Width += l.spacing * float32(len(objects)-1)
	return minSize
}

func (l *customHBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for _, obj := range objects {
		objSize := obj.MinSize()
		// 精确控制尺寸，移除任何可能的额外空间
		obj.Resize(fyne.NewSize(objSize.Width, size.Height))
		obj.Move(fyne.NewPos(x, 0))
		x += objSize.Width + l.spacing
	}
}
