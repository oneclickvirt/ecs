package disk

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/oneclickvirt/defaultset"
	"github.com/shirou/gopsutil/disk"
)

// WinsatTest 通过windows自带系统工具测试IO
func WinsatTest(language string, enableMultiCheck bool, testPath string) string {
	var result string
	parts, err := disk.Partitions(true)
	if err == nil {
		if language == "en" {
			result += "Test Disk               Random Read[Score]       Sequential Read[Score]  Sequential Write[Score]\n"
		} else {
			result += "测试的硬盘                随机写入[得分]            顺序读取[得分]            顺序写入[得分]\n"
		}
		if testPath == "" {
			if enableMultiCheck {
				for _, f := range parts {
					result += fmt.Sprintf("%-20s", f.Device) + "    "
					result += getDiskPerformance(f.Device)
				}
			} else {
				result += fmt.Sprintf("%-20s", "C:") + "    "
				result += getDiskPerformance("C:")
			}
		} else {
			result += fmt.Sprintf("%-20s", testPath) + "    "
			result += getDiskPerformance(testPath)
		}
	}
	return result
}

// execDDTest 执行dd命令测试硬盘IO，并回传结果和测试错误
func execDDTest(ifKey, ofKey, bs, blockCount string) (string, error) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var tempText string
	cmd2 := exec.Command("sudo", "dd", "if="+ifKey, "of="+ofKey, "bs="+bs, "count="+blockCount, "oflag=direct")
	stderr2, err := cmd2.StderrPipe()
	if err != nil {
		if EnableLoger {
			Logger.Info("failed to get StderrPipe: " + err.Error())
		}
		return "", err
	}
	if err := cmd2.Start(); err != nil {
		if EnableLoger {
			Logger.Info("failed to start command: " + err.Error())
		}
		return "", err
	}
	outputBytes, err := io.ReadAll(stderr2)
	if err != nil {
		if EnableLoger {
			Logger.Info("failed to read stderr: " + err.Error())
		}
		return "", err
	}

	tempText = string(outputBytes)
	return tempText, nil
}

// ddTest1 无重试机制
func ddTest1(path, deviceName, blockFile, blockName, blockCount, bs string) string {
	var result string
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	// 写入测试
	tempText, err := execDDTest("/dev/zero", path+blockFile, bs, blockCount)
	defer os.Remove(path + blockFile)
	if err != nil {
		if EnableLoger {
			Logger.Info("Write test error: " + err.Error())
		}
	} else {
		result += fmt.Sprintf("%-10s", strings.TrimSpace(deviceName)) + "    " + fmt.Sprintf("%-15s", blockName) + "    "
		result += parseResultDD(tempText, blockCount)
		time.Sleep(1 * time.Second)
	}
	// 清理缓存, 避免影响测试结果
	syncCmd := exec.Command("sync")
	err = syncCmd.Run()
	if err != nil {
		if EnableLoger {
			Logger.Info("sync command failed: " + err.Error())
		}
	}
	// 读取测试
	tempText, err = execDDTest(path+blockFile, "/dev/null", bs, blockCount)
	defer os.Remove(path + blockFile)
	if err != nil {
		if EnableLoger {
			Logger.Info("Read test error: " + err.Error())
		}
	}
	if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
		if err != nil && EnableLoger {
			Logger.Info("Read test (first attempt) error: " + err.Error())
		}
		time.Sleep(1 * time.Second)
		tempText, err = execDDTest(path+blockFile, path+"/read"+blockFile, bs, blockCount)
		defer os.Remove(path + "/read" + blockFile)
		if err != nil && EnableLoger {
			Logger.Info("Read test (second attempt) error: " + err.Error())
		}
	}
	result += parseResultDD(tempText, blockCount)
	result += "\n"
	return result
}

// ddTest2 有重试机制，重试至于 /tmp 目录
func ddTest2(blockFile, blockName, blockCount, bs string) string {
	var result string
	var testFilePath string
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	// 写入测试
	tempText, err := execDDTest("/dev/zero", "/root/"+blockFile, bs, blockCount)
	defer os.Remove("/root/" + blockFile)
	if err != nil {
		if EnableLoger {
			Logger.Info("execDDTest error for /root/ path: " + err.Error())
		}
	}
	if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
		time.Sleep(1 * time.Second)
		tempText, err = execDDTest("/dev/zero", "/tmp/"+blockFile, bs, blockCount)
		defer os.Remove("/tmp/" + blockFile)
		if err != nil {
			if EnableLoger {
				Logger.Info("execDDTest error for /tmp/ path: " + err.Error())
			}
		}
		testFilePath = "/tmp/"
		result += fmt.Sprintf("%-10s", "/tmp") + "    " + fmt.Sprintf("%-15s", blockName) + "    "
	} else {
		testFilePath = "/root/"
		result += fmt.Sprintf("%-10s", "/root") + "    " + fmt.Sprintf("%-15s", blockName) + "    "
	}
	result += parseResultDD(tempText, blockCount)
	// 清理缓存, 避免影响测试结果
	syncCmd := exec.Command("sync")
	err = syncCmd.Run()
	if err != nil {
		if EnableLoger {
			Logger.Info("sync command failed: " + err.Error())
		}
	}
	// 读取测试
	time.Sleep(1 * time.Second)
	tempText, err = execDDTest("/root/"+blockFile, "/dev/null", bs, blockCount)
	defer os.Remove("/root/" + blockFile)
	if err != nil {
		if EnableLoger {
			Logger.Info("execDDTest read error for /root/ path: " + err.Error())
		}
	}
	if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
		time.Sleep(1 * time.Second)
		tempText, err = execDDTest(testFilePath+blockFile, "/tmp/read"+blockFile, bs, blockCount)
		defer os.Remove(testFilePath + blockFile)
		defer os.Remove("/tmp/read" + blockFile)
		if err != nil {
			if EnableLoger {
				Logger.Info("execDDTest read error for /tmp/ path: " + err.Error())
			}
		}
	}
	result += parseResultDD(tempText, blockCount)
	result += "\n"
	return result
}

