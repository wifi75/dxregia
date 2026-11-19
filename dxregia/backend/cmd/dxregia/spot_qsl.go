package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const qslRefreshInterval = 72 * time.Hour
const maxQSLSourceBytes = 8 << 20

type qslDirectory struct {
	lotw, eqsl atomic.Pointer[map[string]struct{}]
}

type qslSnapshot struct {
	lotw, eqsl *map[string]struct{}
}

func (q *qslDirectory) snapshot() qslSnapshot {
	if q == nil {
		return qslSnapshot{}
	}
	return qslSnapshot{q.lotw.Load(), q.eqsl.Load()}
}

func (s qslSnapshot) enrich(spot *publicSpot) {
	call := strings.ToUpper(spot.Call)
	if s.lotw != nil {
		_, spot.LoTW = (*s.lotw)[call]
	}
	if s.eqsl != nil {
		_, spot.EQSL = (*s.eqsl)[call]
	}
}

func readQSLDirectory(data []byte, lotw bool) (map[string]struct{}, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !lotw && (!scanner.Scan() || !strings.HasPrefix(scanner.Text(), "List of AG members as of ")) {
		return nil, fmt.Errorf("missing eQSL AG header")
	}
	calls := make(map[string]struct{})
	for scanner.Scan() {
		call := strings.TrimSpace(scanner.Text())
		if lotw {
			fields := strings.Split(call, ",")
			if len(fields) != 3 {
				continue
			}
			if _, err := time.Parse("2006-01-02", fields[1]); err != nil {
				continue
			}
			call = fields[0]
		}
		call = strings.ToUpper(call)
		if len(call) >= 2 && researchCall.MatchString(call) {
			calls[call] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(calls) == 0 {
		return nil, fmt.Errorf("empty QSL directory")
	}
	return calls, nil
}

func loadQSLDirectory(ctx context.Context, root string) *qslDirectory {
	q := &qslDirectory{}
	if root == "" || os.Getenv("DXREGIA_QSL_DIRECTORIES") == "off" {
		return q
	}
	client := &http.Client{Timeout: 25 * time.Second}
	for _, source := range []struct {
		name, url   string
		lotw        bool
		destination *atomic.Pointer[map[string]struct{}]
	}{
		{"lotw", "https://lotw.arrl.org/lotw-user-activity.csv", true, &q.lotw},
		{"eqsl", "https://www.eqsl.cc/qslcard/DownloadedFiles/AGMemberList.txt", false, &q.eqsl},
	} {
		path := filepath.Join(root, "local_data", "dxregia-"+source.name+"-members.txt")
		var due time.Time
		if info, err := os.Stat(path); err == nil && info.Size() <= maxQSLSourceBytes {
			if data, err := os.ReadFile(path); err == nil {
				if calls, err := readQSLDirectory(data, source.lotw); err == nil {
					source.destination.Store(&calls)
					due = info.ModTime().Add(qslRefreshInterval)
				}
			}
		}
		go func() {
			for {
				timer := time.NewTimer(time.Until(due))
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				data, err := fetchQSLDirectory(ctx, client, source.url)
				if err == nil {
					var calls map[string]struct{}
					calls, err = readQSLDirectory(data, source.lotw)
					if err == nil {
						source.destination.Store(&calls)
						err = writeNodeFile(path, data)
					}
				}
				due = time.Now().Add(qslRefreshInterval)
				if err != nil && ctx.Err() == nil {
					log.Printf("QSL directory %s: %v", source.name, err)
					due = time.Now().Add(6 * time.Hour)
				}
			}
		}()
	}
	return q
}

func fetchQSLDirectory(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "DXRegia (+https://github.com/wifi75/dxregia)")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxQSLSourceBytes+1))
	if len(data) > maxQSLSourceBytes {
		return nil, fmt.Errorf("QSL directory exceeds size limit")
	}
	return data, err
}
