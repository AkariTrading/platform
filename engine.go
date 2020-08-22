package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func runAtEngine(id string, isTest bool) ([]byte, error) {

	query := url.Values{}
	query.Set("isTest", strconv.FormatBool(isTest))

	ip, err := bestNode(getNodes())

	fmt.Println(ip)

	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("scriptVersions/%s", id), ip+DebugEnginePort)

	url, _ := url.Parse(fmt.Sprintf("http://%s/scriptVersions/%s", ip+DebugEnginePort, id))

	res, err := engineClient.Do(&http.Request{
		Method: "POST",
		URL:    url,
	})

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if res.StatusCode == http.StatusOK {
		return nil, nil
	}

	return body, errors.New("engine error")
}

func stopAtEngine(ip string, id string) ([]byte, error) {

	fmt.Println(fmt.Sprintf("scriptVersions/%s", id), ip+DebugEnginePort)

	url, _ := url.Parse(fmt.Sprintf("http://%s/scriptVersions/%s", ip+DebugEnginePort, id))

	res, err := engineClient.Do(&http.Request{
		Method: "DELETE",
		URL:    url,
	})

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if res.StatusCode == http.StatusOK {
		return nil, nil
	}

	return body, errors.New("engine error")
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
