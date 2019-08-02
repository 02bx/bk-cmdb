package logics

import (
	"context"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
)

func (lgc *Logics) InitInnerChart(ctx context.Context) {
	opt := mapstr.MapStr{}
	result, err := lgc.CoreAPI.CoreService().Operation().SearchOperationChart(ctx, lgc.header, opt)
	if err != nil {
		blog.Errorf("search chart config fail, err: %v", err)
		return
	}

	if result.Data.Count > 0 {
		return
	}

	configID := make([]uint64, 0)
	for _, chart := range InnerChartsArr {
		result, err := lgc.CoreAPI.CoreService().Operation().CreateOperationChart(ctx, lgc.header, InnerCharts[chart])
		if err != nil {
			blog.Errorf("init inner chart fail, err: %v", err)
			return
		}
		configID = append(configID, result.Data)
	}

	position := metadata.ChartPosition{}
	position.Position.Host = configID[2:6]
	position.Position.Inst = configID[6:]
	position.OwnerID = "0"

	if _, err := lgc.CoreAPI.CoreService().Operation().UpdateOperationChartPosition(ctx, lgc.header, position); err != nil {
		blog.Error("init inner chart position fail, err: %v", err)
		return
	}
}

var (
	BizModuleHostChart = metadata.ChartConfig{
		ReportType: common.BizModuleHostChart,
	}

	HostOsChart = metadata.ChartConfig{
		ReportType: common.HostOSChart,
		Name:       "按操作系统类型统计",
		ObjID:      "host",
		Width:      "50",
		ChartType:  "pie",
		Field:      "bk_os_type",
		XAxisCount: 10,
	}

	HostBizChart = metadata.ChartConfig{
		ReportType: common.HostBizChart,
		Name:       "按业务统计",
		ObjID:      "host",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	HostCloudChart = metadata.ChartConfig{
		ReportType: common.HostCloudChart,
		Name:       "按云区域统计",
		Width:      "100",
		ObjID:      "host",
		ChartType:  "bar",
		Field:      common.BKCloudIDField,
		XAxisCount: 20,
	}

	HostChangeBizChart = metadata.ChartConfig{
		ReportType: common.HostChangeBizChart,
		Name:       "主机数量变化趋势",
		Width:      "100",
		XAxisCount: 20,
	}

	ModelAndInstCountChart = metadata.ChartConfig{
		ReportType: common.ModelAndInstCount,
	}

	ModelInstChart = metadata.ChartConfig{
		ReportType: common.ModelInstChart,
		Name:       "实例数量统计",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	ModelInstChangeChart = metadata.ChartConfig{
		ReportType: common.ModelInstChangeChart,
		Name:       "实例变更统计",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	InnerCharts = map[string]metadata.ChartConfig{
		common.BizModuleHostChart:   BizModuleHostChart,
		common.ModelAndInstCount:    ModelAndInstCountChart,
		common.HostOSChart:          HostOsChart,
		common.HostBizChart:         HostBizChart,
		common.HostCloudChart:       HostCloudChart,
		common.HostChangeBizChart:   HostChangeBizChart,
		common.ModelInstChart:       ModelInstChart,
		common.ModelInstChangeChart: ModelInstChangeChart,
	}

	InnerChartsArr = []string{
		common.BizModuleHostChart,
		common.ModelAndInstCount,
		common.HostOSChart,
		common.HostBizChart,
		common.HostCloudChart,
		common.HostChangeBizChart,
		common.ModelInstChart,
		common.ModelInstChangeChart,
	}
)
