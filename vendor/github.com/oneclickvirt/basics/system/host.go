package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/libp2p/go-nat"
	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/system/utils"
	"github.com/shirou/gopsutil/v4/host"
)

func getVmTypeFromSDV(path string) string {
	cmd := exec.Command(fmt.Sprintf("%s/systemd-detect-virt", path))
	output, err := cmd.Output()
	if err == nil {
		switch strings.TrimSpace(strings.ReplaceAll(string(output), "\n", "")) {
		case "kvm":
			return "KVM"
		case "xen":
			return "Xen Hypervisor"
		case "microsoft":
			return "Microsoft Hyper-V"
		case "vmware":
			return "VMware"
		case "oracle":
			return "Oracle VirtualBox"
		case "parallels":
			return "Parallels"
		case "qemu":
			return "QEMU"
		case "amazon":
			return "Amazon Virtualization"
		case "docker":
			return "Docker"
		case "openvz":
			return "OpenVZ (Virutozzo)"
		case "lxc":
			return "LXC"
		case "lxc-libvirt":
			return "LXC (Based on libvirt)"
		case "uml":
			return "User-mode Linux"
		case "systemd-nspawn":
			return "Systemd nspawn"
		case "bochs":
			return "BOCHS"
		case "rkt":
			return "RKT"
		case "zvm":
			return "S390 Z/VM"
		case "none":
			return "Dedicated"
		}
	}
	return ""
}

func getHostInfo() (string, string, string, string, string, string, string, string, error) {
	var Platform, Kernal, Arch, VmType, NatType, CurrentTimeZone string
	var cachedBootTime time.Time
	hi, err := host.Info()
	if err != nil {
		fmt.Println("host.Info error:", err)
	} else {
		if hi.VirtualizationRole == "guest" {
			cpuType = "Virtual"
		} else {
			cpuType = "Physical"
		}
		Platform = hi.Platform + " " + hi.PlatformVersion
		Arch = hi.KernelArch
		// 查询虚拟化类型和内核
		if runtime.GOOS != "windows" {
			cmd := exec.Command("uname", "-r")
			output, err := cmd.Output()
			if err == nil {
				Kernal = strings.TrimSpace(strings.ReplaceAll(string(output), "\n", ""))
			}
			path, exit := utils.GetPATH("systemd-detect-virt")
			if exit {
				VmType = getVmTypeFromSDV(path)
			}
			if VmType == "" {
				_, err := os.Stat("/.dockerenv")
				if os.IsExist(err) {
					VmType = "Docker"
				}
				_, err = os.Stat("/dev/lxss")
				if os.IsExist(err) {
					VmType = "Windows Subsystem for Linux"
				}
				if VmType == "" {
					VmType = "Dedicated (No visible signage)"
				}
			}
		} else {
			VmType = hi.VirtualizationSystem
		}
		// 系统运行时长查询 /proc/uptime
		cachedBootTime = time.Unix(int64(hi.BootTime), 0)
	}
	uptimeDuration := time.Since(cachedBootTime)
	days := int(uptimeDuration.Hours() / 24)
	uptimeDuration -= time.Duration(days*24) * time.Hour
	hours := int(uptimeDuration.Hours())
	uptimeDuration -= time.Duration(hours) * time.Hour
	minutes := int(uptimeDuration.Minutes())
	uptimeFormatted := fmt.Sprintf("%d days, %02d hours, %02d minutes", days, hours, minutes)
	// windows 查询虚拟化类型 使用 wmic
	if VmType == "" && runtime.GOOS == "windows" {
		VmType = utils.CheckVMTypeWithWIMC()
	}
	// MAC需要额外获取信息进行判断
	if runtime.GOOS == "darwin" {
		if len(model.MacOSInfo) > 0 {
			for _, line := range model.MacOSInfo {
				if strings.Contains(line, "Model Name") {
					VmType = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
		}
	}
	// 查询NAT类型
	NatType = getNatType()
	if NatType == "Inconclusive" {
		ctx := context.Background()
		gateway, err := nat.DiscoverGateway(ctx)
		if err == nil {
			natType := gateway.Type()
			NatType = natType
		}
	}
	// 获取当前系统的本地时区
	CurrentTimeZone = utils.GetTimeZone()
	return cpuType, uptimeFormatted, Platform, Kernal, Arch, VmType, NatType, CurrentTimeZone, nil
}
