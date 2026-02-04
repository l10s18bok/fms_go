// Mock 서버 - 리소스 모니터링 테스트용
// 실행: go run mock_server.go
// 장비 추가 시 IP: 127.0.0.1, 포트: 8080
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/device-report", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] /device-report 요청 받음\n", time.Now().Format("15:04:05"))

		response := map[string]interface{}{
			"totalUsed": map[string]interface{}{
				"cpu":          rand.Float64()*50 + 10, // 10~60%
				"memory":       rand.Float64()*40 + 30, // 30~70%
				"disk":         rand.Float64()*30 + 40, // 40~70%
				"rxBps":        rand.Float64() * 10000000,
				"txBps":        rand.Float64() * 5000000,
				"rxBpsDisplay": fmt.Sprintf("%.1f MB/s", rand.Float64()*10),
				"txBpsDisplay": fmt.Sprintf("%.1f MB/s", rand.Float64()*5),
				"rxPps":        rand.Int63n(10000),
				"txPps":        rand.Int63n(5000),
			},
			"cpuall": map[string]interface{}{
				"user":   rand.Float64() * 30,
				"kernel": rand.Float64() * 15,
				"idle":   rand.Float64()*30 + 50,
				"iowait": rand.Float64() * 5,
				"ni":     rand.Float64() * 1,
				"hwIrq":  rand.Float64() * 1,
				"swIrq":  rand.Float64() * 1,
				"steal":  0,
				"avg1":   rand.Float64() * 2,
				"avg5":   rand.Float64() * 1.5,
				"avg15":  rand.Float64() * 1,
				"usage":  rand.Float64() * 50,
				"total":  100,
			},
			"memory": map[string]interface{}{
				"total":            int64(17179869184),
				"free":             int64(5478546637),
				"used":             int64(11701322547),
				"available":        int64(8589934592),
				"totalDisplay":     "16.0 GB",
				"freeDisplay":      "5.1 GB",
				"usedDisplay":      "10.9 GB",
				"availableDisplay": "8.0 GB",
				"usedPercent":      68.1,
				"buffCache":        int64(2684354560),
				"swapTotal":        int64(4294967296),
				"swapFree":         int64(3758096384),
				"swapUsed":         int64(536870912),
				"buffCacheDisplay": "2.5 GB",
				"swapTotalDisplay": "4.0 GB",
				"swapFreeDisplay":  "3.5 GB",
				"swapUsedDisplay":  "0.5 GB",
			},
			"disk": []map[string]interface{}{
				{
					"filesystem":  "/dev/sda1",
					"mountedOn":   "/",
					"size":        int64(107374182400),
					"used":        int64(59055800320),
					"avail":       int64(48318382080),
					"free":        int64(48318382080),
					"sizeDisplay": "100 GB",
					"usedDisplay": "55 GB",
					"freeDisplay": "45 GB",
					"usePercent":  55.0,
				},
				{
					"filesystem":  "/dev/sdb1",
					"mountedOn":   "/data",
					"size":        int64(536870912000),
					"used":        int64(214748364800),
					"avail":       int64(322122547200),
					"free":        int64(322122547200),
					"sizeDisplay": "500 GB",
					"usedDisplay": "200 GB",
					"freeDisplay": "300 GB",
					"usePercent":  40.0,
				},
			},
			"interface": []map[string]interface{}{
				{
					"ifname":         "eth0",
					"ip4":            "192.168.1.100",
					"ip6":            "fe80::1",
					"rxPackets":      rand.Int63n(1000000),
					"rxBytes":        rand.Int63n(10737418240),
					"rxBytesDisplay": "10 GB",
					"rxErrors":       0,
					"rxDropped":      0,
					"txPackets":      rand.Int63n(500000),
					"txBytes":        rand.Int63n(5368709120),
					"txBytesDisplay": "5 GB",
					"txErrors":       0,
					"txDropped":      0,
					"rxBps":          rand.Float64() * 10000000,
					"txBps":          rand.Float64() * 5000000,
					"rxBpsDisplay":   fmt.Sprintf("%.1f MB/s", rand.Float64()*10),
					"txBpsDisplay":   fmt.Sprintf("%.1f MB/s", rand.Float64()*5),
					"rxPps":          rand.Int63n(10000),
					"txPps":          rand.Int63n(5000),
				},
				{
					"ifname":         "lo",
					"ip4":            "127.0.0.1",
					"ip6":            "::1",
					"rxPackets":      rand.Int63n(100000),
					"rxBytes":        rand.Int63n(1073741824),
					"rxBytesDisplay": "1 GB",
					"rxErrors":       0,
					"rxDropped":      0,
					"txPackets":      rand.Int63n(100000),
					"txBytes":        rand.Int63n(1073741824),
					"txBytesDisplay": "1 GB",
					"txErrors":       0,
					"txDropped":      0,
					"rxBps":          rand.Float64() * 1000000,
					"txBps":          rand.Float64() * 1000000,
					"rxBpsDisplay":   "0.5 MB/s",
					"txBpsDisplay":   "0.5 MB/s",
					"rxPps":          rand.Int63n(1000),
					"txPps":          rand.Int63n(1000),
				},
			},
			"process": []map[string]interface{}{
				{
					"name":       "nginx",
					"version":    "1.18.0",
					"user":       "www-data",
					"pid":        1234,
					"cpuPercent": rand.Float64() * 5,
					"memPercent": rand.Float64() * 2,
					"vsz":        int64(104857600),
					"rss":        int64(52428800),
					"vszDisplay": "100 MB",
					"rssDisplay": "50 MB",
					"start":      "Jan01",
					"runTime":    "10:23:45",
					"status":     []string{"S"},
					"cmd":        "nginx: master process",
					"tty":        "?",
				},
				{
					"name":       "agent",
					"version":    "2.0.1",
					"user":       "root",
					"pid":        5678,
					"cpuPercent": rand.Float64() * 2,
					"memPercent": rand.Float64() * 1,
					"vsz":        int64(52428800),
					"rss":        int64(31457280),
					"vszDisplay": "50 MB",
					"rssDisplay": "30 MB",
					"start":      "Jan01",
					"runTime":    "5:12:30",
					"status":     []string{"S"},
					"cmd":        "./agent",
					"tty":        "?",
				},
				{
					"name":       "mysql",
					"version":    "8.0.32",
					"user":       "mysql",
					"pid":        9012,
					"cpuPercent": rand.Float64() * 10,
					"memPercent": rand.Float64() * 5,
					"vsz":        int64(1073741824),
					"rss":        int64(536870912),
					"vszDisplay": "1 GB",
					"rssDisplay": "512 MB",
					"start":      "Jan01",
					"runTime":    "100:00:00",
					"status":     []string{"S"},
					"cmd":        "mysqld",
					"tty":        "?",
				},
				// name과 version이 없는 프로세스 (표시되지 않아야 함)
				{
					"user":       "root",
					"pid":        1,
					"cpuPercent": 0.1,
					"memPercent": 0.1,
					"vsz":        int64(10485760),
					"rss":        int64(5242880),
					"vszDisplay": "10 MB",
					"rssDisplay": "5 MB",
					"start":      "Jan01",
					"runTime":    "200:00:00",
					"status":     []string{"S"},
					"cmd":        "/sbin/init",
					"tty":        "?",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		fmt.Println("  -> 응답 완료")
	})

	fmt.Println("========================================")
	fmt.Println("  리소스 모니터링 Mock 서버")
	fmt.Println("========================================")
	fmt.Println("  URL: http://localhost:8080/device-report")
	fmt.Println("")
	fmt.Println("  fms_fyne 장비 추가 설정:")
	fmt.Println("    - 장비 IP: 127.0.0.1")
	fmt.Println("    - 장비보고 포트: 8080")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Println("서버 시작됨. Ctrl+C로 종료.")
	fmt.Println("")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 시작 실패: %v\n", err)
	}
}
