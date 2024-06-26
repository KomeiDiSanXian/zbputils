package control

import (
	"image"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/disintegration/imaging"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/rendercard"

	"github.com/FloatTech/zbputils/img/text"
)

const (
	bannerpath = "data/zbpbanner/"
	kanbanpath = "data/Control/"
	bannerurl  = "https://gitcode.net/u011570312/zbpbanner/-/raw/main/"
)

type card struct {
	batchsize  int
	n          int
	pluginlist []plugininfo
	cardlist   []image.Image
	drawcard   func([]plugininfo, []image.Image)
	wg         *sync.WaitGroup
}

type plugininfo struct {
	name   string
	brief  string
	banner string
	status bool
}

var (
	// 标题缓存
	titlecache image.Image
	// 卡片缓存
	cardcache = ttl.NewCache[uint64, image.Image](time.Hour * 12)
	// 阴影缓存
	fullpageshadowcache image.Image
	// lnperpg 每页行数
	lnperpg = 9
	// 字体 GlowSans 数据
	glowsd []byte
	// 字体 Sukura 数据
	sakurad []byte
)

func init() {
	err := os.MkdirAll(bannerpath, 0755)
	if err != nil {
		panic(err)
	}
	_, err = file.GetLazyData(kanbanpath+"kanban.png", Md5File, false)
	if err != nil {
		panic(err)
	}
	glowsd, err = file.GetLazyData(text.GlowSansFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
	// 原来的字体缺失中文，更换为这个
	sakurad, err = file.GetLazyData(text.SakuraFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
}

func drawservicesof(gid int64) (imgs []image.Image, err error) {
	pluginlist := make([]plugininfo, len(priomap))
	ForEachByPrio(func(i int, manager *ctrl.Control[*zero.Ctx]) bool {
		pluginlist[i] = plugininfo{
			name:   manager.Service,
			brief:  manager.Options.Brief,
			banner: manager.Options.Banner,
			status: manager.IsEnabledIn(gid),
		}
		return true
	})
	// 分页
	if len(pluginlist) < lnperpg*3 {
		// 如果单页显示数量超出了总数量
		lnperpg = math.Ceil(len(pluginlist), 3)
	}
	if titlecache == nil {
		titlecache, err = (&rendercard.Title{
			Line:          lnperpg,
			LeftTitle:     "服务列表",
			LeftSubtitle:  "service_list",
			RightTitle:    "FloatTech",
			RightSubtitle: "RemiliaBot",
			TitleFontData: glowsd,
			TextFontData:  sakurad,
			ImagePath:     kanbanpath + "kanban.png",
		}).DrawTitle()
		if err != nil {
			return
		}
	}

	cardlist := make([]image.Image, len(pluginlist))
	n := runtime.NumCPU()

	if len(pluginlist) <= n {
		err = drawSingleCard(pluginlist, cardlist, false)
	} else {
		wg := sync.WaitGroup{}
		card{
			batchsize:  len(pluginlist) / n,
			n:          n,
			pluginlist: pluginlist,
			cardlist:   cardlist,
			drawcard: func(info []plugininfo, cards []image.Image) {
				defer wg.Done()
				err = drawSingleCard(info, cards, true)
			},
			wg: &wg,
		}.drawCards()
	}

	wg := sync.WaitGroup{}
	cardnum := lnperpg * 3
	page := math.Ceil(len(pluginlist), cardnum)
	imgs = make([]image.Image, page)
	x, y := 30+2, 30+300+30+6+4
	if fullpageshadowcache == nil {
		fullpageshadowcache = drawShadow(x, y, cardnum)
	}
	wg.Add(page)
	for l := 0; l < page; l++ { // 页数
		go func(l int, islast bool) {
			defer wg.Done()
			x, y := 30+2, 30+300+30+6+4
			var shadowimg image.Image
			if islast && len(pluginlist)-cardnum*l < cardnum {
				shadowimg = drawShadow(x, y, len(pluginlist)-cardnum*l)
			} else {
				shadowimg = fullpageshadowcache
			}
			onePageCardNum := cardnum * l
			cardnums := math.Min(cardnum, len(pluginlist)-onePageCardNum)
			imgs[l] = drawShadowIntoOnePage(shadowimg, cardnums, onePageCardNum, cardlist)
		}(l, l == page-1)
	}
	wg.Wait()
	return
}

func drawShadowIntoOnePage(shadowimg image.Image, cardnums, cardInList int, cardlist []image.Image) image.Image {
	one := gg.NewContextForImage(titlecache)
	x, y := 30, 30+300+30
	one.DrawImage(shadowimg, 0, 0)
	for i := 0; i < cardnums; i++ {
		one.DrawImage(cardlist[cardInList+i], x, y)
		x += 384 + 30
		if (i+1)%3 == 0 {
			x = 30
			y += 256 + 30
		}
	}
	return one.Image()
}

func drawShadow(x, y, cardnum int) image.Image {
	shadow := gg.NewContextForImage(rendercard.Transparency(titlecache, 0))
	shadow.SetRGBA255(0, 0, 0, 192)
	for i := 0; i < cardnum; i++ {
		shadow.DrawRoundedRectangle(float64(x), float64(y), 384-4, 256-4, 0)
		shadow.Fill()
		x += 384 + 30
		if (i+1)%3 == 0 {
			x = 30 + 2
			y += 256 + 30
		}
	}
	return imaging.Blur(shadow.Image(), 7)
}

/*
func downloadCard(info plugininfo) (banner string, err error) {
	switch {
	case strings.HasPrefix(info.banner, "http"):
		err = file.DownloadTo(info.banner, bannerpath+info.name+".png")
		if err != nil {
			return
		}
		banner = bannerpath + info.name + ".png"
	case info.banner != "":
		banner = info.banner
	default:
		_, err = file.GetCustomLazyData(bannerurl, bannerpath+info.name+".png")
		if err == nil {
			banner = bannerpath + info.name + ".png"
		}
	}
	return
}
*/

func drawSingleCard(pluginInfo []plugininfo, cardlist []image.Image, roundedRectangle bool) error {
	for k, info := range pluginInfo {
		// 不下载可以更快
		// banner, _ := downloadCard(info)
		c := &rendercard.Title{
			IsEnabled:     info.status,
			LeftTitle:     info.name,
			LeftSubtitle:  info.brief,
			ImagePath:     "", // banner
			TitleFontData: sakurad,
			TextFontData:  glowsd,
		}
		h := c.Sum64()
		card := cardcache.Get(h)
		if card == nil {
			var err error
			card, err = c.DrawCard()
			if err != nil {
				return err
			}
			if roundedRectangle {
				card = rendercard.Fillet(card, 8)
			}
			cardcache.Set(h, card)
		}
		cardlist[k] = card
	}
	return nil
}

func (c card) drawCards() {
	c.wg.Add(c.n)
	for i := 0; i < c.n; i++ {
		a := i * c.batchsize
		b := (i + 1) * c.batchsize
		go c.drawcard(c.pluginlist[a:b], c.cardlist[a:b])
	}
	if c.batchsize*c.n < len(c.pluginlist) {
		c.wg.Add(1)
		d := len(c.pluginlist) - c.batchsize*c.n
		go c.drawcard(c.pluginlist[d:], c.cardlist[d:])
	}
	c.wg.Wait()
}