// DDTest 通过 dd 命令测试硬盘IO
func DDTest(language string, enableMultiCheck bool, testPath string) string {
	var (
		result      string
		devices     []string
		mountPoints []string
	)
	parts, err := disk.Partitions(false)
	if err == nil {
		for _, f := range parts {
			if !strings.Contains(f.Device, "vda") && !strings.Contains(f.Device, "snap") && !strings.Contains(f.Device, "loop") {
				if isWritableMountpoint(f.Mountpoint) {
					devices = append(devices, f.Device)
					mountPoints = append(mountPoints, f.Mountpoint)
				}
			}
		}
	}
	if language == "en" {
		result += "Test Path     Block Size         Direct Write(IOPS)                Direct Read(IOPS)\n"
	} else {
		result += "测试路径      块大小             直接写入(IOPS)                    直接读取(IOPS)\n"
	}
	blockNames := []string{"100MB-4K Block", "1GB-1M Block"}
	blockCounts := []string{"25600", "1000"}
	blockSizes := []string{"4k", "1M"}
	blockFiles := []string{"100MB.test", "1GB.test"}
	for ind, bs := range blockSizes {
		if testPath == "" {
			if enableMultiCheck {
				for index, path := range mountPoints {
					result += ddTest1(path, devices[index], blockFiles[ind], blockNames[ind], blockCounts[ind], bs)
				}
			} else {
				result += ddTest2(blockFiles[ind], blockNames[ind], blockCounts[ind], bs)
			}
		} else {
			result += ddTest1(testPath, testPath, blockFiles[ind], blockNames[ind], blockCounts[ind], bs)
		}
	}
	return result
}

// buildFioFile 生成对应文件
func buildFioFile(path, fioSize string) (string, error) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	// https://github.com/masonr/yet-another-bench-script/blob/0ad4c4e85694dbcf0958d8045c2399dbd0f9298c/yabs.sh#L435
	// fio --name=setup --ioengine=libaio --rw=read --bs=64k --iodepth=64 --numjobs=2 --size=512MB --runtime=1 --gtod_reduce=1 --filename=/tmp/test.fio --direct=1 --minimal
	var tempText string
	cmd1 := exec.Command("sudo", "fio", "--name=setup", "--ioengine=libaio", "--rw=read", "--bs=64k", "--iodepth=64", "--numjobs=2", "--size="+fioSize, "--runtime=1", "--gtod_reduce=1",
		"--filename="+path+"/test.fio", "--direct=1", "--minimal")
	stderr1, err := cmd1.StderrPipe()
	if err != nil {
		if EnableLoger {
			Logger.Info("failed to get stderr pipe: " + err.Error())
		}
		return "", err
	}
	if err := cmd1.Start(); err != nil {
		if EnableLoger {
			Logger.Info("failed to start fio command: " + err.Error())
		}
		return "", err
	}
	outputBytes, err := io.ReadAll(stderr1)
	if err != nil {
		if EnableLoger {
			Logger.Info("failed to read stderr: " + err.Error())
		}
		return "", err
	}
	tempText = string(outputBytes)
	return tempText, nil
}

