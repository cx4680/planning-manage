package util

import (
	"sort"

	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
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

// CalcNfvServerNumber 计算NFV节点数量
func CalcNfvServerNumber(serverNumber int, masterNumber int) int {
	var nfvServerNumber int
	if serverNumber <= 200 {
		if serverNumber+masterNumber <= 200 {
			nfvServerNumber = 2
		} else if serverNumber+masterNumber > 200 && serverNumber+masterNumber <= 500 {
			nfvServerNumber = 4
		} else if serverNumber+masterNumber > 500 && serverNumber+masterNumber <= 2000 {
			nfvServerNumber = 8
		} else {
			nfvServerNumber = 16
		}
	}
	if serverNumber > 200 && serverNumber <= 500 {
		if serverNumber+masterNumber > 200 && serverNumber+masterNumber <= 500 {
			nfvServerNumber = 4
		} else if serverNumber+masterNumber > 500 && serverNumber+masterNumber <= 2000 {
			nfvServerNumber = 8
		} else {
			nfvServerNumber = 16
		}
	}
	if serverNumber > 500 && serverNumber <= 2000 {
		if serverNumber+masterNumber > 500 && serverNumber+masterNumber <= 2000 {
			nfvServerNumber = 8
		} else {
			nfvServerNumber = 16
		}
	}
	if serverNumber >= 2000 {
		nfvServerNumber = 16
	}
	return nfvServerNumber
}

func CalcMasterServerNumber(pureIaaS bool, serverNumber int, azManageList []*entity.AzManage, cellManage *entity.CellManage) int {
	var masterNumber int
	if pureIaaS {
		if len(azManageList) > 1 {
			if cellManage.Type == constant.CellTypeControl {
				if serverNumber <= 495 {
					masterNumber = 5
				} else if serverNumber <= 1991 {
					masterNumber = 9
				} else {
					masterNumber = 15
				}
			} else {
				if serverNumber <= 197 {
					masterNumber = 3
				} else if serverNumber <= 495 {
					masterNumber = 5
				} else if serverNumber <= 1991 {
					masterNumber = 9
				} else {
					masterNumber = 15
				}
			}
		} else {
			if serverNumber <= 197 {
				masterNumber = 3
			} else if serverNumber <= 495 {
				masterNumber = 5
			} else if serverNumber <= 1991 {
				masterNumber = 9
			} else {
				masterNumber = 15
			}
		}
	} else {
		if serverNumber <= 195 {
			masterNumber = 5
		} else if serverNumber <= 493 {
			masterNumber = 7
		} else if serverNumber <= 1991 {
			masterNumber = 9
		} else {
			masterNumber = 15
		}
	}
	return masterNumber
}
