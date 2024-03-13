package util

import (
	"sort"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"
)

func Pack(items []Item, boxSize Rectangle) []Box {
	log.Infof("pack items: %v", items)
	log.Infof("pack boxSize: %v", boxSize)
	sort.Sort(ByArea(items)) // 按照面积从大到小排序物品
	var boxes []Box
	for i := 0; i < 10000; i++ {
		var box Box
		var usedWidth, usedHeight float64
		packAll := packBox(items, usedWidth, usedHeight, boxSize, &box)
		if packAll {
			break
		}
		boxes = append(boxes, box)
	}
	return boxes
}

func packBox(items []Item, usedWidth, usedHeight float64, boxSize Rectangle, box *Box) bool {
	packAll := true
	for i := range items {
		if items[i].Number == 0 {
			continue
		}
		packAll = false
		for j := items[i].Number; j > 0; j-- {
			if usedWidth+items[i].Size.Width <= boxSize.Width && usedHeight+items[i].Size.Height <= boxSize.Height { // 如果当前物品能放入箱子
				items[i].Number--
				usedWidth += items[i].Size.Width
				usedHeight += items[i].Size.Height
				box.Items = append(box.Items, items[i])
			} else {
				break
			}
		}
	}
	return packAll
}

type Rectangle struct {
	Width  float64 // 矩形宽度
	Height float64 // 矩形高度
}

type Item struct {
	Size   Rectangle
	Number int
}

type Box struct {
	Items []Item
}

// 按照面积降序排序矩形
func (r ByArea) Len() int      { return len(r) }
func (r ByArea) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ByArea) Less(i, j int) bool {
	return r[i].Size.Width*r[i].Size.Height > r[j].Size.Width*r[j].Size.Height
}

type ByArea []Item
