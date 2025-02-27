package cloud_product

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
)

func ListVersion(context *gin.Context) {
	param := context.Query("projectId")
	projectId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[ListVersion] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	// 根据项目id查询云平台类型
	versionList, err := getVersionListByProjectId(projectId)
	if err != nil {
		log.Errorf("[ListVersion] getVersionListByProjectId error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, versionList)
	return
}

func ListCloudProductBaseline(context *gin.Context) {
	param := context.Query("versionId")
	versionId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[ListVersion] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	baselineResponseList, err := getCloudProductBaseListByVersionId(versionId)
	if err != nil {
		log.Errorf("[ListCloudProductBaseline] getCloudProductBaseListByVersionId error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, baselineResponseList)
	return
}

func Save(context *gin.Context) {
	var request CloudProductPlanningRequest
	err := context.BindJSON(&request)
	if err != nil {
		log.Errorf("[Save] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if request.PlanId == 0 || len(request.ProductList) < 1 {
		log.Errorf("[Save] invalid param error, request:%v, %v", request, err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	var productIdList []int64
	for _, product := range request.ProductList {
		productIdList = append(productIdList, product.ProductId)
	}
	// 必选云产品校验
	baselineResponseList, err := getCloudProductBaseListByVersionId(request.VersionId)
	if err != nil {
		log.Errorf("[ListCloudProductBaseline] getCloudProductBaseListByVersionId error", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	for _, baseline := range baselineResponseList {
		if baseline.WhetherRequired == 1 && !contains(baseline.ID, productIdList) {
			log.Error("[Save] invalid param error, 必选云产品未选中")
			result.FailureWithMsg(context, errorcodes.CloudProductRequiredError, http.StatusBadRequest, "必选云产品未选中")
			return
		}
	}
	// 依赖云产品校验
	dependList, err := getDependProductIds()
	for _, productId := range productIdList {
		for _, depend := range dependList {
			// 判断productIdList是否包含depend.DependId
			if productId == depend.ID && !contains(depend.DependId, productIdList) {
				log.Error("[Save] invalid param error, 选择的云产品有依赖项未选中")
				result.FailureWithMsg(context, errorcodes.CloudProductDependError, http.StatusBadRequest, "选择的云产品有依赖项未选中")
				return
			}
		}
	}

	err = saveCloudProductPlanning(request, context.GetString(constant.CurrentUserId))
	if err != nil {
		log.Errorf("[Save] cloudProductPlanning error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, nil)
	return
}

func contains(ele int64, arr []int64) bool {
	for _, data := range arr {
		if data == ele {
			return true
		}
	}
	return false
}

func List(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[List] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	cloudProductPlannings, err := ListCloudProductPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[List] cloudProductPlanning error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(context, cloudProductPlannings)
	return
}

func Export(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[Export] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	fileName, exportResponseDataList, err := exportCloudProductPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[Export] cloudProductPlanning error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	excel.NormalDownLoad(fileName, "云产品清单", "", false, exportResponseDataList, context.Writer)
	return
}
