package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type spotFilterRule struct {
	Action, Band, Mode, Continent, Call string
}

type spotFilterSide struct {
	User       string         `json:"user"`
	Expression string         `json:"asc"`
	Editor     spotFilterRule `json:"dxregia"`
	Stamp      string         `json:"dxregia_stamp"`
}

func (side spotFilterSide) stamp() string {
	hash := sha256.Sum256([]byte(side.User + "\x00" + side.Expression))
	return fmt.Sprintf("%x", hash)
}

func (side spotFilterSide) editable(action string) bool {
	return side.Editor.Action == action && side.Stamp == side.stamp()
}

type spotFilterSlot struct {
	Number       int
	Action, Text string
	Editor       *spotFilterRule
}

var (
	spotFilterCall    = regexp.MustCompile("^[A-Z0-9]{2,10}(?:-[0-9]{1,2})?$")
	filterBandComment = regexp.MustCompile(`(?m)#.*$`)
	filterBandBlock   = regexp.MustCompile(`(?s)'([a-z0-9]+)'\s*=>\s*bless\s*\(\s*\{([^}]+)\}`)
	filterBandRange   = regexp.MustCompile(`([a-z][a-z0-9]*)\s*=>\s*\[([ ,0-9.\t\r\n]+)\]`)
)

func spotFilterOptions(root string) map[string]map[string][]float64 {
	data, _ := os.ReadFile(spotDataPath(root, "bands.pl"))
	bands := map[string]map[string][]float64{}
	for _, match := range filterBandBlock.FindAllStringSubmatch(filterBandComment.ReplaceAllString(string(data), ""), -1) {
		modes := map[string][]float64{}
		for _, field := range filterBandRange.FindAllStringSubmatch(match[2], -1) {
			var values []float64
			for _, item := range strings.Split(field[2], ",") {
				if number, err := strconv.ParseFloat(strings.TrimSpace(item), 64); err == nil {
					values = append(values, number)
				}
			}
			if len(values) > 0 && len(values)%2 == 0 {
				modes[field[1]] = values
			}
		}
		if len(modes["band"]) > 0 {
			bands[match[1]] = modes
		}
	}
	return bands
}

func (rule spotFilterRule) build(bands map[string]map[string][]float64, directory dxccDirectory) (spotFilterSide, error) {
	var user, expression []string
	if rule.Action != "accept" && rule.Action != "reject" {
		return spotFilterSide{}, errors.New("Azione del filtro non valida.")
	}
	if rule.Band != "" || rule.Mode != "" {
		mode := rule.Mode
		if mode == "" {
			mode = "band"
		}
		var labels, ranges []string
		for band, modes := range bands {
			if rule.Band != "" && rule.Band != band {
				continue
			}
			values := modes[mode]
			if len(values) == 0 {
				continue
			}
			label := band
			if mode != "band" {
				label += "/" + mode
			}
			labels = append(labels, label)
			for i := 0; i < len(values); i += 2 {
				ranges = append(ranges, fmt.Sprintf("($r->[0] >= %g && $r->[0] <= %g)", values[i], values[i+1]))
			}
		}
		if len(ranges) == 0 {
			return spotFilterSide{}, errors.New("Banda o segmento di modo non disponibile nel piano DXSpider.")
		}
		slices.Sort(labels)
		slices.Sort(ranges)
		user = append(user, "on "+strings.Join(labels, ","))
		expression = append(expression, "("+strings.Join(ranges, " || ")+")")
	}
	if rule.Continent != "" {
		var codes []int
		for code, entry := range directory.entries {
			if code > 0 && code != 666 && entry.Continent == rule.Continent {
				codes = append(codes, code)
			}
		}
		if len(codes) == 0 {
			return spotFilterSide{}, errors.New("Continente DX non disponibile nel catalogo.")
		}
		slices.Sort(codes)
		var labels, terms []string
		for _, code := range codes {
			labels = append(labels, strconv.Itoa(code))
			terms = append(terms, fmt.Sprintf("$r->[5] == %d", code))
		}
		user = append(user, "call_dxcc "+strings.Join(labels, ","))
		expression = append(expression, "("+strings.Join(terms, " || ")+")")
	}
	if rule.Call != "" {
		rule.Call = strings.ToUpper(strings.TrimSpace(rule.Call))
		if !researchCall.MatchString(rule.Call) || len(rule.Call) < 2 {
			return spotFilterSide{}, errors.New("Nominativo DX non valido.")
		}
		user = append(user, "call "+rule.Call)
		expression = append(expression, "uc($r->[1]) eq '"+rule.Call+"'")
	}
	if len(expression) == 0 {
		return spotFilterSide{}, errors.New("Scegli almeno un criterio per la regola.")
	}
	side := spotFilterSide{User: strings.Join(user, " and "), Expression: strings.Join(expression, " && "), Editor: rule}
	side.Stamp = side.stamp()
	return side, nil
}

