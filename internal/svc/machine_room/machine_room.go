package machine_room

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
)

func PageCabinets(context *gin.Context) {
	request := &PageRequest{Current: 1, PageSize: 10}
	if err := context.ShouldBindQuery(&request); err != nil {
		log.Errorf("page cabinets bind param error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 {
		result.Failure(context, "方案id参数为空", http.StatusBadRequest)
		return
	}
	list, count, err := QueryCabinetsPage(request)
	if err != nil {
		log.Errorf("page cabinets error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.SuccessPage(context, count, list)
}

func DownloadCabinetTemplate(context *gin.Context) {
	file, err := excelize.OpenFile("template/机房勘察模版.xlsx")
	if err != nil {
		log.Errorf("download cabinet template error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	excel.DownLoadExcel("机房勘察模版", context.Writer, file)
	return
}

func ImportCabinet(context *gin.Context) {
	file, err := context.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if file.Size == 0 {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	planIdStr := context.PostForm("planId")
	if planIdStr == "" {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	planId, err := strconv.ParseInt(planIdStr, 10, 64)
	if err != nil {
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cabinets, err := QueryCabinetsByPlanId(planId)
	if err != nil && err != gorm.ErrRecordNotFound {
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "cabinet", time.Now().Unix(), rand.Uint32())
	if err := context.SaveUploadedFile(file, filePath); err != nil {
		log.Error(err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Errorf("excelize openFile error: %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		if err := f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
	}()
	var cabinetExcelList []CabinetExcel
	if err := excel.ImportBySheet(f, &cabinetExcelList, "机房勘察模版", 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if len(cabinetExcelList) > 0 {
		if len(cabinets) > 0 {
			// 先删除，再新增
			if err := DeleteCabinets(cabinets); err != nil {
				log.Errorf("delete cabinets error: %v", err)
				result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
				return
			}
		}
		cabinets = []entity.CabinetInfo{}
		now := time.Now()
		for _, cabinetExcel := range cabinetExcelList {
			var cabinetType int
			switch cabinetExcel.CabinetType {
			case constant.CabinetTypeNetworkCn:
				cabinetType = constant.CabinetTypeNetwork
				break
			case constant.CabinetTypeBusinessCn:
				cabinetType = constant.CabinetTypeBusiness
				break
			case constant.CabinetTypeStorageCn:
				cabinetType = constant.CabinetTypeStorage
				break
			default:
				result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
				return
			}
			cabinets = append(cabinets, entity.CabinetInfo{
				PlanId:                planId,
				MachineRoomAbbr:       cabinetExcel.MachineRoomAbbr,
				MachineRoomNum:        cabinetExcel.MachineRoomNum,
				ColumnNum:             cabinetExcel.ColumnNum,
				CabinetNum:            cabinetExcel.CabinetNum,
				OriginalNum:           cabinetExcel.OriginalNum,
				CabinetType:           cabinetType,
				BusinessAttribute:     cabinetExcel.BusinessAttribute,
				CabinetAsw:            cabinetExcel.CabinetAsw,
				TotalPower:            cabinetExcel.TotalPower,
				ResidualPower:         cabinetExcel.ResidualPower,
				TotalSlotNum:          cabinetExcel.TotalSlotNum,
				IdleSlotRange:         cabinetExcel.IdleSlotRange,
				MaxRackServerNum:      cabinetExcel.MaxRackServerNum,
				ResidualRackServerNum: cabinetExcel.ResidualRackServerNum,
				RackServerSlot:        cabinetExcel.RackServerSlot,
				ResidualRackAswPort:   cabinetExcel.ResidualRackAswPort,
				CreateTime:            now,
			})
		}
		if err := BatchCreateCabinets(cabinets); err != nil {
			result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
			return
		}
		// TODO 处理机柜空闲槽位关联数据、机柜已上架服务器槽位关联数据、机柜剩余可上架ASW端口关联数据
	}
	result.Success(context, nil)
	return
}