// execFioTest 使用fio测试文件进行测试
func execFioTest(path, devicename, fioSize string) (string, error) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var result string
	// 测试
	blockSizes := []string{"4k", "64k", "512k", "1m"}
	for _, BS := range blockSizes {
		// timeout 35 fio --name=rand_rw_4k --ioengine=libaio --rw=randrw --rwmixread=50 --bs=4k --iodepth=64 --numjobs=2 --size=512MB --runtime=30 --gtod_reduce=1 --direct=1 --filename="/tmp/test.fio" --group_reporting --minimal
		cmd2 := exec.Command("timeout", "35", "sudo", "fio", "--name=rand_rw_"+BS, "--ioengine=libaio", "--rw=randrw", "--rwmixread=50", "--bs="+BS, "--iodepth=64", "--numjobs=2", "--size="+fioSize, "--runtime=30", "--gtod_reduce=1", "--direct=1", "--filename="+path+"/test.fio", "--group_reporting", "--minimal")
		output, err := cmd2.Output()
		if err != nil {
			if EnableLoger {
				Logger.Info("failed to execute fio command: " + err.Error())
			}
			return "", err
		} else {
			tempText := string(output)
			tempList := strings.Split(tempText, "\n")
			for _, l := range tempList {
				if strings.Contains(l, "rand_rw_"+BS) {
					tpList := strings.Split(l, ";")
					// IOPS
					DISK_IOPS_R := tpList[8]
					DISK_IOPS_W := tpList[49]
					DISK_IOPS_R_INT, _ := strconv.Atoi(DISK_IOPS_R)
					DISK_IOPS_W_INT, _ := strconv.Atoi(DISK_IOPS_W)
					DISK_IOPS := DISK_IOPS_R_INT + DISK_IOPS_W_INT
					// Speed
					DISK_TEST_R := tpList[7]
					DISK_TEST_W := tpList[48]
					DISK_TEST_R_INT, _ := strconv.ParseFloat(DISK_TEST_R, 64)
					DISK_TEST_W_INT, _ := strconv.ParseFloat(DISK_TEST_W, 64)
					DISK_TEST := DISK_TEST_R_INT + DISK_TEST_W_INT
					// 拼接输出文本
					result += fmt.Sprintf("%-10s", devicename) + "    "
					result += fmt.Sprintf("%-5s", BS) + "    "
					result += fmt.Sprintf("%-20s", formatSpeed(DISK_TEST_R, "string")+"("+formatIOPS(DISK_IOPS_R, "string")+")") + "    "
					result += fmt.Sprintf("%-20s", formatSpeed(DISK_TEST_W, "string")+"("+formatIOPS(DISK_IOPS_W, "string")+")") + "    "
					result += fmt.Sprintf("%-20s", formatSpeed(DISK_TEST, "float64")+"("+formatIOPS(DISK_IOPS, "int")+")") + "    "
					result += "\n"
				}
			}
		}
	}
	return result, nil
}

// FioTest 通过fio测试硬盘
func FioTest(language string, enableMultiCheck bool, testPath string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	cmd := exec.Command("fio", "-v")
	_, err := cmd.CombinedOutput()
	if err != nil {
		if EnableLoger {
			Logger.Info("failed to match fio version: " + err.Error())
		}
		return ""
	}
	var (
		result, fioSize string
		devices         []string
		mountPoints     []string
	)
	parts, err := disk.Partitions(false)
	if err == nil {
		for _, f := range parts {
			if !strings.Contains(f.Device, "vda") && !strings.Contains(f.Device, "snap") && !strings.Contains(f.Device, "loop") {
				if isWritableMountpoint(f.Mountpoint) {
					devices = append(devices, f.Device)
					mountPoints = append(mountPoints, f.Mountpoint)
				}
			}
		}
	}
	if language == "en" {
		result += "Test Path     Block    Read(IOPS)              Write(IOPS)             Total(IOPS)\n"
	} else {
		result += "测试路径      块大小   读测试(IOPS)            写测试(IOPS)            总和(IOPS)\n"
	}
	// 生成测试文件
	if runtime.GOARCH == "amd64" || runtime.GOARCH == "x86" || runtime.GOARCH == "x86_64" {
		fioSize = "2G"
	} else {
		fioSize = "512MB"
	}
	if testPath == "" {
		if enableMultiCheck {
			for index, path := range mountPoints {
				_, err := buildFioFile(path, fioSize)
				defer os.Remove(path + "/test.fio")
				if err == nil {
					time.Sleep(1 * time.Second)
					tempResult, err := execFioTest(path, strings.TrimSpace(devices[index]), fioSize)
					if err == nil {
						result += tempResult
					}
				}
			}
		} else {
			var buildPath string
			tempText, err := buildFioFile("/root", fioSize)
			defer os.Remove("/root/test.fio")
			if err != nil || strings.Contains(tempText, "failed") || strings.Contains(tempText, "Permission denied") {
				_, err = buildFioFile("/tmp", fioSize)
				if err == nil {
					buildPath = "/tmp"
				}
			} else {
				buildPath = "/root"
			}
			if buildPath != "" {
				time.Sleep(1 * time.Second)
				tempResult, err := execFioTest(buildPath, buildPath, fioSize)
				if err == nil {
					result += tempResult
				}
			}
		}
	} else {
		tempText, err := buildFioFile(testPath, fioSize)
		defer os.Remove(testPath + "/test.fio")
		if err != nil || strings.Contains(tempText, "failed") || strings.Contains(tempText, "Permission denied") {
			return tempText
		}
		time.Sleep(1 * time.Second)
		tempResult, err := execFioTest(testPath, testPath, fioSize)
		if err == nil {
			result += tempResult
		}
	}
	return result
}