func readSpotFilterSlots(data []byte) []spotFilterSlot {
	var file map[string]json.RawMessage
	if json.Unmarshal(data, &file) != nil {
		return nil
	}
	var rows []spotFilterSlot
	for i := 0; i <= 9; i++ {
		var slot map[string]spotFilterSide
		if json.Unmarshal(file[fmt.Sprintf("filter%d", i)], &slot) != nil {
			continue
		}
		for _, action := range []string{"reject", "accept"} {
			if side, ok := slot[action]; ok {
				row := spotFilterSlot{Number: i, Action: action, Text: side.User}
				if side.editable(action) {
					row.Editor = &side.Editor
				}
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func validSpotFilterName(name string) bool {
	target := strings.TrimSuffix(strings.TrimPrefix(name, "in_"), ".pl")
	return strings.HasSuffix(name, ".pl") &&
		(target == "node_default" || target == "user_default" || spotFilterCall.MatchString(strings.ToUpper(target)))
}

func changeSpotFilter(path, name string, number int, action, previous string, side *spotFilterSide) error {
	file := map[string]json.RawMessage{}
	data, err := os.ReadFile(path)
	if err == nil {
		if json.Unmarshal(data, &file) != nil || file == nil {
			return errors.New("Il filtro usa un formato precedente. Gestiscilo dalla console DXSpider.")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	key := fmt.Sprintf("filter%d", number)
	slot := map[string]json.RawMessage{}
	if raw := file[key]; raw != nil {
		if err := json.Unmarshal(raw, &slot); err != nil || slot == nil {
			return errors.New("La riga del filtro non è leggibile.")
		}
	}
	if side == nil {
		if slot[action] == nil {
			return errors.New("Regola non trovata.")
		}
		delete(slot, action)
	} else {
		if previous != "" {
			var current spotFilterSide
			if json.Unmarshal(slot[previous], &current) != nil || !current.editable(previous) {
				return errors.New("La regola originale è cambiata. Aggiorna l’elenco dei filtri.")
			}
		}
		if previous != action && slot[action] != nil {
			return errors.New("Questo numero contiene già una regola con la stessa azione.")
		}
		if previous != "" {
			delete(slot, previous)
		}
		slot[action], _ = json.Marshal(side)
	}
	if len(slot) == 0 {
		delete(file, key)
	} else {
		file[key], _ = json.Marshal(slot)
	}
	if len(file) == 2 && file["name"] != nil && file["sort"] != nil {
		return os.Remove(path)
	}
	file["name"], _ = json.Marshal(name)
	file["sort"] = json.RawMessage(`"spots"`)
	encoded, err := json.Marshal(file)
	if err != nil {
		return err
	}
	return writeNodeFile(path, encoded)
}

func readSpotFilterChange(w http.ResponseWriter, r *http.Request) (string, int, bool) {
	name := r.PathValue("name")
	number, err := strconv.Atoi(r.PathValue("number"))
	if !validSpotFilterName(name) || err != nil || number < 0 || number > 9 {
		http.Error(w, "Destinatario o numero di regola non valido.", http.StatusBadRequest)
		return "", 0, false
	}
	return name, number, true
}
