// +build freebsd openbsd netbsd dragonflybsd darwin

package ipset

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

var (
	ipLists = make(map[string][]net.IP)
)

func initLib() (err error) {
	// 调用 pfctl 命令获取所有 pf 表的列表
	cmd := exec.Command("pfctl", "-s", "Tables")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return errors.New("Error running pfctl command: " + err.Error())
	}

	// 按行解析输出，每行表示一个表名
	tableNames := strings.Split(stdout.String(), "\n")

	for _, tableName := range tableNames {
		// 跳过空行
		if tableName == "" {
			continue
		}

		// 调用 pfctl 命令获取表中的 IP 地址
		cmd = exec.Command("pfctl", "-t", tableName, "-T", "show")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err = cmd.Run()
		if err != nil {
			return errors.New("Error running pfctl command for table " + tableName + ": " + err.Error())
		}

		// 按行解析输出，每行表示一个 IP 地址
		ipLines := strings.Split(stdout.String(), "\n")

		// 创建 IP 列表并将其添加到 ipLists
		ipList := make([]net.IP, 0)

		for _, ipLine := range ipLines {
			// 跳过空行
			if ipLine == "" {
				continue
			}

			ip := net.ParseIP(ipLine)
			if ip != nil {
				ipList = append(ipList, ip)
			}
		}

		ipLists[tableName] = ipList
	}

	return nil
}

func addIP(ip net.IP, list string) error {
	// 获取指定名称的 IP 列表
	ipList, exists := ipLists[list]

	// 如果列表不存在，则创建一个新列表
	if !exists {
		ipLists[list] = []net.IP{ip}
	} else {
		// 检查 IP 地址是否已经存在于列表中
		for _, existingIP := range ipList {
			if ip.Equal(existingIP) {
				return errors.New("IP address already exists in the list")
			}
		}

		// 将 IP 地址添加到列表中
		ipLists[list] = append(ipLists[list], ip)
	}

	// 使用 pfctl 命令将 IP 地址添加到指定的 pf 表中
	cmd := exec.Command("pfctl", "-t", list, "-T", "add", ip.String())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.New("Error running pfctl command: " + err.Error() + " - " + stderr.String())
	}

	return nil
}

func flushSet(list string) error {
	if _, exists := ipLists[list]; exists {
		delete(ipLists, list)
		return nil
	}
	return errors.New("List not found")
}

func shutdownLib() error {
	return c.Close()
}
