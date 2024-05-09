package harvester

var (
	ApiVersionHarvesterKey string = "harvesterhci.io/v1beta1"
	ApiVersionKubevirt     string = "kubevirt.io/v1"
)

var (
	KindVirtualMachineImage string = "VirtualMachineImage"
	KindVirtualMachine      string = "VirtualMachine"
	KindVolume              string = "PersistentVolumeClaim"
)

var (
	VirtualMachineSpecRunStrategy string = "RerunOnFailure"
	StorageClassName              string = "harvester-longhorn"
)
