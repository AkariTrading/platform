package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/akaritrading/libs/redis"
	"github.com/akaritrading/platform/pkg/engine"
)

var engineClient = &http.Client{
	Timeout: time.Second * 30,
}

func stopAtEngine(ip string, id uint, isTest bool) {

}

func runAtEngine(id uint, isTest bool) bool {

	query := url.Values{}
	query.Set("isTest", strconv.FormatBool(isTest))

	ip, err := bestNode(getNodes())
	if err != nil {
		return false
	}

	fmt.Println(fmt.Sprintf("scriptVersions/%d", id), ip+DebugEnginePort)

	url, _ := url.Parse(fmt.Sprintf("http://%s/scriptVersions/%d", ip+DebugEnginePort, id))

	req := &http.Request{
		Method: "POST",
		URL:    url,
	}

	res, err := engineClient.Do(req)

	fmt.Println(res)

	if err != nil {
		log.Fatal(err)
		return false
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	if res.StatusCode == http.StatusOK {
		return true
	}

	return false
}

func bestNode(stats map[string]engine.MachineStat) (string, error) {

	var cpulevel = 10.0
	var memlevel = 10.0

	if len(stats) == 0 {
		return "", errors.New("no nodes found")
	}

	for {
		for ip, stat := range stats {
			if stat.CpuUsedPercent <= cpulevel {
				if stat.MemoryUsedPercent <= memlevel {
					return ip, nil
				}
			}
			cpulevel += 10.0
			memlevel += 10.0
		}
	}
}

func getNodes() map[string]engine.MachineStat {

	var ret = make(map[string]engine.MachineStat)

	stats, err := redis.StringMap(redisHandle.Do(redis.GetHash, engine.MachineStatsRedisKey))

	if err != nil {
		log.Fatal(err)
		return ret
	}

	var fieldToDelete []string

	for ip, statJSON := range stats {

		var stat engine.MachineStat
		if err := json.Unmarshal([]byte(statJSON), &stat); err != nil {
			log.Fatal(err) // change
			break
		}

		if time.Since(stat.UpdatedAt) > time.Minute*2 {
			// log node may be dead, remove
			fieldToDelete = append(fieldToDelete, ip)
		}

		ret[ip] = stat

	}

	if len(fieldToDelete) > 0 {
		conn := redisHandle.Conn()
		for _, field := range fieldToDelete {
			conn.Send(redis.DeleteField, engine.MachineStatsRedisKey, field)
		}
		conn.Flush()
		conn.Close()
	}

	return ret
}
