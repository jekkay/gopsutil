// +build openbsd

package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/jekkay/gopsutil/internal/common"
	"github.com/jekkay/gopsutil/process"
	"golang.org/x/sys/unix"
)

const (
	UTNameSize = 32 /* see MAXLOGNAME in <sys/param.h> */
	UTLineSize = 8
	UTHostSize = 16
)

func Info() (*InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) (*InfoStat, error) {
	ret := &InfoStat{
		OS:             runtime.GOOS,
		PlatformFamily: "openbsd",
	}

	hostname, err := os.Hostname()
	if err == nil {
		ret.Hostname = hostname
	}

	kernelArch, err := kernelArch()
	if err == nil {
		ret.KernelArch = kernelArch
	}

	platform, family, version, err := PlatformInformation()
	if err == nil {
		ret.Platform = platform
		ret.PlatformFamily = family
		ret.PlatformVersion = version
	}
	system, role, err := Virtualization()
	if err == nil {
		ret.VirtualizationSystem = system
		ret.VirtualizationRole = role
	}

	procs, err := process.Pids()
	if err == nil {
		ret.Procs = uint64(len(procs))
	}

	boot, err := BootTime()
	if err == nil {
		ret.BootTime = boot
		ret.Uptime = uptime(boot)
	}

	return ret, nil
}

func PlatformInformation() (string, string, string, error) {
	return PlatformInformationWithContext(context.Background())
}

func PlatformInformationWithContext(ctx context.Context) (string, string, string, error) {
	platform := ""
	family := ""
	version := ""

	p, err := unix.Sysctl("kern.ostype")
	if err == nil {
		platform = strings.ToLower(p)
	}
	v, err := unix.Sysctl("kern.osrelease")
	if err == nil {
		version = strings.ToLower(v)
	}

	return platform, family, version, nil
}

func Virtualization() (string, string, error) {
	return VirtualizationWithContext(context.Background())
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}

func Users() ([]UserStat, error) {
	return UsersWithContext(context.Background())
}

func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	var ret []UserStat
	utmpfile := "/var/run/utmp"
	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	u := Utmp{}
	entrySize := int(unsafe.Sizeof(u))
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]
		var u Utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil || u.Time == 0 {
			continue
		}
		user := UserStat{
			User:     common.IntToString(u.Name[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(u.Time),
		}

		ret = append(ret, user)
	}

	return ret, nil
}

func SensorsTemperatures() ([]TemperatureStat, error) {
	return SensorsTemperaturesWithContext(context.Background())
}

func SensorsTemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}

func KernelVersion() (string, error) {
	return KernelVersionWithContext(context.Background())
}

func KernelVersionWithContext(ctx context.Context) (string, error) {
	_, _, version, err := PlatformInformation()
	return version, err
}
