package entity

const ServerBaselineTable = "server_baseline"

type ServerBaseline struct {
	Id                  int64  `gorm:"column:id" json:"id"`                                     // 主键id
	VersionId           int64  `gorm:"column:version_id" json:"versionId"`                      // 版本id
	Arch                string `gorm:"column:arch" json:"Arch"`                                 // 硬件架构
	NetworkInterface    string `gorm:"column:network_interface" json:"networkInterface"`        // 网络接口
	BomCode             string `gorm:"column:bom_code" json:"bomCode"`                          // BOM编码
	ConfigurationInfo   string `gorm:"column:configuration_info" json:"configurationInfo"`      // 配置概要
	Spec                string `gorm:"column:spec" json:"spec"`                                 // 规格
	CpuType             string `gorm:"column:cpu_type" json:"cpuType"`                          // CPU类型
	Cpu                 int    `gorm:"column:cpu" json:"cpu"`                                   // CPU核数
	Gpu                 string `gorm:"column:gpu" json:"gpu"`                                   // GPU
	Memory              int    `gorm:"column:memory" json:"memory"`                             // 内存
	SystemDiskType      string `gorm:"column:system_disk_type" json:"systemDiskType"`           // 系统盘类型
	SystemDisk          string `gorm:"column:system_disk" json:"systemDisk"`                    // 系统盘
	StorageDiskType     string `gorm:"column:storage_disk_type" json:"storageDiskType"`         // 存储盘类型
	StorageDiskNum      int    `gorm:"column:storage_disk_num" json:"storageDiskNum"`           // 存储盘数量
	StorageDiskCapacity int    `gorm:"column:storage_disk_capacity" json:"storageDiskCapacity"` // 存储盘单盘容量（G）
	RamDisk             string `gorm:"column:ram_disk" json:"ramDisk"`                          // 缓存盘
	NetworkCardNum      int    `gorm:"column:network_card_num" json:"networkCardNum"`           // 网卡数量
	Power               int    `gorm:"column:power" json:"power"`                               // 功率
}

func (entity *ServerBaseline) TableName() string {
	return ServerBaselineTable
}
